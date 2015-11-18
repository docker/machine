package log

import (
	"fmt"
	bs "github.com/bugsnag/bugsnag-go"
	"github.com/docker/machine/log/bugsnag"
	"github.com/docker/machine/version"
	"net"
	"os"
	"runtime"
	"strings"
)

// Empty logger for bugsnag
type void struct {
}

func (d *void) Printf(fmtString string, args ...interface{}) {
}

// SetBugsnag configures the bugsnag hook for logrus
func SetBugsnag(apiKey string) error {
	stage := fmt.Sprintf("%s (%s)", runtime.GOOS, runtime.GOARCH)

	bs.Configure(bs.Configuration{
		APIKey: apiKey,
		// XXX we need to abuse bugsnag metrics to get the OS/ARCH information as a usable filter
		// Can do that with either "stage" or "hostname"
		ReleaseStage: stage,
		//		Hostname: stage,
		ProjectPackages: []string{"github.com/docker/machine/[^v]*"},
		AppVersion:      version.Version + " (" + version.GitCommit + ")",
		Logger:          new(void),
		Synchronous:     true,
	})

	hook, err := bugsnag.NewBugsnagHook()
	if err != nil {
		return err
	}

	// Add runtime informations
	hook.Add("app", "compiler", fmt.Sprintf("%s (%s)", runtime.Compiler, runtime.Version()))

	hook.Add("device", "os", runtime.GOOS)
	hook.Add("device", "arch", runtime.GOARCH)

	// Add environ
	for _, v := range os.Environ() {
		idx := strings.Index(v, "=")
		k := v[:idx]
		v = v[idx+1:]
		hook.Add("environ", k, v)
	}

	// Add network information
	t, _ := net.Interfaces()
	for _, value := range t {
		hook.Add("network", value.Name, value)
		addr, _ := value.Addrs()
		for _, a := range addr {
			hook.Add("network", fmt.Sprintf("%s-network", value.Name), a.Network())
			hook.Add("network", fmt.Sprintf("%s-address", value.Name), a.String())
		}
	}

	// Anything else?

	logger.Hooks.Add(hook)

	return nil
}
