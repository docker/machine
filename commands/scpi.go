package commands

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/docker/machine/libmachine"
)

func cmdScpi(c CommandLine, api libmachine.API) error {
	args := c.Args()
	if len(args) != 2 {
		c.ShowHelp()
		return errWrongNumberArguments
	}

	src := args[0]
	dest := args[1]
	hostInfoLoader := &storeHostInfoLoader{api}

	dockerExec, machineExec, err := getDockerAndMachineExecs()
	if err != nil {
		return err
	}

	srcHostName, srcImage, err := parseScpiTargetInfo(src, true, hostInfoLoader)
	if err != nil {
		return err
	}
	srcHostArgs, err := getHostArgs(srcHostName, machineExec)
	if err != nil {
		return err
	}

	destHostName, _, err := parseScpiTargetInfo(dest, false, hostInfoLoader)
	if err != nil {
		return err
	}
	destHostArgs, err := getHostArgs(destHostName, machineExec)
	if err != nil {
		return err
	}

	saveFilename := generateSaveFilename(srcImage)
	if err = saveImage(srcHostArgs, saveFilename, srcImage, dockerExec); err != nil {
		return err
	}

	if err = loadAndCleanup(destHostArgs, saveFilename, dockerExec); err != nil {
		return err
	}
	return nil
}

func saveImage(srcHostArgs []string, saveFilename, srcImage, dockerExec string) error {
	saveCmdArgs := append(srcHostArgs, []string{"save", "-o", saveFilename, srcImage}...)
	saveCmd := exec.Command(dockerExec, saveCmdArgs...)
	if err := runCmdWithStdIo(*saveCmd); err != nil {
		return err
	}
	return nil
}

func loadAndCleanup(destHostArgs []string, saveFilename, dockerExec string) error {
	// Load image on destination
	loadCmdArgs := append(destHostArgs, []string{"load", "-i", saveFilename}...)
	loadCmd := exec.Command(dockerExec, loadCmdArgs...)
	if err := runCmdWithStdIo(*loadCmd); err != nil {
		return err
	}

	// Cleanup
	rmCmd := exec.Command("rm", saveFilename)
	if err := runCmdWithStdIo(*rmCmd); err != nil {
		return err
	}

	return nil
}

func getDockerAndMachineExecs() (string, string, error) {
	dockerExec, err := exec.LookPath("docker")
	if err != nil {
		return "", "", err
	}

	machineExec, err := exec.LookPath("docker-machine")
	if err != nil {
		return "", "", err
	}

	return dockerExec, machineExec, nil
}

func parseScpiTargetInfo(hostAndImage string, isSource bool, hostInfoLoader HostInfoLoader) (string, string, error) {
	var (
		hostName = hostAndImage
		image    string
	)

	if isSource {
		// Validate path. e.g. "default:redis:2.8.23"
		if !strings.Contains(hostAndImage, ":") {
			return "", "", errors.New("Invalid source image")
		}

		// Split hostname and image name
		parts := strings.SplitN(hostAndImage, ":", 2)
		hostName = parts[0]
		image = parts[1]

		if !strings.Contains(image, ":") {
			image = fmt.Sprintf("%s:latest", image)
		}
	}

	// Validate host
	if _, err := hostInfoLoader.load(hostName); err != nil {
		return "", "", fmt.Errorf("Error loading host: %s", err)
	}

	return hostName, image, nil
}

func getHostArgs(hostName, machineExec string) ([]string, error) {
	// Host flags
	rawHostArgs, err := exec.Command(machineExec, "config", hostName).Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(string(rawHostArgs), " "), nil
}

func generateSaveFilename(srcImage string) string {
	filename := strings.Replace(strings.Replace(srcImage, "/", "_", -1), ":", "__", -1)
	return fmt.Sprintf("/tmp/%s.tar", filename)
}
