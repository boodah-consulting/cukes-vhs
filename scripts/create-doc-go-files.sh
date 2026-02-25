#!/bin/bash
#
# Create doc.go files for packages that are missing them
#
# Usage: ./scripts/create-doc-go-files.sh [--dry-run]
#
# Options:
#   --dry-run  Show what would be created without creating files
#

set -euo pipefail

DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
    DRY_RUN=true
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# Get package name from directory path
get_package_name() {
    local dir="$1"
    basename "$dir"
}

# Check if package already has doc.go
has_doc_go() {
    local dir="$1"
    [[ -f "$dir/doc.go" ]]
}

# Get package brief description from existing files
extract_package_description() {
    local dir="$1"
    local package_name
    package_name=$(get_package_name "$dir")
    
    # Try to extract from existing package comments
    local go_file
    go_file=$(find "$dir" -maxdepth 1 -name "*.go" -not -name "*_test.go" -not -name "doc.go" | head -n1)
    
    if [[ -n "$go_file" ]]; then
        # Look for package comment in first few lines
        local comment
        comment=$(sed -n '1,10p' "$go_file" | grep "^// Package $package_name" | head -n1)
        if [[ -n "$comment" ]]; then
            echo "$comment" | sed "s|^// Package $package_name ||"
            return
        fi
    fi
    
    # Fallback to generic description
    echo "provides $package_name functionality"
}

# Generate doc.go content
generate_doc_go() {
    local dir="$1"
    local package_name
    package_name=$(get_package_name "$dir")
    local description
    description=$(extract_package_description "$dir")
    
    cat << EOF
// Package $package_name $description.
//
// # Overview
//
// This package contains functionality related to $package_name.
// Add more detailed description here as the package evolves.
//
// # Usage Example
//
//	// Add usage example here
//	$package_name.SomeFunction()
//
package $package_name
EOF
}

# Main function
main() {
    local created=0
    local skipped=0
    local failed=0
    
    log_info "Scanning for packages missing doc.go files..."
    
    # Find all Go packages (directories containing .go files)
    while IFS= read -r -d '' dir; do
        if has_doc_go "$dir"; then
            log_info "✓ $(get_package_name "$dir") already has doc.go"
            ((skipped++))
            continue
        fi
        
        local package_name
        package_name=$(get_package_name "$dir")
        
        if [[ "$DRY_RUN" == "true" ]]; then
            log_info "Would create: $dir/doc.go for package '$package_name'"
        else
            local doc_content
            doc_content=$(generate_doc_go "$dir")
            
            if echo "$doc_content" > "$dir/doc.go"; then
                log_info "✓ Created: $dir/doc.go for package '$package_name'"
                ((created++))
            else
                log_error "✗ Failed to create: $dir/doc.go"
                ((failed++))
            fi
        fi
    done < <(find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" | \
             xargs dirname | sort -u | while read -r dir; do echo -n "$dir\0"; done)
    
    echo
    log_info "Summary:"
    log_info "  Created: $created"
    log_info "  Skipped: $skipped"
    log_info "  Failed: $failed"
    
    if [[ "$DRY_RUN" != "true" && "$created" -gt 0 ]]; then
        echo
        log_warn "Remember to:"
        log_warn "1. Review generated doc.go files for accuracy"
        log_warn "2. Add proper descriptions and usage examples"
        log_warn "3. Run 'make check-docblocks' to validate documentation"
    fi
}

# Script entry point
main "$@"