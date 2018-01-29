package egoscale

import (
	"testing"
)

func TestServiceOfferings(t *testing.T) {
	var _ Command = (*ListServiceOfferings)(nil)
}

func TestListServiceOfferings(t *testing.T) {
	req := &ListServiceOfferings{}
	if req.name() != "listServiceOfferings" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListServiceOfferingsResponse)
}
