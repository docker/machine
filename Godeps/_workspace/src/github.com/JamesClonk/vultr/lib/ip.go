package lib

import "net/url"

// IPv4 information of a virtual machine
type IPv4 struct {
	IP         string `json:"ip"`
	Netmask    string `json:"netmask"`
	Gateway    string `json:"gateway"`
	Type       string `json:"type"`
	ReverseDNS string `json:"reverse"`
}

// IPv6 information of a virtual machine
type IPv6 struct {
	IP          string `json:"ip"`
	Network     string `json:"network"`
	NetworkSize string `json:"network_size"`
	Type        string `json:"type"`
}

// ReverseDNSIPv6 information of a virtual machine
type ReverseDNSIPv6 struct {
	IP         string `json:"ip"`
	ReverseDNS string `json:"reverse"`
}

func (c *Client) ListIPv4(id string) (list []IPv4, err error) {
	var ipMap map[string][]IPv4
	if err := c.get(`server/list_ipv4?SUBID=`+id, &ipMap); err != nil {
		return nil, err
	}

	for _, iplist := range ipMap {
		for _, ip := range iplist {
			list = append(list, ip)
		}
	}
	return list, nil
}

func (c *Client) ListIPv6(id string) (list []IPv6, err error) {
	var ipMap map[string][]IPv6
	if err := c.get(`server/list_ipv6?SUBID=`+id, &ipMap); err != nil {
		return nil, err
	}

	for _, iplist := range ipMap {
		for _, ip := range iplist {
			list = append(list, ip)
		}
	}
	return list, nil
}

func (c *Client) ListIPv6ReverseDNS(id string) (list []ReverseDNSIPv6, err error) {
	var ipMap map[string][]ReverseDNSIPv6
	if err := c.get(`server/reverse_list_ipv6?SUBID=`+id, &ipMap); err != nil {
		return nil, err
	}

	for _, iplist := range ipMap {
		for _, ip := range iplist {
			list = append(list, ip)
		}
	}
	return list, nil
}

func (c *Client) DeleteIPv6ReverseDNS(id string, ip string) error {
	values := url.Values{
		"SUBID": {id},
		"ip":    {ip},
	}

	if err := c.post(`server/reverse_delete_ipv6`, values, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) SetIPv6ReverseDNS(id, ip, entry string) error {
	values := url.Values{
		"SUBID": {id},
		"ip":    {ip},
		"entry": {entry},
	}

	if err := c.post(`server/reverse_set_ipv6`, values, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) DefaultIPv4ReverseDNS(id, ip string) error {
	values := url.Values{
		"SUBID": {id},
		"ip":    {ip},
	}

	if err := c.post(`server/reverse_default_ipv4`, values, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) SetIPv4ReverseDNS(id, ip, entry string) error {
	values := url.Values{
		"SUBID": {id},
		"ip":    {ip},
		"entry": {entry},
	}

	if err := c.post(`server/reverse_set_ipv4`, values, nil); err != nil {
		return err
	}
	return nil
}
