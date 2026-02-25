# AGENTS.md

This document provides AI agents with the necessary context, conventions, and patterns to work effectively on the **cukes-vhs** codebase.

## Overview
**cukes-vhs** is a Go-based CLI tool and library that converts Gherkin (Cucumber) feature files into VHS tape files for automated terminal recordings using [charmbracelet/vhs](https://github.com/charmbracelet/vhs).

## Directory Structure
- `cmd/cukes-vhs/` - CLI entry point and main package.
- `internal/cukesvhs/` - Core library including the parser, generator, renderer, validator, and templates.
- `internal/cli/` - Implementation of CLI command logic.
- `tools/analyzers/` - Custom static analyzers (e.g., docblocks).
- `scripts/` - Shell scripts for build, test, and CI processes.
- `.git-hooks/` - Git hooks for enforcing conventions and AI attribution.
- `demos/` - Examples of generated VHS outputs and demo recordings.

## Testing Conventions
- **Framework**: All tests use the [Ginkgo](https://onsi.github.io/ginkgo/) BDD framework with [Gomega](https://onsi.github.io/gomega/) assertions.
- **Organisation**: Test files are named `*_test.go` and reside in the same package as the code they test.
- **Suite Setup**: Each package contains a `suite_test.go` for Ginkgo suite initialisation.
- **BDD Style**: Use `Describe`, `Context`, and `It` blocks to define behaviour clearly and descriptively.
- **Golden Baselines**: The project uses golden baseline testing to validate generated tape outputs against known good versions.
- **Execution**: Run the full test suite using `make test`.

## Commit Conventions
- **Conventional Commits**: All commits must follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.
- **Allowed Scopes**: `parser`, `generator`, `template`, `validator`, `golden`, `renderer`, `mapping`, `cli`, `ci`, `deps`, `repo`, `lint`.
- **Allowed Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`, `ci`, `build`, `revert`.
- **Formatting**: Use lowercase subjects, no trailing periods, and a maximum of 72 characters for the subject line.
- **AI Attribution**: For AI-generated changes, use `make ai-commit FILE=<path>` to ensure proper attribution and co-authoring metadata.

## Key Make Targets
- `make test` - Runs all unit and integration tests.
- `make test-race` - Runs tests with the Go race detector enabled.
- `make coverage` - Generates a test coverage report (95% threshold).
- `make build` - Builds the `cukes-vhs` CLI binary.
- `make lint` - Runs `golangci-lint` with the project's configuration.
- `make check-compliance` - Performs a full suite of compliance and quality checks.
- `make ai-commit FILE=<path>` - Creates an AI-attributed commit from the provided message file.

## CI Workflows
- **`ci.yml`**: The main CI pipeline that runs tests on Ubuntu, macOS, and Windows. It also handles linting and security scanning (`gosec`).
- **`pr-validation.yml`**: Validates pull requests for commit message format, conventional commits, and AI attribution.
- **`release.yml`**: Automates the release process using `semantic-release`.

## Windows Compatibility Notes
- **Testing**: Certain tests are skipped on Windows due to `chmod` limitations or the absence of the `vhs` binary in the environment.
- **Helpers**: Use the `skipIfWindows()` helper function within test suites to handle platform-specific exclusions gracefully.
- **Path Handling**: Always use `filepath.ToSlash()` or the `path` package when working with paths intended for cross-platform compatibility to avoid backslash issues.
