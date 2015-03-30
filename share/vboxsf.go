package share

import (
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/machine/drivers"
	vbox "github.com/docker/machine/drivers/virtualbox"
)

type VBoxSharedFolder struct {
	Options ShareOptions
}

func (sf VBoxSharedFolder) ContractFulfilled(d drivers.Driver) (bool, error) {
	if d.DriverName() != "virtualbox" {
		return false, nil
	}
	cmd, err := drivers.GetSSHCommandFromDriver(d, "lsmod | grep -i vbox")
	if err != nil {
		return false, err
	}
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (sf VBoxSharedFolder) Create(d drivers.Driver) error {
	log.Info("Stopping machine to create the share...")

	if err := d.Stop(); err != nil {
		return fmt.Errorf("Error stopping the VM: %s")
	}

	// let VBoxService do nice magic automounting (when it's used)
	if err := vbox.Vbm("guestproperty", "set", d.GetMachineName(), "/VirtualBox/GuestAdd/SharedFolders/MountPrefix", "/"); err != nil {
		return err
	}
	if err := vbox.Vbm("guestproperty", "set", d.GetMachineName(), "/VirtualBox/GuestAdd/SharedFolders/MountDir", "/"); err != nil {
		return err
	}

	if sf.Options.SrcPath != "" {
		if _, err := os.Stat(sf.Options.SrcPath); err != nil && !os.IsNotExist(err) {
			return err
		} else if !os.IsNotExist(err) {
			if sf.Options.Name == "" {
				// parts of the VBox internal code are buggy with share names that start with "/"
				sf.Options.Name = strings.TrimLeft(sf.Options.SrcPath, "/")
				// TODO do some basic Windows -> MSYS path conversion
				// ie, s!^([a-z]+):[/\\]+!\1/!; s!\\!/!g
			}

			// woo, sf.Options.SrcPath exists!  let's carry on!
			if err := vbox.Vbm("sharedfolder", "add", d.GetMachineName(), "--name", sf.Options.Name, "--hostpath", sf.Options.SrcPath, "--automount"); err != nil {
				return err
			}

			// enable symlinks
			if err := vbox.Vbm("setextradata", d.GetMachineName(), "VBoxInternal2/SharedFoldersEnableSymlinksCreate/"+sf.Options.Name, "1"); err != nil {
				return err
			}
		}
	}

	if err := d.Start(); err != nil {
		return fmt.Errorf("Error starting the VM: %s")
	}

	return nil
}

func (sf VBoxSharedFolder) Mount(d drivers.Driver) error {
	cmdFmtString := "sudo mkdir -p %s && sudo mount -t vboxsf -o uid=%d,gid=%d %s %s"
	mountCmd := fmt.Sprintf(cmdFmtString, sf.Options.SrcPath, sf.Options.DestUid, sf.Options.DestGid, sf.Options.Name, sf.Options.SrcPath)
	cmd, err := drivers.GetSSHCommandFromDriver(d, mountCmd)
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (sf VBoxSharedFolder) Destroy(d drivers.Driver) error {
	// lol how do I do this
	return nil
}

func (sf VBoxSharedFolder) GetOptions() ShareOptions {
	return sf.Options
}
