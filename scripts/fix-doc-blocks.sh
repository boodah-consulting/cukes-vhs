#!/usr/bin/env python3
"""
Fix documentation blocks in Go files by adding missing Expected, Returns, and Side effects sections.

Reads from stdin or file, outputs fixed version to stdout.

Usage:
    cat file.go | ./fix-doc-blocks.sh > file_fixed.go
    ./fix-doc-blocks.sh file.go > file_fixed.go
    ./fix-doc-blocks.sh --in-place file.go

The script:
1. Finds exported functions (func Name with capital letter)
2. Checks if they have complete documentation (Expected, Returns, Side effects)
3. Adds missing sections based on function signature
4. Preserves existing documentation
"""

import re
import sys
import argparse
from typing import List, Tuple, Optional


def parse_function_signature(line: str) -> Tuple[Optional[List[str]], Optional[List[str]]]:
    """
    Parse a Go function signature to extract parameters and return values.
    
    Returns:
        Tuple of (parameters, return_values) or (None, None) if not a function
    """
    # Match function declarations: func Name(...) [ReturnType]
    # Handle methods too: func (r *Receiver) Name(...)
    pattern = r'^func\s+(?:\([^)]+\)\s+)?(\w+)\s*\(([^)]*)\)\s*(\S+)?\s*{$'
    match = re.match(pattern, line.strip())
    
    if not match:
        # Try simpler pattern for single return type without braces on same line
        pattern_simple = r'^func\s+(?:\([^)]+\)\s+)?(\w+)\s*\(([^)]*)\)\s*(\S+)\s*$'
        match = re.match(pattern_simple, line.strip())
    
    if not match:
        return None, None
    
    func_name = match.group(1)
    params_str = match.group(2).strip()
    return_str = match.group(3) if match.group(3) else None
    
    # Parse parameters
    params = []
    if params_str:
        # Split by comma, but handle type pairs
        # "name type, name2 type2" -> extract type info
        param_parts = [p.strip() for p in params_str.split(',') if p.strip()]
        for part in param_parts:
            # Skip context.Context as it's common
            if 'context.Context' in part:
                continue
            # Extract just the type part
            tokens = part.split()
            if len(tokens) >= 2:
                params.append(tokens[-1])  # Last token is the type
    
    # Parse return values
    returns = []
    if return_str:
        # Handle multiple returns: (type1, type2)
        if return_str.startswith('(') and return_str.endswith(')'):
            inner = return_str[1:-1]
            returns = [r.strip() for r in inner.split(',') if r.strip()]
        else:
            returns = [return_str]
    
    return params, returns


def has_complete_documentation(comment_lines: List[str]) -> bool:
    """Check if comment already has all required sections."""
    comment_text = '\n'.join(comment_lines)
    has_expected = 'Expected:' in comment_text
    has_returns = 'Returns:' in comment_text
    has_side_effects = 'Side effects:' in comment_text
    
    # Must have all three for complete documentation
    return has_expected and has_returns and has_side_effects


def generate_documentation(func_name: str, params: List[str], returns: List[str], 
                          existing_comment: List[str]) -> List[str]:
    """
    Generate complete documentation for a function.
    
    Preserves the first line of existing comment (the brief description)
    and adds the required sections.
    """
    # Get the brief description from existing comment
    brief = func_name
    if existing_comment:
        first_line = existing_comment[0].strip()
        if first_line.startswith('//'):
            brief = first_line[2:].strip()
            # Remove function name prefix if present
            if brief.startswith(func_name):
                brief = brief[len(func_name):].strip()
            if brief.startswith('creates'):
                brief = f"{func_name} creates{brief[7:]}"
            elif brief.startswith('returns'):
                brief = f"{func_name} returns{brief[7:]}"
            elif brief.startswith('provides'):
                brief = f"{func_name} provides{brief[8:]}"
            elif not brief.startswith(func_name):
                brief = f"{func_name} {brief[0].lower()}{brief[1:]}"
    
    # Build new documentation
    doc_lines = [f"// {brief}"]
    
    # Add Expected section if there are parameters
    if params:
        doc_lines.append("//")
        doc_lines.append("// Expected:")
        for param in params:
            param_name = param.split('.')[-1]  # Get last part of qualified name
            param_name = param_name.replace('*', '')
            param_name = param_name.lower()
            
            # Special handling for common patterns
            if 'theme' in param_name.lower():
                doc_lines.append(f"//   - th must be a valid theme instance (can be nil).")
            elif 'formatter' in param_name.lower():
                doc_lines.append(f"//   - formatter must be a non-nil function.")
            elif 'config' in param_name.lower():
                doc_lines.append(f"//   - config must be a valid configuration object.")
            elif 'string' in param_name.lower():
                doc_lines.append(f"//   - Must be a valid string.")
            else:
                doc_lines.append(f"//   - {param_name} must be valid.")
    
    # Add Returns section if there are return values
    if returns:
        doc_lines.append("//")
        doc_lines.append("// Returns:")
        for ret in returns:
            if ret.startswith('*'):
                doc_lines.append(f"//   - A fully initialized {ret[1:]} ready for use.")
            else:
                doc_lines.append(f"//   - A {ret} value.")
    
    # Always add Side effects section
    doc_lines.append("//")
    doc_lines.append("// Side effects:")
    doc_lines.append("//   - None.")
    
    return doc_lines


def process_go_file(content: str) -> str:
    """Process Go file content and fix documentation."""
    lines = content.split('\n')
    result = []
    i = 0
    
    while i < len(lines):
        line = lines[i]
        
        # Check if this is a comment line
        if line.strip().startswith('//'):
            # Collect all consecutive comment lines
            comment_lines = []
            comment_start = i
            while i < len(lines) and lines[i].strip().startswith('//'):
                comment_lines.append(lines[i])
                i += 1
            
            # Check if next line is a function declaration
            if i < len(lines):
                next_line = lines[i]
                params, returns = parse_function_signature(next_line)
                
                if params is not None and returns is not None:
                    # This is a function declaration after the comment
                    # Extract function name
                    func_match = re.match(r'^func\s+(?:\([^)]+\)\s+)?(\w+)', next_line.strip())
                    if func_match:
                        func_name = func_match.group(1)
                        
                        # Check if it's an exported function (starts with uppercase)
                        if func_name[0].isupper():
                            # Check if documentation is complete
                            if not has_complete_documentation(comment_lines):
                                # Generate new documentation
                                new_doc = generate_documentation(func_name, params, returns, comment_lines)
                                result.extend(new_doc)
                            else:
                                result.extend(comment_lines)
                        else:
                            result.extend(comment_lines)
                else:
                    result.extend(comment_lines)
            else:
                result.extend(comment_lines)
        else:
            result.append(line)
            i += 1
    
    return '\n'.join(result)


def main():
    parser = argparse.ArgumentParser(
        description='Fix documentation blocks in Go files by adding Expected, Returns, and Side effects sections.'
    )
    parser.add_argument('file', nargs='?', help='Go file to process (default: stdin)')
    parser.add_argument('--in-place', '-i', action='store_true', 
                       help='Edit file in place instead of outputting to stdout')
    
    args = parser.parse_args()
    
    # Read input
    if args.file:
        with open(args.file, 'r') as f:
            content = f.read()
    else:
        content = sys.stdin.read()
    
    # Process
    fixed_content = process_go_file(content)
    
    # Output
    if args.in_place and args.file:
        with open(args.file, 'w') as f:
            f.write(fixed_content)
    else:
        print(fixed_content, end='')


if __name__ == '__main__':
    main()
