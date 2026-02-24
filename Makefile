# cukes-vhs Makefile
# Build, test, lint, and CI automation for the cukes-vhs CLI tool.

VERSION := $(shell cat VERSION 2>/dev/null || echo "dev")
GO_VERSION := 1.25.4
GOLANGCI_LINT_VERSION := v2.8.0

.DEFAULT_GOAL := help

.PHONY: build test test-race coverage fmt vet staticcheck gosec golangci-lint \
	check-docblocks pre-commit check-compliance ci-local ci-install-tools \
	ai-commit install-git-hooks clean help

build: ## Build the cukes-vhs binary
	go build -ldflags "-X main.version=$(VERSION)" -o cukes-vhs ./cmd/cukes-vhs

test: ## Run tests with Ginkgo
	ginkgo -v --skip-package=testdata ./...

test-race: ## Run tests with race detector
	ginkgo -v --race --skip-package=testdata ./...

coverage: ## Generate test coverage report
	bash scripts/test-coverage.sh

fmt: ## Format Go source files
	gofmt -w .

vet: ## Run go vet
	go vet ./...

staticcheck: ## Run staticcheck linter
	staticcheck ./...

gosec: ## Run gosec security scanner
	gosec ./...

golangci-lint: ## Run golangci-lint
	golangci-lint run ./...

check-docblocks: ## Check docblock compliance in vhsgen
	go build -o /tmp/docblocks ./cmd/docblocks && go vet -vettool=/tmp/docblocks ./internal/vhsgen/...

pre-commit: fmt vet staticcheck test-race ## Run pre-commit checks

check-compliance: ## Run compliance checks
	bash scripts/check-compliance.sh

ci-local: ## Run local CI pipeline
	bash scripts/ci-local.sh

ci-install-tools: ## Install CI toolchain (ginkgo, staticcheck, gosec, golangci-lint)
	go install github.com/onsi/ginkgo/v2/ginkgo@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)

ai-commit: ## Create an AI-attributed commit (FILE=path NO_VERIFY=1 AMEND=1)
	bash scripts/ai-commit.sh "$(FILE)" "$(NO_VERIFY)" "$(AMEND)"

install-git-hooks: ## Install git hooks
	bash scripts/install-git-hooks.sh

clean: ## Remove build artefacts and coverage files
	rm -f cukes-vhs
	rm -f coverage.out coverage.html

help: ## Show this help
	@printf "\n\033[1mcukes-vhs\033[0m — available targets:\n\n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@printf "\n"
