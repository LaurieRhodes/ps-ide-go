# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Chroma library integration for professional syntax highlighting
- Support for 150+ programming languages (PowerShell default)
- ChromaSyntaxHighlighter with grammar-based tokenization
- Configurable syntax engine selection (Chroma vs Regex)
- SyntaxHighlighterInterface for engine abstraction
- Enhanced font rendering system for crisp, clear text
- Font rendering module (font_rendering.go) with GTK optimization
- System-level font configuration (hinting, antialiasing, subpixel)
- Comprehensive documentation for Chroma and font rendering
- Syntax highlighting test file (test/syntax_test.ps1)
- **Tab closing features** (tab_events.go):
  - Middle-click on tabs to close (experimental)
  - Ctrl+W keyboard shortcut to close active tab
  - Right-click context menu with Close/Close Others/Close All
  - Save prompts integrate with all closing methods
  - Bulk operations (Close Other Tabs, Close All Tabs)
  - Copy Full Path menu option
- **PowerShell Snippets System** (snippets.go):
  - 18 built-in PowerShell code templates
  - Ctrl+J keyboard shortcut (Windows ISE compatible)
  - Snippet selector dialog with live preview
  - Smart indentation matching
  - Editor right-click context menu
  - Snippets include: cmdlets, functions, classes, loops, error handling, etc.

### Changed
- ScriptTab now uses SyntaxHighlighterInterface instead of concrete type
- CreateSyntaxHighlighter() factory function for engine selection
- Default font size increased from 9pt to 11pt
- Font family priority: Consolas → Liberation Mono → DejaVu Sans Mono
- Binary size increased to 11MB (from 7.6MB) with Chroma
- Pure white background (#FFFFFF) for better contrast
- Enhanced text spacing (2px above/below lines)
- Console text colors brightened for visibility
- Console background fixed to dark blue (#012456)
- Global CSS modified to exclude console from generic textview rules

### Improved
- Text rendering quality matches Windows PowerShell ISE
- More vibrant syntax highlighting colors
- Better font hinting and antialiasing
- Crisp, clear text at all zoom levels
- Reduced eye strain with larger default font
- Console text now bright white on dark blue (Windows ISE style)
- Tab management superior to Windows PowerShell ISE

### Fixed
- Console CSS priority conflicts resolved
- TextTag colors properly applied with correct CSS specificity
- Screen-level CSS properly excludes console textview
- ANSI color codes brightened for visibility

### Maintained
- Original regex-based highlighter available as fallback
- Backward compatibility with existing functionality
- Same performance characteristics for incremental updates
- No breaking changes to API or configuration
- All existing save prompts work with new tab closing features

### TODO
- Implement multiple file tabs support
- Add search and replace functionality
- Implement keyboard shortcuts
- Create integrated PowerShell console
- Add recent files menu
- Implement line numbers display
- Add syntax validation feedback
- Create customizable themes
- Add IntelliSense/autocomplete

## [1.0.0] - 2026-02-05

### Added
- Professional CI/CD pipeline with GitHub Actions
- Automated multi-platform releases (Linux, Windows, macOS)
- Comprehensive linting with golangci-lint
- Pre-publish checks script (lint.sh)
- Enhanced Makefile with lint, test, and CI targets
- README.md with comprehensive documentation
- Cross-platform build support (amd64, arm64)
- Version information in binary via ldflags

### Changed
- Updated build process to include version and build time
- Improved project structure and organization
- Enhanced development tooling and workflows

### Infrastructure
- GitHub Actions CI workflow for testing and linting
- GitHub Actions Release workflow for automated releases
- golangci-lint configuration
- SHA256 checksums for all release binaries

## [0.1.0] - 2025-10-13

### Added
- Initial project creation
- Basic MVP implementation
- Core packages: editor, executor, highlighter, ui, config
- Documentation (README, DEVELOPMENT guide)
- Build and installation scripts
- Sample PowerShell script for testing
