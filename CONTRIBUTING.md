# Contributing to cukes-vhs

Thank you for your interest in contributing to cukes-vhs. This document provides guidelines for setting up your development environment and contributing to the project.

## Development Environment

### Prerequisites

- Go 1.23 or later
- Node.js (for commitlint)
- golangci-lint v2.8.0 or later
- make

### Setup

Clone the repository and install dependencies:

```bash
git clone https://github.com/boodah-consulting/cukes-vhs.git
cd cukes-vhs
make install-git-hooks
make ci-install-tools
npm install
```

## Running Tests

Run the full test suite:

```bash
make test          # Standard test run
make test-race     # With race detector
make coverage      # With coverage report (95% threshold)
```

Run a single test by description:

```bash
make individual-test TEST="description"
```

Run BDD tests only:

```bash
make bdd
```

## Linting

Run linters before committing:

```bash
make lint          # Run golangci-lint (51 linters)
make vet           # Run go vet
make staticcheck   # Run staticcheck
```

Run full compliance checks:

```bash
make check-compliance
```

## Code Style

This project follows strict code style conventions. See `AGENTS.md` for detailed rules. Key points:

- **Imports**: Grouped (stdlib, external, internal) and alphabetical
- **Naming**: Files use `snake_case.go`, packages use `lowercase`, types/public use `PascalCase`, private use `camelCase`
- **Forbidden**: Comments inside function bodies, `TODO`/`FIXME`/`HACK` markers
- **Language**: British English throughout (behaviour, colour, initialise, analyse, etc.)

## Commit Conventions

This project follows [Conventional Commits](https://www.conventionalcommits.org/).

### Format

```
type(scope): description

Allowed types: feat, fix, docs, style, refactor, test, chore, perf, ci, build, revert
Allowed scopes: parser, generator, template, validator, golden, renderer, mapping, cli, ci, deps, repo, lint
```

### AI-Attributed Commits

For AI-generated changes, use:

```bash
make ai-commit FILE=/tmp/commit.txt
```

This ensures proper attribution and co-authoring metadata.

## Testing Conventions

- All tests use [Ginkgo](https://onsi.github.io/ginkgo/) BDD framework with [Gomega](https://onsi.github.io/gomega/) assertions
- Test files are named `*_test.go` and reside in the same package as the code they test
- Each package contains a `suite_test.go` for Ginkgo suite initialisation
- Use `Describe`, `Context`, and `It` blocks to define behaviour clearly

## Pull Requests

Before submitting a pull request:

1. Run `make check-compliance` to ensure all checks pass
2. Write tests for new functionality
3. Update documentation as needed
4. Follow commit conventions
5. Ensure all tests pass

## Questions?

If you have questions or need help, please open an issue on GitHub.
