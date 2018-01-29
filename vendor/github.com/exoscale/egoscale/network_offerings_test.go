package egoscale

import (
	"testing"
)

func TestNetworkOfferings(t *testing.T) {
	var _ Command = (*ListNetworkOfferings)(nil)
}

func TestListNetworkOfferings(t *testing.T) {
	req := &ListNetworkOfferings{}
	if req.name() != "listNetworkOfferings" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListNetworkOfferingsResponse)
}
