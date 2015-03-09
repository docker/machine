package hypervisor

// HypervisorType represents the type of hypervisor a machine is using
type HypervisorType int

const (
	None HypervisorType = iota
	Local
	Remote
)

var hypervisorTypes = []string{
	"",
	"Local",
	"Remote",
}

// Given a type, returns its string representation
func (h HypervisorType) String() string {
	if int(h) >= 0 && int(h) < len(hypervisorTypes) {
		return hypervisorTypes[h]
	} else {
		return ""
	}
}
