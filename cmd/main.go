// Copyright 2020 Jim Schubert
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"

	"changelog"
	"changelog/model"
)

var version = ""
var date = ""
var commit = ""
var projectName = ""

var opts struct {
	Owner string `short:"o" long:"owner" description:"GitHub Owner/Org name (required)" required:"true" env:"GITHUB_OWNER"`

	Repo string `short:"r" long:"repo" description:"GitHub Repo name (required)" required:"true" env:"GITHUB_REPO"`

	From string `short:"f" long:"from" description:"Begin changelog from this commit or tag"`

	To string `short:"t" long:"to" description:"End changelog at this commit or tag" default:"HEAD"`

	Config *string `short:"c" long:"config" description:"Config file location for more advanced options beyond defaults"`

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

	config := model.LoadOrNewConfig(opts.Config, opts.Owner, opts.Repo)

	changes := changelog.Changelog{
		Config: config,
		From:   opts.From,
		To:     opts.To,
	}

	err = changes.Generate(os.Stdout)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "generation failed: %s", err)
	}
}
