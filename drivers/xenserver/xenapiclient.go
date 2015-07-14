package xenserver

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/nilshell/xmlrpc"
	xsclient "github.com/xenserver/go-xenserver-client"
)

type XenAPIClient struct {
	xsclient.XenAPIClient
}

func assertUnique(objs []interface{}, obj_label, name_label string) (obj interface{}, err error) {
	switch len(objs) {
	case 1:
		return objs[0], nil
	case 0:
		return nil, fmt.Errorf("Unable to get a %s named %v", obj_label, name_label)
	default:
		return nil, fmt.Errorf("Too many %ss returned for name %v", obj_label, name_label)
	}
}

// Get Unique VM By NameLabel
func (c *XenAPIClient) GetUniqueVMByNameLabel(name_label string) (vm *xsclient.VM, err error) {
	vms, err := c.GetVMByNameLabel(name_label)
	if err != nil {
		return nil, err
	}

	var objs []interface{}
	for _, v := range vms {
		objs = append(objs, v)
	}

	obj, err := assertUnique(objs, "VM", name_label)

	if err != nil {
		return nil, err
	}

	return obj.(*xsclient.VM), nil
}

// Get Unique SR By NameLabel
func (c *XenAPIClient) GetUniqueSRByNameLabel(name_label string) (sr *xsclient.SR, err error) {
	srs, err := c.GetSRByNameLabel(name_label)
	if err != nil {
		return nil, err
	}

	var objs []interface{}
	for _, v := range srs {
		objs = append(objs, v)
	}

	obj, err := assertUnique(objs, "SR", name_label)

	if err != nil {
		return nil, err
	}

	return obj.(*xsclient.SR), nil
}

// Get Unique Host By NameLabel
func (c *XenAPIClient) GetUniqueHostByNameLabel(name_label string) (host *xsclient.Host, err error) {
	hosts, err := c.GetHostByNameLabel(name_label)
	if err != nil {
		return nil, err
	}

	var objs []interface{}
	for _, v := range hosts {
		objs = append(objs, v)
	}

	obj, err := assertUnique(objs, "Host", name_label)

	if err != nil {
		return nil, err
	}

	return obj.(*xsclient.Host), nil
}

// Get Unique Network By NameLabel
func (c *XenAPIClient) GetUniqueNetworkByNameLabel(name_label string) (network *xsclient.Network, err error) {
	networks, err := c.GetNetworkByNameLabel(name_label)
	if err != nil {
		return nil, err
	}

	var objs []interface{}
	for _, v := range networks {
		objs = append(objs, v)
	}

	obj, err := assertUnique(objs, "Network", name_label)

	if err != nil {
		return nil, err
	}

	return obj.(*xsclient.Network), nil
}

func NewXenAPIClient(host, username, password string) (c XenAPIClient) {
	c.Host = host
	c.Url = "https://" + host
	c.Username = username
	c.Password = password
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	c.RPC, _ = xmlrpc.NewClient(c.Url, tr)
	return
}
