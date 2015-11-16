package host

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/docker/machine/drivers/none"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/version"
)

type RawDataDriver struct {
	*none.Driver
	data []byte // passed directly back when invoking json.Marshal on this type
}

func (r *RawDataDriver) MarshalJSON() ([]byte, error) {
	// now marshal it back up
	return r.data, nil
}

func getMigratedHostMetadata(data []byte) (*Metadata, error) {
	// HostMetadata is for a "first pass" so we can then load the driver
	var (
		hostMetadata *MetadataV0
	)

	if err := json.Unmarshal(data, &hostMetadata); err != nil {
		return &Metadata{}, err
	}

	migratedHostMetadata := MigrateHostMetadataV0ToHostMetadataV1(hostMetadata)

	return migratedHostMetadata, nil
}

func MigrateHost(h *Host, data []byte) (*Host, bool, error) {
	var (
		migrationNeeded    = false
		migrationPerformed = false
		hostV1             *V1
		hostV2             *V2
	)

	migratedHostMetadata, err := getMigratedHostMetadata(data)
	if err != nil {
		return nil, false, err
	}

	globalStorePath := filepath.Dir(filepath.Dir(migratedHostMetadata.HostOptions.AuthOptions.StorePath))

	driver := none.NewDriver(h.Name, globalStorePath)

	if migratedHostMetadata.ConfigVersion > version.ConfigVersion {
		return nil, false, errors.New("Config version is from the future, please upgrade your Docker Machine client.")
	}

	if migratedHostMetadata.ConfigVersion == version.ConfigVersion {
		h.Driver = driver
		if err := json.Unmarshal(data, &h); err != nil {
			return nil, migrationPerformed, fmt.Errorf("Error unmarshalling most recent host version: %s", err)
		}

		// We are config version 3, so we definitely should have a
		// RawDriver field.  However, it's possible some might use
		// older clients after already migrating, so check if it exists
		// and create one if not.  The following code is an (admittedly
		// fragile) attempt to account for the fact that the above code
		// to forbid loading from future clients was not introduced
		// sooner.
		if h.RawDriver == nil {
			log.Warn("It looks like you have used an older Docker Machine binary to interact with hosts after using a 0.5.0 binary.")
			log.Warn("Please be advised that doing so can result in erratic behavior due to migrated configuration settings.")
			log.Warn("Machine will attempt to re-migrate the configuration settings, but safety is not guaranteed.")
			migrationNeeded = true

			// Treat the data as config version 1, even though it
			// says "latest".
			migratedHostMetadata.ConfigVersion = 1
		}
	} else {
		migrationNeeded = true
	}

	if migrationNeeded {
		migrationPerformed = true
		for h.ConfigVersion = migratedHostMetadata.ConfigVersion; h.ConfigVersion < version.ConfigVersion; h.ConfigVersion++ {
			log.Debugf("Migrating to config v%d", h.ConfigVersion)
			switch h.ConfigVersion {
			case 0:
				hostV0 := &V0{
					Driver: driver,
				}
				if err := json.Unmarshal(data, &hostV0); err != nil {
					return nil, migrationPerformed, fmt.Errorf("Error unmarshalling host config version 0: %s", err)
				}
				hostV1 = MigrateHostV0ToHostV1(hostV0)
			case 1:
				if hostV1 == nil {
					hostV1 = &V1{
						Driver: driver,
					}
					if err := json.Unmarshal(data, &hostV1); err != nil {
						return nil, migrationPerformed, fmt.Errorf("Error unmarshalling host config version 1: %s", err)
					}
				}
				hostV2 = MigrateHostV1ToHostV2(hostV1)
			case 2:
				if hostV2 == nil {
					hostV2 = &V2{
						Driver: driver,
					}
					if err := json.Unmarshal(data, &hostV2); err != nil {
						return nil, migrationPerformed, fmt.Errorf("Error unmarshalling host config version 2: %s", err)
					}
				}
				h = MigrateHostV2ToHostV3(hostV2, data, globalStorePath)
				h.Driver = RawDataDriver{driver, nil}
			case 3:
			}
		}
	}

	return h, migrationPerformed, nil
}
