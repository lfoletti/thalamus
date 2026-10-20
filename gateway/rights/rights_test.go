package rights

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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

// A brain needs its address, its key file and its collection.
func TestABrainNeedsItsThreeFields(t *testing.T) {
	for _, field := range []string{"addr: 127.0.0.1:47821", "keys_file: /srv/acme/keys.json", "collection: team-acme"} {
		bad := strings.Replace(sample, field, "", 1)
		_, err := Parse([]byte(bad))
		name := strings.SplitN(field, ":", 2)[0]
		if err == nil || !strings.Contains(err.Error(), name) {
			t.Fatalf("without %s: got %v", name, err)
		}
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
