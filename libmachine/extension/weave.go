package extension

import (
	"fmt"
	"strings"

	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/log"
)

var (
	weaveName    = "weave"
	weaveVersion = "latest"
)

func init() {
	RegisterExtension(weaveName, &RegisteredExtension{
		New: NewWeaveExtension,
	})
}

// NewWeaveExtension returns the generic extension for weave
func NewWeaveExtension() Extension {
	return &WeaveExtension{
		GenericExtension{
			extensionName: weaveName,
			version:       weaveVersion,
		},
	}
}

// WeaveExtension is a generic struct
type WeaveExtension struct {
	GenericExtension
}

// Install will install run the Weave install workflow
func (extension *WeaveExtension) Install(provisioner provision.Provisioner, hostInfo *ExtensionParams, extInfo *ExtensionInfo) error {
	if extInfo.version != weaveVersion {
		weaveVersion = extInfo.version
	}

	log.Debugf("%s: downloading", strings.ToUpper(extInfo.name))
	if _, err := provisioner.SSHCommand("sudo curl -L git.io/weave -o /usr/local/bin/weave"); err != nil {
		return err
	}
	log.Debugf("%s: setting permissions", strings.ToUpper(extInfo.name))
	if _, err := provisioner.SSHCommand("sudo chmod a+x /usr/local/bin/weave"); err != nil {
		return err
	}

	switch hostInfo.OsID {
	case "ubuntu", "debian", "centos", "redhat", "coreos":
		log.Debugf("%s: found supported OS: %s", strings.ToUpper(extInfo.name), hostInfo.OsID)
	default:
		return fmt.Errorf("%s not supported on: %s", strings.ToUpper(extInfo.name), hostInfo.OsID)
	}

	if extInfo.params != nil {
		weaveLaunch(provisioner, extInfo)
	} else {
		log.Debugf("%s: launching first weave node", strings.ToUpper(extInfo.name))
		if _, err := provisioner.SSHCommand("sudo weave launch"); err != nil {
			return err
		}
	}

	return nil
}

func weaveLaunch(provisioner provision.Provisioner, extInfo *ExtensionInfo) error {
	for k, v := range extInfo.params {
		if k == "peer" {
			log.Debugf("%s: Launching Peer Connection to: %s", strings.ToUpper(extInfo.name), v)
			if _, err := provisioner.SSHCommand(fmt.Sprintf("sudo weave launch %s", v)); err != nil {
				return err
			}
		}
	}
	return nil
}
