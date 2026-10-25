// Package version names the build of a binary: the release tag when the
// release set it, the module version when go install fetched a release, the
// commit otherwise.
package version

import "runtime/debug"

// Tag is set by the release build:
//
//	go build -ldflags "-X github.com/lfoletti/thalamus/gateway/version.Tag=v0.5.0"
//
// tools/release/build.sh and the Dockerfile pass it. A local build leaves it
// empty and is a development build. -X ignores a symbol it does not find, so
// the release workflow runs the built binary and checks what it prints.
var Tag string

// readBuildInfo is debug.ReadBuildInfo, replaced in the tests.
var readBuildInfo = debug.ReadBuildInfo

// String names this build in full: the tag, else the module version of a
// `go install …@v0.5.0`, else "dev" with the commit the binary was built
// from and "modified" when the tree was dirty (any untracked file counts,
// a binary built into the tree included). A build with -buildvcs=false, or
// outside a git tree, is "dev" alone; -trimpath keeps the commit.
func String() string { return describe(Tag, readBuildInfo) }

// Public names this build for anyone: the tag or the module version, else
// "dev" without the commit. GET /v1/version answers it without a key, and a
// development deployment does not tell the world which commit it runs.
func Public() string {
	if v := release(Tag, readBuildInfo); v != "" {
		return v
	}
	return "dev"
}

// release is the tag, else the module version when it names a release.
func release(tag string, info func() (*debug.BuildInfo, bool)) string {
	if tag != "" {
		return tag
	}
	if bi, ok := info(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return ""
}

func describe(tag string, info func() (*debug.BuildInfo, bool)) string {
	if v := release(tag, info); v != "" {
		return v
	}
	bi, ok := info()
	if !ok {
		return "dev"
	}
	var rev, modified string
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			modified = s.Value
		}
	}
	if rev == "" {
		return "dev"
	}
	if len(rev) > 12 {
		rev = rev[:12]
	}
	if modified == "true" {
		return "dev " + rev + " modified"
	}
	return "dev " + rev
}
