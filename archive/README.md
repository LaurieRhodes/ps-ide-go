# Archive Directory

This directory contains temporary files and legacy code from the development process. These files are kept for reference but are not part of the active codebase.

## Contents

All files in this directory were intermediate attempts, test files, or temporary code snippets created during debugging and development. They include:

- **Fix files** (`fix*.go`, `*_fix.go`, `*_fixed.go`) - Temporary code attempts to fix various issues
- **Text snippets** (`*.txt`) - Code fragments saved during development
- **Interface fixes** (`vte_interface_fix.go`, etc.) - Multiple attempts to resolve GTK type conversion issues

## Why These Files Exist

During development, we encountered several challenging bugs, particularly with GTK signal handler type conversions. These files represent the iterative process of finding solutions.

## Should You Use These Files?

**NO** - The working, current code is in `cmd/ps-ide/`. These archived files are here only for historical reference and should not be used.

## The Final Working Solution

The GTK signal handler issue was finally resolved by using `interface{}` as the first parameter type in signal handlers. See `cmd/ps-ide/terminal.go` for the correct implementation.

## Can These Be Deleted?

Yes, these files can be safely deleted if you want to clean up the repository. They are not referenced by any active code.
