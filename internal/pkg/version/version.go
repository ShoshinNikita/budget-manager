// Package version contains global variables which must be set during the build process
// For example, version and commit hash
package version

// nolint:gochecknoglobals
var (
	// Version is a version of the app. It must be set during the build process with -ldflags flag
	Version = "unknown"
	// GitHash is the last commit hash. It must be set during the build process with -ldflags flag
	GitHash = "unknown"
)
