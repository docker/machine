/*
Copyright (c) 2014 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package flags

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/net/context"
)

type NetworkFlag struct {
	*DatacenterFlag

	register sync.Once
	name     string
	net      object.NetworkReference
	adapter  string
	address  string
}

func NewNetworkFlag() *NetworkFlag {
	return &NetworkFlag{}
}

func (flag *NetworkFlag) Register(f *flag.FlagSet) {
	flag.register.Do(func() {
		env := "GOVC_NETWORK"
		value := os.Getenv(env)
		flag.Set(value)
		usage := fmt.Sprintf("Network [%s]", env)
		f.Var(flag, "net", usage)
		f.StringVar(&flag.adapter, "net.adapter", "e1000", "Network adapter type")
		f.StringVar(&flag.address, "net.address", "", "Network hardware address")
	})
}

func (flag *NetworkFlag) Process() error { return nil }

func (flag *NetworkFlag) String() string {
	return flag.name
}

func (flag *NetworkFlag) Set(name string) error {
	flag.name = name
	return nil
}

func (flag *NetworkFlag) Network() (object.NetworkReference, error) {
	if flag.net != nil {
		return flag.net, nil
	}

	finder, err := flag.Finder()
	if err != nil {
		return nil, err
	}

	if flag.name == "" {
		flag.net, err = finder.DefaultNetwork(context.TODO())
	} else {
		flag.net, err = finder.Network(context.TODO(), flag.name)
	}

	return flag.net, err
}

func (flag *NetworkFlag) Device() (types.BaseVirtualDevice, error) {
	net, err := flag.Network()
	if err != nil {
		return nil, err
	}

	backing, err := net.EthernetCardBackingInfo(context.TODO())
	if err != nil {
		return nil, err
	}

	device, err := object.EthernetCardTypes().CreateEthernetCard(flag.adapter, backing)
	if err != nil {
		return nil, err
	}

	if flag.address != "" {
		card := device.(types.BaseVirtualEthernetCard).GetVirtualEthernetCard()
		card.AddressType = string(types.VirtualEthernetCardMacTypeManual)
		card.MacAddress = flag.address
	}

	return device, nil
}
