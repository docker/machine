package share

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/machine/drivers"
	vbox "github.com/docker/machine/drivers/virtualbox"
)

type NfsSharedFolder struct {
	Options ShareOptions
}

type EtcExportsTemplateContext struct {
	Name      string
	Options   ShareOptions
	MachineIP string
}

func (ns NfsSharedFolder) ContractFulfilled(d drivers.Driver) (bool, error) {
	switch runtime.GOOS {
	case "windows":
		return false, errors.New("The NFS share driver is not supported on Windows")
	case "darwin":
		nfsdPath, err := exec.LookPath("nfsd")
		if err != nil || nfsdPath == "" {
			return false, fmt.Errorf("nfsd not found locally: %s", err)
		}
	case "linux":
		// TODO: lol this is probably bad
		cmd := exec.Command("sudo", "modprobe", "nfs")
		if err := cmd.Run(); err != nil {
			return false, fmt.Errorf("Seems that the NFS module is not installed locally: %s", err)
		}
	}

	// TODO
	/*
		provisioner, err := provision.DetectProvisioner(d)
		if err != nil {
			return false, err
		}

		if _, ok := provisioner.(provision.Boot2DockerProvisioner); !ok {
			return false, errors.New("NFS share driver only supported with local VMs using boot2docker")
		}
	*/

	return true, nil
}

func (ns NfsSharedFolder) Create(d drivers.Driver) error {
	var (
		buf            bytes.Buffer
		tmpl           *template.Template
		err            error
		nfsdRestartCmd *exec.Cmd
	)

	switch runtime.GOOS {
	case "darwin":
		tmpl, err = template.New("export").Parse(`
# docker-machine-begin-{{.Name}}-{{.Options.Name}}
"{{.Options.SrcPath}}" {{.MachineIP}} -alldirs -mapall=501:1000
# docker-machine-end-{{.Name}}-{{.Options.Name}}
`)

		if err != nil {
			return err
		}
		nfsdRestartCmd = exec.Command("sudo", "nfsd", "restart")
	case "linux":
		tmpl, err = template.New("export").Parse(`
# docker-machine-begin-{{.Name}}-{{.Options.Name}}
{{.Options.SrcPath}} {{.MachineIP}}(rw,no_root_squash,no_subtree_check)
# docker-machine-end-{{.Name}}-{{.Options.Name}}
`)
		if err != nil {
			return err
		}
		nfsdRestartCmd = exec.Command("sudo", "systemctl", "restart", "nfs-kernel-server")
	}

	machineIP, err := d.GetIP()
	if err != nil {
		return err
	}

	tmplContext := EtcExportsTemplateContext{
		Name:      d.GetMachineName(),
		Options:   ns.Options,
		MachineIP: machineIP,
	}

	if err := tmpl.Execute(&buf, tmplContext); err != nil {
		return err
	}

	appendCmd := exec.Command("sudo", "tee", "-a", "/etc/exports")
	appendCmd.Stdin = &buf

	if err := appendCmd.Run(); err != nil {
		return err
	}

	if err := nfsdRestartCmd.Run(); err != nil {
		return err
	}

	return nil
}

func (ns NfsSharedFolder) Mount(d drivers.Driver) error {
	hostonlyAdapterIP, err := vbox.GetHostOnlyNetworkIPv4ByMachineName(d.GetMachineName())
	if err != nil {
		return err
	}
	cmdFmtString := "sudo /usr/local/etc/init.d/nfs-client start && sudo mkdir -p %s && sudo mount -t nfs -o vers=3,nolock,udp %s:%s %s"
	mountCmd := fmt.Sprintf(cmdFmtString, ns.Options.SrcPath, hostonlyAdapterIP, ns.Options.SrcPath, ns.Options.DestPath)
	log.Debug(mountCmd)
	cmd, err := drivers.GetSSHCommandFromDriver(d, mountCmd)
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (ns NfsSharedFolder) Destroy(d drivers.Driver) error {
	// TODO
	return nil
}

func (ns NfsSharedFolder) GetOptions() ShareOptions {
	return ns.Options
}
