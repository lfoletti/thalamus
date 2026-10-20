// Package rights reads rights.yaml: the brains an organisation runs, the
// groups of its employees, and who may read or write each brain. It is the
// one place the gateway decides access; the engine then enforces it through
// the confined key the gateway issued.
package rights

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/liliang-cn/cortexdb/v2/pkg/authz"
	"gopkg.in/yaml.v3"
)

// Access is what a user may do on a brain. Write implies Read.
type Access int

const (
	None Access = iota
	Read
	Write
)

func (a Access) String() string {
	switch a {
	case Read:
		return "read"
	case Write:
		return "write"
	}
	return "none"
}

// AllowsRead reports whether the access lets the user read the brain.
func (a Access) AllowsRead() bool { return a == Read || a == Write }

// Clearance is the CortexDB clearance a key issued for this access carries.
// None has no key.
func (a Access) Clearance() authz.Clearance {
	if a == Write {
		return authz.ReadWrite
	}
	return authz.ReadOnly
}

// Brain is one CortexDB server and, when the server is shared between
// brains, the collection this brain lives in. Without a collection the
// brain is the whole server: its keys are not confined, and the engine
// opens the graph tools to them, which a confined key never gets.
type Brain struct {
	Addr     string `yaml:"addr"`
	KeysFile string `yaml:"keys_file"`
	// Collection is empty when the brain is the whole server.
	Collection string   `yaml:"collection"`
	Readers    []string `yaml:"readers"`
	Writers    []string `yaml:"writers"`
	// RankWeight scales the brain's ranks in a federated recall; nil means 1.
	RankWeight *float64 `yaml:"weight"`
}

// Weight is the rank weight of the brain, 1 when the file says nothing.
func (b Brain) Weight() float64 {
	if b.RankWeight == nil {
		return 1
	}
	return *b.RankWeight
}

// WholeServer reports whether the brain is its server's only brain.
func (b Brain) WholeServer() bool { return b.Collection == "" }

// Rights is the parsed, validated file.
type Rights struct {
	Brains map[string]Brain    `yaml:"brains"`
	Groups map[string][]string `yaml:"groups"`
}

// Grant is one brain a user may reach, with the access level.
type Grant struct {
	Brain  string
	Access Access
}

// Load reads and validates a rights file.
func Load(path string) (*Rights, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read rights file: %w", err)
	}
	return Parse(data)
}

// Parse validates the file's contents. Unknown fields are an error rather
// than ignored: a misspelled "writers" would otherwise load as a brain nobody
// may write to, and a misspelled "readers" as one nobody may read.
func Parse(data []byte) (*Rights, error) {
	var r Rights
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&r); err != nil {
		return nil, fmt.Errorf("parse rights file: %w", err)
	}
	if len(r.Brains) == 0 {
		return nil, errors.New("parse rights file: no brains declared")
	}
	for name, b := range r.Brains {
		for field, value := range map[string]string{"addr": b.Addr, "keys_file": b.KeysFile} {
			if strings.TrimSpace(value) == "" {
				return nil, fmt.Errorf("parse rights file: brain %q has no %s", name, field)
			}
		}
		for _, m := range append(append([]string{}, b.Readers...), b.Writers...) {
			if err := r.checkMember(m); err != nil {
				return nil, fmt.Errorf("parse rights file: brain %q: %w", name, err)
			}
		}
		if b.Weight() <= 0 {
			return nil, fmt.Errorf("parse rights file: brain %q: weight must be positive", name)
		}
	}
	return &r, nil
}

func (r *Rights) checkMember(m string) error {
	kind, name, ok := strings.Cut(m, ":")
	if !ok || name == "" {
		return fmt.Errorf("member %q must be group:<name> or user:<login>", m)
	}
	switch kind {
	case "user":
		return nil
	case "group":
		if _, declared := r.Groups[name]; !declared {
			return fmt.Errorf("group %q is not declared", name)
		}
		return nil
	}
	return fmt.Errorf("member %q must be group:<name> or user:<login>", m)
}

// Access says what user may do on brain. An unknown brain is an error, so a
// typo is never read as "no access".
func (r *Rights) Access(user, brain string) (Access, error) {
	b, ok := r.Brains[brain]
	if !ok {
		return None, fmt.Errorf("unknown brain %q", brain)
	}
	if r.listed(user, b.Writers) {
		return Write, nil
	}
	if r.listed(user, b.Readers) {
		return Read, nil
	}
	return None, nil
}

func (r *Rights) listed(user string, members []string) bool {
	for _, m := range members {
		kind, name, _ := strings.Cut(m, ":")
		switch kind {
		case "user":
			if name == user {
				return true
			}
		case "group":
			for _, u := range r.Groups[name] {
				if u == user {
					return true
				}
			}
		}
	}
	return false
}

// BrainsOf lists the brains the user may at least read, sorted by name.
func (r *Rights) BrainsOf(user string) []Grant {
	var out []Grant
	for name := range r.Brains {
		a, _ := r.Access(user, name)
		if a.AllowsRead() {
			out = append(out, Grant{Brain: name, Access: a})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Brain < out[j].Brain })
	return out
}
