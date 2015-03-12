package provider

// ProviderType represents the type of a provider for a machine
type ProviderType int

const (
	None ProviderType = iota
	Local
	Remote
)

var providerTypes = []string{
	"",
	"Local",
	"Remote",
}

// Given a type, returns its string representation
func (t ProviderType) String() string {
	if int(t) >= 0 && int(t) < len(providerTypes) {
		return providerTypes[t]
	} else {
		return ""
	}
}
