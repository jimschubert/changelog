# Contributing

:art: Thanks for considering a contribution! Please understand this document fully before any contribution.

You must agree to the [code of conduct](./CODE_OF_CONDUCT.md). It's pretty standard.

Before contributing new features or functionality, please first open an issue to discuss the change.

## Setup

This project is written in [Go](https://golang.org/). 

You'll need the following installed:

* [Go 1.14+](https://golang.org/doc/install)
* [goreleaser](https://goreleaser.com/) (optional)
* [Docker](https://www.docker.com/) (optional)

Then, clone the repository:

```bash
git clone git@github.com:jimschubert/changelog.git
```

The project uses [Modules](https://github.com/golang/go/wiki/Modules), so in your new directory run:

```bash
go mod tidy
```

Make sure it all works before contributing:

```bash
go test -v ./...
```

## Commit

Please use well-formatted commit messages, and try to follow [Go's Good Commit Message Guidelines](https://golang.org/doc/contribute.html#commit_messages).

## Pull Request

Fork the project and push your changes to a clearly named branch. Open a pull request against the main repository's `master` branch.

GitHub Workflows are performed on every Pull Request to verify your change. Pull Requests will not be merged with failing validations unless the error 
is previously known, actively worked on, and the new code will have no additional negative impact on the error.

## Sponsoring

Contributions to this repository are excluded from financial sponsorships provided by users of the software. One or more maintainers may be added
to accept sponsors, but any financial sponsorship is up to the discretion of those sponsoring and not those being sponsored. If you'd like to sponsor
my work, please visit [GitHub Sponsors](https://github.com/sponsors/jimschubert).
