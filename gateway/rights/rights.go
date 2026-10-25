// Package rights reads rights.yaml: the brains an organisation runs, the
// groups of its users, and who may read or write each brain. It is the
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

// Access is what a user may do on a brain. Write implies Read; Curate
// implies Write and adds the review of what others propose.
type Access int

const (
	None Access = iota
	Read
	Write
	Curate
)

func (a Access) String() string {
	switch a {
	case Read:
		return "read"
	case Write:
		return "write"
	case Curate:
		return "curate"
	}
	return "none"
}

// AllowsRead reports whether the access lets the user read the brain.
func (a Access) AllowsRead() bool { return a >= Read }

// AllowsWrite reports whether the access lets the user write to the brain.
func (a Access) AllowsWrite() bool { return a >= Write }

// Clearance is the CortexDB clearance a key issued for this access carries.
// None has no key.
func (a Access) Clearance() authz.Clearance {
	if a.AllowsWrite() {
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
	// Curators write, and review what the writers propose.
	Curators []string `yaml:"curators"`
	// RankWeight scales the brain's ranks in a federated recall; nil means 1.
	RankWeight *float64 `yaml:"weight"`
	// AuditReads asks for a named audit line per read of this brain. Without
	// it the reads are counted per day, without a name.
	AuditReads bool `yaml:"audit_reads"`
	// Pyramid is the collection of the brain's syntheses on the same
	// server, when a producer writes them; nil when there is none.
	Pyramid *Pyramid `yaml:"pyramid"`
	// ExcerptWords bounds what th_get and th_descend serve of a source
	// document: its first words, and the address of the original. A
	// publisher sets it when the sources are not its to hand out whole. 0
	// serves them whole.
	ExcerptWords int `yaml:"excerpt_words"`
}

// Pyramid names the collection of a brain's syntheses and the metadata key
// under which a synthesis lists the ids of the documents it cites.
type Pyramid struct {
	Collection string `yaml:"collection"`
	Sources    string `yaml:"sources"`
}

// SourcesKey is the metadata key of the cited documents, "sources" when the
// file says nothing.
func (p *Pyramid) SourcesKey() string {
	if p.Sources == "" {
		return "sources"
	}
	return p.Sources
}

// Collections are the collections a key of the brain is confined to: the
// brain's, then its pyramid's. A whole-server brain confines nothing.
func (b Brain) Collections() []string {
	if b.Collection == "" {
		return nil
	}
	out := []string{b.Collection}
	if b.Pyramid != nil {
		out = append(out, b.Pyramid.Collection)
	}
	return out
}

// Weight is the rank weight of the brain, 1 when the file says nothing.
func (b Brain) Weight() float64 {
	if b.RankWeight == nil {
		return 1
	}
	return *b.RankWeight
}

// members lists every member of the brain's three lists.
func (b Brain) members() []string {
	out := append([]string{}, b.Readers...)
	out = append(out, b.Writers...)
	return append(out, b.Curators...)
}

// WholeServer reports whether the brain is its server's only brain.
func (b Brain) WholeServer() bool { return b.Collection == "" }

// Identity names the OpenID Connect issuer the gateway trusts for
// POST /v1/login: an id_token of this issuer, for this audience, gives the
// login from Claim ("email" by default), with the domain stripped when
// StripDomain is set.
type Identity struct {
	Issuer      string `yaml:"issuer"`
	Audience    string `yaml:"audience"`
	Claim       string `yaml:"claim"`
	StripDomain bool   `yaml:"strip_domain"`
	// ClientID is the client the provider knows the users' tools by
	// (th-gateway login, the extension); GET /v1/identity hands it out.
	ClientID string `yaml:"client_id"`
}

// Rights is the parsed, validated file.
type Rights struct {
	Brains map[string]Brain    `yaml:"brains"`
	Groups map[string][]string `yaml:"groups"`
	// Identity is optional: without it, /v1/login answers 404.
	Identity *Identity `yaml:"identity"`
	// Audit is optional: without it, the audit keeps every line.
	Audit *AuditPolicy `yaml:"audit"`
}

// AuditPolicy says how long the audit file keeps its lines.
type AuditPolicy struct {
	RetentionDays int `yaml:"retention_days"`
}

// RetentionDays is how many days the audit keeps a line; 0 keeps every line.
func (r *Rights) RetentionDays() int {
	if r.Audit == nil {
		return 0
	}
	return r.Audit.RetentionDays
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
		for _, m := range b.members() {
			if err := r.checkMember(m); err != nil {
				return nil, fmt.Errorf("parse rights file: brain %q: %w", name, err)
			}
		}
		if b.Weight() <= 0 {
			return nil, fmt.Errorf("parse rights file: brain %q: weight must be positive", name)
		}
		if p := b.Pyramid; p != nil && strings.TrimSpace(p.Collection) == "" {
			return nil, fmt.Errorf("parse rights file: brain %q: the pyramid needs a collection", name)
		}
		if b.ExcerptWords < 0 {
			return nil, fmt.Errorf("parse rights file: brain %q: excerpt_words must not be negative", name)
		}
		if p := b.Pyramid; p != nil && p.Collection == b.Collection {
			return nil, fmt.Errorf("parse rights file: brain %q: the pyramid's collection must differ from the brain's", name)
		}
	}
	if r.Identity != nil && (strings.TrimSpace(r.Identity.Issuer) == "" || strings.TrimSpace(r.Identity.Audience) == "") {
		return nil, errors.New("identity needs issuer and audience")
	}
	if r.RetentionDays() < 0 {
		return nil, errors.New("parse rights file: audit retention_days must not be negative")
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
	if r.listed(user, b.Curators) {
		return Curate, nil
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

// Users lists every login the file names, through a group or directly,
// sorted and once each.
func (r *Rights) Users() []string {
	seen := map[string]bool{}
	for _, members := range r.Groups {
		for _, u := range members {
			seen[u] = true
		}
	}
	for _, b := range r.Brains {
		for _, m := range b.members() {
			if kind, name, _ := strings.Cut(m, ":"); kind == "user" {
				seen[name] = true
			}
		}
	}
	out := make([]string, 0, len(seen))
	for u := range seen {
		out = append(out, u)
	}
	sort.Strings(out)
	return out
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
