package virtualbox

import (
	"bufio"
	"io"
	"strings"
)

type VirtualDisk struct {
	UUID string
	Path string
}

func parseDiskInfo(r io.Reader) (*VirtualDisk, error) {
	s := bufio.NewScanner(r)
	disk := &VirtualDisk{}
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
		case "SATA-1-0":
			disk.Path = val
		case "SATA-ImageUUID-1-0":
			disk.UUID = val
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return disk, nil
}

func getVMDiskInfo(name string) (*VirtualDisk, error) {
	out, err := vbmOut("showvminfo", name, "--machinereadable")
	if err != nil {
		return nil, err
	}
	r := strings.NewReader(out)
	return parseDiskInfo(r)
}
