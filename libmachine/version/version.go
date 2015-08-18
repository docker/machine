package version

var (
	// ApiVersion dictates which version of the libmachine API this is.
	ApiVersion = 1

	// ConfigVersion dictates which version of the config.json format is
	// used. It needs to be bumped if there is a breaking change, and
	// therefore migration, introduced to the config file format.
	ConfigVersion = 2
)
