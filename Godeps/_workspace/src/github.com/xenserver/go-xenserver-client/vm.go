package client

import (
	"fmt"
	"github.com/nilshell/xmlrpc"
	"strconv"
)

type VM XenAPIObject

func (self *VM) Clone(name_label string) (new_instance *VM, err error) {
	new_instance = new(VM)

	result := APIResult{}
	err = self.Client.APICall(&result, "VM.clone", self.Ref, name_label)
	if err != nil {
		return nil, err
	}
	new_instance.Ref = result.Value.(string)
	new_instance.Client = self.Client
	return
}

func (self *VM) Copy(new_name string, targetSr *SR) (new_instance *VM, err error) {
	new_instance = new(VM)

	result := APIResult{}
	err = self.Client.APICall(&result, "VM.copy", self.Ref, new_name, targetSr.Ref)
	if err != nil {
		return nil, err
	}
	new_instance.Ref = result.Value.(string)
	new_instance.Client = self.Client
	return
}

func (self *VM) Snapshot(label string) (snapshot *VM, err error) {
	snapshot = new(VM)

	result := APIResult{}
	err = self.Client.APICall(&result, "VM.snapshot", self.Ref, label)
	if err != nil {
		return nil, err
	}
	snapshot.Ref = result.Value.(string)
	snapshot.Client = self.Client
	return
}

func (self *VM) Provision() (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.provision", self.Ref)
	if err != nil {
		return err
	}
	return
}

func (self *VM) Destroy() (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.destroy", self.Ref)
	if err != nil {
		return err
	}
	return
}

func (self *VM) Start(paused, force bool) (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.start", self.Ref, paused, force)
	if err != nil {
		return err
	}
	return
}

func (self *VM) StartOn(host *Host, paused, force bool) (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.start_on", self.Ref, host.Ref, paused, force)
	if err != nil {
		return err
	}
	return
}

func (self *VM) CleanShutdown() (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.clean_shutdown", self.Ref)
	if err != nil {
		return err
	}
	return
}

func (self *VM) HardShutdown() (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.hard_shutdown", self.Ref)
	if err != nil {
		return err
	}
	return
}

func (self *VM) CleanReboot() (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.clean_reboot", self.Ref)
	if err != nil {
		return err
	}
	return
}

func (self *VM) HardReboot() (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.hard_reboot", self.Ref)
	if err != nil {
		return err
	}
	return
}

func (self *VM) Unpause() (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.unpause", self.Ref)
	if err != nil {
		return err
	}
	return
}

func (self *VM) Resume(paused, force bool) (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.resume", self.Ref, paused, force)
	if err != nil {
		return err
	}
	return
}

func (self *VM) GetHVMBootPolicy() (bootOrder string, err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.get_HVM_boot_policy", self.Ref)
	if err != nil {
		return "", err
	}
	bootOrder = ""
	if result.Value != nil {
		bootOrder = result.Value.(string)
	}

	return bootOrder, nil
}

func (self *VM) SetHVMBoot(policy, bootOrder string) (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.set_HVM_boot_policy", self.Ref, policy)
	if err != nil {
		return err
	}
	result = APIResult{}
	params := make(xmlrpc.Struct)
	params["order"] = bootOrder
	err = self.Client.APICall(&result, "VM.set_HVM_boot_params", self.Ref, params)
	if err != nil {
		return err
	}
	return
}

func (self *VM) SetPVBootloader(pv_bootloader, pv_args string) (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.set_PV_bootloader", self.Ref, pv_bootloader)
	if err != nil {
		return err
	}
	result = APIResult{}
	err = self.Client.APICall(&result, "VM.set_PV_bootloader_args", self.Ref, pv_args)
	if err != nil {
		return err
	}
	return
}

func (self *VM) GetDomainId() (domid string, err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.get_domid", self.Ref)
	if err != nil {
		return "", err
	}
	domid = result.Value.(string)
	return domid, nil
}

func (self *VM) GetResidentOn() (host *Host, err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.get_resident_on", self.Ref)
	if err != nil {
		return nil, err
	}

	host = new(Host)
	host.Ref = result.Value.(string)
	host.Client = self.Client

	return host, nil
}

func (self *VM) GetPowerState() (state string, err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.get_power_state", self.Ref)
	if err != nil {
		return "", err
	}
	state = result.Value.(string)
	return state, nil
}

func (self *VM) GetUuid() (uuid string, err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.get_uuid", self.Ref)
	if err != nil {
		return "", err
	}
	uuid = result.Value.(string)
	return uuid, nil
}

func (self *VM) GetVBDs() (vbds []VBD, err error) {
	vbds = make([]VBD, 0)
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.get_VBDs", self.Ref)
	if err != nil {
		return vbds, err
	}
	for _, elem := range result.Value.([]interface{}) {
		vbd := VBD{}
		vbd.Ref = elem.(string)
		vbd.Client = self.Client
		vbds = append(vbds, vbd)
	}

	return vbds, nil
}

func (self *VM) GetAllowedVBDDevices() (devices []string, err error) {
	var device string
	devices = make([]string, 0)
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.get_allowed_VBD_devices", self.Ref)
	if err != nil {
		return devices, err
	}
	for _, elem := range result.Value.([]interface{}) {
		device = elem.(string)
		devices = append(devices, device)
	}

	return devices, nil
}

func (self *VM) GetVIFs() (vifs []VIF, err error) {
	vifs = make([]VIF, 0)
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.get_VIFs", self.Ref)
	if err != nil {
		return vifs, err
	}
	for _, elem := range result.Value.([]interface{}) {
		vif := VIF{}
		vif.Ref = elem.(string)
		vif.Client = self.Client
		vifs = append(vifs, vif)
	}

	return vifs, nil
}

func (self *VM) GetAllowedVIFDevices() (devices []string, err error) {
	var device string
	devices = make([]string, 0)
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.get_allowed_VIF_devices", self.Ref)
	if err != nil {
		return devices, err
	}
	for _, elem := range result.Value.([]interface{}) {
		device = elem.(string)
		devices = append(devices, device)
	}

	return devices, nil
}

func (self *VM) GetDisks() (vdis []*VDI, err error) {
	// Return just data disks (non-isos)
	vdis = make([]*VDI, 0)
	vbds, err := self.GetVBDs()
	if err != nil {
		return nil, err
	}

	for _, vbd := range vbds {
		rec, err := vbd.GetRecord()
		if err != nil {
			return nil, err
		}
		if rec["type"] == "Disk" {

			vdi, err := vbd.GetVDI()
			if err != nil {
				return nil, err
			}
			vdis = append(vdis, vdi)

		}
	}
	return vdis, nil
}

func (self *VM) GetGuestMetricsRef() (ref string, err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.get_guest_metrics", self.Ref)
	if err != nil {
		return "", nil
	}
	ref = result.Value.(string)
	return ref, err
}

func (self *VM) GetGuestMetrics() (metrics map[string]interface{}, err error) {
	metrics_ref, err := self.GetGuestMetricsRef()
	if err != nil {
		return nil, err
	}
	if metrics_ref == "OpaqueRef:NULL" {
		return nil, nil
	}

	result := APIResult{}
	err = self.Client.APICall(&result, "VM_guest_metrics.get_record", metrics_ref)
	if err != nil {
		return nil, err
	}
	return result.Value.(xmlrpc.Struct), nil
}

func (self *VM) SetStaticMemoryRange(min, max uint64) (err error) {
	result := APIResult{}
	strMin := fmt.Sprintf("%d", min)
	strMax := fmt.Sprintf("%d", max)
	err = self.Client.APICall(&result, "VM.set_memory_limits", self.Ref, strMin, strMax, strMin, strMax)
	if err != nil {
		return err
	}
	return
}

func (self *VM) ConnectVdi(vdi *VDI, vdiType VDIType, userdevice string) (err error) {

	// 1. Create a VBD
	if userdevice == "" {
		userdevice = "autodetect"
	}

	vbd_rec := make(xmlrpc.Struct)
	vbd_rec["VM"] = self.Ref
	vbd_rec["VDI"] = vdi.Ref
	vbd_rec["userdevice"] = userdevice
	vbd_rec["empty"] = false
	vbd_rec["other_config"] = make(xmlrpc.Struct)
	vbd_rec["qos_algorithm_type"] = ""
	vbd_rec["qos_algorithm_params"] = make(xmlrpc.Struct)

	switch vdiType {
	case CD:
		vbd_rec["mode"] = "RO"
		vbd_rec["bootable"] = true
		vbd_rec["unpluggable"] = false
		vbd_rec["type"] = "CD"
	case Disk:
		vbd_rec["mode"] = "RW"
		vbd_rec["bootable"] = false
		vbd_rec["unpluggable"] = false
		vbd_rec["type"] = "Disk"
	case Floppy:
		vbd_rec["mode"] = "RW"
		vbd_rec["bootable"] = false
		vbd_rec["unpluggable"] = true
		vbd_rec["type"] = "Floppy"
	}

	result := APIResult{}
	err = self.Client.APICall(&result, "VBD.create", vbd_rec)

	if err != nil {
		return err
	}

	vbd_ref := result.Value.(string)

	result = APIResult{}
	err = self.Client.APICall(&result, "VBD.get_uuid", vbd_ref)

	/*
	   // 2. Plug VBD (Non need - the VM hasn't booted.
	   // @todo - check VM state
	   result = APIResult{}
	   err = self.Client.APICall(&result, "VBD.plug", vbd_ref)

	   if err != nil {
	       return err
	   }
	*/
	return
}

func (self *VM) DisconnectVdi(vdi *VDI) error {
	vbds, err := self.GetVBDs()
	if err != nil {
		return fmt.Errorf("Unable to get VM VBDs: %s", err.Error())
	}

	for _, vbd := range vbds {
		rec, err := vbd.GetRecord()
		if err != nil {
			return fmt.Errorf("Could not get record for VBD '%s': %s", vbd.Ref, err.Error())
		}

		if recVdi, ok := rec["VDI"].(string); ok {
			if recVdi == vdi.Ref {
				_ = vbd.Unplug()
				err = vbd.Destroy()
				if err != nil {
					return fmt.Errorf("Could not destroy VBD '%s': %s", vbd.Ref, err.Error())
				}

				return nil
			}
		}
	}

	return fmt.Errorf("Could not find VBD for VDI '%s'", vdi.Ref)
}

func (self *VM) SetPlatform(params map[string]string) (err error) {
	result := APIResult{}
	platform_rec := make(xmlrpc.Struct)
	for key, value := range params {
		platform_rec[key] = value
	}

	err = self.Client.APICall(&result, "VM.set_platform", self.Ref, platform_rec)

	if err != nil {
		return err
	}
	return
}

func (self *VM) ConnectNetwork(network *Network, device string) (vif *VIF, err error) {
	// Create the VIF

	vif_rec := make(xmlrpc.Struct)
	vif_rec["network"] = network.Ref
	vif_rec["VM"] = self.Ref
	vif_rec["MAC"] = ""
	vif_rec["device"] = device
	vif_rec["MTU"] = "1504"
	vif_rec["other_config"] = make(xmlrpc.Struct)
	vif_rec["MAC_autogenerated"] = true
	vif_rec["locking_mode"] = "network_default"
	vif_rec["qos_algorithm_type"] = ""
	vif_rec["qos_algorithm_params"] = make(xmlrpc.Struct)

	result := APIResult{}
	err = self.Client.APICall(&result, "VIF.create", vif_rec)

	if err != nil {
		return nil, err
	}

	vif = new(VIF)
	vif.Ref = result.Value.(string)
	vif.Client = self.Client

	return vif, nil
}

//      Setters

func (self *VM) SetVCpuMax(vcpus uint) (err error) {
	result := APIResult{}
	strVcpu := fmt.Sprintf("%d", vcpus)

	err = self.Client.APICall(&result, "VM.set_VCPUs_max", self.Ref, strVcpu)

	if err != nil {
		return err
	}
	return
}

func (self *VM) SetVCpuAtStartup(vcpus uint) (err error) {
	result := APIResult{}
	strVcpu := fmt.Sprintf("%d", vcpus)

	err = self.Client.APICall(&result, "VM.set_VCPUs_at_startup", self.Ref, strVcpu)

	if err != nil {
		return err
	}
	return
}

func (self *VM) SetIsATemplate(is_a_template bool) (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.set_is_a_template", self.Ref, is_a_template)
	if err != nil {
		return err
	}
	return
}

func (self *VM) SetOtherConfig(other_config map[string]string) (err error) {
	result := APIResult{}
	other_config_rec := make(xmlrpc.Struct)
	for key, value := range other_config {
		other_config_rec[key] = value
	}
	err = self.Client.APICall(&result, "VM.set_other_config", self.Ref, other_config_rec)
	if err != nil {
		return err
	}
	return
}

func (self *VM) SetNameLabel(name_label string) (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.set_name_label", self.Ref, name_label)
	if err != nil {
		return err
	}
	return
}

func (self *VM) SetDescription(description string) (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.set_name_description", self.Ref, description)
	if err != nil {
		return err
	}
	return
}

func (self *VM) SetVCPUsMax(vcpus uint) (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.set_VCPUs_max", self.Ref, strconv.Itoa(int(vcpus)))
	if err != nil {
		return err
	}
	return
}

func (self *VM) SetVCPUsAtStartup(vcpus uint) (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.set_VCPUs_at_startup", self.Ref, strconv.Itoa(int(vcpus)))
	if err != nil {
		return err
	}
	return
}

func (self *VM) SetSuspendSR(vdi *VDI) (err error) {
	result := APIResult{}
	var vdi_uuid string
	vdi_uuid, err = vdi.GetUuid()
	if err != nil {
		return err
	}
	err = self.Client.APICall(&result, "VM.set_suspend_SR", self.Ref, vdi_uuid)
	if err != nil {
		return err
	}
	return
}

func (self *VM) SetHaAlwaysRun(ha_always_run bool) (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VM.set_ha_always_run", self.Ref, ha_always_run)
	if err != nil {
		return err
	}
	return
}
