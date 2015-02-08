package drivers

import "fmt"
import "github.com/codegangsta/cli"

func DistinctFlags(flags *[]cli.Flag) {
	var unique []cli.Flag

	find := func(flag cli.Flag) error {
		for _, v := range unique {
			if flag == v {
				return nil
			}
		}

		return fmt.Errorf("Not found")
	}

	for _, v := range *flags {
		if err := find(v); err != nil {
			unique = append(unique, v)
		}
	}

	*flags = unique
}
