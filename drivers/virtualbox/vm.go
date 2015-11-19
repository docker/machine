package virtualbox

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

type VM struct {
	CPUs   int
	Memory int
}

func getVMInfo(name string, vbox VBoxManager) (*VM, error) {
	out, err := vbox.vbmOut("showvminfo", name, "--machinereadable")
	if err != nil {
		return nil, err
	}

	r := strings.NewReader(out)
	return parseVMInfo(r)
}

func parseVMInfo(r io.Reader) (*VM, error) {
	s := bufio.NewScanner(r)
	vm := &VM{}
	for s.Scan() {
		line := s.Text()
		if line == "" {
			continue
		}
		res := reEqualLine.FindStringSubmatch(line)
		if res == nil {
			continue
		}
		switch key, val := res[1], res[2]; key {
		case "cpus":
			v, err := strconv.Atoi(val)
			if err != nil {
				return nil, err
			}
			vm.CPUs = v
		case "memory":
			v, err := strconv.Atoi(val)
			if err != nil {
				return nil, err
			}
			vm.Memory = v
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return vm, nil
}
