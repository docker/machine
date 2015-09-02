package serviceaction

import "testing"

func TestActionValue(t *testing.T) {
	if Restart.String() != "restart" {
		t.Fatal("Expected %q but got %q", "install", Restart.String())
	}
}
