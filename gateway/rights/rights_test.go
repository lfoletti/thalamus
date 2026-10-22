package rights

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/liliang-cn/cortexdb/v2/pkg/authz"
)

const sample = `
brains:
  team/acme:
    addr: 127.0.0.1:47821
    keys_file: /srv/acme/keys.json
    collection: team-acme
    readers: [group:acme, user:guest]
    writers: [group:acme-devs]
  org/reference:
    addr: 127.0.0.1:47822
    keys_file: /srv/ref/keys.json
    collection: reference
    readers: [group:acme, group:acme-devs]
    writers: [user:kmlead]
    weight: 0.8
groups:
  acme: [lfoletti, mdupont]
  acme-devs: [lfoletti]
`

func parse(t *testing.T, text string) *Rights {
	t.Helper()
	r, err := Parse([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// Writing implies reading: a writer who is on no readers list still reads.
// A reader who is on no writers list reads only.
func TestAWriterMayAlsoRead(t *testing.T) {
	r := parse(t, sample)
	for _, c := range []struct {
		user, brain string
		want        Access
	}{
		{"kmlead", "org/reference", Write},
		{"mdupont", "org/reference", Read},
		{"lfoletti", "team/acme", Write},
		{"mdupont", "team/acme", Read},
		{"guest", "team/acme", Read},
	} {
		got, err := r.Access(c.user, c.brain)
		if err != nil || got != c.want {
			t.Fatalf("%s on %s: got %v, %v; want %v", c.user, c.brain, got, err, c.want)
		}
	}
	if !Write.AllowsRead() || !Read.AllowsRead() || None.AllowsRead() {
		t.Fatal("AllowsRead is wrong")
	}
}

// A user on no list of a brain has no access to it, and the answer is not an
// error: the brain exists, the user is simply not admitted.
func TestAUserOutsideEveryListHasNoAccess(t *testing.T) {
	r := parse(t, sample)
	got, err := r.Access("stranger", "team/acme")
	if err != nil || got != None {
		t.Fatalf("got %v, %v", got, err)
	}
	if got, err := r.Access("guest", "org/reference"); err != nil || got != None {
		t.Fatalf("guest on reference: got %v, %v", got, err)
	}
}

// A brain that is not declared is an error, so a typo in a brain name is
// never read as "no access".
func TestAnUnknownBrainIsAnError(t *testing.T) {
	r := parse(t, sample)
	if _, err := r.Access("lfoletti", "team/acm"); err == nil || !strings.Contains(err.Error(), "unknown brain") {
		t.Fatalf("got %v", err)
	}
}

// BrainsOf lists the brains a user may read, sorted, with the access level.
func TestBrainsOfListsTheUsersBrainsInOrder(t *testing.T) {
	r := parse(t, sample)
	got := r.BrainsOf("lfoletti")
	if len(got) != 2 || got[0].Brain != "org/reference" || got[0].Access != Read || got[1].Brain != "team/acme" || got[1].Access != Write {
		t.Fatalf("got %+v", got)
	}
	if got := r.BrainsOf("stranger"); len(got) != 0 {
		t.Fatalf("stranger: got %+v", got)
	}
}

// A member is written group:<name> or user:<login>. Anything else, and a
// group that is not declared, is refused at parse time, not at lookup time.
func TestAMemberMustBeAGroupOrAUser(t *testing.T) {
	bad := strings.Replace(sample, "user:guest", "guest", 1)
	if _, err := Parse([]byte(bad)); err == nil || !strings.Contains(err.Error(), `member "guest"`) {
		t.Fatalf("bare member: got %v", err)
	}
	bad = strings.Replace(sample, "group:acme-devs]", "group:acme-dev]", 1)
	if _, err := Parse([]byte(bad)); err == nil || !strings.Contains(err.Error(), `group "acme-dev"`) {
		t.Fatalf("undeclared group: got %v", err)
	}
}

// A brain needs its address and its key file. The collection is optional:
// without it, the brain is the whole server and its keys are not confined,
// which is what opens the graph tools to them.
func TestABrainNeedsItsAddressAndKeyFile(t *testing.T) {
	for _, field := range []string{"addr: 127.0.0.1:47821", "keys_file: /srv/acme/keys.json"} {
		bad := strings.Replace(sample, field, "", 1)
		_, err := Parse([]byte(bad))
		name := strings.SplitN(field, ":", 2)[0]
		if err == nil || !strings.Contains(err.Error(), name) {
			t.Fatalf("without %s: got %v", name, err)
		}
	}
	r := parse(t, strings.Replace(sample, "collection: team-acme", "", 1))
	if b := r.Brains["team/acme"]; b.Collection != "" || !b.WholeServer() || r.Brains["org/reference"].WholeServer() {
		t.Fatalf("whole server: %+v", b)
	}
}

// A misspelled field is refused rather than ignored: "writer" instead of
// "writers" would otherwise load as a brain nobody may write to.
func TestAMisspelledFieldIsRefused(t *testing.T) {
	bad := strings.Replace(sample, "writers: [group:acme-devs]", "writer: [group:acme-devs]", 1)
	if _, err := Parse([]byte(bad)); err == nil || !strings.Contains(err.Error(), "writer") {
		t.Fatalf("got %v", err)
	}
	if _, err := Parse([]byte("groups:\n  acme: [a]\n")); err == nil || !strings.Contains(err.Error(), "no brains") {
		t.Fatalf("no brains: got %v", err)
	}
}

// Load reads the file; a missing file is an error that names it.
func TestLoadReadsTheFileAndNamesItOnFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rights.yaml")
	if err := os.WriteFile(path, []byte(sample), 0o600); err != nil {
		t.Fatal(err)
	}
	r, err := Load(path)
	if err != nil || len(r.Brains) != 2 {
		t.Fatalf("got %v, %v", r, err)
	}
	if _, err := Load(filepath.Join(dir, "absent.yaml")); err == nil || !strings.Contains(err.Error(), "absent.yaml") {
		t.Fatalf("got %v", err)
	}
	b := r.Brains["team/acme"]
	if b.Addr != "127.0.0.1:47821" || b.KeysFile != "/srv/acme/keys.json" || b.Collection != "team-acme" {
		t.Fatalf("brain: %+v", b)
	}
}

// Access has a name and the clearance of the key issued for it; None has
// no key, so its clearance is the harmless one.
func TestAccessNamesItselfAndItsClearance(t *testing.T) {
	if None.String() != "none" || Read.String() != "read" || Write.String() != "write" {
		t.Fatalf("%v %v %v", None, Read, Write)
	}
	if Write.Clearance() != authz.ReadWrite || Read.Clearance() != authz.ReadOnly || None.Clearance() != authz.ReadOnly {
		t.Fatal("clearance is wrong")
	}
}

// A member of a kind other than user or group, or with an empty name, is
// refused.
func TestAMemberOfAnotherKindIsRefused(t *testing.T) {
	for _, m := range []string{"team:acme", "user:", "group:"} {
		bad := strings.Replace(sample, "user:guest", m, 1)
		if _, err := Parse([]byte(bad)); err == nil || !strings.Contains(err.Error(), `member "`+m+`"`) {
			t.Fatalf("%s: got %v", m, err)
		}
	}
}

// A brain's weight scales its ranks in a federated recall; it is 1 when
// the file says nothing, and must be positive.
func TestABrainHasAPositiveWeight(t *testing.T) {
	r := parse(t, sample)
	if w := r.Brains["team/acme"].Weight(); w != 1 {
		t.Fatalf("default weight %v", w)
	}
	if w := r.Brains["org/reference"].Weight(); w != 0.8 {
		t.Fatalf("weight %v", w)
	}
	for _, bad := range []string{"weight: 0", "weight: -1"} {
		text := strings.Replace(sample, "weight: 0.8", bad, 1)
		if _, err := Parse([]byte(text)); err == nil || !strings.Contains(err.Error(), "weight") {
			t.Fatalf("%s: got %v", bad, err)
		}
	}
}

// Users lists every login the file names, through groups or directly,
// sorted and without duplicates.
func TestUsersListsEveryLoginOnce(t *testing.T) {
	r := parse(t, sample)
	if got := strings.Join(r.Users(), ","); got != "guest,kmlead,lfoletti,mdupont" {
		t.Fatalf("got %s", got)
	}
}

// A curator writes and reviews; the role sits above writer. A curator who
// is on no other list still reads and writes.
func TestACuratorWritesAndReviews(t *testing.T) {
	text := strings.Replace(sample, "    writers: [group:acme-devs]\n", "    writers: [group:acme-devs]\n    curators: [user:kmlead]\n", 1)
	r := parse(t, text)
	got, err := r.Access("kmlead", "team/acme")
	if err != nil || got != Curate || !got.AllowsRead() || !got.AllowsWrite() || got.Clearance() != authz.ReadWrite || got.String() != "curate" {
		t.Fatalf("got %v, %v", got, err)
	}
	if Write.AllowsWrite() != true || Read.AllowsWrite() || None.AllowsWrite() {
		t.Fatal("AllowsWrite is wrong")
	}
	if a, _ := r.Access("lfoletti", "team/acme"); a != Write {
		t.Fatalf("a writer is %v", a)
	}
	bad := strings.Replace(text, "curators: [user:kmlead]", "curators: [kmlead]", 1)
	if _, err := Parse([]byte(bad)); err == nil || !strings.Contains(err.Error(), `member "kmlead"`) {
		t.Fatalf("bare curator: got %v", err)
	}
	if got := strings.Join(r.Users(), ","); got != "guest,kmlead,lfoletti,mdupont" {
		t.Fatalf("users %s", got)
	}
}

// A source reloads the file when it changes: a right withdrawn in the file
// is withdrawn at the next read, without a restart. A broken file leaves
// the previous rights in force and warns once; a removed file does the
// same. The first load must succeed.
func TestASourceReloadsTheFileWhenItChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rights.yaml")
	write := func(text string, when time.Time) {
		t.Helper()
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, when, when); err != nil {
			t.Fatal(err)
		}
	}
	t0 := time.Now().Add(-time.Hour)
	write(sample, t0)
	var warnings []string
	src, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	src.Warn = func(err error) { warnings = append(warnings, err.Error()) }
	first := src.Current()
	if a, _ := first.Access("lfoletti", "team/acme"); a != Write {
		t.Fatalf("first: %v", a)
	}
	if src.Current() != first {
		t.Fatal("an unchanged file was reloaded")
	}
	write(strings.Replace(sample, "    writers: [group:acme-devs]\n", "", 1), t0.Add(time.Second))
	second := src.Current()
	if second == first {
		t.Fatal("a changed file was not reloaded")
	}
	if a, _ := second.Access("lfoletti", "team/acme"); a != Read {
		t.Fatalf("second: %v", a)
	}
	write("brains: [", t0.Add(2*time.Second))
	for i := 0; i < 2; i++ {
		if src.Current() != second {
			t.Fatal("a broken file replaced the rights")
		}
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "previous rights stay") {
		t.Fatalf("warnings: %q", warnings)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if src.Current() != second {
			t.Fatal("a removed file replaced the rights")
		}
	}
	if len(warnings) != 2 || !strings.Contains(warnings[1], "previous rights stay") {
		t.Fatalf("warnings: %q", warnings)
	}
	// A directory in its place: the stat passes, the read fails.
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if src.Current() != second || len(warnings) != 3 {
			t.Fatalf("a directory replaced the rights: %q", warnings)
		}
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	write(sample, t0.Add(3*time.Second))
	third := src.Current()
	if a, _ := third.Access("lfoletti", "team/acme"); a != Write {
		t.Fatal("the repaired file was not reloaded")
	}
	// The same text written again: a new time, the same rights.
	write(sample, t0.Add(5*time.Second))
	if src.Current() != third {
		t.Fatal("the same text was parsed again")
	}
	if _, err := Open(filepath.Join(t.TempDir(), "absent.yaml")); err == nil {
		t.Fatal("an absent file opened")
	}
	if _, err := Open(t.TempDir()); err == nil {
		t.Fatal("a directory opened")
	}
	write("brains: [", t0.Add(4*time.Second))
	if _, err := Open(path); err == nil {
		t.Fatal("a broken file opened")
	}
	// A static source serves one value and never reads a file.
	s := Static(first)
	for i := 0; i < 2; i++ {
		if s.Current() != first {
			t.Fatal("static")
		}
	}
}

// The identity block names the OpenID Connect issuer the gateway trusts,
// the audience, the claim that gives the login and whether to strip the
// domain. Issuer and audience are required when the block is present.
func TestTheIdentityBlockIsReadAndChecked(t *testing.T) {
	r := parse(t, sample+"identity:\n  issuer: https://login.example\n  audience: thalamus\n  claim: preferred_username\n  strip_domain: true\n  client_id: thalamus-cli\n")
	if r.Identity == nil || r.Identity.Issuer != "https://login.example" || r.Identity.Audience != "thalamus" || r.Identity.Claim != "preferred_username" || !r.Identity.StripDomain || r.Identity.ClientID != "thalamus-cli" {
		t.Fatalf("identity: %+v", r.Identity)
	}
	if parse(t, sample).Identity != nil {
		t.Fatal("an absent block is not nil")
	}
	for _, bad := range []string{"identity:\n  audience: thalamus\n", "identity:\n  issuer: https://login.example\n", "identity:\n  issuer: ' '\n  audience: thalamus\n"} {
		if _, err := Parse([]byte(sample + bad)); err == nil || !strings.Contains(err.Error(), "identity") {
			t.Fatalf("%q: %v", bad, err)
		}
	}
}

// The audit block sets how long the audit keeps its lines, and a brain may
// ask for named reads. Without them the file keeps every line and counts
// the reads; a negative retention is refused.
func TestTheAuditBlockIsReadAndChecked(t *testing.T) {
	r := parse(t, sample)
	if r.RetentionDays() != 0 || r.Brains["team/acme"].AuditReads {
		t.Fatalf("defaults: %d, %v", r.RetentionDays(), r.Brains["team/acme"].AuditReads)
	}
	r = parse(t, strings.Replace(sample, "    weight: 0.8\n", "    weight: 0.8\n    audit_reads: true\n", 1)+"audit:\n  retention_days: 365\n")
	if r.RetentionDays() != 365 || !r.Brains["org/reference"].AuditReads || r.Brains["team/acme"].AuditReads {
		t.Fatalf("read: %d, %+v", r.RetentionDays(), r.Brains)
	}
	if _, err := Parse([]byte(sample + "audit:\n  retention_days: -1\n")); err == nil || !strings.Contains(err.Error(), "retention_days") {
		t.Fatalf("negative retention: %v", err)
	}
}

// A brain may declare the collection of its pyramid of syntheses, on the
// same server. The key of a confined brain then holds both collections; a
// whole-server key stays unconfined. The metadata key that lists the cited
// sources is "sources" unless the file names another.
func TestABrainMayDeclareItsPyramid(t *testing.T) {
	r := parse(t, strings.Replace(sample, "    collection: team-acme\n", "    collection: team-acme\n    pyramid:\n      collection: acme-pyramid\n      sources: preuves\n", 1))
	b := r.Brains["team/acme"]
	if b.Pyramid == nil || b.Pyramid.Collection != "acme-pyramid" || b.Pyramid.SourcesKey() != "preuves" || strings.Join(b.Collections(), ",") != "team-acme,acme-pyramid" {
		t.Fatalf("pyramid: %+v %v", b.Pyramid, b.Collections())
	}
	if ref := r.Brains["org/reference"]; ref.Pyramid != nil || strings.Join(ref.Collections(), ",") != "reference" {
		t.Fatalf("no pyramid: %+v %v", ref.Pyramid, ref.Collections())
	}
	plain := parse(t, strings.Replace(sample, "    collection: team-acme\n", "    collection: team-acme\n    pyramid: {collection: acme-pyramid}\n", 1))
	if p := plain.Brains["team/acme"].Pyramid; p.SourcesKey() != "sources" {
		t.Fatalf("default sources key: %q", p.SourcesKey())
	}
	whole := parse(t, strings.Replace(sample, "    collection: team-acme\n", "    pyramid: {collection: acme-pyramid}\n", 1))
	if c := whole.Brains["team/acme"].Collections(); c != nil {
		t.Fatalf("a whole-server key is not confined: %v", c)
	}
	for _, c := range []struct{ block, want string }{
		{"    pyramid: {collection: team-acme}\n", "must differ"},
		{"    pyramid: {sources: preuves}\n", "pyramid needs a collection"},
	} {
		_, err := Parse([]byte(strings.Replace(sample, "    collection: team-acme\n", "    collection: team-acme\n"+c.block, 1)))
		if err == nil || !strings.Contains(err.Error(), c.want) || !strings.Contains(err.Error(), "team/acme") {
			t.Fatalf("%q: %v", c.block, err)
		}
	}
}
