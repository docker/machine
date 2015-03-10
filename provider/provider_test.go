package provider

import (
	"testing"
)

func TestProviderType(t *testing.T) {
	if None.String() != "" {
		t.Fatal("None provider type should be empty string")
	}
	if Local.String() != "Local" {
		t.Fatal("Local provider type should be 'Local'")
	}
	if Remote.String() != "Remote" {
		t.Fatal("Remote provider type should be 'Remote'")
	}
}
