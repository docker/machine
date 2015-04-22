package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
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
	tmplString := c.String("format")
	if tmplString != "" {
		var tmpl *template.Template
		var err error
		if tmpl, err = template.New("").Funcs(funcMap).Parse(tmplString); err != nil {
			log.Fatalf("Template parsing error: %v\n", err)
		}

		jsonHost, err := json.Marshal(getHost(c))
		if err != nil {
			log.Fatal(err)
		}
		obj := make(map[string]interface{})
		if err := json.Unmarshal(jsonHost, &obj); err != nil {
			log.Fatal(err)
		}

		if err := tmpl.Execute(os.Stderr, obj); err != nil {
			log.Fatal(err)
		}
		os.Stderr.Write([]byte{'\n'})
	} else {
		prettyJSON, err := json.MarshalIndent(getHost(c), "", "    ")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(prettyJSON))
	}
}
