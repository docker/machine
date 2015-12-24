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

package host

import (
	"flag"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"golang.org/x/net/context"
)

type info struct {
	*flags.ClientFlag
	*flags.OutputFlag
	*flags.HostSystemFlag
}

func init() {
	cli.Register("host.info", &info{})
}

func (c *info) Register(f *flag.FlagSet) {}

func (c *info) Process() error { return nil }

func (c *info) Run(f *flag.FlagSet) error {
	client, err := c.Client()
	if err != nil {
		return err
	}

	var hosts []*object.HostSystem

	// We could do without the -host flag, leaving it for compat
	host, err := c.HostSystemIfSpecified()
	if err != nil {
		return err
	}

	// Default only if there is a single host
	if host == nil && f.NArg() == 0 {
		host, err = c.HostSystem()
		if err != nil {
			return err
		}
	}

	if host != nil {
		hosts = append(hosts, host)
	} else {
		hosts, err = c.HostSystems(f.Args())
		if err != nil {
			return err
		}
	}

	var res infoResult
	var props []string

	if c.OutputFlag.JSON {
		props = nil // Load everything
	} else {
		props = []string{"summary"} // Load summary
	}

	for _, host := range hosts {
		var h mo.HostSystem

		pc := property.DefaultCollector(client)
		err = pc.RetrieveOne(context.TODO(), host.Reference(), props, &h)
		if err != nil {
			return err
		}

		res.HostSystems = append(res.HostSystems, h)
	}

	return c.WriteResult(&res)
}

type infoResult struct {
	HostSystems []mo.HostSystem
}

func (r *infoResult) Write(w io.Writer) error {
	tw := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)

	for _, host := range r.HostSystems {
		s := host.Summary
		h := s.Hardware
		z := s.QuickStats
		ncpu := int(h.NumCpuPkgs * h.NumCpuCores)
		cpuUsage := 100 * float64(z.OverallCpuUsage) / float64(ncpu*h.CpuMhz)
		memUsage := 100 * float64(z.OverallMemoryUsage<<20) / float64(h.MemorySize)

		fmt.Fprintf(tw, "Name:\t%s\n", s.Config.Name)
		fmt.Fprintf(tw, "  Manufacturer:\t%s\n", h.Vendor)
		fmt.Fprintf(tw, "  Logical CPUs:\t%d CPUs @ %dMHz\n", ncpu, h.CpuMhz)
		fmt.Fprintf(tw, "  Processor type:\t%s\n", h.CpuModel)
		fmt.Fprintf(tw, "  CPU usage:\t%d MHz (%.1f%%)\n", z.OverallCpuUsage, cpuUsage)
		fmt.Fprintf(tw, "  Memory:\t%dMB\n", h.MemorySize/(1024*1024))
		fmt.Fprintf(tw, "  Memory usage:\t%d MB (%.1f%%)\n", z.OverallMemoryUsage, memUsage)
		fmt.Fprintf(tw, "  Boot time:\t%s\n", s.Runtime.BootTime)
	}

	return tw.Flush()
}
