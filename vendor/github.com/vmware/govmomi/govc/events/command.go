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

package events

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/vmware/govmomi/event"
	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/list"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/net/context"
)

type events struct {
	*flags.DatacenterFlag

	Max int
}

func init() {
	cli.Register("events", &events{})
}

func (cmd *events) Register(f *flag.FlagSet) {
	f.IntVar(&cmd.Max, "n", 25, "Output the last N events")
}

func (cmd *events) Process() error { return nil }

func (cmd *events) Usage() string {
	return "[PATH]..."
}

func (cmd *events) Run(f *flag.FlagSet) error {
	ctx := context.TODO()

	finder, err := cmd.Finder()
	if err != nil {
		return err
	}

	c, err := cmd.Client()
	if err != nil {
		return err
	}

	m := event.NewManager(c)

	var objs []list.Element

	args := f.Args()
	if len(args) == 0 {
		args = []string{"."}
	}

	for _, arg := range args {
		es, err := finder.ManagedObjectList(ctx, arg)
		if err != nil {
			return err
		}

		objs = append(objs, es...)
	}

	var events []types.BaseEvent

	for _, o := range objs {
		filter := types.EventFilterSpec{
			Entity: &types.EventFilterSpecByEntity{
				Entity:    o.Object.Reference(),
				Recursion: types.EventFilterSpecRecursionOptionAll,
			},
		}

		collector, err := m.CreateCollectorForEvents(ctx, filter)
		if err != nil {
			return fmt.Errorf("[%s] %s", o.Path, err)
		}
		defer collector.Destroy(ctx)

		err = collector.SetPageSize(ctx, cmd.Max)
		if err != nil {
			return err
		}

		page, err := collector.LatestPage(ctx)
		if err != nil {
			return err
		}

		events = append(events, page...)
	}

	event.Sort(events)

	tw := tabwriter.NewWriter(os.Stdout, 3, 0, 2, ' ', 0)

	for _, e := range events {
		cat, err := m.EventCategory(ctx, e)
		if err != nil {
			return err
		}

		event := e.GetEvent()
		msg := strings.TrimSpace(event.FullFormattedMessage)

		if t, ok := e.(*types.TaskEvent); ok {
			msg = fmt.Sprintf("%s (target=%s %s)", msg, t.Info.Entity.Type, t.Info.EntityName)
		}

		fmt.Fprintf(tw, "[%s]\t[%s]\t%s\n",
			event.CreatedTime.Local().Format(time.ANSIC),
			cat, msg)
	}

	return tw.Flush()
}
