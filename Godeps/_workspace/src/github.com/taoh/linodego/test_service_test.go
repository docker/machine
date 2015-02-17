package linodego

import (
	"testing"
)

func TestEcho(t *testing.T) {
	client := NewClient(APIKey, nil)
	v := &TestResponse{}
	if err := client.Test.Echo("foo", "bar", v); err != nil {
		t.Fatal(err)
	}
	if v.Data["foo"] != "bar" {
		t.Fatalf("Expected bar, got %s", v.Data["foo"])
	}
}
