package egoscale

import (
	"testing"
)

func TestZones(t *testing.T) {
	var _ Command = (*ListZones)(nil)
}

func TestListZones(t *testing.T) {
	req := &ListZones{}
	if req.name() != "listZones" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListZonesResponse)
}
