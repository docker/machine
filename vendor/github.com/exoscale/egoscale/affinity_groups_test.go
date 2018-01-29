package egoscale

import (
	"testing"
)

func TestAffinityGroups(t *testing.T) {
	var _ AsyncCommand = (*CreateAffinityGroup)(nil)
	var _ AsyncCommand = (*DeleteAffinityGroup)(nil)
	var _ Command = (*ListAffinityGroupTypes)(nil)
	var _ Command = (*ListAffinityGroups)(nil)
	var _ AsyncCommand = (*UpdateVMAffinityGroup)(nil)
}

func TestCreateAffinityGroup(t *testing.T) {
	req := &CreateAffinityGroup{}
	if req.name() != "createAffinityGroup" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*CreateAffinityGroupResponse)
}

func TestDeleteAffinityGroup(t *testing.T) {
	req := &DeleteAffinityGroup{}
	if req.name() != "deleteAffinityGroup" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}

func TestListAffinityGroups(t *testing.T) {
	req := &ListAffinityGroups{}
	if req.name() != "listAffinityGroups" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListAffinityGroupsResponse)
}

func TestListAffinityGroupTypes(t *testing.T) {
	req := &ListAffinityGroupTypes{}
	if req.name() != "listAffinityGroupTypes" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListAffinityGroupTypesResponse)
}

func TestUpdateVMAffinityGroup(t *testing.T) {
	req := &UpdateVMAffinityGroup{}
	if req.name() != "updateVMAffinityGroup" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*UpdateVMAffinityGroupResponse)
}
