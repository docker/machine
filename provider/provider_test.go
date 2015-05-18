package provider

import (
	"fmt"
	"math"
	"testing"
)

func TestProviderType(t *testing.T) {
	if Invalid.String() != "Invalid" {
		t.Fatal("Invalid provider type should be 'Invalid'")
	}
	if None.String() != "" {
		t.Fatal("None provider type should be empty string")
	}
	if Local.String() != "Local" {
		t.Fatal("Local provider type should be 'Local'")
	}
	if Remote.String() != "Remote" {
		t.Fatal("Remote provider type should be 'Remote'")
	}
	maxProviderTypeString := fmt.Sprintf("ProviderType%d", math.MaxUint16)
	if ProviderType(math.MaxUint16).String() != maxProviderTypeString {
		t.Fatalf("Unknown provider type should be %s", maxProviderTypeString)
	}
}
