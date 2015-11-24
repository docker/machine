package virtualbox

type VirtualDisk struct {
	UUID string
	Path string
}

func getVMDiskInfo(name string, vbox VBoxManager) (*VirtualDisk, error) {
	out, err := vbox.vbmOut("showvminfo", name, "--machinereadable")
	if err != nil {
		return nil, err
	}

	disk := &VirtualDisk{}

	err = parseKeyValues(out, reEqualQuoteLine, func(key, val string) error {
		switch key {
		case "SATA-1-0":
			disk.Path = val
		case "SATA-ImageUUID-1-0":
			disk.UUID = val
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return disk, nil
}
