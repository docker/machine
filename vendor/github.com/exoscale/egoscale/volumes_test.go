package egoscale

import (
	"testing"
)

func TestVolumes(t *testing.T) {
	var _ Command = (*ListVolumes)(nil)
	var _ AsyncCommand = (*ResizeVolume)(nil)
}

func TestListVolumes(t *testing.T) {
	req := &ListVolumes{}
	if req.name() != "listVolumes" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListVolumesResponse)
}

func TestResizeVolume(t *testing.T) {
	req := &ResizeVolume{}
	if req.name() != "resizeVolume" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*ResizeVolumeResponse)
}
