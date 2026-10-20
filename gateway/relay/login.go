package relay

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/lfoletti/thalamus/gateway/identity"
	"github.com/lfoletti/thalamus/gateway/keys"
	"github.com/lfoletti/thalamus/gateway/rights"
)

// POST /v1/login: the employee presents the id_token of the organisation's
// identity provider, the gateway derives the login, issues the keys the
// rights allow where they are missing, and returns them. No key is handed
// out by hand any more. The identity block of rights.yaml names the
// provider; without it the route does not exist.

// loginAnswer is what the employee receives: one key for the federated
// endpoint (the first brain's), and one per brain.
type loginAnswer struct {
	Login  string       `json:"login"`
	Key    string       `json:"key"`
	Brains []brainLogin `json:"brains"`
}

type brainLogin struct {
	Brain  string `json:"brain"`
	Access string `json:"access"`
	Key    string `json:"key"`
}

// verifier returns the verifier of the current identity block, built when
// the block changes (the rights file reloads).
func (s *Server) verifier(id *rights.Identity) *identity.Verifier {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.idp != nil && s.idpFor == *id {
		return s.idp
	}
	s.idp = &identity.Verifier{Issuer: id.Issuer, Audience: id.Audience, Claim: id.Claim, StripDomain: id.StripDomain}
	s.idpFor = *id
	return s.idp
}

// serveIdentity answers GET /v1/identity: the provider, the audience and
// the client id, for the tools that sign the employee in.
func (s *Server) serveIdentity(w http.ResponseWriter, r *http.Request) {
	id := s.rights().Identity
	if id == nil {
		restError(w, http.StatusNotFound, errors.New("no identity provider is configured: the rights file has no identity block"))
		return
	}
	if r.Method != http.MethodGet {
		restError(w, http.StatusMethodNotAllowed, errors.New("use GET"))
		return
	}
	restJSON(w, http.StatusOK, struct {
		Issuer   string `json:"issuer"`
		Audience string `json:"audience"`
		ClientID string `json:"client_id"`
	}{id.Issuer, id.Audience, id.ClientID})
}

func (s *Server) serveLogin(w http.ResponseWriter, r *http.Request) {
	id := s.rights().Identity
	if id == nil {
		restError(w, http.StatusNotFound, errors.New("no identity provider is configured: the rights file has no identity block"))
		return
	}
	if r.Method != http.MethodPost {
		restError(w, http.StatusMethodNotAllowed, errors.New("use POST"))
		return
	}
	token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !found || token == "" {
		restError(w, http.StatusUnauthorized, errors.New("an id_token of the identity provider is required, as a bearer"))
		return
	}
	login, err := s.verifier(id).Login(r.Context(), token)
	if err != nil {
		restError(w, http.StatusUnauthorized, err)
		return
	}
	grants := s.rights().BrainsOf(login)
	if len(grants) == 0 {
		restError(w, http.StatusForbidden, fmt.Errorf("%s has no brain in the rights", login))
		return
	}
	answer := loginAnswer{Login: login}
	var names []string
	for _, g := range grants {
		b := s.rights().Brains[g.Brain]
		secret, err := keys.SecretOf(b.KeysFile, login, g.Brain)
		if err != nil {
			k, err := keys.Issue(b.KeysFile, login, g.Brain, b.Collection, g.Access)
			if err != nil {
				_ = s.audit.Record(Entry{User: login, Tool: "login", Error: err.Error()})
				restError(w, http.StatusInternalServerError, fmt.Errorf("key file of %s: %w", g.Brain, err))
				return
			}
			secret = k.Secret
		}
		answer.Brains = append(answer.Brains, brainLogin{Brain: g.Brain, Access: g.Access.String(), Key: secret})
		names = append(names, g.Brain)
	}
	answer.Key = answer.Brains[0].Key
	_ = s.audit.Record(Entry{User: login, Tool: "login", IDs: map[string]any{"brains": names}, OK: true})
	restJSON(w, http.StatusOK, answer)
}
