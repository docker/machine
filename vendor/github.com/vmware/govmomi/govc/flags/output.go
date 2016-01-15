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
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/vmware/govmomi/vim25/progress"
)

type OutputWriter interface {
	Write(io.Writer) error
}

type OutputFlag struct {
	JSON bool
	TTY  bool
}

func (flag *OutputFlag) Register(f *flag.FlagSet) {
	f.BoolVar(&flag.JSON, "json", false, "Enable JSON output")
}

func (flag *OutputFlag) Process() error {
	if !flag.JSON {
		// Assume we have a tty if not outputting JSON
		flag.TTY = true
	}

	return nil
}

// Log outputs the specified string, prefixed with the current time.
// A newline is not automatically added. If the specified string
// starts with a '\r', the current line is cleared first.
func (flag *OutputFlag) Log(s string) (int, error) {
	if len(s) > 0 && s[0] == '\r' {
		flag.Write([]byte{'\r', 033, '[', 'K'})
		s = s[1:]
	}

	return flag.WriteString(time.Now().Format("[02-01-06 15:04:05] ") + s)
}

func (flag *OutputFlag) Write(b []byte) (int, error) {
	if !flag.TTY {
		return 0, nil
	}

	n, err := os.Stdout.Write(b)
	os.Stdout.Sync()
	return n, err
}

func (flag *OutputFlag) WriteString(s string) (int, error) {
	return flag.Write([]byte(s))
}

func (flag *OutputFlag) WriteResult(result OutputWriter) error {
	var err error
	var out = os.Stdout

	if flag.JSON {
		enc := json.NewEncoder(out)
		err = enc.Encode(result)
	} else {
		err = result.Write(out)
	}

	return err
}

type progressLogger struct {
	flag   *OutputFlag
	prefix string

	wg sync.WaitGroup

	sink chan chan progress.Report
	done chan struct{}
}

func newProgressLogger(flag *OutputFlag, prefix string) *progressLogger {
	p := &progressLogger{
		flag:   flag,
		prefix: prefix,

		sink: make(chan chan progress.Report),
		done: make(chan struct{}),
	}

	p.wg.Add(1)

	go p.loopA()

	return p
}

// loopA runs before Sink() has been called.
func (p *progressLogger) loopA() {
	var err error

	defer p.wg.Done()

	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()

	for stop := false; !stop; {
		select {
		case ch := <-p.sink:
			err = p.loopB(tick, ch)
			stop = true
		case <-p.done:
			stop = true
		case <-tick.C:
			line := fmt.Sprintf("\r%s", p.prefix)
			p.flag.Log(line)
		}
	}

	if err != nil && err != io.EOF {
		p.flag.Log(fmt.Sprintf("\r%sError: %s\n", p.prefix, err))
	} else {
		p.flag.Log(fmt.Sprintf("\r%sOK\n", p.prefix))
	}
}

// loopA runs after Sink() has been called.
func (p *progressLogger) loopB(tick *time.Ticker, ch <-chan progress.Report) error {
	var r progress.Report
	var ok bool
	var err error

	for ok = true; ok; {
		select {
		case r, ok = <-ch:
			if !ok {
				break
			}
			err = r.Error()
		case <-tick.C:
			line := fmt.Sprintf("\r%s", p.prefix)
			if r != nil {
				line += fmt.Sprintf("(%.0f%%", r.Percentage())
				detail := r.Detail()
				if detail != "" {
					line += fmt.Sprintf(", %s", detail)
				}
				line += ")"
			}
			p.flag.Log(line)
		}
	}

	return err
}

func (p *progressLogger) Sink() chan<- progress.Report {
	ch := make(chan progress.Report)
	p.sink <- ch
	return ch
}

func (p *progressLogger) Wait() {
	close(p.done)
	p.wg.Wait()
}

func (flag *OutputFlag) ProgressLogger(prefix string) *progressLogger {
	return newProgressLogger(flag, prefix)
}
