package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/docker/machine/cli"
)

var funcMap = template.FuncMap{
	"json": func(v interface{}) string {
		a, _ := json.Marshal(v)
		return string(a)
	},
	"prettyjson": func(v interface{}) string {
		a, _ := json.MarshalIndent(v, "", "    ")
		return string(a)
	},
}

func cmdInspect(c *cli.Context) error {
	if len(c.Args()) == 0 {
		cli.ShowCommandHelp(c, "inspect")
		return ErrExpectedOneMachine
	}

	host, err := getFirstArgHost(c)
	if err != nil {
		return err
	}

	tmplString := c.String("format")
	if tmplString != "" {
		var tmpl *template.Template
		var err error
		if tmpl, err = template.New("").Funcs(funcMap).Parse(tmplString); err != nil {
			return fmt.Errorf("Template parsing error: %v\n", err)
		}

		jsonHost, err := json.Marshal(host)
		if err != nil {
			return err
		}

		obj := make(map[string]interface{})
		if err := json.Unmarshal(jsonHost, &obj); err != nil {
			return err
		}

		if err := tmpl.Execute(os.Stdout, obj); err != nil {
			return err
		}

		os.Stdout.Write([]byte{'\n'})
	} else {
		prettyJSON, err := json.MarshalIndent(host, "", "    ")
		if err != nil {
			return err
		}

		fmt.Println(string(prettyJSON))
	}

	return nil
}
