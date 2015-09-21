package extension

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/log"
)

func appendEnvFile(provisioner provision.Provisioner, extInfo *ExtensionInfo) error {
	for k, v := range extInfo.envs {
		log.Debugf("%s: Setting Environment Variable: %s", strings.ToUpper(extInfo.name), k)
		if _, err := provisioner.SSHCommand(fmt.Sprintf("sudo -E bash -c 'echo %s=%s >> /etc/environment'", k, v)); err != nil {
			return err
		}
	}
	return nil
}

func fileTransfer(provisioner provision.Provisioner, hostInfo *ExtensionParams, extInfo *ExtensionInfo, files files, checkFileKey string) error {
	for fileKey, v := range files {
		if checkFileKey != "" && fileKey != checkFileKey {
			continue
		}
		var source, destination string
		for key, value := range v {
			switch key {
			case "source":
				source = value
			case "destination":
				destination = value
			}
		}

		if _, err := os.Stat(source); os.IsNotExist(err) {
			return fmt.Errorf("No such file or directory: %s", source)
		}

		destDir := filepath.Dir(destination)

		//check if the destination directory exists, if it doesn't, create it
		log.Debugf("%s: Creating directory if it doesn't exist: %s", strings.ToUpper(extInfo.name), destDir)
		if _, err := provisioner.SSHCommand(fmt.Sprintf("sudo mkdir -p %s", destDir)); err != nil {
			return err
		}

		dmFile := os.Args[0]
		dmPath, _ := filepath.Abs(dmFile)
		log.Debugf("%s: Using docker machine binary at: %s", strings.ToUpper(extInfo.name), dmPath)
		app := dmPath
		arg0 := "scp"
		arg1 := source
		arg2 := fmt.Sprintf("%v:%v", strings.TrimSpace(hostInfo.Hostname), destination)
		//call docker-machine scp to transfer the local file to a directory where it has writeable access
		log.Debugf("%s: Transferring %s to destination: %s", strings.ToUpper(extInfo.name), arg1, arg2)
		if _, err := exec.Command(app, arg0, arg1, arg2).Output(); err != nil {
			return err
		}
	}
	return nil
}

func returnFilePathString(fullpath string) (file, path string) {
	fullPathSlice := strings.SplitAfterN(fullpath, "/", 100)
	file = fullPathSlice[len(fullPathSlice)-1]
	pathSlice := fullPathSlice[:len(fullPathSlice)-1]
	path = strings.Join(pathSlice[:], "")
	return file, path
}

func execRemoteCommand(provisioner provision.Provisioner, extInfo *ExtensionInfo) error {
	for _, val := range extInfo.run {
		log.Debugf("%s: Running command: %s", strings.ToUpper(extInfo.name), val)
		if _, err := provisioner.SSHCommand(val); err != nil {
			return err
		}
	}
	return nil
}
