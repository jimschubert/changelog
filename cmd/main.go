package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"os"
	"strings"
)

var version = ""
var date = ""
var commit = ""
var projectName = ""

var opts struct {
	Owner string `short:"o" long:"owner" description:"GitHub Owner/Org name (required)" required:"true"`

	Repo string `short:"r" long:"repo" description:"GitHub Repo name (required)" required:"true"`

	From string `short:"f" long:"from" description:"Begin changelog from this commit or tag"`

	To string `short:"t" long:"to" description:"End changelog at this commit or tag"`

	Config string `short:"c" long:"config" description:"Config file location for more advanced options"`

	Version bool `short:"v" long:"version" description:"Display version information"`
}

const parseArgs = flags.HelpFlag | flags.PassDoubleDash

func main() {
	parser := flags.NewParser(&opts, parseArgs)
	_, err := parser.Parse()
	if err != nil {
		flagError := err.(*flags.Error)
		if flagError.Type == flags.ErrHelp {
			parser.WriteHelp(os.Stdout)
			return
		}

		if flagError.Type == flags.ErrUnknownFlag {
			_, _ = fmt.Fprintf(os.Stderr, "%s. Please use --help for available options.\n", strings.Replace(flagError.Message, "unknown", "Unknown", 1))
			return
		}
		_, _ = fmt.Fprintf(os.Stderr, "Error parsing command line options: %s\n", err)
		return
	}

	if opts.Version {
		fmt.Printf("%s %s (%s)\n", projectName, version, commit)
		return
	}

}
