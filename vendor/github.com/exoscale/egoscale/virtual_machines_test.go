package egoscale

import (
	"testing"
)

func TestVirtualMachines(t *testing.T) {
	var _ AsyncCommand = (*DeployVirtualMachine)(nil)
	var _ AsyncCommand = (*DestroyVirtualMachine)(nil)
	var _ AsyncCommand = (*RebootVirtualMachine)(nil)
	var _ AsyncCommand = (*StartVirtualMachine)(nil)
	var _ AsyncCommand = (*StopVirtualMachine)(nil)
	var _ AsyncCommand = (*ResetPasswordForVirtualMachine)(nil)
	var _ Command = (*UpdateVirtualMachine)(nil)
	var _ Command = (*ListVirtualMachines)(nil)
	var _ Command = (*GetVMPassword)(nil)
	var _ AsyncCommand = (*RestoreVirtualMachine)(nil)
	var _ Command = (*ChangeServiceForVirtualMachine)(nil)
	var _ AsyncCommand = (*ScaleVirtualMachine)(nil)
	var _ Command = (*RecoverVirtualMachine)(nil)
	var _ AsyncCommand = (*ExpungeVirtualMachine)(nil)
	var _ AsyncCommand = (*AddNicToVirtualMachine)(nil)
	var _ AsyncCommand = (*RemoveNicFromVirtualMachine)(nil)
	var _ AsyncCommand = (*UpdateDefaultNicForVirtualMachine)(nil)
}

func TestDeployVirtualMachine(t *testing.T) {
	req := &DeployVirtualMachine{}
	if req.name() != "deployVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*DeployVirtualMachineResponse)
}

func TestDestroyVirtualMachine(t *testing.T) {
	req := &DestroyVirtualMachine{}
	if req.name() != "destroyVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*DestroyVirtualMachineResponse)
}

func TestRebootVirtualMachine(t *testing.T) {
	req := &RebootVirtualMachine{}
	if req.name() != "rebootVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*RebootVirtualMachineResponse)
}

func TestStartVirtualMachine(t *testing.T) {
	req := &StartVirtualMachine{}
	if req.name() != "startVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*StartVirtualMachineResponse)
}

func TestStopVirtualMachine(t *testing.T) {
	req := &StopVirtualMachine{}
	if req.name() != "stopVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*StopVirtualMachineResponse)
}

func TestResetPasswordForVirtualMachine(t *testing.T) {
	req := &ResetPasswordForVirtualMachine{}
	if req.name() != "resetPasswordForVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*ResetPasswordForVirtualMachineResponse)
}

func TestUpdateVirtualMachine(t *testing.T) {
	req := &UpdateVirtualMachine{}
	if req.name() != "updateVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*UpdateVirtualMachineResponse)
}

func TestListVirtualMachines(t *testing.T) {
	req := &ListVirtualMachines{}
	if req.name() != "listVirtualMachines" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListVirtualMachinesResponse)
}

func TestGetVMPassword(t *testing.T) {
	req := &GetVMPassword{}
	if req.name() != "getVMPassword" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*GetVMPasswordResponse)
}

func TestRestoreVirtualMachine(t *testing.T) {
	req := &RestoreVirtualMachine{}
	if req.name() != "restoreVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*RestoreVirtualMachineResponse)
}

func TestChangeServiceForVirtualMachine(t *testing.T) {
	req := &ChangeServiceForVirtualMachine{}
	if req.name() != "changeServiceForVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ChangeServiceForVirtualMachineResponse)
}

func TestScaleVirtualMachine(t *testing.T) {
	req := &ScaleVirtualMachine{}
	if req.name() != "scaleVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}

func TestRecoverVirtualMachine(t *testing.T) {
	req := &RecoverVirtualMachine{}
	if req.name() != "recoverVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*RecoverVirtualMachineResponse)
}

func TestExpungeVirtualMachine(t *testing.T) {
	req := &ExpungeVirtualMachine{}
	if req.name() != "expungeVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}

func TestAddNicToVirtualMachine(t *testing.T) {
	req := &AddNicToVirtualMachine{}
	if req.name() != "addNicToVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*AddNicToVirtualMachineResponse)
}

func TestRemoveNicFromVirtualMachine(t *testing.T) {
	req := &RemoveNicFromVirtualMachine{}
	if req.name() != "removeNicFromVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*RemoveNicFromVirtualMachineResponse)
}

func TestUpdateDefaultNicForVirtualMachine(t *testing.T) {
	req := &UpdateDefaultNicForVirtualMachine{}
	if req.name() != "updateDefaultNicForVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*UpdateDefaultNicForVirtualMachineResponse)
}
