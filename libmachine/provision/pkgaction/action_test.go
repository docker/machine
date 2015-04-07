package pkgaction

import "testing"

func TestActionValue(t *testing.T) {
	if Install.String() != "install" {
		t.Fatal("Expected %q but got %q", "install", Install.String())
	}
}
