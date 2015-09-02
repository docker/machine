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

func cmdInspect(c *cli.Context) {
	if len(c.Args()) == 0 {
		cli.ShowCommandHelp(c, "inspect")
		fatal("You must specify a machine name")
	}

	tmplString := c.String("format")
	if tmplString != "" {
		var tmpl *template.Template
		var err error
		if tmpl, err = template.New("").Funcs(funcMap).Parse(tmplString); err != nil {
			fatalf("Template parsing error: %v\n", err)
		}

		jsonHost, err := json.Marshal(getFirstArgHost(c))
		if err != nil {
			fatal(err)
		}
		obj := make(map[string]interface{})
		if err := json.Unmarshal(jsonHost, &obj); err != nil {
			fatal(err)
		}

		if err := tmpl.Execute(os.Stdout, obj); err != nil {
			fatal(err)
		}
		os.Stdout.Write([]byte{'\n'})
	} else {
		prettyJSON, err := json.MarshalIndent(getFirstArgHost(c), "", "    ")
		if err != nil {
			fatal(err)
		}

		fmt.Println(string(prettyJSON))
	}
}
