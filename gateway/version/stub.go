package version

import "runtime/debug"

// SetForTest fixes the tag and the build info for a test of another
// package, and returns what restores them. Tests only.
func SetForTest(tag string, read func() (*debug.BuildInfo, bool)) (restore func()) {
	oldTag, oldRead := Tag, readBuildInfo
	Tag, readBuildInfo = tag, read
	return func() { Tag, readBuildInfo = oldTag, oldRead }
}
