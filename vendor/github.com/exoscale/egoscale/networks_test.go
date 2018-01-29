package egoscale

import (
	"testing"
)

func TestListNetworksIsACommand(t *testing.T) {
	var _ Command = (*CreateNetwork)(nil)
	var _ AsyncCommand = (*DeleteNetwork)(nil)
	var _ Command = (*ListNetworks)(nil)
	var _ AsyncCommand = (*RestartNetwork)(nil)
	var _ AsyncCommand = (*UpdateNetwork)(nil)
}

func TestListNetworks(t *testing.T) {
	req := &ListNetworks{}
	if req.name() != "listNetworks" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListNetworksResponse)
}

func TestCreateNetwork(t *testing.T) {
	req := &CreateNetwork{}
	if req.name() != "createNetwork" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*CreateNetworkResponse)
}

func TestRestartNetwork(t *testing.T) {
	req := &RestartNetwork{}
	if req.name() != "restartNetwork" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*RestartNetworkResponse)
}

func TestUpdateNetwork(t *testing.T) {
	req := &UpdateNetwork{}
	if req.name() != "updateNetwork" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*UpdateNetworkResponse)
}

func TestDeleteNetwork(t *testing.T) {
	req := &DeleteNetwork{}
	if req.name() != "deleteNetwork" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}
