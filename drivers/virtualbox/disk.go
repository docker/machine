package virtualbox

import (
	"bufio"
	"strings"
)

type VirtualDisk struct {
	UUID string
	Path string
}

func getVMDiskInfo(name string, vbox VBoxManager) (*VirtualDisk, error) {
	out, err := vbox.vbmOut("showvminfo", name, "--machinereadable")
	if err != nil {
		return nil, err
	}

	return parseDiskInfo(out)
}

func parseDiskInfo(out string) (*VirtualDisk, error) {
	disk := &VirtualDisk{}

	r := strings.NewReader(out)
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := s.Text()
		if line == "" {
			continue
		}

		res := reEqualQuoteLine.FindStringSubmatch(line)
		if res == nil {
			continue
		}

		key, val := res[1], res[2]
		switch key {
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
