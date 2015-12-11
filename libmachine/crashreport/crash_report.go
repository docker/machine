package crashreport

import (
	"fmt"
	"os"
	"runtime"

	"bytes"

	"os/exec"

	"path/filepath"

	"errors"

	"github.com/bugsnag/bugsnag-go"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/version"
)

const (
	defaultAPIKey  = "a9697f9a010c33ee218a65e5b1f3b0c1"
	noreportAPIKey = "no-report"
)

var apiKey string

// Configure the apikey for bugnag
func Configure(key string) {
	apiKey = defaultAPIKey
	if key != "" {
		apiKey = key
	}
}

// Send through http the crash report to bugsnag need a call to Configure(apiKey) before
func Send(err error, context string, driverName string, command string) error {
	if noReportFileExist() || apiKey == noreportAPIKey {
		log.Debug("Opting out of crash reporting.")
		return nil
	}

	if apiKey == "" {
		return errors.New("Not sending report since no api key has been set.")
	}

	bugsnag.Configure(bugsnag.Configuration{
		APIKey: apiKey,
		// XXX we need to abuse bugsnag metrics to get the OS/ARCH information as a usable filter
		// Can do that with either "stage" or "hostname"
		ReleaseStage:    fmt.Sprintf("%s (%s)", runtime.GOOS, runtime.GOARCH),
		ProjectPackages: []string{"github.com/docker/machine/[^v]*"},
		AppVersion:      version.FullVersion(),
		Synchronous:     true,
		PanicHandler:    func() {},
		Logger:          new(logger),
	})

	metaData := bugsnag.MetaData{}

	metaData.Add("app", "compiler", fmt.Sprintf("%s (%s)", runtime.Compiler, runtime.Version()))
	metaData.Add("device", "os", runtime.GOOS)
	metaData.Add("device", "arch", runtime.GOARCH)

	detectRunningShell(&metaData)
	detectUname(&metaData)

	var buffer bytes.Buffer
	for _, message := range log.History() {
		buffer.WriteString(message + "\n")
	}
	metaData.Add("history", "trace", buffer.String())
	return bugsnag.Notify(err, metaData, bugsnag.SeverityError, bugsnag.Context{String: context}, bugsnag.ErrorClass{Name: fmt.Sprintf("%s/%s", driverName, command)})
}

func noReportFileExist() bool {
	optOutFilePath := filepath.Join(mcndirs.GetBaseDir(), "no-error-report")
	if _, err := os.Stat(optOutFilePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func detectRunningShell(metaData *bugsnag.MetaData) {
	shell := os.Getenv("SHELL")
	if shell != "" {
		metaData.Add("device", "shell", shell)
	}
	shell = os.Getenv("__fish_bin_dir")
	if shell != "" {
		metaData.Add("device", "shell", shell)
	}
}

func detectUname(metaData *bugsnag.MetaData) {
	cmd := exec.Command("uname", "-s")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	metaData.Add("device", "uname", string(output))
}
