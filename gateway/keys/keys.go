// Package keys issues and revokes the confined CortexDB keys the gateway
// hands to employees, in the key file of the brain's server. The key is the
// employee's credential on that brain: the engine enforces the collection
// and the clearance, the gateway only decides who gets which key.
package keys

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/liliang-cn/cortexdb/v2/pkg/authz"

	"github.com/lfoletti/granular-cortex/gateway/rights"
)

// Issued is the key handed to the employee.
type Issued struct {
	ID         string
	Secret     string
	Clearance  authz.Clearance
	Collection string
}

const idPrefix = "gc/"

// ID is the key id the gateway writes for a user on a brain. It names both,
// so the key file itself says who holds what.
func ID(user, brain string) string { return idPrefix + user + "/" + brain }

// holderOf parses an id the gateway wrote back into user and brain.
func holderOf(id string) (user, brain string, ok bool) {
	rest, found := strings.CutPrefix(id, idPrefix)
	if !found {
		return "", "", false
	}
	user, brain, ok = strings.Cut(rest, "/")
	return user, brain, ok && user != "" && brain != ""
}

type keyFile struct {
	Keys []authz.Key `json:"keys"`
}

// read loads the server's key file. A missing file is an empty file; a file
// that cannot be read or parsed is an error that names it, and is never
// replaced.
func read(path string) (*keyFile, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &keyFile{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var f keyFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &f, nil
}

// write replaces the file in one step: the new contents land in a file next
// to it, then take its name. A failure leaves the old file as it was. The
// file is private to its owner: it holds secrets. Before anything lands, the
// server's own parser must accept the contents: the gateway never writes a
// key file the server refuses to load, except the empty one after the last
// revocation.
func write(path string, f *keyFile) error {
	data, _ := json.MarshalIndent(f, "", " ")
	if len(f.Keys) > 0 {
		if _, err := authz.Parse(data); err != nil {
			return fmt.Errorf("the key file the gateway would write is one the server refuses: %w", err)
		}
	}
	tmp := path + ".tmp"
	err := os.WriteFile(tmp, append(data, '\n'), 0o600)
	if err == nil {
		err = os.Rename(tmp, path)
	}
	if err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// Issue writes a key for user on brain into the server's key file, confined
// to the brain's collection, with the clearance of the access. A user with
// no access gets no key. A key already issued to the user for the brain is
// not issued twice: revoke it first.
func Issue(keysFile, user, brain, collection string, access rights.Access) (Issued, error) {
	if !access.AllowsRead() {
		return Issued{}, fmt.Errorf("%s has no access to %s: no key issued", user, brain)
	}
	f, err := read(keysFile)
	if err != nil {
		return Issued{}, err
	}
	id := ID(user, brain)
	for _, k := range f.Keys {
		if k.ID == id {
			return Issued{}, fmt.Errorf("a key is already issued to %s for %s (%s): revoke it first", user, brain, id)
		}
	}
	secret := rand.Text() + rand.Text()
	issued := Issued{ID: id, Secret: secret, Clearance: access.Clearance(), Collection: collection}
	f.Keys = append(f.Keys, authz.Key{ID: id, Secret: secret, Clearance: issued.Clearance, Scope: authz.Scope{Collections: []string{collection}}})
	if err := write(keysFile, f); err != nil {
		return Issued{}, err
	}
	return issued, nil
}

// Revoke removes the user's key for the brain from the server's key file.
func Revoke(keysFile, user, brain string) error {
	f, err := read(keysFile)
	if err != nil {
		return err
	}
	id := ID(user, brain)
	kept := f.Keys[:0:0]
	for _, k := range f.Keys {
		if k.ID != id {
			kept = append(kept, k)
		}
	}
	if len(kept) == len(f.Keys) {
		return fmt.Errorf("no key issued to %s for %s (%s)", user, brain, id)
	}
	f.Keys = kept
	return write(keysFile, f)
}

// Holder maps a secret presented to the gateway back to the user and brain
// the key was issued for. A secret the file does not hold, or a key the
// gateway did not issue (an operator key, for one), is refused.
func Holder(keysFile, secret string) (user, brain string, err error) {
	f, err := read(keysFile)
	if err != nil {
		return "", "", err
	}
	for _, k := range f.Keys {
		if k.Secret != secret {
			continue
		}
		user, brain, ok := holderOf(k.ID)
		if !ok {
			return "", "", fmt.Errorf("key %q was not issued by the gateway", k.ID)
		}
		return user, brain, nil
	}
	return "", "", errors.New("unknown key")
}
