// Package identity turns an OpenID Connect id_token into a login. The
// gateway trusts one issuer: it fetches the issuer's signing keys from the
// discovery document, checks the RS256 signature, the issuer, the audience
// and the expiry, and reads the login from one claim. No library: a JWT
// with RS256 is a signature over two base64 parts, and that is all the
// gateway needs.
package identity

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Verifier checks tokens of one issuer for one audience.
type Verifier struct {
	Issuer   string
	Audience string
	// Claim is the claim that gives the login; "email" by default.
	Claim string
	// StripDomain keeps what is before the @ of the claim, lowercased.
	StripDomain bool
	// HTTP fetches the discovery document and the keys; http.DefaultClient
	// when nil.
	HTTP *http.Client

	mu   sync.Mutex
	keys map[string]*rsa.PublicKey // by kid, from the last good fetch
}

// Login verifies the token and returns the login it carries.
func (v *Verifier) Login(ctx context.Context, token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", errors.New("token: not three parts")
	}
	var head struct{ Alg, Kid string }
	if err := decodePart(parts[0], &head); err != nil {
		return "", fmt.Errorf("token header: %w", err)
	}
	if head.Alg != "RS256" {
		return "", fmt.Errorf("token: algorithm %q is not RS256", head.Alg)
	}
	var claims map[string]any
	if err := decodePart(parts[1], &claims); err != nil {
		return "", fmt.Errorf("token claims: %w", err)
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return "", fmt.Errorf("token signature: %w", err)
	}
	key, err := v.key(ctx, head.Kid)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, sum[:], sig); err != nil {
		return "", errors.New("token: the signature does not verify")
	}
	if claims["iss"] != v.Issuer {
		return "", fmt.Errorf("token: issuer %v is not %s", claims["iss"], v.Issuer)
	}
	if !hasAudience(claims["aud"], v.Audience) {
		return "", fmt.Errorf("token: audience %v is not %s", claims["aud"], v.Audience)
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		return "", errors.New("token: no expiry")
	}
	if time.Now().After(time.Unix(int64(exp), 0)) {
		return "", errors.New("token: expired")
	}
	claim := v.Claim
	if claim == "" {
		claim = "email"
	}
	login, _ := claims[claim].(string)
	login = strings.TrimSpace(login)
	if login == "" {
		return "", fmt.Errorf("token: claim %q is missing or empty", claim)
	}
	if v.StripDomain {
		login, _, _ = strings.Cut(login, "@")
		login = strings.ToLower(login)
	}
	return login, nil
}

func decodePart(part string, v any) error {
	data, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func hasAudience(aud any, want string) bool {
	switch a := aud.(type) {
	case string:
		return a == want
	case []any:
		for _, x := range a {
			if x == want {
				return true
			}
		}
	}
	return false
}

// key returns the public key of a kid, from the keys kept or from a fresh
// fetch when the kid is new.
func (v *Verifier) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if k, ok := v.keys[kid]; ok {
		return k, nil
	}
	keys, err := v.fetch(ctx)
	if err != nil {
		return nil, err
	}
	v.keys = keys
	k, ok := keys[kid]
	if !ok {
		return nil, fmt.Errorf("token: key %q is not one the issuer publishes", kid)
	}
	return k, nil
}

// fetch reads the discovery document, then the JWKS, and keeps the RSA
// keys whose numbers decode. A key with bad numbers is left out, so that a
// token under it is refused as an unknown key.
func (v *Verifier) fetch(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	var disc struct {
		JWKSURI string `json:"jwks_uri"`
	}
	if err := v.getJSON(ctx, strings.TrimRight(v.Issuer, "/")+"/.well-known/openid-configuration", &disc); err != nil {
		return nil, fmt.Errorf("issuer discovery: %w", err)
	}
	if disc.JWKSURI == "" {
		return nil, errors.New("issuer discovery: no jwks_uri")
	}
	var set struct {
		Keys []struct{ Kty, Kid, N, E string } `json:"keys"`
	}
	if err := v.getJSON(ctx, disc.JWKSURI, &set); err != nil {
		return nil, fmt.Errorf("issuer jwks: %w", err)
	}
	keys := map[string]*rsa.PublicKey{}
	for _, k := range set.Keys {
		if k.Kty != "RSA" {
			continue
		}
		n, errN := base64.RawURLEncoding.DecodeString(k.N)
		e, errE := base64.RawURLEncoding.DecodeString(k.E)
		if errN != nil || errE != nil {
			continue
		}
		keys[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: int(new(big.Int).SetBytes(e).Int64())}
	}
	return keys, nil
}

func (v *Verifier) getJSON(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	c := v.HTTP
	if c == nil {
		c = http.DefaultClient
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()
	return json.NewDecoder(res.Body).Decode(out)
}
