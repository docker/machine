package egoscale

import (
	"testing"
)

func TestAddressess(t *testing.T) {
	var _ AsyncCommand = (*AssociateIPAddress)(nil)
	var _ AsyncCommand = (*DisassociateIPAddress)(nil)
	var _ Command = (*ListPublicIPAddresses)(nil)
	var _ AsyncCommand = (*UpdateIPAddress)(nil)
}

func TestAssociateIPAddress(t *testing.T) {
	req := &AssociateIPAddress{}
	if req.name() != "associateIpAddress" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*AssociateIPAddressResponse)
}

func TestDisassociateIPAddress(t *testing.T) {
	req := &DisassociateIPAddress{}
	if req.name() != "disassociateIpAddress" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}

func TestListPublicIPAddresses(t *testing.T) {
	req := &ListPublicIPAddresses{}
	if req.name() != "listPublicIpAddresses" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListPublicIPAddressesResponse)
}

func TestUpdateIPAddress(t *testing.T) {
	req := &UpdateIPAddress{}
	if req.name() != "updateIpAddress" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*UpdateIPAddressResponse)
}
