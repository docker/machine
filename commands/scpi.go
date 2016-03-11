package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
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

	srcHost, srcImage, err := getScpiHostAndImage(src, true, api)
	if err != nil {
		return err
	}

	destHost, _, err := getScpiHostAndImage(dest, false, api)
	if err != nil {
		return err
	}

	// Establish connection to source and destination hosts
	srcClient, err := srcHost.CreateSSHClient()
	if err != nil {
		return err
	}
	destClient, err := destHost.CreateSSHClient()
	if err != nil {
		return err
	}

	saveFilename := generateSaveFilename(srcImage)
	if err = srcClient.Shell(fmt.Sprintf("docker save -o %s %s", saveFilename, srcImage)); err != nil {
		return err
	}

	// Copy archive from source to destination
	cpSrc := fmt.Sprintf("%s:%s", srcHost.Name, saveFilename)
	cpDest := fmt.Sprintf("%s:%s", destHost.Name, saveFilename)
	cmd, err := getScpCmd(cpSrc, cpDest, false, hostInfoLoader)
	if err != nil {
		return err
	}
	if err = runCmdWithStdIo(*cmd); err != nil {
		return err
	}

	// Restore image on destination
	if err = destClient.Shell(fmt.Sprintf("docker load -i %s", saveFilename)); err != nil {
		return err
	}

	// cleanup
	rmCmd := fmt.Sprintf("rm %s", saveFilename)
	srcClient.Shell(rmCmd)
	destClient.Shell(rmCmd)

	return nil
}

func getScpiHostAndImage(hostAndPath string, isSource bool, api libmachine.API) (*host.Host, string, error) {
	var (
		hostName = hostAndPath
		image    string
	)

	if isSource {
		// Validate path. e.g. "default:redis:2.8.23"
		if !strings.Contains(hostAndPath, ":") {
			return nil, "", errors.New("Invalid source image")
		}

		// Split hostname and image name
		parts := strings.SplitN(hostAndPath, ":", 2)
		hostName = parts[0]
		image = parts[1]

		if !strings.Contains(image, ":") {
			image = fmt.Sprintf("%s:latest", image)
		}
	}

	// Remote path
	host, err := api.Load(hostName)
	if err != nil {
		return nil, "", fmt.Errorf("Error loading host: %s", err)
	}

	return host, image, nil
}

func generateSaveFilename(srcImage string) string {
	filename := strings.Replace(strings.Replace(srcImage, "/", "_", -1), ":", "__", -1)
	return fmt.Sprintf("/tmp/%s.tar", filename)
}
