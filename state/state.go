package state

// State represents the state of a hosts
type State int

const (
	None State = iota
	Running
	Paused
	Saved
	Stopped
	Stopping
	Starting
	Error
)

var states = []string{
	"",
	"Running",
	"Paused",
	"Saved",
	"Stopped",
	"Stopping",
	"Starting",
	"Error",
}

func (s State) String() string {
	if int(s) < len(states) {
		return states[s]
	}
	return ""
}
