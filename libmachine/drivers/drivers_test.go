package drivers

import (
	"testing"

	"github.com/codegangsta/cli"
)

func TestGetCreateFlags(t *testing.T) {
	Register("foo", &RegisteredDriver{
		GetCreateFlags: func() []cli.Flag {
			return []cli.Flag{
				cli.StringFlag{
					Name:   "a",
					Value:  "",
					Usage:  "",
					EnvVar: "",
				},
				cli.StringFlag{
					Name:   "b",
					Value:  "",
					Usage:  "",
					EnvVar: "",
				},
				cli.StringFlag{
					Name:   "c",
					Value:  "",
					Usage:  "",
					EnvVar: "",
				},
			}
		},
	})
	Register("bar", &RegisteredDriver{
		GetCreateFlags: func() []cli.Flag {
			return []cli.Flag{
				cli.StringFlag{
					Name:   "d",
					Value:  "",
					Usage:  "",
					EnvVar: "",
				},
				cli.StringFlag{
					Name:   "e",
					Value:  "",
					Usage:  "",
					EnvVar: "",
				},
				cli.StringFlag{
					Name:   "f",
					Value:  "",
					Usage:  "",
					EnvVar: "",
				},
			}
		},
	})

	expected := []string{"-a \t", "-b \t", "-c \t", "-d \t", "-e \t", "-f \t"}

	// test a few times to catch offset issue
	// if it crops up again
	for i := 0; i < 5; i++ {
		flags := GetCreateFlags()
		for j, e := range expected {
			if flags[j].String() != e {
				t.Fatal("Flags are out of order")
			}
		}
	}
}
