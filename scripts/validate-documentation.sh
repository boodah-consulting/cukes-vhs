#!/bin/bash
#
# Validate documentation across the codebase
#
# Usage: ./scripts/validate-documentation.sh [--fix] [--verbose]
#
# Options:
#   --fix     Attempt to fix issues automatically
#   --verbose Show detailed output
#

set -euo pipefail

FIX=false
VERBOSE=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --fix)
            FIX=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--fix] [--verbose]"
            exit 1
            ;;
    esac
done

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

ERROR_COUNT=0
WARNING_COUNT=0

log_info() {
    if [[ "$VERBOSE" == "true" ]]; then
        echo -e "${GREEN}[INFO]${NC} $1"
    else
        echo -e "${GREEN}✓${NC} $1"
    fi
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" >&2
    ((WARNING_COUNT++))
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
    ((ERROR_COUNT++))
}

log_section() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

# Check 1: Packages with missing doc.go files
check_missing_doc_go() {
    log_section "Checking for missing doc.go files"
    
    local missing=0
    while IFS= read -r -d '' dir; do
        if [[ ! -f "$dir/doc.go" ]]; then
            log_error "Package $(basename "$dir") missing doc.go file"
            ((missing++))
        else
            log_info "✓ $(basename "$dir") has doc.go"
        fi
    done < <(find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" | \
             xargs dirname | sort -u | while read -r dir; do echo -n "$dir\0"; done)
    
    if [[ "$missing" -eq 0 ]]; then
        log_info "✓ All packages have doc.go files"
    fi
}

# Check 2: Package comment validation
check_package_comments() {
    log_section "Checking package comment quality"
    
    local issues=0
    while IFS= read -r -d '' doc_file; do
        if [[ -f "$doc_file" ]]; then
            local content
            content=$(cat "$doc_file")
            
            # Check for basic structure
            if ! grep -q "^// Package " "$doc_file"; then
                log_error "$(basename "$(dirname "$doc_file")"): Missing package comment header"
                ((issues++))
            fi
            
            if ! grep -q "^package " "$doc_file"; then
                log_error "$(basename "$(dirname "$doc_file")"): Missing package declaration"
                ((issues++))
            fi
            
            # Check for minimum length (should be more than just a header)
            local line_count
            line_count=$(grep -c "^// " "$doc_file" || true)
            if [[ "$line_count" -lt 3 ]]; then
                log_warn "$(basename "$(dirname "$doc_file")"): Package comment too brief"
            fi
            
            log_info "✓ $(basename "$(dirname "$doc_file")"): Package comment OK"
        fi
    done < <(find . -name "doc.go" -not -path "./vendor/*" -not -path "./.git/*" | \
             while read -r file; do echo -n "$file\0"; done)
    
    if [[ "$issues" -eq 0 ]]; then
        log_info "✓ All package comments are valid"
    fi
}

# Check 3: Exported identifier documentation
check_exported_docs() {
    log_section "Checking exported identifier documentation"
    
    local missing_docs=0
    local incomplete_docs=0
    
    while IFS= read -r -d '' go_file; do
        # Skip test files and doc.go
        if [[ "$go_file" =~ _test\.go$ ]] || [[ "$go_file" =~ doc\.go$ ]]; then
            continue
        fi
        
        # Find exported functions, types, variables, constants
        local exported_funcs
        exported_funcs=$(grep -n "^func [A-Z]" "$go_file" | cut -d: -f1 || true)
        
        local exported_types
        exported_types=$(grep -n "^type [A-Z]" "$go_file" | cut -d: -f1 || true)
        
        local exported_vars
        exported_vars=$(grep -n "^var [A-Z]" "$go_file" | cut -d: -f1 || true)
        
        local exported_consts
        exported_consts=$(grep -n "^const [A-Z]" "$go_file" | cut -d: -f1 || true)
        
        # Check each exported identifier
        for line_num in $exported_funcs $exported_types $exported_vars $exported_consts; do
            # Look for documentation comment before this line
            local comment_start=$((line_num - 1))
            local has_comment=false
            local has_complete_doc=false
            
            # Look backwards for comment lines
            while [[ $comment_start -gt 0 ]]; do
                local line_content
                line_content=$(sed -n "${comment_start}p" "$go_file")
                
                if [[ "$line_content" =~ ^[[:space:]]*// ]]; then
                    has_comment=true
                    comment_start=$((comment_start - 1))
                elif [[ "$line_content" =~ ^[[:space:]]*$ ]]; then
                    comment_start=$((comment_start - 1))
                else
                    break
                fi
            done
            
            if [[ "$has_comment" == "false" ]]; then
                log_error "$(basename "$go_file"):$line_num: Missing documentation for exported identifier"
                ((missing_docs++))
            else
                # Check for complete documentation (Expected, Returns, Side effects for functions)
                local func_name
                func_name=$(sed -n "${line_num}p" "$go_file" | sed -n 's/^func \([^(]*\).*/\1/p')
                
                if [[ -n "$func_name" ]]; then
                    local comment_block
                    comment_block=$(sed -n "$((comment_start + 2)),$((line_num - 1))p" "$go_file")
                    
                    if grep -q "Expected:" <<< "$comment_block" && \
                       grep -q "Returns:" <<< "$comment_block" && \
                       grep -q "Side effects:" <<< "$comment_block"; then
                        has_complete_doc=true
                    fi
                    
                    if [[ "$has_complete_doc" == "false" ]]; then
                        log_warn "$(basename "$go_file"):$line_num: Incomplete documentation for function $func_name"
                        ((incomplete_docs++))
                    fi
                fi
            fi
        done
    done < <(find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" | \
             while read -r file; do echo -n "$file\0"; done)
    
    if [[ "$missing_docs" -eq 0 && "$incomplete_docs" -eq 0 ]]; then
        log_info "✓ All exported identifiers have complete documentation"
    else
        log_error "Found $missing_docs missing and $incomplete_docs incomplete documentation items"
    fi
}

# Check 4: Inline comments prohibition
check_inline_comments() {
    log_section "Checking for prohibited inline comments"
    
    local inline_comments=0
    
    while IFS= read -r -d '' go_file; do
        # Skip test files and doc.go
        if [[ "$go_file" =~ _test\.go$ ]] || [[ "$go_file" =~ doc\.go$ ]]; then
            continue
        fi
        
        # Look for inline comments (excluding e2e test files)
        if [[ ! "$go_file" =~ _e2e_test\.go$ ]]; then
            local found_inline
            found_inline=$(grep -n "^[[:space:]]*[^/].*//" "$go_file" | head -5 || true)
            
            if [[ -n "$found_inline" ]]; then
                while IFS= read -r line; do
                    log_warn "$(basename "$go_file"):$line: Inline comment found (should be removed)"
                    ((inline_comments++))
                done <<< "$found_inline"
            fi
        fi
    done < <(find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" | \
             while read -r file; do echo -n "$file\0"; done)
    
    if [[ "$inline_comments" -eq 0 ]]; then
        log_info "✓ No inline comments found"
    else
        log_warn "Found $inline_comments inline comments"
    fi
}

# Auto-fix issues
fix_issues() {
    if [[ "$FIX" != "true" ]]; then
        return
    fi
    
    log_section "Attempting to fix issues"
    
    # Create missing doc.go files
    log_info "Creating missing doc.go files..."
    ./scripts/create-doc-go-files.sh
    
    # Fix documentation blocks
    log_info "Fixing documentation blocks..."
    find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" -not -name "*_test.go" -not -name "doc.go" | \
    while read -r file; do
        if [[ -s "$file" ]]; then
            ./scripts/fix-doc-blocks.sh "$file" > "${file}.tmp" && \
            mv "${file}.tmp" "$file" && \
            log_info "✓ Fixed: $file"
        fi
    done
}

# Run all checks
main() {
    log_section "Documentation Validation"
    
    check_missing_doc_go
    check_package_comments
    check_exported_docs
    check_inline_comments
    
    if [[ "$FIX" == "true" ]]; then
        fix_issues
    fi
    
    log_section "Summary"
    log_info "Errors: $ERROR_COUNT"
    log_info "Warnings: $WARNING_COUNT"
    
    if [[ "$ERROR_COUNT" -gt 0 ]]; then
        echo
        log_error "Documentation validation failed with $ERROR_COUNT errors"
        echo
        if [[ "$FIX" != "true" ]]; then
            echo "Try running with --fix to attempt automatic fixes"
        fi
        exit 1
    else
        log_info "✓ Documentation validation passed"
        exit 0
    fi
}

# Script entry point
main "$@"