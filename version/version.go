package version

var (
	// ConfigVersion dictates which version of the config.json format is
	// used. It needs to be bumped if there is a breaking change, and
	// therefore migration, introduced to the config file format.
	ConfigVersion = 1

	// Version should be updated by hand at each release
	Version = "0.5.0-dev"

	// GitCommit will be overwritten automatically by the build system
	GitCommit = "HEAD"
)
