/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcloudair

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/url"
	"os"
	"strconv"

	types "github.com/vmware/govcloudair/types/v56"
)

type VApp struct {
	VApp *types.VApp
	c    *Client
}

func NewVApp(c *Client) *VApp {
	return &VApp{
		VApp: new(types.VApp),
		c:    c,
	}
}

func (v *VApp) Refresh() error {

	if v.VApp.HREF == "" {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	u, _ := url.ParseRequestURI(v.VApp.HREF)

	req := v.c.NewRequest(map[string]string{}, "GET", *u, nil)

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error retrieving task: %s", err)
	}

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	v.VApp = &types.VApp{}

	if err = decodeBody(resp, v.VApp); err != nil {
		return fmt.Errorf("error decoding task response: %s", err)
	}

	// The request was successful
	return nil
}

func (v *VApp) ComposeVApp(orgvdcnetwork OrgVDCNetwork, vapptemplate VAppTemplate, name string, description string) (Task, error) {

	if vapptemplate.VAppTemplate.Children == nil || orgvdcnetwork.OrgVDCNetwork == nil {
		return Task{}, fmt.Errorf("can't compose a new vApp, objects passed are not valid")
	}

	// Build request XML
	vcomp := &types.ComposeVAppParams{
		Ovf:         "http://schemas.dmtf.org/ovf/envelope/1",
		Xsi:         "http://www.w3.org/2001/XMLSchema-instance",
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Deploy:      false,
		Name:        name,
		PowerOn:     false,
		Description: description,
		InstantiationParams: &types.InstantiationParams{
			NetworkConfigSection: &types.NetworkConfigSection{
				Info: "Configuration parameters for logical networks",
				NetworkConfig: &types.VAppNetworkConfiguration{
					NetworkName: orgvdcnetwork.OrgVDCNetwork.Name,
					Configuration: &types.NetworkConfiguration{
						FenceMode: "bridged",
						ParentNetwork: &types.Reference{
							HREF: orgvdcnetwork.OrgVDCNetwork.HREF,
							Name: orgvdcnetwork.OrgVDCNetwork.Name,
							Type: orgvdcnetwork.OrgVDCNetwork.Type,
						},
					},
				},
			},
		},
		SourcedItem: &types.SourcedCompositionItemParam{
			Source: &types.Reference{
				HREF: vapptemplate.VAppTemplate.Children.VM[0].HREF,
				Name: vapptemplate.VAppTemplate.Children.VM[0].Name,
			},
			InstantiationParams: &types.InstantiationParams{
				NetworkConnectionSection: &types.NetworkConnectionSection{
					Type: vapptemplate.VAppTemplate.Children.VM[0].NetworkConnectionSection.Type,
					HREF: vapptemplate.VAppTemplate.Children.VM[0].NetworkConnectionSection.HREF,
					Info: "Network config for sourced item",
					PrimaryNetworkConnectionIndex: vapptemplate.VAppTemplate.Children.VM[0].NetworkConnectionSection.PrimaryNetworkConnectionIndex,
					NetworkConnection: &types.NetworkConnection{
						Network:                 orgvdcnetwork.OrgVDCNetwork.Name,
						NetworkConnectionIndex:  vapptemplate.VAppTemplate.Children.VM[0].NetworkConnectionSection.NetworkConnection.NetworkConnectionIndex,
						IsConnected:             true,
						IPAddressAllocationMode: "POOL",
					},
				},
			},
			NetworkAssignment: &types.NetworkAssignment{
				InnerNetwork:     orgvdcnetwork.OrgVDCNetwork.Name,
				ContainerNetwork: orgvdcnetwork.OrgVDCNetwork.Name,
			},
		},
	}

	output, err := xml.MarshalIndent(vcomp, "  ", "    ")
	if err != nil {
		return Task{}, fmt.Errorf("error marshaling vapp compose: %s", err)
	}

	debug := os.Getenv("GOVCLOUDAIR_DEBUG")

	if debug == "true" {
		fmt.Printf("\n\nXML DEBUG: %s\n\n", string(output))
	}

	b := bytes.NewBufferString(xml.Header + string(output))

	s := v.c.VCDVDCHREF
	s.Path += "/action/composeVApp"

	req := v.c.NewRequest(map[string]string{}, "POST", s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.composeVAppParams+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error instantiating a new vApp: %s", err)
	}

	if err = decodeBody(resp, v.VApp); err != nil {
		return Task{}, fmt.Errorf("error decoding vApp response: %s", err)
	}

	task := NewTask(v.c)
	task.Task = v.VApp.Tasks.Task[0]

	// The request was successful
	return *task, nil

}

func (v *VApp) PowerOn() (Task, error) {

	s, _ := url.ParseRequestURI(v.VApp.HREF)
	s.Path += "/power/action/powerOn"

	req := v.c.NewRequest(map[string]string{}, "POST", *s, nil)

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error powering on vApp: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) PowerOff() (Task, error) {

	s, _ := url.ParseRequestURI(v.VApp.HREF)
	s.Path += "/power/action/powerOff"

	req := v.c.NewRequest(map[string]string{}, "POST", *s, nil)

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error powering off vApp: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) Reboot() (Task, error) {

	s, _ := url.ParseRequestURI(v.VApp.HREF)
	s.Path += "/power/action/reboot"

	req := v.c.NewRequest(map[string]string{}, "POST", *s, nil)

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error rebooting vApp: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) Reset() (Task, error) {

	s, _ := url.ParseRequestURI(v.VApp.HREF)
	s.Path += "/power/action/reset"

	req := v.c.NewRequest(map[string]string{}, "POST", *s, nil)

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error resetting vApp: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) Suspend() (Task, error) {

	s, _ := url.ParseRequestURI(v.VApp.HREF)
	s.Path += "/power/action/suspend"

	req := v.c.NewRequest(map[string]string{}, "POST", *s, nil)

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error suspending vApp: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) Shutdown() (Task, error) {

	s, _ := url.ParseRequestURI(v.VApp.HREF)
	s.Path += "/power/action/shutdown"

	req := v.c.NewRequest(map[string]string{}, "POST", *s, nil)

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error shutting down vApp: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) Undeploy() (Task, error) {

	vu := &types.UndeployVAppParams{
		Xmlns:               "http://www.vmware.com/vcloud/v1.5",
		UndeployPowerAction: "powerOff",
	}

	output, err := xml.MarshalIndent(vu, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	debug := os.Getenv("GOVCLOUDAIR_DEBUG")

	if debug == "true" {
		fmt.Printf("\n\nXML DEBUG: %s\n\n", string(output))
	}

	b := bytes.NewBufferString(xml.Header + string(output))

	s, _ := url.ParseRequestURI(v.VApp.HREF)
	s.Path += "/action/undeploy"

	req := v.c.NewRequest(map[string]string{}, "POST", *s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.undeployVAppParams+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error undeploy vApp: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) Deploy() (Task, error) {

	vu := &types.DeployVAppParams{
		Xmlns:   "http://www.vmware.com/vcloud/v1.5",
		PowerOn: false,
	}

	output, err := xml.MarshalIndent(vu, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	debug := os.Getenv("GOVCLOUDAIR_DEBUG")

	if debug == "true" {
		fmt.Printf("\n\nXML DEBUG: %s\n\n", string(output))
	}

	b := bytes.NewBufferString(xml.Header + string(output))

	s, _ := url.ParseRequestURI(v.VApp.HREF)
	s.Path += "/action/deploy"

	req := v.c.NewRequest(map[string]string{}, "POST", *s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.deployVAppParams+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error undeploy vApp: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) Delete() (Task, error) {

	s, _ := url.ParseRequestURI(v.VApp.HREF)

	req := v.c.NewRequest(map[string]string{}, "DELETE", *s, nil)

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error deleting vApp: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) RunCustomizationScript(computername, script string) (Task, error) {

	err := v.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing vapp before running customization: %v", err)
	}

	// Check if VApp Children is populated
	if v.VApp.Children == nil {
		return Task{}, fmt.Errorf("vApp doesn't contain any children, aborting customization")
	}

	vu := &types.GuestCustomizationSection{
		Ovf:   "http://schemas.dmtf.org/ovf/envelope/1",
		Xsi:   "http://www.w3.org/2001/XMLSchema-instance",
		Xmlns: "http://www.vmware.com/vcloud/v1.5",

		HREF:                v.VApp.Children.VM[0].HREF,
		Type:                "application/vnd.vmware.vcloud.guestCustomizationSection+xml",
		Info:                "Specifies Guest OS Customization Settings",
		Enabled:             true,
		ComputerName:        computername,
		CustomizationScript: script,
	}

	output, err := xml.MarshalIndent(vu, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	debug := os.Getenv("GOVCLOUDAIR_DEBUG")

	if debug == "true" {
		fmt.Printf("\n\nXML DEBUG: %s\n\n", string(output))
	}

	b := bytes.NewBufferString(xml.Header + string(output))

	s, _ := url.ParseRequestURI(v.VApp.Children.VM[0].HREF)
	s.Path += "/guestCustomizationSection/"

	req := v.c.NewRequest(map[string]string{}, "PUT", *s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.guestCustomizationSection+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error customizing VM: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) GetStatus() (string, error) {
	err := v.Refresh()
	if err != nil {
		return "", fmt.Errorf("error refreshing vapp: %v", err)
	}
	return types.VAppStatuses[v.VApp.Status], nil
}

func (v *VApp) ChangeCPUcount(size int) (Task, error) {

	err := v.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing vapp before running customization: %v", err)
	}

	// Check if VApp Children is populated
	if v.VApp.Children == nil {
		return Task{}, fmt.Errorf("vApp doesn't contain any children, aborting customization")
	}

	newcpu := &types.OVFItem{
		XmlnsRasd:       "http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData",
		XmlnsVCloud:     "http://www.vmware.com/vcloud/v1.5",
		XmlnsXsi:        "http://www.w3.org/2001/XMLSchema-instance",
		VCloudHREF:      v.VApp.Children.VM[0].HREF + "/virtualHardwareSection/cpu",
		VCloudType:      "application/vnd.vmware.vcloud.rasdItem+xml",
		AllocationUnits: "hertz * 10^6",
		Description:     "Number of Virtual CPUs",
		ElementName:     strconv.Itoa(size) + " virtual CPU(s)",
		InstanceID:      4,
		Reservation:     0,
		ResourceType:    3,
		VirtualQuantity: size,
		Weight:          0,
		Link: &types.Link{
			HREF: v.VApp.Children.VM[0].HREF + "/virtualHardwareSection/cpu",
			Rel:  "edit",
			Type: "application/vnd.vmware.vcloud.rasdItem+xml",
		},
	}

	output, err := xml.MarshalIndent(newcpu, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	debug := os.Getenv("GOVCLOUDAIR_DEBUG")

	if debug == "true" {
		fmt.Printf("\n\nXML DEBUG: %s\n\n", string(output))
	}

	b := bytes.NewBufferString(xml.Header + string(output))

	s, _ := url.ParseRequestURI(v.VApp.Children.VM[0].HREF)
	s.Path += "/virtualHardwareSection/cpu"

	req := v.c.NewRequest(map[string]string{}, "PUT", *s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.rasdItem+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error customizing VM: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) ChangeMemorySize(size int) (Task, error) {

	err := v.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing vapp before running customization: %v", err)
	}

	// Check if VApp Children is populated
	if v.VApp.Children == nil {
		return Task{}, fmt.Errorf("vApp doesn't contain any children, aborting customization")
	}

	newmem := &types.OVFItem{
		XmlnsRasd:       "http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData",
		XmlnsVCloud:     "http://www.vmware.com/vcloud/v1.5",
		XmlnsXsi:        "http://www.w3.org/2001/XMLSchema-instance",
		VCloudHREF:      v.VApp.Children.VM[0].HREF + "/virtualHardwareSection/memory",
		VCloudType:      "application/vnd.vmware.vcloud.rasdItem+xml",
		AllocationUnits: "byte * 2^20",
		Description:     "Memory Size",
		ElementName:     strconv.Itoa(size) + " MB of memory",
		InstanceID:      5,
		Reservation:     0,
		ResourceType:    4,
		VirtualQuantity: size,
		Weight:          0,
		Link: &types.Link{
			HREF: v.VApp.Children.VM[0].HREF + "/virtualHardwareSection/memory",
			Rel:  "edit",
			Type: "application/vnd.vmware.vcloud.rasdItem+xml",
		},
	}

	output, err := xml.MarshalIndent(newmem, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	debug := os.Getenv("GOVCLOUDAIR_DEBUG")

	if debug == "true" {
		fmt.Printf("\n\nXML DEBUG: %s\n\n", string(output))
	}

	b := bytes.NewBufferString(xml.Header + string(output))

	s, _ := url.ParseRequestURI(v.VApp.Children.VM[0].HREF)
	s.Path += "/virtualHardwareSection/memory"

	req := v.c.NewRequest(map[string]string{}, "PUT", *s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.rasdItem+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error customizing VM: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}
