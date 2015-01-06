package drivers

import (
	"testing"

	"github.com/codegangsta/cli"
)

func TestGetCreateFlags(t *testing.T) {
	Register("foo", &RegisteredDriver{
		New: func(storePath string) (Driver, error) { return nil, nil },
		GetCreateFlags: func() []cli.Flag {
			return []cli.Flag{
				cli.StringFlag{"a", "", "", ""},
				cli.StringFlag{"b", "", "", ""},
				cli.StringFlag{"c", "", "", ""},
			}
		},
	})
	Register("bar", &RegisteredDriver{
		New: func(storePath string) (Driver, error) { return nil, nil },
		GetCreateFlags: func() []cli.Flag {
			return []cli.Flag{
				cli.StringFlag{"d", "", "", ""},
				cli.StringFlag{"e", "", "", ""},
				cli.StringFlag{"f", "", "", ""},
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
