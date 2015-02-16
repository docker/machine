package pkgaction

type ServiceAction int

const (
	Restart ServiceAction = iota
	Start
	Stop
)

var serviceActions = []string{
	"restart",
	"start",
	"stop",
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
)

var packageActions = []string{
	"install",
	"remove",
}

func (s PackageAction) String() string {
	if int(s) >= 0 && int(s) < len(packageActions) {
		return packageActions[s]
	}

	return ""
}
