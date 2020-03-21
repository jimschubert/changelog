# GitHub Changelog Generator

[![MIT License](https://img.shields.io/badge/license-MIT-blue)](./LICENSE)
![Go Version](https://img.shields.io/github/go-mod/go-version/jimschubert/changelog)
![Go](https://github.com/jimschubert/changelog/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/jimschubert/changelog)](https://goreportcard.com/report/github.com/jimschubert/changelog)
![Docker Pulls](https://img.shields.io/docker/pulls/jimschubert/changelog)
[![Lines of Code](https://sonarcloud.io/api/project_badges/measure?project=jimschubert_changelog&metric=ncloc)](https://sonarcloud.io/dashboard?id=jimschubert_changelog)

Changelog is a cross-platform changelog generator for GitHub repositories. It queries between any two git branches or tags as supported by the [GitHub Commits API](https://developer.github.com/v3/repos/commits/#compare-two-commits),
and generates a changelog of commits between the two. It supports [templates](https://golang.org/pkg/text/template/) for those who want more control over generated output.

Changelog's more advanced features support excluding commits from the changelog and grouping commits by heading based on regular expression patterns.

## Usage

A `GITHUB_TOKEN` environment variable must be provided for changelog to operate effectively.

```
Usage:
  changelog [OPTIONS]

Application Options:
  -o, --owner=   GitHub Owner/Org name (required) [$GITHUB_OWNER]
  -r, --repo=    GitHub Repo name (required) [$GITHUB_REPO]
  -f, --from=    Begin changelog from this commit or tag
  -t, --to=      End changelog at this commit or tag (default: HEAD)
  -c, --config=  Config file location for more advanced options beyond defaults
  -v, --version  Display version information

Help Options:
  -h, --help     Show this help message
```

The changelog output is written to standard output and can be redirected to overwrite or append to a file.

### Basic

The changelog generator doesn't assume a start or end tag, and doesn't evaluate existing tags to determine tag order. If `from` and `to` options are not provided, your changelog will result in the single latest commit on `master`.

You may specify `GITHUB_OWNER` and `GITHUB_REPO` as environment variables for use in CI.

#### Examples

**Output for single latest commit on master**

```bash
./changelog -o jimschubert -r changelog -f master~1 -t master
```

**Output from some version to latest master**

```bash
./changelog -o jimschubert -r changelog -f v0.1
```

**Sample output from one version to another**

```bash
./changelog -o jimschubert -r kopper -f v0.0.2 -t v0.0.3

## v0.0.3

* [d12243c81d](https://github.com/jimschubert/kopper/commit/d12243c81d6b4b45547929d97e49277d1cae4110) Bump version 0.0.3 ([jimschubert](https://github.com/jimschubert))
* [41f8fafd25](https://github.com/jimschubert/kopper/commit/41f8fafd25ff336f0de2f16c91ec199aec577843) Support name/description in TypedArgumentParser ([jimschubert](https://github.com/jimschubert))
* [91b5f02d99](https://github.com/jimschubert/kopper/commit/91b5f02d9918e769493fbeb22fe0ff884ac99b67) Support writing directly to PrintStream ([jimschubert](https://github.com/jimschubert))
* [ec986bff99](https://github.com/jimschubert/kopper/commit/ec986bff995507c92f10f74b6acae840eb5ab1dc) 0.0.3-SNAPSHOT ([jimschubert](https://github.com/jimschubert))


<em>For more details, see <a href="https://github.com/jimschubert/kopper/compare/v0.0.2...v0.0.3">v0.0.2..v0.0.3</a></em>
```

#### Default Template

The default template used in basic usage will output Markdown. The template is defined as:

```
## {{.Version}}

{{range .Items -}}
* [{{.CommitHashShort}}]({{.CommitURL}}) {{.Title}} ({{if .IsPull}}[contributed]({{.PullURL}}) by {{end}}[{{.Author}}]({{.AuthorURL}}))
{{end}}

<em>For more details, see <a href="{{.CompareURL}}">{{.PreviousVersion}}..{{.Version}}</a></em>
```

Currently, you must define an external JSON configuration file to override the default template.

## Install

Latest binary releases are available via [GitHub Releases](https://github.com/jimschubert/changelog/releases).

The preferred way to run changelog is via Docker. For example:

```bash
docker pull jimschubert/changelog:latest
docker run -e GITHUB_TOKEN=yourtoken \
           -e GITHUB_OWNER=jimschubert \
           -e GITHUB_REPO=changelog \
   jimschubert/changelog:latest -f v0.1 -t v0.2 >> CHANGELOG.md
```

## Advanced

More advanced scenarios require an external JSON configuration object which can be loaded by the `--config` option. The following example properties are supported by the config (comments added inline for brevity):

```json5
{
  // "commits" or "prs", defaults to commits. "prs" will soon allow for resolving labels 
  // from pull requests
  "resolve": "commits",

  // "asc" or "desc", determines the order of commits in the output
  "sort": "asc",
  
  // GitHub user or org name
  "owner": "jimschubert",  
   
  // Repository name
  "repo": "changelog",

  // Enterprise GitHub base url
  "enterprise": "https://ghe.example.com",

  // Path to custom template following Go Text template syntax
  "template": "/path/to/your/template",

  // Group commits by headings based on patterns supporting Perl syntax regex or plain text
  "groupings": [
    { "name":  "Contributions", "patterns":  [ "(?i)\\bfeat\\b" ] }
  ],

  // Exclude commits based on this set of patterns or texts
  // (useful for common maintenance commit messages)
  "exclude": [
    "^(?i)release\\s+\\d+\\.\\d+\\.\\d+",
    "^(?i)minor fix\\b",
    "^(?i)wip\\b"
  ]
}
```

### Grouping

Grouping is done by the `name` property of the groupings array objects. You will want to provide a custom template to display groupings.

For example, create a directory at `/tmp/changelog` to contain a sample JSON and template.

Save the follow **template** as `template.tmpl`:

```gotemplate
## {{.Version}}

{{range $key, $value := .Grouped -}}
### {{ $key }}

{{range $value -}}
* [{{.CommitHashShort}}]({{.CommitURL}}) {{.Title}} ({{if .IsPull}}[contributed]({{.PullURL}}) by {{end}}[{{.Author}}]({{.AuthorURL}}))
{{end}}
{{end}}

<em>For more details, see <a href="{{.CompareURL}}">{{.PreviousVersion}}..{{.Version}}</a></em>
```

And save the following as `config.json` (note: template currently requires a full path to the template file):

```json
{
  "sort": "desc",
  "template": "/tmp/changelog/template.tmpl",
  "groupings": [
    {
      "name": "Fixes",
      "patterns": [
        "(?i)\\bbug\\b",
        "(?i)\\bfix\\b"
      ]
    },
    {
      "name": "Features",
      "patterns": [
        "^(?i)feat:\\b",
        "^(?i)add:\\b"
      ]
    },
    {
      "name": "Cleanup",
      "patterns": [
        "(?i)\\brefactor:\\b"
      ]
    },
    {
      "name": "Other Contributions",
      "patterns": [
        ".?"
      ]
    }
  ],
  "exclude": [
    "(?i)readme\\b",
    "(?i)\\b>\\b",
    "(?i)\\btypo\\b",
    "^usage$",
    "Bump dependencies",
    "minor",
    "slight"
  ]
}
```

Now, run this against [cli/cli](https://github.com/cli/cli) v0.5.6 and v0.5.7. Via Docker:

```bash
docker run -e GITHUB_TOKEN=yourtoken \
   -e GITHUB_OWNER=cli \
   -e GITHUB_REPO=cli \
   jimschubert/changelog:latest -f v0.5.6 -t v0.5.7 >> /tmp/changelog/CHANGELOG.md
```

And via cli:

```bash
export GITHUB_TOKEN=your-token 
./changelog -o cli -r cli -f v0.5.6 -t v0.5.7 \
    -c /tmp/changelog/config.json >> /tmp/changelog/CHANGELOG.md
```

This changelog output in `/tmp/changelog/CHANGELOG.md` should look like this:

```text
## v0.5.7

### Fixes

* [4c3e498021](https://github.com/cli/cli/commit/4c3e498021997b40d3c78f8c858ed734f819b064) Fix column alignment and truncation for Eastern Asian languages ([mislav](https://github.com/mislav))
* [4ee995dafd](https://github.com/cli/cli/commit/4ee995dafdf98730c292c63c1b8a0fab5f2198d1) fix(486): Getting issue list on no remotes specified ([yashLadha](https://github.com/yashLadha))
* [f9649ebddd](https://github.com/cli/cli/commit/f9649ebddd1b6a9731046c98cd8019a245c82fde) Merge pull request #521 from yashLadha/bug/issue_list_on_no_remote ([mislav](https://github.com/mislav))

### Other Contributions

* [4727fc4659](https://github.com/cli/cli/commit/4727fc465982d3029324fc5b77ee37e28c29a2b3) Ensure descriptive error when no github.com remotes found ([mislav](https://github.com/mislav))
* [69304ce9af](https://github.com/cli/cli/commit/69304ce9af6100e49bb6a128a81639d48ac590ec) Merge pull request #518 from cli/eastern-asian ([mislav](https://github.com/mislav))
* [1a82e39ba9](https://github.com/cli/cli/commit/1a82e39ba9627654aca22e9608d5b81589855d41) Respect title & body from arguments to `pr create -w` ([mislav](https://github.com/mislav))
* [b5d0b7c640](https://github.com/cli/cli/commit/b5d0b7c640ad897f395a72074a0f4b31787e5826) Merge pull request #523 from cli/title-body-web ([mislav](https://github.com/mislav))



<em>For more details, see <a href="https://github.com/cli/cli/compare/v0.5.6...v0.5.7">v0.5.6..v0.5.7</a></em>
```

### Debugging

You may debug select operations such as groupings and exclusions by exporting `LOG_LEVEL=debug`.
