package host

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/docker/machine/drivers/driverfactory"
	"github.com/docker/machine/libmachine/version"
)

func getMigratedHostMetadata(data []byte) (*HostMetadata, error) {
	// HostMetadata is for a "first pass" so we can then load the driver
	var (
		hostMetadata *HostMetadataV0
	)

	if err := json.Unmarshal(data, &hostMetadata); err != nil {
		return &HostMetadata{}, err
	}

	migratedHostMetadata := MigrateHostMetadataV0ToHostMetadataV1(hostMetadata)

	return migratedHostMetadata, nil
}

func MigrateHost(h *Host, data []byte) (*Host, bool, error) {
	var (
		migrationPerformed = false
		hostV1             *HostV1
	)

	migratedHostMetadata, err := getMigratedHostMetadata(data)
	if err != nil {
		return &Host{}, false, err
	}

	globalStorePath := filepath.Dir(filepath.Dir(migratedHostMetadata.HostOptions.AuthOptions.StorePath))

	driver, err := driverfactory.NewDriver(migratedHostMetadata.DriverName, h.Name, globalStorePath)
	if err != nil {
		return &Host{}, false, err
	}

	if migratedHostMetadata.ConfigVersion == version.ConfigVersion {
		h.Driver = driver
		if err := json.Unmarshal(data, &h); err != nil {
			return &Host{}, migrationPerformed, fmt.Errorf("Error unmarshalling most recent host version: %s", err)
		}
	} else {
		migrationPerformed = true
		for h.ConfigVersion = migratedHostMetadata.ConfigVersion; h.ConfigVersion < version.ConfigVersion; h.ConfigVersion++ {
			switch h.ConfigVersion {
			case 0:
				hostV0 := &HostV0{
					Driver: driver,
				}
				if err := json.Unmarshal(data, &hostV0); err != nil {
					return &Host{}, migrationPerformed, fmt.Errorf("Error unmarshalling host config version 0: %s", err)
				}
				hostV1 = MigrateHostV0ToHostV1(hostV0)
			case 1:
				if hostV1 == nil {
					hostV1 = &HostV1{
						Driver: driver,
					}
					if err := json.Unmarshal(data, &hostV1); err != nil {
						return &Host{}, migrationPerformed, fmt.Errorf("Error unmarshalling host config version 1: %s", err)
					}
				}
				h = MigrateHostV1ToHostV2(hostV1)
			}
		}

	}

	return h, migrationPerformed, nil
}
