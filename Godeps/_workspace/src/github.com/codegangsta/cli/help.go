package cli

import "fmt"

// The text template for the Default help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.Name}} {{if .Flags}}[global options] {{end}}command{{if .Flags}} [command options]{{end}} [arguments...]

VERSION:
   {{.Version}}{{if or .Author .Email}}

AUTHOR:{{if .Author}}
  {{.Author}}{{if .Email}} - <{{.Email}}>{{end}}{{else}}
  {{.Email}}{{end}}{{end}}

COMMANDS:
   {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
   {{end}}{{if .Flags}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{end}}
`

// The text template for the command help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var CommandHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   command {{.Name}}{{if .Flags}} [command options]{{end}} [arguments...]{{if .Description}}

DESCRIPTION:
   {{.Description}}{{end}}{{if .Flags}}

OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{ end }}
`

// The text template for the subcommand help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var SubcommandHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.Name}} command{{if .Flags}} [command options]{{end}} [arguments...]

COMMANDS:
   {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
   {{end}}{{if .Flags}}
OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{end}}
`

var helpCommand = Command{
	Name:      "help",
	ShortName: "h",
	Usage:     "Shows a list of commands or help for one command",
	Action: func(c *Context) {
		args := c.Args()
		if args.Present() {
			ShowCommandHelp(c, args.First())
		} else {
			ShowAppHelp(c)
		}
	},
}

var helpSubcommand = Command{
	Name:      "help",
	ShortName: "h",
	Usage:     "Shows a list of commands or help for one command",
	Action: func(c *Context) {
		args := c.Args()
		if args.Present() {
			ShowCommandHelp(c, args.First())
		} else {
			ShowSubcommandHelp(c)
		}
	},
}

// Prints help for the App
type helpPrinter func(templ string, data interface{})

var HelpPrinter helpPrinter = nil

// Prints version for the App
var VersionPrinter = printVersion

func ShowAppHelp(c *Context) {
	HelpPrinter(AppHelpTemplate, c.App)
}

// Prints the list of subcommands as the default app completion method
func DefaultAppComplete(c *Context) {
	for _, command := range c.App.Commands {
		fmt.Fprintln(c.App.Writer, command.Name)
		if command.ShortName != "" {
			fmt.Fprintln(c.App.Writer, command.ShortName)
		}
	}
}

// Prints help for the given command
func ShowCommandHelp(c *Context, command string) {
	for _, c := range c.App.Commands {
		if c.HasName(command) {
			HelpPrinter(CommandHelpTemplate, c)
			return
		}
	}

	if c.App.CommandNotFound != nil {
		c.App.CommandNotFound(c, command)
	} else {
		fmt.Fprintf(c.App.Writer, "No help topic for '%v'\n", command)
	}
}

// Prints help for the given subcommand
func ShowSubcommandHelp(c *Context) {
	ShowCommandHelp(c, c.Command.Name)
}

// Prints the version number of the App
func ShowVersion(c *Context) {
	VersionPrinter(c)
}

func printVersion(c *Context) {
	fmt.Fprintf(c.App.Writer, "%v version %v\n", c.App.Name, c.App.Version)
}

// Prints the lists of commands within a given context
func ShowCompletions(c *Context) {
	a := c.App
	if a != nil && a.BashComplete != nil {
		a.BashComplete(c)
	}
}

// Prints the custom completions for a given command
func ShowCommandCompletions(ctx *Context, command string) {
	c := ctx.App.Command(command)
	if c != nil && c.BashComplete != nil {
		c.BashComplete(ctx)
	}
}

func checkVersion(c *Context) bool {
	if c.GlobalBool("version") {
		ShowVersion(c)
		return true
	}

	return false
}

func checkHelp(c *Context) bool {
	if c.GlobalBool("h") || c.GlobalBool("help") {
		ShowAppHelp(c)
		return true
	}

	return false
}

func checkCommandHelp(c *Context, name string) bool {
	if c.Bool("h") || c.Bool("help") {
		ShowCommandHelp(c, name)
		return true
	}

	return false
}

func checkSubcommandHelp(c *Context) bool {
	if c.GlobalBool("h") || c.GlobalBool("help") {
		ShowSubcommandHelp(c)
		return true
	}

	return false
}

func checkCompletions(c *Context) bool {
	if (c.GlobalBool(BashCompletionFlag.Name) || c.Bool(BashCompletionFlag.Name)) && c.App.EnableBashCompletion {
		ShowCompletions(c)
		return true
	}

	return false
}

func checkCommandCompletions(c *Context, name string) bool {
	if c.Bool(BashCompletionFlag.Name) && c.App.EnableBashCompletion {
		ShowCommandCompletions(c, name)
		return true
	}

	return false
}
