package commands

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
)

func cmdIp(c *cli.Context) {
	ip, err := getHost(c).Driver.GetIP()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(ip)
}
