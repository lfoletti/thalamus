package version

import (
	"runtime/debug"
	"testing"
)

func info(settings ...debug.BuildSetting) func() (*debug.BuildInfo, bool) {
	return func() (*debug.BuildInfo, bool) { return &debug.BuildInfo{Settings: settings}, true }
}

func module(v string, settings ...debug.BuildSetting) func() (*debug.BuildInfo, bool) {
	return func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: v}, Settings: settings}, true
	}
}

// Stub sets the tag and the build info of this test binary, and restores them.
func Stub(t *testing.T, tag string, read func() (*debug.BuildInfo, bool)) {
	t.Helper()
	t.Cleanup(SetForTest(tag, read))
}

// A build that set the tag is named by the tag alone: the build info is
// not even asked.
func TestTheTagNamesTheBuild(t *testing.T) {
	asked := func() (*debug.BuildInfo, bool) {
		t.Fatal("the build info was asked although the tag is set")
		return nil, false
	}
	if got := describe("v0.5.0", asked); got != "v0.5.0" {
		t.Fatalf("got %q", got)
	}
	if got := release("v0.5.0", asked); got != "v0.5.0" {
		t.Fatalf("release: got %q", got)
	}
}

// A binary fetched by go install carries the module version in its build
// info: that version names the build, before the commit.
func TestTheModuleVersionNamesAGoInstallBuild(t *testing.T) {
	rev := debug.BuildSetting{Key: "vcs.revision", Value: "abcdef1234567890"}
	for _, c := range []struct {
		name string
		info func() (*debug.BuildInfo, bool)
		want string
	}{
		{"a release", module("v0.5.0"), "v0.5.0"},
		{"a pseudo-version", module("v0.5.1-0.20261025120000-abcdef123456"), "v0.5.1-0.20261025120000-abcdef123456"},
		{"a local build says (devel)", module("(devel)", rev), "dev abcdef123456"},
		{"no module version", module("", rev), "dev abcdef123456"},
	} {
		if got := describe("", c.info); got != c.want {
			t.Fatalf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

// Without a tag, the build is "dev" with the commit it was built from,
// shortened to 12 characters, and "modified" when the tree was dirty.
// Without build info, or without a commit in it, it is "dev".
func TestWithoutATagTheCommitNamesTheBuild(t *testing.T) {
	for _, c := range []struct {
		name string
		info func() (*debug.BuildInfo, bool)
		want string
	}{
		{"no build info", func() (*debug.BuildInfo, bool) { return nil, false }, "dev"},
		{"no commit", info(debug.BuildSetting{Key: "GOOS", Value: "linux"}), "dev"},
		{"a commit", info(debug.BuildSetting{Key: "vcs.revision", Value: "abcdef1234567890abcdef"}), "dev abcdef123456"},
		{"a short commit", info(debug.BuildSetting{Key: "vcs.revision", Value: "abc"}), "dev abc"},
		{"a modified tree", info(debug.BuildSetting{Key: "vcs.revision", Value: "abcdef1234567890"}, debug.BuildSetting{Key: "vcs.modified", Value: "true"}), "dev abcdef123456 modified"},
		{"a clean tree", info(debug.BuildSetting{Key: "vcs.revision", Value: "abcdef1234567890"}, debug.BuildSetting{Key: "vcs.modified", Value: "false"}), "dev abcdef123456"},
	} {
		if got := describe("", c.info); got != c.want {
			t.Fatalf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

// String reads the tag of this build, then the build info of the binary.
func TestStringNamesThisBuild(t *testing.T) {
	Stub(t, "v9.9.9", func() (*debug.BuildInfo, bool) { return nil, false })
	if got := String(); got != "v9.9.9" {
		t.Fatalf("with a tag: got %q", got)
	}
	Stub(t, "", info(debug.BuildSetting{Key: "vcs.revision", Value: "0123456789abcdef"}, debug.BuildSetting{Key: "vcs.modified", Value: "true"}))
	if got := String(); got != "dev 0123456789ab modified" {
		t.Fatalf("without a tag: got %q", got)
	}
}

// Public names a release in full and a development build by "dev" alone:
// no commit leaves the gateway without a key.
func TestPublicNamesTheReleaseAndHidesTheCommit(t *testing.T) {
	Stub(t, "v9.9.9", func() (*debug.BuildInfo, bool) { return nil, false })
	if got := Public(); got != "v9.9.9" {
		t.Fatalf("with a tag: got %q", got)
	}
	Stub(t, "", module("v0.5.0"))
	if got := Public(); got != "v0.5.0" {
		t.Fatalf("go install: got %q", got)
	}
	Stub(t, "", info(debug.BuildSetting{Key: "vcs.revision", Value: "0123456789abcdef"}, debug.BuildSetting{Key: "vcs.modified", Value: "true"}))
	if got := Public(); got != "dev" {
		t.Fatalf("development build: got %q", got)
	}
}
