package version

import "fmt"

var (
	// Version should be updated by hand at each release
	Version = "0.5.2"

	// GitCommit will be overwritten automatically by the build system
	GitCommit = "HEAD"
)

// FullVersion formats the version to be printed
func FullVersion() string {
	return fmt.Sprintf("%s ( %s )", Version, GitCommit)
}
