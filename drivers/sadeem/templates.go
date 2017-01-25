package sadeem

import (
	"net"
	"net/http"
	"net/url"
)

type ListResp struct {
	Href string        `json:"href"`
	Data []*ListObject `json:"data"`
}

type ListObject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Client struct {
	client    *http.Client
	BaseURL   *url.URL
	UserAgent string
	ApiKey    string
	ClientId  string
}

type VMResp struct {
	Href string `json:"href"`
	Data *VM    `json:"data"`
}
type VM struct {
	Hostname   string     `json:"hostname"`
	Id         string     `json:"id"`
	CpuShare   int        `json:"cpu_shares"`
	VCpu       int        `json:"vcpu"`
	Memory     int        `json:"ram"`
	Datacenter string     `json:"datacenter_id"`
	Template   string     `json:"template_value"`
	Status     string     `json:"status"`
	Network    []*Network `json:"network"`
}

type SSHKey struct {
	Id     string `json:"id,omitempty"`
	Key    string `json:"key"`
	Status string `json:"status,omitempty"`
	Name   string `json:"name"`
}

type CreateResponce struct {
	Id string `json:"id"`
}

type ErrorMessage struct {
	Error ErrorMessageBody `json:"error"`
}

type ErrorMessageBody struct {
	Code    int    `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

type Network struct {
	ID      uint   `json:"id"`
	IP      net.IP `json:"ip"`
	Gateway net.IP `json:"gateway"`
	Netmask net.IP `json:"netmask"`
	Mac     string `json: "mac"`
}

type VMNew struct {
	Hostname       string   `json:"hostname"`
	DatacenterId   string   `json:"dc_id"`
	ServiceOfferId string   `json:"service_offer_id"`
	TemplateId     string   `json:"template_id"`
	SSHKeyId       []string `json:"ssh_key_id,omitempty"`
	UserScript     string   `json:"user_script,omitempty"`
	PrivateNetwork bool     `json:"private_network,omitempty"`
	Firewall       []string `json:"firewall,omitempty"`
	EnableFirewall bool     `json:"enable_firewall,omitempty"`
}
