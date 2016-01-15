package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

// re-implementation of private function in https://github.com/golang/go/blob/master/src/syscall/syscall_windows.go#L945
func getProcessEntry(pid int) (pe *syscall.ProcessEntry32, err error) {
	snapshot, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer syscall.CloseHandle(syscall.Handle(snapshot))

	var processEntry syscall.ProcessEntry32
	processEntry.Size = uint32(unsafe.Sizeof(processEntry))
	err = syscall.Process32First(snapshot, &processEntry)
	if err != nil {
		return nil, err
	}

	for {
		if processEntry.ProcessID == uint32(pid) {
			pe = &processEntry
			return
		}

		err = syscall.Process32Next(snapshot, &processEntry)
		if err != nil {
			return nil, err
		}
	}
}

// startedBy returns the exe file name of the parent process.
func startedBy() (exefile string, err error) {
	ppid := os.Getppid()

	pe, err := getProcessEntry(ppid)
	if err != nil {
		return "", err
	}

	name := syscall.UTF16ToString(pe.ExeFile[:])
	return name, nil
}

func Detect() (string, error) {
	shell := os.Getenv("SHELL")

	if shell == "" {
		shell, err := startedBy()
		if err != nil {
			return "cmd", err // defaulting to cmd
		}
		if strings.Contains(strings.ToLower(shell), "powershell") {
			return "powershell", nil
		} else if strings.Contains(strings.ToLower(shell), "cmd") {
			return "cmd", nil
		} else {
			fmt.Printf("You can further specify your shell with either 'cmd' or 'powershell' with the --shell flag.\n\n")
			return "cmd", nil // this could be either powershell or cmd, defaulting to cmd
		}
	}

	if os.Getenv("__fish_bin_dir") != "" {
		return "fish", nil
	}

	return filepath.Base(shell), nil
}
