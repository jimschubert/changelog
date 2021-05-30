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
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"

	"github.com/jimschubert/changelog"
	"github.com/jimschubert/changelog/model"
)

var version = ""
var date = ""
var commit = ""
var projectName = ""

type Options struct {
	Owner string `short:"o" long:"owner" description:"GitHub Owner/Org name" env:"GITHUB_OWNER" default:""`

	Repo string `short:"r" long:"repo" description:"GitHub Repo name" env:"GITHUB_REPO" default:""`

	From string `short:"f" long:"from" description:"Begin changelog from this commit or tag"`

	To string `short:"t" long:"to" description:"End changelog at this commit or tag" default:"master"`

	Config *string `short:"c" long:"config" description:"Config file location for more advanced options beyond defaults"`

	Local *bool `short:"l" long:"local" description:"Prefer local commits when gathering commit logs (as opposed to querying via API)"`

	MaxCommits *int `long:"max" description:"The maximum number of commits to include"`

	Version bool `short:"v" long:"version" description:"Display version information"`
}

const parseArgs = flags.HelpFlag | flags.PassDoubleDash

var opts Options
var parser = flags.NewParser(&opts, parseArgs)
var commandCompleted = errors.New("completed")
var commandError = errors.New("command failed")

func main() {
	parser := flags.NewParser(&opts, parseArgs)
	parser.SubcommandsOptional = true
	_, err := parser.Parse()
	handleError(err)

	if opts.Version {
		fmt.Printf("%s %s (%s)\n", projectName, version, commit)
		return
	}

	initLogging()

	config := model.LoadOrNewConfig(opts.Config, opts.Owner, opts.Repo)
	err = validateConfig(config)
	handleError(err)

	config.MaxCommits = opts.MaxCommits
	if opts.Local != nil {
		config.PreferLocal = opts.Local
	}

	log.WithFields(log.Fields{"config": config}).Debug("Loaded config.")

	changes := changelog.Changelog{
		Config: config,
		From:   opts.From,
		To:     opts.To,
	}

	err = changes.Generate(os.Stdout)
	handleError(err)
}

func validateConfig(opts *model.Config) error {
	var required []string

	// owner/repo are only required if config is empty
	// if they're empty by this point, nag user about CLI opts (regardless of whether they've used a config file)
	if opts.Owner == "" {
		required = append(required, "'-o, --owner'")
	}
	if opts.Repo == "" {
		required = append(required, "'-r, --repo'")
	}

	if len(required) > 0 {
		msg := fmt.Sprintf("the required arguments %s and %s were not provided",
			strings.Join(required[:len(required)-1], ", "), required[len(required)-1])
		return &flags.Error{ Type: flags.ErrRequired, Message: msg }
	}

	return nil
}

func handleError(err error) {
	if err != nil {
		if errors.Is(err, commandCompleted) {
			os.Exit(0)
		}

		if flagError, ok := err.(*flags.Error); ok {
			if flagError.Type == flags.ErrHelp {
				parser.WriteHelp(os.Stdout)
				os.Exit(0)
			}

			if flagError.Type == flags.ErrUnknownFlag {
				_, _ = fmt.Fprintf(os.Stderr, "%s. Please use --help for available options.\n", strings.Replace(flagError.Message, "unknown", "Unknown", 1))
				os.Exit(1)
			}
			_, _ = fmt.Fprintf(os.Stderr, "Error parsing command line options: %s\n", err)
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "generation failed: %s", err)
		}

		os.Exit(1)
	}
}

func initLogging() {
	logLevel, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		logLevel = "error"
	}
	ll, err := log.ParseLevel(logLevel)
	if err != nil {
		ll = log.DebugLevel
	}
	log.SetLevel(ll)
	log.SetOutput(os.Stderr)
}
