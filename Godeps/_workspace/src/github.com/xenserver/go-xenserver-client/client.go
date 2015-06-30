package client

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/nilshell/xmlrpc"
)

type XenAPIClient struct {
	Session  interface{}
	Host     string
	Url      string
	Username string
	Password string
	RPC      *xmlrpc.Client
}

type APIResult struct {
	Status           string
	Value            interface{}
	ErrorDescription string
}

type XenAPIObject struct {
	Ref    string
	Client *XenAPIClient
}

func (c *XenAPIClient) RPCCall(result interface{}, method string, params []interface{}) (err error) {
	log.Debugf("RPCCall method=%v params=%v\n", method, params)
	p := new(xmlrpc.Params)
	p.Params = params
	err = c.RPC.Call(method, *p, result)
	return err
}

func (client *XenAPIClient) Login() (err error) {
	//Do loging call
	result := xmlrpc.Struct{}

	params := make([]interface{}, 2)
	params[0] = client.Username
	params[1] = client.Password

	err = client.RPCCall(&result, "session.login_with_password", params)
	if err == nil {
		// err might not be set properly, so check the reference
		if result["Value"] == nil {
			return errors.New("Invalid credentials supplied")
		}
	}
	client.Session = result["Value"]
	return err
}

func (client *XenAPIClient) APICall(result *APIResult, method string, params ...interface{}) (err error) {
	if client.Session == nil {
		log.Errorf("no session\n")
		return fmt.Errorf("No session. Unable to make call")
	}

	//Make a params slice which will include the session
	p := make([]interface{}, len(params)+1)
	p[0] = client.Session

	if params != nil {
		for idx, element := range params {
			p[idx+1] = element
		}
	}

	res := xmlrpc.Struct{}

	err = client.RPCCall(&res, method, p)

	if err != nil {
		return err
	}

	result.Status = res["Status"].(string)

	if result.Status != "Success" {
		log.Errorf("Encountered an API error: %v %v", result.Status, res["ErrorDescription"])
		return fmt.Errorf("API Error: %s", res["ErrorDescription"])
	} else {
		result.Value = res["Value"]
	}
	return
}

func (client *XenAPIClient) GetHosts() (hosts []*Host, err error) {
	hosts = make([]*Host, 0)
	result := APIResult{}
	err = client.APICall(&result, "host.get_all")
	if err != nil {
		return hosts, err
	}
	for _, elem := range result.Value.([]interface{}) {
		host := new(Host)
		host.Ref = elem.(string)
		host.Client = client
		hosts = append(hosts, host)
	}
	return hosts, nil
}

func (client *XenAPIClient) GetPools() (pools []*Pool, err error) {
	pools = make([]*Pool, 0)
	result := APIResult{}
	err = client.APICall(&result, "pool.get_all")
	if err != nil {
		return pools, err
	}

	for _, elem := range result.Value.([]interface{}) {
		pool := new(Pool)
		pool.Ref = elem.(string)
		pool.Client = client
		pools = append(pools, pool)
	}

	return pools, nil
}

func (client *XenAPIClient) GetDefaultSR() (sr *SR, err error) {
	pools, err := client.GetPools()

	if err != nil {
		return nil, err
	}

	pool_rec, err := pools[0].GetRecord()

	if err != nil {
		return nil, err
	}

	if pool_rec["default_SR"] == "" {
		return nil, errors.New("No default_SR specified for the pool.")
	}

	sr = new(SR)
	sr.Ref = pool_rec["default_SR"].(string)
	sr.Client = client

	return sr, nil
}

func (client *XenAPIClient) GetVMByUuid(vm_uuid string) (vm *VM, err error) {
	vm = new(VM)
	result := APIResult{}
	err = client.APICall(&result, "VM.get_by_uuid", vm_uuid)
	if err != nil {
		return nil, err
	}
	vm.Ref = result.Value.(string)
	vm.Client = client
	return
}

func (client *XenAPIClient) GetHostByUuid(host_uuid string) (host *Host, err error) {
	host = new(Host)
	result := APIResult{}
	err = client.APICall(&result, "host.get_by_uuid", host_uuid)
	if err != nil {
		return nil, err
	}
	host.Ref = result.Value.(string)
	host.Client = client
	return
}

func (client *XenAPIClient) GetVMByNameLabel(name_label string) (vms []*VM, err error) {
	vms = make([]*VM, 0)
	result := APIResult{}
	err = client.APICall(&result, "VM.get_by_name_label", name_label)
	if err != nil {
		return vms, err
	}

	for _, elem := range result.Value.([]interface{}) {
		vm := new(VM)
		vm.Ref = elem.(string)
		vm.Client = client
		vms = append(vms, vm)
	}

	return vms, nil
}

func (client *XenAPIClient) GetHostByNameLabel(name_label string) (hosts []*Host, err error) {
	hosts = make([]*Host, 0)
	result := APIResult{}
	err = client.APICall(&result, "host.get_by_name_label", name_label)
	if err != nil {
		return hosts, err
	}

	for _, elem := range result.Value.([]interface{}) {
		host := new(Host)
		host.Ref = elem.(string)
		host.Client = client
		hosts = append(hosts, host)
	}

	return hosts, nil
}

func (client *XenAPIClient) GetSRByNameLabel(name_label string) (srs []*SR, err error) {
	srs = make([]*SR, 0)
	result := APIResult{}
	err = client.APICall(&result, "SR.get_by_name_label", name_label)
	if err != nil {
		return srs, err
	}

	for _, elem := range result.Value.([]interface{}) {
		sr := new(SR)
		sr.Ref = elem.(string)
		sr.Client = client
		srs = append(srs, sr)
	}

	return srs, nil
}

func (client *XenAPIClient) GetNetworks() (networks []*Network, err error) {
	networks = make([]*Network, 0)
	result := APIResult{}
	err = client.APICall(&result, "network.get_all")
	if err != nil {
		return nil, err
	}

	for _, elem := range result.Value.([]interface{}) {
		network := new(Network)
		network.Ref = elem.(string)
		network.Client = client
		networks = append(networks, network)
	}

	return networks, nil
}

func (client *XenAPIClient) GetNetworkByUuid(network_uuid string) (network *Network, err error) {
	network = new(Network)
	result := APIResult{}
	err = client.APICall(&result, "network.get_by_uuid", network_uuid)
	if err != nil {
		return nil, err
	}
	network.Ref = result.Value.(string)
	network.Client = client
	return
}

func (client *XenAPIClient) GetNetworkByNameLabel(name_label string) (networks []*Network, err error) {
	networks = make([]*Network, 0)
	result := APIResult{}
	err = client.APICall(&result, "network.get_by_name_label", name_label)
	if err != nil {
		return networks, err
	}

	for _, elem := range result.Value.([]interface{}) {
		network := new(Network)
		network.Ref = elem.(string)
		network.Client = client
		networks = append(networks, network)
	}

	return networks, nil
}

func (client *XenAPIClient) GetVdiByNameLabel(name_label string) (vdis []*VDI, err error) {
	vdis = make([]*VDI, 0)
	result := APIResult{}
	err = client.APICall(&result, "VDI.get_by_name_label", name_label)
	if err != nil {
		return vdis, err
	}

	for _, elem := range result.Value.([]interface{}) {
		vdi := new(VDI)
		vdi.Ref = elem.(string)
		vdi.Client = client
		vdis = append(vdis, vdi)
	}

	return vdis, nil
}

func (client *XenAPIClient) GetSRByUuid(sr_uuid string) (sr *SR, err error) {
	sr = new(SR)
	result := APIResult{}
	err = client.APICall(&result, "SR.get_by_uuid", sr_uuid)
	if err != nil {
		return nil, err
	}
	sr.Ref = result.Value.(string)
	sr.Client = client
	return
}

func (client *XenAPIClient) GetVdiByUuid(vdi_uuid string) (vdi *VDI, err error) {
	vdi = new(VDI)
	result := APIResult{}
	err = client.APICall(&result, "VDI.get_by_uuid", vdi_uuid)
	if err != nil {
		return nil, err
	}
	vdi.Ref = result.Value.(string)
	vdi.Client = client
	return
}

func (client *XenAPIClient) GetPIFs() (pifs []*PIF, err error) {
	pifs = make([]*PIF, 0)
	result := APIResult{}
	err = client.APICall(&result, "PIF.get_all")
	if err != nil {
		return pifs, err
	}
	for _, elem := range result.Value.([]interface{}) {
		pif := new(PIF)
		pif.Ref = elem.(string)
		pif.Client = client
		pifs = append(pifs, pif)
	}

	return pifs, nil
}

func (client *XenAPIClient) CreateTask() (task *Task, err error) {
	result := APIResult{}
	err = client.APICall(&result, "task.create", "packer-task", "Packer task")

	if err != nil {
		return
	}

	task = new(Task)
	task.Ref = result.Value.(string)
	task.Client = client
	return
}

func (client *XenAPIClient) CreateNetwork(name_label string, name_description string, bridge string) (network *Network, err error) {
	network = new(Network)

	net_rec := make(xmlrpc.Struct)
	net_rec["name_label"] = name_label
	net_rec["name_description"] = name_description
	net_rec["bridge"] = bridge
	net_rec["other_config"] = make(xmlrpc.Struct)

	result := APIResult{}
	err = client.APICall(&result, "network.create", net_rec)
	if err != nil {
		return nil, err
	}
	network.Ref = result.Value.(string)
	network.Client = client

	return network, nil
}

func NewXenAPIClient(host, username, password string) (client XenAPIClient) {
	client.Host = host
	client.Url = "http://" + host
	client.Username = username
	client.Password = password
	client.RPC, _ = xmlrpc.NewClient(client.Url, nil)
	return
}
