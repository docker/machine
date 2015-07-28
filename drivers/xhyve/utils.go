package xhyve

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/docker/machine/log"
)

var (
	//	ErrMachineExist     = errors.New("machine already exists")
	//	ErrMachineNotExist  = errors.New("machine does not exist")
	ErrDdNotFound      = errors.New("xhyve not found")
	ErrUuidgenNotFound = errors.New("uuidgen not found")
	//	ErrUuid2macNotFound = errors.New("uuid2mac not found")
	ErrHdiutilNotFound = errors.New("hdiutil not found")
	ErrVBMNotFound     = errors.New("VBoxManage not found")
	vboxManageCmd      = setVBoxManageCmd()
)

func dd(input string, output string, block string, count int) (string, string, error) { // TODO
	cmd := exec.Command("dd", fmt.Sprintf("if=%s", input), fmt.Sprintf("of=%s", output), fmt.Sprintf("bs=%s", block), fmt.Sprintf("count=%d", count))
	if os.Getenv("DEBUG") != "" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	err := cmd.Run()
	if err != nil {
		if ee, ok := err.(*exec.Error); ok && ee == exec.ErrNotFound {
			err = ErrDdNotFound
		}
	}

	fmt.Println(cmd)
	return stdout.String(), stderr.String(), err
}

func uuidgen() string {
	cmd := exec.Command("uuidgen")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	log.Debugf("executing: %v", cmd)

	err := cmd.Run()
	if err != nil {
		if ee, ok := err.(*exec.Error); ok && ee == exec.ErrNotFound {
			err = ErrUuidgenNotFound
		}
	}

	out := stdout.String()
	out = strings.Replace(out, "\n", "", -1)
	return out
}

func hdiutil(args ...string) error {
	cmd := exec.Command("hdiutil", args...)

	log.Debugf("executing: %v %v", cmd, strings.Join(args, " "))

	err := cmd.Run()
	if err != nil {
		log.Error(err)
	}

	return nil
}

// detect the VBoxManage cmd's path if needed
func setVBoxManageCmd() string {
	cmd := "VBoxManage"
	if path, err := exec.LookPath(cmd); err == nil {
		return path
	}
	if runtime.GOOS == "windows" {
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
		// look at HKEY_LOCAL_MACHINE\SOFTWARE\Oracle\VirtualBox\InstallDir
		p := "C:\\Program Files\\Oracle\\VirtualBox"
		if path, err := exec.LookPath(filepath.Join(p, cmd)); err == nil {
			return path
		}
	}
	return cmd
}

func vbm(args ...string) error {
	_, _, err := vbmOutErr(args...)
	return err
}

func vbmOut(args ...string) (string, error) {
	stdout, _, err := vbmOutErr(args...)
	return stdout, err
}

func vbmOutErr(args ...string) (string, string, error) {
	cmd := exec.Command(vboxManageCmd, args...)
	log.Debugf("executing: %v %v", vboxManageCmd, strings.Join(args, " "))
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	stderrStr := stderr.String()
	log.Debugf("STDOUT: %v", stdout.String())
	log.Debugf("STDERR: %v", stderrStr)
	if err != nil {
		if ee, ok := err.(*exec.Error); ok && ee == exec.ErrNotFound {
			err = ErrVBMNotFound
		}
	} else {
		// VBoxManage will sometimes not set the return code, but has a fatal error
		// such as VBoxManage.exe: error: VT-x is not available. (VERR_VMX_NO_VMX)
		if strings.Contains(stderrStr, "error:") {
			err = fmt.Errorf("%v %v failed: %v", vboxManageCmd, strings.Join(args, " "), stderrStr)
		}
	}
	return stdout.String(), stderrStr, err
}

func vboxVersionDetect() (string, error) {
	ver, err := vbmOut("-v")
	if err != nil {
		return "", err
	}
	return ver, err
}
