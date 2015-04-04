package lib

import (
	"fmt"
	"net/url"
)

// Server (virtual machine) on Vultr account
type Server struct {
	ID               string  `json:"SUBID"`
	Name             string  `json:"label"`
	OS               string  `json:"os"`
	RAM              string  `json:"ram"`
	Disk             string  `json:"disk"`
	MainIP           string  `json:"main_ip"`
	VCpus            int     `json:"vcpu_count,string"`
	Location         string  `json:"location"`
	RegionID         int     `json:"DCID,string"`
	DefaultPassword  string  `json:"default_password"`
	Created          string  `json:"date_created"`
	PendingCharges   float64 `json:"pending_charges"`
	Status           string  `json:"status"`
	Cost             string  `json:"cost_per_month"`
	CurrentBandwidth float64 `json:"current_bandwidth_gb"`
	AllowedBandwidth float64 `json:"allowed_bandwidth_gb,string"`
	NetmaskV4        string  `json:"netmask_v4"`
	GatewayV4        string  `json:"gateway_v4"`
	PowerStatus      string  `json:"power_status"`
	PlanID           int     `json:"VPSPLANID,string"`
	NetworkV6        string  `json:"v6_network"`
	MainIPV6         string  `json:"v6_main_ip"`
	NetworkSizeV6    string  `json:"v6_network_size"`
	InternalIP       string  `json:"internal_ip"`
	KVMUrl           string  `json:"kvm_url"`
	AutoBackups      string  `json:"auto_backups"`
}

type ServerOptions struct {
	IPXEChainURL      string
	ISO               int
	Script            int
	Snapshot          string
	SSHKey            string
	IPV6              bool
	PrivateNetworking bool
	AutoBackups       bool
}

func (c *Client) GetServers() (servers []Server, err error) {
	var serverMap map[string]Server
	if err := c.get(`server/list`, &serverMap); err != nil {
		return nil, err
	}

	for _, server := range serverMap {
		servers = append(servers, server)
	}
	return servers, nil
}

func (c *Client) GetServer(id string) (server Server, err error) {
	if err := c.get(`server/list?SUBID=`+id, &server); err != nil {
		return Server{}, err
	}
	return server, nil
}

func (c *Client) CreateServer(name string, regionID, planID, osID int, options *ServerOptions) (Server, error) {
	values := url.Values{
		"label":     {name},
		"DCID":      {fmt.Sprintf("%v", regionID)},
		"VPSPLANID": {fmt.Sprintf("%v", planID)},
		"OSID":      {fmt.Sprintf("%v", osID)},
	}

	if options != nil {
		if options.IPXEChainURL != "" {
			values.Add("ipxe_chain_url", options.IPXEChainURL)
		}

		if options.ISO != 0 {
			values.Add("ISOID", fmt.Sprintf("%v", options.ISO))
		}

		if options.Script != 0 {
			values.Add("SCRIPTID", fmt.Sprintf("%v", options.Script))
		}

		if options.Snapshot != "" {
			values.Add("SNAPSHOTID", options.Snapshot)
		}

		if options.SSHKey != "" {
			values.Add("SSHKEYID", options.SSHKey)
		}

		values.Add("enable_ipv6", "no")
		if options.IPV6 {
			values.Set("enable_ipv6", "yes")
		}

		values.Add("enable_private_network", "no")
		if options.PrivateNetworking {
			values.Set("enable_private_network", "yes")
		}

		values.Add("auto_backups", "no")
		if options.AutoBackups {
			values.Set("auto_backups", "yes")
		}
	}

	var server Server
	if err := c.post(`server/create`, values, &server); err != nil {
		return Server{}, err
	}
	server.Name = name
	server.RegionID = regionID
	server.PlanID = planID

	return server, nil
}

func (c *Client) RenameServer(id, name string) error {
	values := url.Values{
		"SUBID": {id},
		"label": {name},
	}

	if err := c.post(`server/label_set`, values, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) StartServer(id string) error {
	values := url.Values{
		"SUBID": {id},
	}

	if err := c.post(`server/start`, values, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) HaltServer(id string) error {
	values := url.Values{
		"SUBID": {id},
	}

	if err := c.post(`server/halt`, values, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) RebootServer(id string) error {
	values := url.Values{
		"SUBID": {id},
	}

	if err := c.post(`server/reboot`, values, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) ReinstallServer(id string) error {
	values := url.Values{
		"SUBID": {id},
	}

	if err := c.post(`server/reinstall`, values, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) ChangeOSofServer(id string, osID int) error {
	values := url.Values{
		"SUBID": {id},
		"OSID":  {fmt.Sprintf("%v", osID)},
	}

	if err := c.post(`server/os_change`, values, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) ListOSforServer(id string) (os []OS, err error) {
	var osMap map[string]OS
	if err := c.get(`server/os_change_list?SUBID=`+id, &osMap); err != nil {
		return nil, err
	}

	for _, o := range osMap {
		os = append(os, o)
	}
	return os, nil
}

func (c *Client) DeleteServer(id string) error {
	values := url.Values{
		"SUBID": {id},
	}

	if err := c.post(`server/destroy`, values, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) BandwidthOfServer(id string) (bandwidth []map[string]string, err error) {
	var bandwidthMap map[string][][]string
	if err := c.get(`server/bandwidth?SUBID=`+id, &bandwidthMap); err != nil {
		return nil, err
	}

	// parse incoming bytes
	for _, b := range bandwidthMap["incoming_bytes"] {
		bMap := make(map[string]string)
		bMap["date"] = b[0]
		bMap["incoming"] = b[1]
		bandwidth = append(bandwidth, bMap)
	}

	// parse outgoing bytes (we'll assume that incoming and outgoing dates are always a match)
	for _, b := range bandwidthMap["outgoing_bytes"] {
		for i := range bandwidth {
			if bandwidth[i]["date"] == b[0] {
				bandwidth[i]["outgoing"] = b[1]
				break
			}
		}
	}

	return bandwidth, nil
}
