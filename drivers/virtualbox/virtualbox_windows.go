package virtualbox

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/docker/machine/libmachine/log"
	"golang.org/x/sys/windows/registry"
)

const PF_VIRT_FIRMWARE_ENABLED = 21

// IsVTXDisabled checks if VT-X is disabled in the BIOS. If it is, the vm will fail to start.
// If we can't be sure it is disabled, we carry on and will check the vm logs after it's started.
func (d *Driver) IsVTXDisabled() bool {
	mod := syscall.NewLazyDLL("kernel32.dll")
	proc := mod.NewProc("IsProcessorFeaturePresent")

	ret, _, err := proc.Call(uintptr(PF_VIRT_FIRMWARE_ENABLED))
	if err != nil {
		log.Debugf("%s: %v", ErrMsgUnableToCheckVTX, err)
		return false
	}

	return ret == 0
}

func detectVBoxManageCmd() string {
	cmd := "VBoxManage"
	if p := os.Getenv("VBOX_INSTALL_PATH"); p != "" {
		if path, err := exec.LookPath(filepath.Join(p, cmd)); err == nil {
			return path
		}
	}

	if p := os.Getenv("VBOX_MSI_INSTALL_PATH"); p != "" {
		if path, err := exec.LookPath(filepath.Join(p, cmd)); err == nil {
			return path
		}
	}

	// Look in default installation path for VirtualBox version > 5
	if path, err := exec.LookPath(filepath.Join("C:\\Program Files\\Oracle\\VirtualBox", cmd)); err == nil {
		return path
	}

	// Look in windows registry
	if p, err := findVBoxInstallDirInRegistry(); err == nil {
		if path, err := exec.LookPath(filepath.Join(p, cmd)); err == nil {
			return path
		}
	}

	return detectVBoxManageCmdInPath() //fallback to path
}

func findVBoxInstallDirInRegistry() (string, error) {
	registryKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Oracle\VirtualBox`, registry.QUERY_VALUE)
	if err != nil {
		errorMessage := fmt.Sprintf("Can't find VirtualBox registry entries, is VirtualBox really installed properly? %s", err)
		log.Debugf(errorMessage)
		return "", fmt.Errorf(errorMessage)
	}

	defer registryKey.Close()

	installDir, _, err := registryKey.GetStringValue("InstallDir")
	if err != nil {
		errorMessage := fmt.Sprintf("Can't find InstallDir registry key within VirtualBox registries entries, is VirtualBox really installed properly? %s", err)
		log.Debugf(errorMessage)
		return "", fmt.Errorf(errorMessage)
	}

	return installDir, nil
}

func getShareDriveAndName() (string, string) {
	return "c/Users", "c:\\Users"
}

func isHyperVInstalled() bool {
	_, err := exec.LookPath("vmms.exe")
	return err == nil
}
