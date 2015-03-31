package commands

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
)

func cmdUrl(c *cli.Context) {
	url, err := getHost(c).GetURL()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(url)
}
