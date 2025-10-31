# Archive Note

## What Happened Here

This archive contains files from two different approaches to building PS-IDE-Go:

### 1. First Approach: Fyne Framework (Archived)
- **Files**: START_HERE.md, PROJECT_SUMMARY.md
- **Status**: Abandoned in favor of GTK3
- **Why**: Needed closer match to Windows ISE appearance and behavior

### 2. Development Artifacts (Archived)
- **Files**: All the .go, .txt fix files
- **Status**: Temporary debugging files
- **Why**: Iterative process of fixing GTK type conversion bugs

## Current Active Approach: GTK3

The **current working code** is in `cmd/ps-ide/` and uses:
- GTK3 for UI (closer to native Windows ISE look)
- VTE for terminal (better PowerShell integration)
- Clean modular structure

## For Next Session

**DO NOT** reference these archived files. Instead:
1. Read `../PROJECT_CONTEXT.md` for complete current state
2. Read `../README.md` for quick overview
3. Work with files in `../cmd/ps-ide/`

All the lessons learned from these archived attempts are documented in PROJECT_CONTEXT.md.
