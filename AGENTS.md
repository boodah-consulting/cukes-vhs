# AGENTS.md

This document provides AI agents with the necessary context, conventions, and patterns to work effectively on the **cukes-vhs** codebase.

## Overview
**cukes-vhs** is a Go-based CLI tool and library that converts Gherkin (Cucumber) feature files into VHS tape files for automated terminal recordings using [charmbracelet/vhs](https://github.com/charmbracelet/vhs).

### ALWAYS
- Follow the architecture and code style guidelines
- Write tests FIRST (TDD)
- Capture knowledge in memory (MCP) and obsidian (MCP) when DISCOVERING
  or CHANGING something
- Use `make ai-commit` for commits (not `git commit`)
- Ask if confidence is ASSUMED or UNKNOWN
- NEVER skip checks or tests
  - Unless explicitly approved by user, NEVER skip `make session-start` or any
    tests
- Refuse if about to do something that violates these rules
- **NEVER declare a task "done" or "complete" - only the USER decides when work is finished**
- **NEVER decide to skip, postpone, or deprioritize work - only the USER makes these decisions**

### Token Efficiency
- Be concise and precise
- No unnecessary words
- Specific, not vague
- Structured output

### Pre-Action Framework

Before ANY action, complete this checklist:

```
1. STOP  - What am I being asked to do?
2. THINK - What do I KNOW vs ASSUME vs UNKNOWN?
3. INVESTIGATE - Use tools, skills, git history to find out
4. CHOOSE - Which skills/tools are appropriate?
5. CONFIDENCE - VERIFIED / SUPPORTED / ASSUMED / UNKNOWN?
6. ACT or ASK - If ASSUMED/UNKNOWN: ask user first
```

## Memory Capture (MANDATORY)

Capture knowledge when you **DISCOVER** or **CHANGE** something:

| Trigger | What to Capture | Example |
|---------|-----------------|---------|
| DISCOVERY | Validation rules, patterns, gotchas | "Event text requires 10+ chars" |
| CHANGE | New rules, modified behaviour | "Changed min from 0 to 10 chars" |

Use `mcp_memory_create_entities` or `mcp_memory_add_observations`.

## Commands

| Task | Command |
|------|---------|
| All tests | `make test` |
| Single test | `make individual-test TEST="description"` |
| BDD tests | `make bdd` |
| Build | `make build` |
| Lint | `make vet && make staticcheck` |
| Compliance | `make check-compliance` |
| Commit | `make ai-commit FILE=/tmp/commit.txt` |

## Code Style

### Imports (grouped, alphabetical)
```go
import (
    "context"                                          // stdlib
    tea "github.com/charmbracelet/bubbletea"          // external
    "github.com/boodah-consulting/cukes-vhs/internal/cukesvhs" // internal
)
```

### Naming
- Files: `snake_case.go` | Packages: `lowercase`
- Types/Public: `PascalCase` | Private: `camelCase`

### Forbidden
- Comments inside function bodies
- `TODO`, `FIXME`, `HACK` markers
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
