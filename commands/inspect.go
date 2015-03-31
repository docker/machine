package commands

import (
	"encoding/json"
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
)

func cmdInspect(c *cli.Context) {
	prettyJSON, err := json.MarshalIndent(getHost(c), "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(prettyJSON))
}
