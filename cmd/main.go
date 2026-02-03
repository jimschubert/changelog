// Copyright 2020-2026 Jim Schubert
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/alecthomas/kong"
	log "github.com/sirupsen/logrus"

	"github.com/jimschubert/changelog"
	"github.com/jimschubert/changelog/model"
)

//nolint:unused
var (
	version     = "dev"
	commit      = "unknown"
	date        = ""
	projectName = "changelog"
)

type Options struct {
	Owner string `short:"o" help:"GitHub Owner/Org name" env:"GITHUB_OWNER" default:""`

	Repo string `short:"r" help:"GitHub Repo name" env:"GITHUB_REPO" default:""`

	From string `short:"f" help:"Begin changelog from this commit or tag"`

	To string `short:"t" help:"End changelog at this commit or tag" default:"master"`

	Config *string `short:"c" help:"Config file location for more advanced options beyond defaults"`

	Local *bool `short:"l" help:"Prefer local commits when gathering commit logs (as opposed to querying via API)"`

	MaxCommits *int `name:"max" help:"The maximum number of commits to include"`

	Version kong.VersionFlag `short:"v" help:"Display version information"`
}

var opts Options

func main() {
	kong.Parse(&opts,
		kong.Name(projectName),
		kong.Description("Generate a changelog from GitHub commits"),
		kong.Vars{"version": fmt.Sprintf("%s %s (%s)", projectName, version, commit)},
	)

	initLogging()

	config := model.LoadOrNewConfig(opts.Config, opts.Owner, opts.Repo)
	err := validateConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

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
	if err != nil {
		fmt.Fprintf(os.Stdout, "generation failed: %s", err)
		os.Exit(1)
	}
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
		if len(required) == 1 {
			return fmt.Errorf("the required argument %s was not provided", required[0])
		}
		return fmt.Errorf("the required arguments %s and %s were not provided",
			required[0], required[1])
	}

	return nil
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
