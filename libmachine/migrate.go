package libmachine

import (
	"encoding/json"
	"fmt"

	"github.com/docker/machine/drivers"
	"github.com/docker/machine/version"
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

func MigrateHost(h *Host, data []byte) (*Host, error) {
	migratedHostMetadata, err := getMigratedHostMetadata(data)
	if err != nil {
		return &Host{}, err
	}

	authOptions := migratedHostMetadata.HostOptions.AuthOptions

	driver, err := drivers.NewDriver(
		migratedHostMetadata.DriverName,
		h.Name,
		h.StorePath,
		authOptions.CaCertPath,
		authOptions.PrivateKeyPath,
	)
	if err != nil {
		return &Host{}, err
	}

	for h.ConfigVersion = migratedHostMetadata.ConfigVersion; h.ConfigVersion < version.ConfigVersion; h.ConfigVersion++ {
		switch h.ConfigVersion {
		case 0:
			hostV0 := &HostV0{
				Driver: driver,
			}
			if err := json.Unmarshal(data, &hostV0); err != nil {
				return &Host{}, fmt.Errorf("Error unmarshalling host config version 0: %s", err)
			}
			h = MigrateHostV0ToHostV1(hostV0)
		default:
		}
	}

	h.Driver = driver
	if err := json.Unmarshal(data, &h); err != nil {
		return &Host{}, fmt.Errorf("Error unmarshalling most recent host version: %s", err)
	}

	return h, nil
}
