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
	"path/filepath"
	"strings"
	"time"

	"github.com/vmware/govmomi/vim25/debug"
)

type DebugFlag struct {
	enable bool
}

func (flag *DebugFlag) Register(f *flag.FlagSet) {
	env := "GOVC_DEBUG"
	enable := false
	switch env := strings.ToLower(os.Getenv(env)); env {
	case "1", "true":
		enable = true
	}

	usage := fmt.Sprintf("Store debug logs [%s]", env)
	f.BoolVar(&flag.enable, "debug", enable, usage)
}

func (flag *DebugFlag) Process() error {
	if flag.enable {
		// Base path for storing debug logs.
		r := os.Getenv("GOVC_DEBUG_PATH")
		if r == "" {
			r = filepath.Join(os.Getenv("HOME"), ".govmomi")
		}
		r = filepath.Join(r, "debug")

		// Path for this particular run.
		now := time.Now().Format("2006-01-02T15-04-05.999999999")
		r = filepath.Join(r, now)

		err := os.MkdirAll(r, 0700)
		if err != nil {
			return err
		}

		p := debug.FileProvider{
			Path: r,
		}

		debug.SetProvider(&p)
	}

	return nil
}
