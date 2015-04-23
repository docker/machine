package virtualbox

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

type VirtualBoxVM struct {
	CPUs   int
	Memory int
}

func parseVMInfo(r io.Reader) (*VirtualBoxVM, error) {
	s := bufio.NewScanner(r)
	vm := &VirtualBoxVM{}
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

func getVMInfo(name string) (*VirtualBoxVM, error) {
	out, err := vbmOut("showvminfo", name, "--machinereadable")
	if err != nil {
		return nil, err
	}
	r := strings.NewReader(out)
	return parseVMInfo(r)
}
