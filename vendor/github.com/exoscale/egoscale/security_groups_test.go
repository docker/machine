package egoscale

import (
	"testing"
)

func TestGroupsRequests(t *testing.T) {
	var _ AsyncCommand = (*AuthorizeSecurityGroupEgress)(nil)
	var _ onBeforeHook = (*AuthorizeSecurityGroupEgress)(nil)
	var _ AsyncCommand = (*AuthorizeSecurityGroupIngress)(nil)
	var _ onBeforeHook = (*AuthorizeSecurityGroupIngress)(nil)
	var _ Command = (*CreateSecurityGroup)(nil)
	var _ Command = (*DeleteSecurityGroup)(nil)
	var _ Command = (*ListSecurityGroups)(nil)
	var _ AsyncCommand = (*RevokeSecurityGroupEgress)(nil)
	var _ AsyncCommand = (*RevokeSecurityGroupIngress)(nil)
}

func TestAuthorizeSecurityGroupEgress(t *testing.T) {
	req := &AuthorizeSecurityGroupEgress{}
	if req.name() != "authorizeSecurityGroupEgress" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*AuthorizeSecurityGroupEgressResponse)
}

func TestAuthorizeSecurityGroupIngress(t *testing.T) {
	req := &AuthorizeSecurityGroupIngress{}
	if req.name() != "authorizeSecurityGroupIngress" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*AuthorizeSecurityGroupIngressResponse)
}

func TestCreateSecurityGroup(t *testing.T) {
	req := &CreateSecurityGroup{}
	if req.name() != "createSecurityGroup" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*CreateSecurityGroupResponse)
}

func TestDeleteSecurityGroup(t *testing.T) {
	req := &DeleteSecurityGroup{}
	if req.name() != "deleteSecurityGroup" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*booleanSyncResponse)
}

func TestListSecurityGroups(t *testing.T) {
	req := &ListSecurityGroups{}
	if req.name() != "listSecurityGroups" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListSecurityGroupsResponse)
}

func TestRevokeSecurityGroupEgress(t *testing.T) {
	req := &RevokeSecurityGroupEgress{}
	if req.name() != "revokeSecurityGroupEgress" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}

func TestRevokeSecurityGroupIngress(t *testing.T) {
	req := &RevokeSecurityGroupIngress{}
	if req.name() != "revokeSecurityGroupIngress" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}
