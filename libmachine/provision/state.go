package provision

type ServiceState int

const (
	Restart ServiceState = iota
	Start
	Stop
)

var serviceStates = []string{
	"restart",
	"start",
	"stop",
}

func (s ServiceState) String() string {
	if int(s) >= 0 && int(s) < len(serviceStates) {
		return serviceStates[s]
	}

	return ""
}

type PackageState int

const (
	Installed PackageState = iota
	Missing
)

var packageStates = []string{
	"installed",
	"missing",
}

func (s PackageState) String() string {
	if int(s) >= 0 && int(s) < len(packageStates) {
		return packageStates[s]
	}

	return ""
}
