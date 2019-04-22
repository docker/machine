package yandex

import (
	"testing"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

func TestNewYandexCloudClient(t *testing.T) {
	type args struct {
		d *Driver
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "use token",
			args: args{
				d: &Driver{
					Token: "some-test-token",
					ServiceAccountKeyFile: "",
				},
			},
			wantErr: false,
		},
		{
			name: "service account key file doest not exist",
			args: args{
				d: &Driver{
					Token: "",
					ServiceAccountKeyFile: "some-not-exist-sa-key-file",
				},
			},
			wantErr: true,
		},
		{
			name: "both auth methods provided",
			args: args{
				d: &Driver{
					Token: "some-test-token",
					ServiceAccountKeyFile: "test-fixtures/fake_service_account_key.json",
				},
			},
			wantErr: true,
		},
		{
			name: "service account key file",
			args: args{
				d: &Driver{
					Token: "",
					ServiceAccountKeyFile: "test-fixtures/fake_service_account_key.json",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewYandexCloudClient(tt.args.d)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewYandexCloudClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestYandexCloudClient_getInstanceIPAddress(t *testing.T) {
	type args struct {
		d        *Driver
		instance *compute.Instance
	}
	tests := []struct {
		name        string
		args        args
		wantAddress string
		wantErr     bool
	}{
		{
			name: "instance with both addresses, want internal address",
			args: args{
				d: &Driver{
					UseIPv6:       false,
					UseInternalIP: true,
				},
				instance: &compute.Instance{
					NetworkInterfaces: []*compute.NetworkInterface{
						{
							Index: "1",
							PrimaryV4Address: &compute.PrimaryAddress{
								Address: "192.168.19.86",
								OneToOneNat: &compute.OneToOneNat{
									Address:   "92.68.12.34",
									IpVersion: compute.IpVersion_IPV4,
								},
							},
							SubnetId:   "some-subnet-id",
							MacAddress: "aa-bb-cc-dd-ee-ff",
						},
					},
				},
			},
			wantAddress: "192.168.19.86",
			wantErr:     false,
		},
		{
			name: "instance with both addresses, want external address",
			args: args{
				d: &Driver{
					UseIPv6:       false,
					UseInternalIP: false,
				},
				instance: &compute.Instance{
					NetworkInterfaces: []*compute.NetworkInterface{
						{
							Index: "1",
							PrimaryV4Address: &compute.PrimaryAddress{
								Address: "192.168.19.86",
								OneToOneNat: &compute.OneToOneNat{
									Address:   "92.68.12.34",
									IpVersion: compute.IpVersion_IPV4,
								},
							},
							SubnetId:   "some-subnet-id",
							MacAddress: "aa-bb-cc-dd-ee-ff",
						},
					},
				},
			},
			wantAddress: "92.68.12.34",
			wantErr:     false,
		},
		{
			name: "instance with internal address, want external address",
			args: args{
				d: &Driver{
					UseIPv6:       false,
					UseInternalIP: false,
				},
				instance: &compute.Instance{
					NetworkInterfaces: []*compute.NetworkInterface{
						{
							Index: "1",
							PrimaryV4Address: &compute.PrimaryAddress{
								Address: "192.168.19.86",
							},
							SubnetId:   "some-subnet-id",
							MacAddress: "aa-bb-cc-dd-ee-ff",
						},
					},
				},
			},
			wantAddress: "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &YandexCloudClient{}
			gotAddress, err := c.getInstanceIPAddress(tt.args.d, tt.args.instance)
			if (err != nil) != tt.wantErr {
				t.Errorf("YandexCloudClient.getInstanceIPAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotAddress != tt.wantAddress {
				t.Errorf("YandexCloudClient.getInstanceIPAddress() = %v, want %v", gotAddress, tt.wantAddress)
			}
		})
	}
}

func TestYandexCloudClient_instanceAddresses(t *testing.T) {
	type args struct {
		instance *compute.Instance
	}
	tests := []struct {
		name        string
		args        args
		wantIpV4Int string
		wantIpV4Ext string
		wantIpV6    string
		wantErr     bool
	}{
		{
			name: "no nics defined",
			args: args{
				instance: &compute.Instance{
					NetworkInterfaces: []*compute.NetworkInterface{},
				},
			},
			wantIpV4Int: "",
			wantIpV4Ext: "",
			wantIpV6:    "",
			wantErr:     true,
		},
		{
			name: "one nic with internal address",
			args: args{
				instance: &compute.Instance{
					NetworkInterfaces: []*compute.NetworkInterface{
						{
							Index: "1",
							PrimaryV4Address: &compute.PrimaryAddress{
								Address: "192.168.19.16",
							},
							SubnetId:   "some-subnet-id",
							MacAddress: "aa-bb-cc-dd-ee-ff",
						},
					},
				},
			},
			wantIpV4Int: "192.168.19.16",
			wantIpV4Ext: "",
			wantIpV6:    "",
			wantErr:     false,
		},
		{
			name: "one nic with internal and external address",
			args: args{
				instance: &compute.Instance{
					NetworkInterfaces: []*compute.NetworkInterface{
						{
							Index: "1",
							PrimaryV4Address: &compute.PrimaryAddress{
								Address: "192.168.19.86",
								OneToOneNat: &compute.OneToOneNat{
									Address:   "92.68.12.34",
									IpVersion: compute.IpVersion_IPV4,
								},
							},
							SubnetId:   "some-subnet-id",
							MacAddress: "aa-bb-cc-dd-ee-ff",
						},
					},
				},
			},
			wantIpV4Int: "192.168.19.86",
			wantIpV4Ext: "92.68.12.34",
			wantIpV6:    "",
			wantErr:     false,
		},
		{
			name: "one nic with ipv6 address",
			args: args{
				instance: &compute.Instance{
					NetworkInterfaces: []*compute.NetworkInterface{
						{
							Index: "1",
							PrimaryV6Address: &compute.PrimaryAddress{
								Address: "2001:db8::370:7348",
							},
							SubnetId:   "some-subnet-id",
							MacAddress: "aa-bb-cc-dd-ee-ff",
						},
					},
				},
			},
			wantIpV4Int: "",
			wantIpV4Ext: "",
			wantIpV6:    "2001:db8::370:7348",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &YandexCloudClient{}
			gotIpV4Int, gotIpV4Ext, gotIpV6, err := c.instanceAddresses(tt.args.instance)
			if (err != nil) != tt.wantErr {
				t.Errorf("YandexCloudClient.instanceAddresses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotIpV4Int != tt.wantIpV4Int {
				t.Errorf("YandexCloudClient.instanceAddresses() gotIpV4Int = %v, want %v", gotIpV4Int, tt.wantIpV4Int)
			}
			if gotIpV4Ext != tt.wantIpV4Ext {
				t.Errorf("YandexCloudClient.instanceAddresses() gotIpV4Ext = %v, want %v", gotIpV4Ext, tt.wantIpV4Ext)
			}
			if gotIpV6 != tt.wantIpV6 {
				t.Errorf("YandexCloudClient.instanceAddresses() gotIpV6 = %v, want %v", gotIpV6, tt.wantIpV6)
			}
		})
	}
}
