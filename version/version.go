package version

import "fmt"

var (
	// Version should be updated by hand at each release
	Version = "0.7.0-dev"

	// GitCommit will be overwritten automatically by the build system
	GitCommit = "HEAD"
)

// FullVersion formats the version to be printed
func FullVersion() string {
	return fmt.Sprintf("%s, build %s", Version, GitCommit)
}
