package egoscale

import (
	"testing"
)

func TestApis(t *testing.T) {
	var _ Command = (*ListAPIs)(nil)
}

func TestListAPIs(t *testing.T) {
	req := &ListAPIs{}
	if req.name() != "listApis" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListAPIsResponse)
}
