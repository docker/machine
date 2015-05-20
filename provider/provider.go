package provider

import "fmt"

// ProviderType represents the type of a provider for a machine
type ProviderType uint

const (
	Invalid ProviderType = iota
	None
	Local
	Remote
)

var providerNames = []string{
	"Invalid",
	"",
	"Local",
	"Remote",
}

// Given a type, returns its string representation
func (t ProviderType) String() string {
	if int(t) < len(providerNames) {
		return providerNames[t]
	}
	return fmt.Sprintf("ProviderType%d", int(t))
}
