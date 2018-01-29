package egoscale

import (
	"testing"
)

func TestSSHKeyPairs(t *testing.T) {
	var _ AsyncCommand = (*ResetSSHKeyForVirtualMachine)(nil)
	var _ Command = (*RegisterSSHKeyPair)(nil)
	var _ Command = (*CreateSSHKeyPair)(nil)
	var _ Command = (*DeleteSSHKeyPair)(nil)
	var _ Command = (*ListSSHKeyPairs)(nil)
}

func TestResetSSHKeyForVirtualMachine(t *testing.T) {
	req := &ResetSSHKeyForVirtualMachine{}
	if req.name() != "resetSSHKeyForVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*ResetSSHKeyForVirtualMachineResponse)
}

func TestRegisterSSHKeyPair(t *testing.T) {
	req := &RegisterSSHKeyPair{}
	if req.name() != "registerSSHKeyPair" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*RegisterSSHKeyPairResponse)
}

func TestCreateSSHKeyPair(t *testing.T) {
	req := &CreateSSHKeyPair{}
	if req.name() != "createSSHKeyPair" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*CreateSSHKeyPairResponse)
}

func TestDeleteSSHKeyPair(t *testing.T) {
	req := &DeleteSSHKeyPair{}
	if req.name() != "deleteSSHKeyPair" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*booleanSyncResponse)
}

func TestListSSHKeyPairs(t *testing.T) {
	req := &ListSSHKeyPairs{}
	if req.name() != "listSSHKeyPairs" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListSSHKeyPairsResponse)
}
