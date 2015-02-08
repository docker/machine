package drivers

import "github.com/codegangsta/cli"

func DistinctFlags(flags *[]cli.Flag) {
	var unique []cli.Flag

	isInCollection := func(flag cli.Flag) bool {
		for _, v := range unique {
			if flag == v {
				return true
			}
		}

		return false
	}

	for _, v := range *flags {
		if ok := isInCollection(v); !ok {
			unique = append(unique, v)
		}
	}

	*flags = unique
}
