package libmachine

import (
	"fmt"
	"path/filepath"

	"github.com/docker/machine/libmachine/cert"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/persist"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/state"
)

func GetDefaultStore() *persist.Filestore {
	homeDir := mcnutils.GetHomeDir()
	certsDir := filepath.Join(homeDir, ".docker", "machine", "certs")
	return &persist.Filestore{
		Path:             homeDir,
		CaCertPath:       certsDir,
		CaPrivateKeyPath: certsDir,
	}
}

// Create is the wrapper method which covers all of the boilerplate around
// actually creating, provisioning, and persisting an instance in the store.
func Create(store persist.Store, h *host.Host) error {
	if err := cert.BootstrapCertificates(h.HostOptions.AuthOptions); err != nil {
		return fmt.Errorf("Error generating certificates: %s", err)
	}

	log.Info("Running pre-create checks...")

	if err := h.Driver.PreCreateCheck(); err != nil {
		return fmt.Errorf("Error with pre-create check: %s", err)
	}

	if err := store.Save(h); err != nil {
		return fmt.Errorf("Error saving host to store before attempting creation: %s", err)
	}

	log.Info("Creating machine...")

	if err := h.Driver.Create(); err != nil {
		return fmt.Errorf("Error in driver during machine creation: %s", err)
	}

	if err := store.Save(h); err != nil {
		return fmt.Errorf("Error saving host to store after attempting creation: %s", err)
	}

	// TODO: Not really a fan of just checking "none" here.
	if h.Driver.DriverName() != "none" {
		log.Info("Waiting for machine to be running, this may take a few minutes...")
		if err := mcnutils.WaitFor(drivers.MachineInState(h.Driver, state.Running)); err != nil {
			return fmt.Errorf("Error waiting for machine to be running: %s", err)
		}

		log.Info("Machine is running, waiting for SSH to be available...")
		if err := host.WaitForSSH(h); err != nil {
			return fmt.Errorf("Error waiting for SSH: %s", err)
		}

		log.Info("Detecting operating system of created instance...")
		provisioner, err := provision.DetectProvisioner(h.Driver)
		if err != nil {
			return fmt.Errorf("Error detecting OS: %s", err)
		}

		log.Info("Provisioning created instance...")
		if err := provisioner.Provision(*h.HostOptions.SwarmOptions, *h.HostOptions.AuthOptions, *h.HostOptions.EngineOptions); err != nil {
			return fmt.Errorf("Error running provisioning: %s", err)
		}
	}

	log.Debug("Reticulating splines...")

	return nil
}

func SetDebug(val bool) {
	log.IsDebug = val
}
