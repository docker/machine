package commands

<<<<<<< HEAD
func cmdKill(c CommandLine) error {
	return runActionWithContext("kill", c)
=======
import (
	"github.com/docker/machine/libmachine/log"
)

func cmdKill(c CommandLine) error {
	runActionWithContext("kill", c)

	hosts, err := getHostsFromContext(c)
	if err != nil {
		return err
	}

	for _, h := range hosts {
		currentState, err := h.Driver.GetState()
		if err != nil {
			log.Errorf("error getting state for host %s: %s", h.Name, err)
		}
		log.Printf("Machine \"%s\" state is %s", h.Name, currentState)
	}
	return nil
>>>>>>> 5e75843... Final commit for modified kill.go
}
