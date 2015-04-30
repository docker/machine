package gandi

type IpAddressInfo struct {
	Id      int    `xmlrpc:"id"`
	Version int    `xmlrpc:"version"`
	Ip      string `xmlrpc:"ip"`
}

type NetworkInterfaceInfo struct {
	Id   int             `xmlrpc:"id"`
	Type string          `xmlrpc:"type"`
	Ips  []IpAddressInfo `xmlrpc:"ips"`
}

type VmInfo struct {
	Id                int                    `xmlrpc:"id"`
	DcId              int                    `xmlrpc:"datacenter_id"`
	Hostname          string                 `xmlrpc:"hostname"`
	State             string                 `xmlrpc:"state"`
	NetworkInterfaces []NetworkInterfaceInfo `xmlrpc:"ifaces"`
}

type VmCreateRequest struct {
	DcId       int    `xmlrpc:"datacenter_id"`
	Hostname   string `xmlrpc:"hostname"`
	Memory     int    `xmlrpc:"memory"`
	Cores      int    `xmlrpc:"cores"`
	IpVersion  int    `xmlrpc:"ip_version"`
	SshKey     string `xmlrpc:"ssh_key"`
	RunCommand string `xmlrpc:"run"`
}

type DiskCreateRequest struct {
	DcId int    `xmlrpc:"datacenter_id"`
	Name string `xmlrpc:"name"`
	Size int    `xmlrpc:"size"`
}

type DatacenterInfo struct {
	Id   int    `xmlrpc:"id"`
	Name string `xmlrpc:"name"`
}

type ImageInfo struct {
	Id     int    `xmlrpc:"id"`
	Name   string `xmlrpc:"label"`
	Size   int    `xmlrpc:"size"`
	Kernel string `xmlrpc:"kernel_version"`
	DiskId int    `xmlrpc:"disk_id"`
}

type ImageFilter struct {
	Name string `xmlrpc:"label"`
	DcId int    `xmlrpc:"datacenter_id"`
}

type OperationInfo struct {
	Id     int    `xmlrpc:"id"`
	Status string `xmlrpc:"step"`
}

type KeyInfo struct {
	Id   int    `xmlrpc:"id"`
	Name string `xmlrpc:"name"`
}
