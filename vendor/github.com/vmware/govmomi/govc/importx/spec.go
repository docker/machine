/*
Copyright (c) 2015 VMware, Inc. All Rights Reserved.

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

package importx

import (
	"encoding/json"
	"flag"
	"fmt"
	"path"
	"strings"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/ovf"
	"github.com/vmware/govmomi/vim25/types"
)

var (
	// all possible ovf property values
	// the first element being the default value

	allDeploymentOptions         = []string{"small", "medium", "large"}
	allDiskProvisioningOptions   = []string{"thin", "monolithicSparse", "monolithicFlat", "twoGbMaxExtentSparse", "twoGbMaxExtentFlat", "seSparse", "eagerZeroedThick", "thick", "sparse", "flat"}
	allIPAllocationPolicyOptions = []string{"dhcpPolicy", "transientPolicy", "fixedPolicy", "fixedAllocatedPolicy"}
	allIPProtocolOptions         = []string{"IPv4", "IPv6"}
)

type spec struct {
	*ovfx
}

func init() {
	cli.Register("import.spec", &spec{})
}

func (cmd *spec) Register(f *flag.FlagSet) {}

func (cmd *spec) Process() error { return nil }

func (cmd *spec) Usage() string {
	return "PATH_TO_OVF or PATH_TO_OVA"
}

func (cmd *spec) Run(f *flag.FlagSet) error {
	fpath := ""
	args := f.Args()
	if len(args) == 1 {
		fpath = f.Arg(0)
	}

	switch path.Ext(fpath) {
	case "":
	case ".ovf":
		cmd.Archive = &FileArchive{fpath}
	case ".ova":
		cmd.Archive = &TapeArchive{fpath}
		fpath = strings.TrimSuffix(path.Base(fpath), path.Ext(fpath)) + ".ovf"
	default:
		return fmt.Errorf("invalid file extension %s", path.Ext(fpath))
	}

	return cmd.Spec(fpath)
}

func (cmd *spec) Map(e *ovf.Envelope) []types.KeyValue {
	if e == nil {
		return nil
	}

	var p []types.KeyValue
	for _, v := range e.VirtualSystem.Product.Property {
		d := ""
		if v.Default != nil {
			d = *v.Default
		}

		p = append(p, types.KeyValue{
			Key:   v.Key,
			Value: d})
	}

	return p
}

func (cmd *spec) Spec(fpath string) error {
	e, err := cmd.ReadEnvelope(fpath)
	if err != nil {
		return err
	}

	var deploymentOptions = allDeploymentOptions
	if e != nil && e.DeploymentOption.Configuration != nil {
		deploymentOptions = nil
		// add default first
		for _, c := range e.DeploymentOption.Configuration {
			if c.Default != nil && *c.Default {
				deploymentOptions = append(deploymentOptions, c.ID)
			}
		}
		for _, c := range e.DeploymentOption.Configuration {
			if c.Default == nil || !*c.Default {
				deploymentOptions = append(deploymentOptions, c.ID)
			}
		}
	}

	o := Options{
		AllDeploymentOptions:         deploymentOptions,
		Deployment:                   deploymentOptions[0],
		AllDiskProvisioningOptions:   allDiskProvisioningOptions,
		DiskProvisioning:             allDiskProvisioningOptions[0],
		AllIPAllocationPolicyOptions: allIPAllocationPolicyOptions,
		IPAllocationPolicy:           allIPAllocationPolicyOptions[0],
		AllIPProtocolOptions:         allIPProtocolOptions,
		IPProtocol:                   allIPProtocolOptions[0],
		PowerOn:                      false,
		WaitForIP:                    false,
		InjectOvfEnv:                 false,
		PropertyMapping:              cmd.Map(e)}

	j, err := json.Marshal(&o)
	if err != nil {
		return err
	}
	fmt.Println(string(j))

	return nil
}
