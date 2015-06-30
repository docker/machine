package client

import (
	"github.com/nilshell/xmlrpc"
)

type Network XenAPIObject

func (self *Network) GetAssignedIPs() (ipMap map[string]string, err error) {
	ipMap = make(map[string]string, 0)
	result := APIResult{}
	err = self.Client.APICall(&result, "network.get_assigned_ips", self.Ref)
	if err != nil {
		return ipMap, err
	}
	for k, v := range result.Value.(xmlrpc.Struct) {
		ipMap[k] = v.(string)
	}
	return ipMap, nil
}

func (self *Network) GetOtherConfig() (otherConfig map[string]string, err error) {
	otherConfig = make(map[string]string, 0)
	result := APIResult{}
	err = self.Client.APICall(&result, "network.get_other_config", self.Ref)
	if err != nil {
		return otherConfig, err
	}
	for k, v := range result.Value.(xmlrpc.Struct) {
		otherConfig[k] = v.(string)
	}
	return otherConfig, nil
}

func (self *Network) IsHostInternalManagementNetwork() (isHostInternalManagementNetwork bool, err error) {
	other_config, err := self.GetOtherConfig()
	if err != nil {
		return false, nil
	}
	value, ok := other_config["is_host_internal_management_network"]
	isHostInternalManagementNetwork = ok && value == "true"
	return isHostInternalManagementNetwork, nil
}

func (self *Network) Destroy() (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "network.destroy", self.Ref)
	if err != nil {
		return err
	}
	return
}
