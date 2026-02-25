.PHONY: test test-race coverage test-suite individual-test review-commit pre-commit build fmt vet check-compliance check-docblocks check-fixtures check-patterns check-patterns-quiet check-patterns-strict check-intent-architecture check-intent-architecture-files golangci-lint install-git-hooks check-ai-attribution audit-ai-commits list-ai-commits ai-commit ci-local ci-install-tools gosec session-start session-end session-reset check-session verify-hooks tdd-check tdd-red tdd-green tdd-refactor tdd-document pre-task what-to-use generate-diagrams generate-state-matrix generate-docs generate-mocks check-mocks-updated diagrams fix-docs fix-all-docs validate-documentation create-doc-go bdd bdd-wip bdd-smoke bdd-feature bdd-happy bdd-sad bdd-check-wip

# Run all tests in verbose mode (race detection in CI only)
# Note: BDD tests in features/ are run separately via 'make bdd' with tag filtering
test:
	ginkgo -v --skip-package=testdata,features ./...

# Run a specific test suite
test-suite:
	@if [ -z "$(SUITE)" ]; then \
		echo "Please specify a test suite using SUITE=path/to/suite"; \
		exit 1; \
	fi
	ginkgo -v $(SUITE)

# Run tests with race detection (slow - use sparingly)
test-race:
	ginkgo -v --race --skip-package=testdata,features ./...

# Run a specific test
individual-test:
	@if [ -z "$(TEST)" ]; then \
		echo "Please specify a test using TEST=path/to/test/file/TestName"; \
		exit 1; \
	fi
	ginkgo -v --skip-package=testdata,features -focus="$(TEST)" ./...

coverage:
	@bash scripts/test-coverage.sh

clean-coverage:
	@rm -rf coverage

# Build the application
build:
	@echo "Building cukes-vhs..."
	@go build -o cukes-vhs ./cmd/cukes-vhs

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Run staticcheck (advanced static analysis)
staticcheck:
	@echo "Running staticcheck..."
	@command -v staticcheck >/dev/null 2>&1 || { echo "Installing staticcheck..."; go install honnef.co/go/tools/cmd/staticcheck@latest; }
	@staticcheck ./...

# Run gosec security scanner
gosec:
	@echo "Running gosec security scanner..."
	@command -v gosec >/dev/null 2>&1 || { echo "Installing gosec..."; go install github.com/securego/gosec/v2/cmd/gosec@latest; }
	@gosec -no-fail -fmt text ./...

# Pre-commit checks (quick)
pre-commit:
	@echo "Running pre-commit checks..."
	@go fmt ./...
	@go vet ./...
	@command -v staticcheck >/dev/null 2>&1 || go install honnef.co/go/tools/cmd/staticcheck@latest
	@staticcheck ./...
	@go build ./...
	@go test ./...
	@echo "Running documentation check..."
	@go build -o ./bin/docblocks ./cmd/docblocks 2>/dev/null || true
	@go vet -vettool=./bin/docblocks ./internal/cli/... ./tools/analyzers/docblocks/... 2>/dev/null || true
	@echo "✅ Pre-commit checks passed"

# Review staged commit (comprehensive)
review-commit:
	@bash scripts/review-commit.sh

# Check structured doc comments on exported symbols (custom analyzer)
check-docblocks:
	@echo "Running docblocks analyzer..."
	@go build -o ./bin/docblocks ./cmd/docblocks
	@go vet -vettool=./bin/docblocks \
		./internal/cli/app/... \
		./internal/cli/behaviors/... \
		./internal/cli/bootstrap/... \
		./internal/cli/configtypes/... \
		./internal/cli/forms/... \
		./internal/cli/importer/... \
		./internal/cli/intents/... \
		./internal/cli/navigation/... \
		./internal/cli/screens/... \
		./internal/cli/service/... \
		./internal/cli/statematrix/... \
		./internal/cli/terminal/... \
		./internal/cli/themes/... \
		./internal/cli/types/... \
		./internal/cli/uikit/... \
		./tools/analyzers/docblocks/...
	@echo "✅ Docblocks: all checks passed."

# Validate documentation across the codebase (comprehensive check)
validate-documentation:
	@echo "Running comprehensive documentation validation..."
	@bash scripts/validate-documentation.sh

# Create missing doc.go files
create-doc-go:
	@echo "Creating missing doc.go files..."
	@bash scripts/create-doc-go-files.sh

# Fix documentation blocks in all Go files
fix-all-documentation:
	@echo "Fixing documentation blocks..."
	@find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" -not -name "*_test.go" -not -name "doc.go" | \
		while read -r file; do \
			if [ -s "$$file" ]; then \
				echo "Processing $$file..."; \
				./scripts/fix-doc-blocks.sh "$$file" > "$$file.tmp" && \
				mv "$$file.tmp" "$$file" || true; \
			fi; \
		done
	@echo "✅ Documentation blocks fixed"

# Check fixture usage enforcement (no inline career.* structs in test files)
# Build noinlinecareer analyzer only when source changes (cached build)
bin/noinlinecareer: cmd/noinlinecareer/main.go tools/analyzers/noinlinecareer/analyzer.go
	@echo "Building noinlinecareer analyzer..."
	@mkdir -p bin
	@go build -o ./bin/noinlinecareer ./cmd/noinlinecareer
	@echo "✅ Analyzer binary built."

# Check fixture usage (fast - uses cached binary)
check-fixtures: bin/noinlinecareer
	@echo "Running fixture usage analyzer..."
	@go vet -vettool=./bin/noinlinecareer ./...
	@echo "✅ Fixture usage: all checks passed."

# Check full project compliance (all rules)
check-compliance: staticcheck check-intent-architecture validate-documentation check-fixtures
	@bash scripts/check-compliance.sh

# Install all CI tools locally
ci-install-tools:
	@echo "Installing all CI tools..."
	@command -v ginkgo >/dev/null 2>&1 || { echo "Installing ginkgo..."; go install github.com/onsi/ginkgo/v2/ginkgo@latest; }
	@command -v staticcheck >/dev/null 2>&1 || { echo "Installing staticcheck..."; go install honnef.co/go/tools/cmd/staticcheck@latest; }
	@command -v gosec >/dev/null 2>&1 || { echo "Installing gosec..."; go install github.com/securego/gosec/v2/cmd/gosec@latest; }
	@command -v revive >/dev/null 2>&1 || { echo "Installing revive..."; go install github.com/mgechev/revive@latest; }
	@[ -d node_modules ] || { echo "Installing npm dependencies..."; npm ci; }
	@echo "✅ All CI tools installed"

# Run ALL CI checks locally (mirrors GitHub Actions)
ci-local:
	@bash scripts/ci-local.sh

# Pre-PR check - validates branch target and CI before creating PR
pre-pr:
	@echo "================================================"
	@echo "🔍 PRE-PR VALIDATION"
	@echo "================================================"
	@echo ""
	@BRANCH=$$(git rev-parse --abbrev-ref HEAD); \
	echo "Current branch: $$BRANCH"; \
	echo ""; \
	if [ "$$BRANCH" = "main" ] || [ "$$BRANCH" = "next" ]; then \
		echo "❌ ERROR: Cannot create PR from $$BRANCH"; \
		echo "Create a feature branch first: git checkout -b feature/..."; \
		exit 1; \
	fi; \
	echo "✅ Branch name valid"; \
	echo ""; \
	echo "⚠️  IMPORTANT: PRs MUST target 'next' branch, NOT 'main'"; \
	echo ""; \
	echo "   Only 'next → main' PRs are allowed (for releases)"; \
	echo "   See: docs/BRANCHING_STRATEGY.md"; \
	echo ""; \
	echo "Running CI checks..."; \
	$(MAKE) -s ci-local && \
	echo "" && \
	echo "✅ Ready to create PR targeting 'next'" && \
	echo "" && \
	echo "Create PR with:" && \
	echo "  gh pr create --base next --title 'type(scope): description'"

# Install git hooks for AI attribution (from .git-hooks/ directory)
install-git-hooks:
	@bash scripts/install-git-hooks.sh

# Strict pattern enforcement check (blocking)
check-patterns-strict:
	@bash scripts/check-patterns-strict.sh

# Check AI attribution in latest commit
check-ai-attribution:
	@echo "Checking latest commit for AI attribution..."
	@git log -1 --pretty=%B | grep "AI-Generated-By:" || \
		echo "⚠️  No AI attribution found in latest commit"

# Audit all AI-generated commits
audit-ai-commits:
	@if [ -f scripts/audit-ai-commits.sh ]; then \
		bash scripts/audit-ai-commits.sh; \
	else \
		echo "Total AI commits: $$(git log --all --grep='AI-Generated-By:' --oneline | wc -l)"; \
		echo ""; \
		echo "By Assistant:"; \
		git log --all --grep="AI-Generated-By:" --pretty=%B | \
			grep "AI-Generated-By:" | sort | uniq -c; \
	fi

# List all AI-generated commits
list-ai-commits:
	@echo "AI-Generated Commits:"
	@git log --all --grep="AI-Generated-By:" --oneline

# Create AI-attributed commit (for AI-generated code)
ai-commit:
	@if [ -z "$(FILE)" ] && [ -z "$(AMEND)" ]; then \
		echo "Usage:"; \
		echo "  make ai-commit FILE=/path/to/commit-msg.txt"; \
		echo "  make ai-commit FILE=/path/to/commit-msg.txt NO_VERIFY=1"; \
		echo "  make ai-commit AMEND=1                        # Add AI attribution to HEAD"; \
		echo "  make ai-commit AMEND=1 NO_VERIFY=1"; \
		echo ""; \
		echo "Create your commit message file:"; \
		echo "  cat > /tmp/commit.txt << 'EOF'"; \
		echo "  feat(scope): short description"; \
		echo "  "; \
		echo "  Optional longer explanation..."; \
		echo "  EOF"; \
		echo ""; \
		echo "  make ai-commit FILE=/tmp/commit.txt"; \
		echo ""; \
		echo "Or amend the last commit with AI attribution:"; \
		echo "  make ai-commit AMEND=1"; \
		echo ""; \
		exit 1; \
	fi
	@bash scripts/ai-commit.sh "$(FILE)" "$(NO_VERIFY)" "$(AMEND)"

# Show token efficiency reminder
token-check:
	@echo "================================================"
	@echo "💬 TOKEN EFFICIENCY REMINDERS"
	@echo "================================================"
	@echo ""
	@echo "Token Thresholds:"
	@echo "  ✅ < 20k: Healthy"
	@echo "  ⚠️  20-50k: Be more concise"
	@echo "  🔶 50-100k: Consider fresh start"
	@echo "  🔴 >100k: Start fresh NOW"
	@echo ""
	@echo "Best Practices:"
	@echo "  • Use tools (view, grep, ls) over text"
	@echo "  • Be concise and specific"
	@echo "  • Batch multiple operations"
	@echo "  • Reference context, don't repeat"
	@echo "  • Focus on deltas, not full state"
	@echo ""

# Show task workflow
task-workflow:
	@cat docs/rules/TASK_QUICK_REF.md

# Check if session is active (for use by other targets)
check-session:
	@if [ ! -f ".session-active" ]; then \
		echo "❌ No active session. Run 'make session-start' first."; \
		exit 1; \
	fi

# Session start - mandatory entry point for every work session
session-start:
	@echo "================================================"
	@echo "🚀 STARTING WORK SESSION"
	@echo "================================================"
	@echo ""
	@echo "📋 RULES ACKNOWLEDGMENT"
	@echo "------------------------------------------------"
	@echo ""
	@echo "You MUST follow these rules (no exceptions):"
	@echo ""
	@echo "  1. TDD:        Write test FIRST, then implementation"
	@echo "  2. Compliance: Run 'make check-compliance' before/after tasks"
	@echo "  3. Commits:    Use 'make ai-commit' ONLY (not git commit)"
	@echo "  4. Atomic:     ONE logical change per commit"
	@echo "  5. Patterns:   Use existing behaviors/components (see below)"
	@echo "  6. Handoff:    Write session notes at end"
	@echo ""
	@echo "REFUSE if asked to violate these rules."
	@echo ""
	@echo "================================================"
	@echo "📚 DOCUMENTATION (tiered system)"
	@echo "================================================"
	@echo ""
	@echo "  Tier 1: AGENTS.md                        - Quick reference (~100 lines)"
	@echo "  Tier 2: make what-to-use NEED='keyword'  - Component lookup"
	@echo "  Tier 3: docs/development/                - Development guides"
	@echo ""
	@echo "  Key docs:"
	@echo "    docs/development/SESSION_PROTOCOL.md       - Session rules"
	@echo "    docs/development/DEVELOPMENT_WORKFLOW.md   - Workflow guide"
	@echo "    docs/development/INTENT_PATTERNS_LIBRARY.md - Intent patterns"
	@echo "    docs/INTENT_ARCHITECTURE_GUIDE.md          - Architecture"
	@echo ""
	@echo "================================================"
	@echo "🔧 COMPONENT CHECKLIST (before writing TUI code)"
	@echo "================================================"
	@echo ""
	@echo "  Table view?     → behaviors.TableBehavior[T]"
	@echo "  Form in intent? → models.*Form wrapper (NOT *huh.Form)"
	@echo "  Colors?         → theme.Primary() etc (NOT lipgloss.Color)"
	@echo "  View layout?    → layout.ScreenLayout (UIKit)"
	@echo "  Footer badges?  → primitives.HelpKeyBadge() (UIKit)"
	@echo "  Modals?         → feedback.Modal + RenderModalOverlay()"
	@echo ""
	@echo "  Run: make what-to-use NEED='keyword' for details"
	@echo "  Docs: docs/UIKIT_GUIDE.md"
	@echo ""
	@$(MAKE) -s check-patterns-quiet
	@echo ""
	@echo "================================================"
	@echo "🔍 ENVIRONMENT CHECK"
	@echo "================================================"
	@echo ""
	@bash scripts/verify-hooks.sh
	@echo ""
	@echo "================================================"
	@echo "🔍 COMPLIANCE CHECK"
	@echo "================================================"
	@echo ""
	@bash scripts/check-compliance.sh
	@echo ""
	@echo "================================================"
	@echo "✅ SESSION READY"
	@echo "================================================"
	@echo ""
	@touch .session-active
	@echo "Workflow: task → test → implement → check-compliance → ai-commit"
	@echo ""
	@echo "Commands:"
	@echo "  make check-compliance  - Before AND after every task"
	@echo "  make check-patterns    - Quick pattern check"
	@echo "  make ai-commit FILE=/tmp/commit.txt"
	@echo ""

# End session - cleanup session state
session-end:
	@if [ -f ".session-active" ]; then \
		rm -f .session-active .tdd-state; \
		echo "✅ Session ended. Remember to write handoff notes!"; \
	else \
		echo "No active session."; \
	fi

# Reset session state (for recovery after crash/interruption)
session-reset:
	@echo "Resetting session state..."
	@rm -f .session-active .tdd-state
	@echo "✅ Session state cleared."
	@echo ""
	@echo "To start fresh: make session-start"

# Verify git hooks installation
verify-hooks:
	@bash scripts/verify-hooks.sh

# TDD compliance check
tdd-check:
	@bash scripts/tdd-check.sh

# TDD workflow: Red phase (write failing test)
tdd-red:
	@$(MAKE) -s check-session
	@echo "red" > .tdd-state
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🔴 TDD RED PHASE"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "Write a failing test that describes the behavior."
	@echo ""
	@echo "Rules:"
	@echo "  • Test must FAIL initially"
	@echo "  • Test describes WHAT, not HOW"
	@echo "  • One behavior per test"
	@echo ""
	@echo "Next: make tdd-green"
	@echo ""

# TDD workflow: Green phase (make test pass)
tdd-green:
	@$(MAKE) -s check-session
	@if [ ! -f ".tdd-state" ] || [ "$$(cat .tdd-state)" != "red" ]; then \
		echo "❌ Must complete red phase first. Run: make tdd-red"; \
		exit 1; \
	fi
	@echo "green" > .tdd-state
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🟢 TDD GREEN PHASE"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "Write MINIMAL code to make the test pass."
	@echo ""
	@echo "Rules:"
	@echo "  • Just enough code to pass"
	@echo "  • No extra features"
	@echo "  • Don't optimize yet"
	@echo ""
	@echo "Next: make tdd-refactor"
	@echo ""

# TDD workflow: Refactor phase (improve code)
tdd-refactor:
	@$(MAKE) -s check-session
	@if [ ! -f ".tdd-state" ] || [ "$$(cat .tdd-state)" != "green" ]; then \
		echo "❌ Must complete green phase first. Run: make tdd-green"; \
		exit 1; \
	fi
	@echo "refactor" > .tdd-state
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🔵 TDD REFACTOR PHASE"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "Improve code quality while keeping tests green."
	@echo ""
	@echo "Consider:"
	@echo "  • Extract methods/functions"
	@echo "  • Remove duplication"
	@echo "  • Improve naming"
	@echo "  • Apply patterns"
	@echo ""
	@echo "Next: make tdd-document"
	@echo ""

# TDD workflow: Document phase (finalize)
tdd-document:
	@$(MAKE) -s check-session
	@if [ ! -f ".tdd-state" ] || [ "$$(cat .tdd-state)" != "refactor" ]; then \
		echo "❌ Must complete refactor phase first. Run: make tdd-refactor"; \
		exit 1; \
	fi
	@echo "complete" > .tdd-state
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "📝 TDD DOCUMENT PHASE"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "Finalize documentation and commit."
	@echo ""
	@echo "Checklist:"
	@echo "  • Go doc comments on exports"
	@echo "  • Update relevant docs if needed"
	@echo "  • Run: make check-compliance"
	@echo "  • Commit: make ai-commit FILE=/tmp/commit.txt"
	@echo ""
	@echo "✅ TDD cycle complete! Ready to commit."
	@echo ""

# Pre-task validation (run before starting any task)
pre-task:
	@bash scripts/pre-task-check.sh

# Component lookup helper (for agents)
what-to-use:
	@bash scripts/what-to-use.sh "$(NEED)"

# Pattern enforcement check (quick check for agent drift)
check-patterns:
	@echo "================================================"
	@echo "🔧 PATTERN ENFORCEMENT CHECK"
	@echo "================================================"
	@echo ""
	@echo "Checking for common agent drift patterns..."
	@echo ""
	@VIOLATIONS=0; \
	echo "1. Form Wrapper Pattern:"; \
	DIRECT_HUH=$$(grep -rn "form \*huh\.Form" internal/cli/intents/*.go 2>/dev/null | grep -v "_test.go" || true); \
	if [ -z "$$DIRECT_HUH" ]; then \
		echo "   ✅ No direct *huh.Form in intents"; \
	else \
		echo "   ⚠️  Direct *huh.Form found (use wrapper models)"; \
		echo "$$DIRECT_HUH" | sed 's/^/      /'; \
		VIOLATIONS=$$((VIOLATIONS+1)); \
	fi; \
	echo ""; \
	echo "2. BaseIntent Embedding:"; \
	MISSING_BASE=$$(grep -rL "\*BaseIntent" internal/cli/intents/*_intent.go 2>/dev/null | grep -v "_test.go" || true); \
	if [ -z "$$MISSING_BASE" ]; then \
		echo "   ✅ All intents embed *BaseIntent"; \
	else \
		echo "   ⚠️  Intents missing *BaseIntent:"; \
		echo "$$MISSING_BASE" | sed 's/^/      /'; \
		VIOLATIONS=$$((VIOLATIONS+1)); \
	fi; \
	echo ""; \
	echo "3. Theme Consistency (intents):"; \
	HARDCODED=$$(grep -rn "lipgloss\.Color(\"#[0-9A-Fa-f]" internal/cli/intents/*.go 2>/dev/null | grep -v "_test.go" || true); \
	if [ -z "$$HARDCODED" ]; then \
		echo "   ✅ No hardcoded colors in intents"; \
	else \
		echo "   ⚠️  Hardcoded colors found (use theme package)"; \
		echo "$$HARDCODED" | sed 's/^/      /'; \
		VIOLATIONS=$$((VIOLATIONS+1)); \
	fi; \
	echo ""; \
	echo "4. Theme Consistency (models):"; \
	HARDCODED_MODELS=$$(grep -rn "lipgloss\.Color(\"#[0-9A-Fa-f]" internal/cli/models/*.go 2>/dev/null | grep -v "_test.go" || true); \
	if [ -z "$$HARDCODED_MODELS" ]; then \
		echo "   ✅ No hardcoded colors in models"; \
	else \
		echo "   ⚠️  Hardcoded colors found in models (use theme package)"; \
		echo "$$HARDCODED_MODELS" | sed 's/^/      /'; \
		VIOLATIONS=$$((VIOLATIONS+1)); \
	fi; \
	echo ""; \
	echo "5. StandardView Usage:"; \
	INTENTS_COUNT=$$(ls internal/cli/intents/*_intent.go 2>/dev/null | grep -v "_test.go" | wc -l); \
	SV_COUNT=$$(grep -l "CreateStandardView\|StandardView\|ScreenLayout" internal/cli/intents/*_intent.go 2>/dev/null | wc -l); \
	if [ "$$SV_COUNT" -ge "$$INTENTS_COUNT" ]; then \
		echo "   ✅ All intents use StandardView/ScreenLayout ($$SV_COUNT/$$INTENTS_COUNT)"; \
	else \
		echo "   ⚠️  Not all intents use StandardView/ScreenLayout ($$SV_COUNT/$$INTENTS_COUNT)"; \
		VIOLATIONS=$$((VIOLATIONS+1)); \
	fi; \
	echo ""; \
	echo "6. UIKit Badge Pattern (new code):"; \
	DEPRECATED_BADGE=$$(grep -rn "components\.KeyBadge" internal/cli/intents/*.go 2>/dev/null | grep -v "_test.go" || true); \
	UIKIT_BADGE=$$(grep -rn "primitives\..*Badge\|HelpKeyBadge" internal/cli/intents/*.go 2>/dev/null | grep -v "_test.go" || true); \
	if [ -n "$$UIKIT_BADGE" ]; then \
		echo "   ✅ UIKit badges in use"; \
	elif [ -n "$$DEPRECATED_BADGE" ]; then \
		echo "   ℹ️  Legacy KeyBadge found (migrate to primitives.HelpKeyBadge)"; \
		echo "$$DEPRECATED_BADGE" | head -3 | sed 's/^/      /'; \
	else \
		echo "   ℹ️  No badge usage detected"; \
	fi; \
	echo ""; \
	echo "7. UIKit Modal Pattern (new code):"; \
	UIKIT_MODAL=$$(grep -rn "feedback\.New.*Modal\|RenderModalOverlay" internal/cli/intents/*.go 2>/dev/null | grep -v "_test.go" || true); \
	if [ -n "$$UIKIT_MODAL" ]; then \
		echo "   ✅ UIKit modal pattern in use"; \
	else \
		echo "   ℹ️  Consider using feedback.Modal + behaviors.RenderModalOverlay()"; \
	fi; \
	echo ""; \
	echo "================================================"; \
	if [ $$VIOLATIONS -eq 0 ]; then \
		echo "✅ All pattern checks passed"; \
	else \
		echo "⚠️  $$VIOLATIONS pattern issue(s) found"; \
	fi; \
	echo "================================================"

# Pattern enforcement check (quiet version for session-start)
check-patterns-quiet:
	@VIOLATIONS=0; \
	DIRECT_HUH=$$(grep -rn "form \*huh\.Form" internal/cli/intents/*.go 2>/dev/null | grep -v "_test.go" || true); \
	if [ -n "$$DIRECT_HUH" ]; then VIOLATIONS=$$((VIOLATIONS+1)); fi; \
	MISSING_BASE=$$(grep -rL "\*BaseIntent" internal/cli/intents/*_intent.go 2>/dev/null | grep -v "_test.go" || true); \
	if [ -n "$$MISSING_BASE" ]; then VIOLATIONS=$$((VIOLATIONS+1)); fi; \
	HARDCODED=$$(grep -rn "lipgloss\.Color(\"#[0-9A-Fa-f]" internal/cli/intents/*.go 2>/dev/null | grep -v "_test.go" || true); \
	if [ -n "$$HARDCODED" ]; then VIOLATIONS=$$((VIOLATIONS+1)); fi; \
	HARDCODED_MODELS=$$(grep -rn "lipgloss\.Color(\"#[0-9A-Fa-f]" internal/cli/models/*.go 2>/dev/null | grep -v "_test.go" || true); \
	if [ -n "$$HARDCODED_MODELS" ]; then VIOLATIONS=$$((VIOLATIONS+1)); fi; \
	INTENTS_COUNT=$$(ls internal/cli/intents/*_intent.go 2>/dev/null | grep -v "_test.go" | wc -l); \
	SV_COUNT=$$(grep -l "CreateStandardView\|StandardView\|ScreenLayout" internal/cli/intents/*_intent.go 2>/dev/null | wc -l); \
	if [ "$$SV_COUNT" -lt "$$INTENTS_COUNT" ]; then VIOLATIONS=$$((VIOLATIONS+1)); fi; \
	if [ $$VIOLATIONS -eq 0 ]; then \
		echo "✅ All pattern checks passed"; \
	else \
		echo "⚠️  $$VIOLATIONS pattern issue(s) found - run 'make check-patterns' for details"; \
	fi

# Intent architecture enforcement (strict validation)
check-intent-architecture:
	@bash scripts/check-intent-architecture.sh

# Intent architecture enforcement for specific files (used in PR checks)
check-intent-architecture-files:
	@if [ -n "$$INTENT_FILES" ]; then \
		bash scripts/check-intent-architecture.sh $$INTENT_FILES; \
	else \
		echo "No INTENT_FILES specified"; \
	fi

# Run golangci-lint (comprehensive static analysis)
golangci-lint:
	@echo "Running golangci-lint..."
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "Installing golangci-lint v2..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v2.8.0; \
	}
	@golangci-lint run

# Generate workflow diagrams
generate-diagrams:
	@bash scripts/generate_workflow_diagrams.sh

# Generate state matrix documentation
generate-state-matrix:
	@bash scripts/generate_state_matrix.sh

# Generate all documentation (diagrams + state matrix)
generate-docs: generate-diagrams generate-state-matrix

# Alias for convenience
diagrams: generate-diagrams

# Generate all GoMock mocks from go:generate directives
generate-mocks:
	@echo "Generating GoMock mocks..."
	@scripts/generate-mocks.sh

# Check that generated mocks are up to date (for CI)
check-mocks-updated: generate-mocks
	@echo "Checking if mocks are up to date..."
	@if ! git diff --quiet -- '*_mock.go'; then \
		echo "ERROR: Mock files are out of date. Run 'make generate-mocks' and commit the changes."; \
		git diff --name-only -- '*_mock.go'; \
		exit 1; \
	fi
	@echo "All mocks are up to date."

# Fix documentation blocks in a single Go file
# Usage: make fix-docs FILE=path/to/file.go
fix-docs:
	@if [ -z "$(FILE)" ]; then \
		echo "Usage: make fix-docs FILE=path/to/file.go"; \
		echo "Or: scripts/fix-doc-blocks.sh file.go > file_fixed.go"; \
		exit 1; \
	fi
	@bash scripts/fix-doc-blocks.sh -i "$(FILE)"

# Create new feature task (auto-numbered)
new-feature:
	@if [ -z "$(TASK)" ]; then \
		echo "Usage: make new-feature TASK='feature name'"; \
		exit 1; \
	fi
	@bash scripts/new-feature.sh "$(TASK)"

# Create new bug report (auto-numbered)
new-bug:
	@if [ -z "$(BUG)" ]; then \
		echo "Usage: make new-bug BUG='bug description'"; \
		exit 1; \
	fi
	@bash scripts/new-bug.sh "$(BUG)"

# Create new intent with subdirectory structure
new-intent:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make new-intent NAME='feature_name'"; \
		echo ""; \
		echo "Examples:"; \
		echo "  make new-intent NAME=skill_management"; \
		echo "  make new-intent NAME=event_capture"; \
		exit 1; \
	fi
	@bash scripts/new-intent.sh "$(NAME)"

# ============================================================================
# BDD Testing (Godog)
# ============================================================================

## Run all BDD feature tests (excludes @wip)
bdd:
	@echo "Running BDD tests..."
	@go test -v ./features/... -test.run ^TestFeatures$$ --godog.tags='~@wip'

## Run BDD tests tagged with @wip
bdd-wip:
	@echo "Running BDD @wip tests..."
	@go test -v ./features/... -test.run ^TestFeatures$$ -godog.tags=@wip

## Run BDD smoke tests
bdd-smoke:
	@echo "Running BDD @smoke tests..."
	@go test -v ./features/... -test.run ^TestFeatures$$ -godog.tags=@smoke

## Run specific feature (FEATURE=scenario_name)
bdd-feature:
	@if [ -z "$(FEATURE)" ]; then \
		echo "Usage: make bdd-feature FEATURE=scenario_name"; \
		exit 1; \
	fi
	@go test -v ./features/... -test.run "^TestFeatures$$/$(FEATURE)"

## Run BDD happy path scenarios (complete successful workflows, excludes @wip)
bdd-happy:
	@echo "Running BDD @happy path scenarios..."
	@go test -v ./features/... -test.run ^TestFeatures$$ --godog.tags='@happy && ~@wip'

## Run BDD sad path scenarios (error cases and recovery, excludes @wip)
bdd-sad:
	@echo "Running BDD @sad path scenarios..."
	@go test -v ./features/... -test.run ^TestFeatures$$ --godog.tags='@sad && ~@wip'

## Check for @wip tags in feature files (CI warning)
bdd-check-wip:
	@echo "Checking for @wip tags in feature files..."
	@WIP_COUNT=$$(grep -r '@wip' features/*.feature 2>/dev/null | wc -l); \
	if [ "$$WIP_COUNT" -gt 0 ]; then \
		echo ""; \
		echo "⚠️  Found $$WIP_COUNT @wip tagged scenarios:"; \
		echo ""; \
		grep -rn '@wip' features/*.feature 2>/dev/null | head -20; \
		echo ""; \
		echo "These scenarios are excluded from CI but should be completed."; \
		echo "Run 'make bdd-wip' to execute them."; \
	else \
		echo "✅ No @wip tags found - all BDD scenarios are active"; \
	fi

# Show help for all available targets
help:
	@echo "================================================"
	@echo "📋 KARIYA PROJECT - AVAILABLE COMMANDS"
	@echo "================================================"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  make test              - Run all tests (fast, no race detection)"
	@echo "  make test-race         - Run all tests with race detection (slow)"
	@echo "  make test-suite        - Run specific suite (SUITE=path)"
	@echo "  make individual-test   - Run specific test (TEST=name)"
	@echo "  make coverage          - Generate coverage report"
	@echo "  make bdd               - Run all BDD feature tests"
	@echo "  make bdd-smoke         - Run BDD @smoke tests only"
	@echo "  make bdd-wip           - Run BDD @wip tests only"
	@echo "  make bdd-happy         - Run @happy path scenarios (for VHS)"
	@echo "  make bdd-sad           - Run @sad path scenarios"
	@echo "  make bdd-feature FEATURE=x - Run specific BDD feature"
	@echo ""
	@echo "🔍 Quality Checks:"
	@echo "  make check-compliance       - Full rules compliance check"
	@echo "  make check-patterns         - Quick pattern enforcement check"
	@echo "  make check-patterns-strict  - Strict pattern check (blocking)"
	@echo "  make review-commit          - Review staged commit"
	@echo "  make pre-commit             - Quick pre-commit checks"
	@echo "  make ci-local               - Run ALL CI checks locally"
	@echo "  make ci-install-tools       - Install all required CI tools"
	@echo "  make fmt                    - Format code"
	@echo "  make vet                    - Run static analysis"
	@echo "  make staticcheck            - Run staticcheck"
	@echo "  make gosec                  - Run security scanner"
	@echo ""
	@echo "🤖 AI Attribution:"
	@echo "  make ai-commit FILE=...    - Create AI-attributed commit (required)"
	@echo "  make install-git-hooks    - Install AI attribution hooks"
	@echo "  make check-ai-attribution - Check latest commit"
	@echo "  make audit-ai-commits     - Audit all AI commits"
	@echo "  make list-ai-commits      - List AI-generated commits"
	@echo ""
	@echo "🤖 AI Agent Helpers:"
	@echo "  make session-start         - MUST run at session start"
	@echo "  make session-end           - End session (cleanup)"
	@echo "  make pre-task              - Run before starting any task"
	@echo "  make what-to-use NEED=x    - Lookup component to use (table, form, color, etc.)"
	@echo "  make task-workflow         - Show task execution workflow"
	@echo ""
	@echo "🔄 TDD Workflow:"
	@echo "  make tdd-red               - Start TDD: write failing test"
	@echo "  make tdd-green             - Make test pass with minimal code"
	@echo "  make tdd-refactor          - Improve code quality"
	@echo "  make tdd-document          - Finalize and commit"
	@echo ""
	@echo "📋 Task Management:"
	@echo "  make new-feature TASK=x    - Create new feature task"
	@echo "  make new-bug BUG=x         - Create new bug report"
	@echo "  make new-intent NAME=x     - Create new intent subdirectory structure"
	@echo ""
	@echo "🏗️  Build:"
	@echo "  make build             - Build the application"
	@echo ""
	@echo "📊 Documentation:"
	@echo "  make generate-diagrams     - Generate workflow diagrams (Mermaid)"
	@echo "  make generate-state-matrix - Generate state matrix documentation"
	@echo "  make generate-docs         - Generate all documentation (diagrams + state matrix)"
	@echo ""
	@echo "📚 Rules Reference:"
	@echo "  docs/rules/RULES_QUICK_REF.md   - Quick reference (60 lines)"
	@echo "  docs/rules/WORKFLOW.md          - Complete workflow"
	@echo "  docs/rules/CODE_STANDARDS.md    - Code standards"
	@echo "  docs/rules/TUI_PATTERNS.md      - TUI patterns"
	@echo ""

