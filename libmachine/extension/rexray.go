package extension

import (
	"fmt"
	"strings"

	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/log"
)

var (
	rexName    = "rexray"
	rexVersion = "stable"
)

func init() {
	RegisterExtension(rexName, &RegisteredExtension{
		New: NewRexrayExtension,
	})
}

func NewRexrayExtension() Extension {
	return &RexrayExtension{
		GenericExtension{
			extensionName: rexName,
			version:       rexVersion,
		},
	}
}

type RexrayExtension struct {
	GenericExtension
}

func (extension *RexrayExtension) Install(provisioner provision.Provisioner, hostInfo *ExtensionParams, extInfo *ExtensionInfo) error {
	if extInfo.version != rexVersion {
		rexVersion = extInfo.version
	}

	switch hostInfo.OsID {
	case "ubuntu", "debian", "centos", "redhat":
		log.Debugf("%s: found supported OS: %s", strings.ToUpper(extInfo.name), hostInfo.OsID)
	default:
		return fmt.Errorf("%s not supported on: %s", strings.ToUpper(extInfo.name), hostInfo.OsID)
	}

	if extInfo.envs != nil {
		appendEnvFile(provisioner, extInfo)
	}

	log.Debugf("%s: installing version: %s", strings.ToUpper(extInfo.name), rexVersion)
	switch rexVersion {
	case "stable", "latest":
		if _, err := provisioner.SSHCommand("curl -sSL https://dl.bintray.com/emccode/rexray/install | sh "); err != nil {
			return err
		}
	case "stupid", "unstable", "staged":
		uNameS, err := provisioner.SSHCommand("uname -s")
		if err != nil {
			return err
		}
		uNameM, err := provisioner.SSHCommand("uname -m")
		if err != nil {
			return err
		}
		log.Debugf("%s: downloading version %s", strings.ToUpper(extInfo.name), rexVersion)
		if _, err := provisioner.SSHCommand(fmt.Sprintf("curl -L 'https://dl.bintray.com/emccode/rexray/%s/rexray-%s-%s.tar.gz' -o 'rexray-%s-%s.tar.gz'", rexVersion, strings.TrimSpace(uNameS), strings.TrimSpace(uNameM), strings.TrimSpace(uNameS), strings.TrimSpace(uNameM))); err != nil {
			return err
		}
		log.Debugf("%s: extracting version %s", strings.ToUpper(extInfo.name), rexVersion)
		if _, err := provisioner.SSHCommand(fmt.Sprintf("tar xzf rexray-%s-%s.tar.gz", strings.TrimSpace(uNameS), strings.TrimSpace(uNameM))); err != nil {
			return err
		}
		log.Debugf("%s: moving binary to /bin", strings.ToUpper(extInfo.name))
		if _, err := provisioner.SSHCommand("sudo mv rexray /bin"); err != nil {
			return err
		}
		log.Debugf("%s: installing service", strings.ToUpper(extInfo.name))
		if strings.TrimSpace(uNameS) != "Darwin" {
			if _, err := provisioner.SSHCommand("sudo /bin/rexray service install"); err != nil {
				return err
			}
		}
		log.Debugf("%s: running version command", strings.ToUpper(extInfo.name))
		if strings.TrimSpace(uNameS) != "Darwin" {
			if _, err := provisioner.SSHCommand("sudo rexray version"); err != nil {
				return err
			}
		}

	}

	if extInfo.files != nil {
		fileTransfer(provisioner, hostInfo, extInfo, extInfo.files, "config.yaml")
	}

	log.Debugf("%s: starting service", strings.ToUpper(extInfo.name))
	provisioner.SSHCommand("sudo service rexray start")

	return nil
}
