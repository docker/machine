package pkgaction

type ServiceAction int

const (
	Restart ServiceAction = iota
	Start
	Stop
	Enable
	Disable
	DaemonReload
)

var serviceActions = []string{
	"restart",
	"start",
	"stop",
	"enable",
	"disable",
	"daemon-reload",
}

func (s ServiceAction) String() string {
	if int(s) >= 0 && int(s) < len(serviceActions) {
		return serviceActions[s]
	}

	return ""
}

type PackageAction int

const (
	Install PackageAction = iota
	Remove
	Upgrade
)

var packageActions = []string{
	"install",
	"remove",
	"upgrade",
}

func (s PackageAction) String() string {
	if int(s) >= 0 && int(s) < len(packageActions) {
		return packageActions[s]
	}

	return ""
}
