package highlighter

import (
	"bytes"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// Highlighter provides syntax highlighting for PowerShell code
type Highlighter struct {
	style     string
	formatter chroma.Formatter
}

// New creates a new Highlighter with the specified style
func New(styleName string) *Highlighter {
	if styleName == "" {
		styleName = "monokai"
	}

	return &Highlighter{
		style:     styleName,
		formatter: formatters.Get("terminal256"),
	}
}

// Highlight returns syntax-highlighted PowerShell code
func (h *Highlighter) Highlight(code string) (string, error) {
	// Get PowerShell lexer
	lexer := lexers.Get("powershell")
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// Get style
	style := styles.Get(h.style)
	if style == nil {
		style = styles.Fallback
	}

	// Tokenize
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return "", err
	}

	// Format to buffer
	var buf bytes.Buffer
	err = h.formatter.Format(&buf, style, iterator)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GetAvailableStyles returns list of available color schemes
func GetAvailableStyles() []string {
	return styles.Names()
}

// SetStyle changes the color scheme
func (h *Highlighter) SetStyle(styleName string) {
	h.style = styleName
}

// ValidateSyntax performs basic PowerShell syntax validation
func ValidateSyntax(code string) []string {
	var errors []string

	// Basic validation checks
	lines := strings.Split(code, "\n")
	openBraces := 0
	openParens := 0

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Skip comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Count braces
		openBraces += strings.Count(line, "{") - strings.Count(line, "}")
		openParens += strings.Count(line, "(") - strings.Count(line, ")")

		// Check for common errors
		if strings.HasSuffix(trimmed, "|") && !strings.HasSuffix(trimmed, "||") {
			errors = append(errors, "Line %d: Pipeline operator '|' should not end a line", lineNum)
		}
	}

	if openBraces != 0 {
		errors = append(errors, "Unmatched braces: %d unclosed '{'", openBraces)
	}

	if openParens != 0 {
		errors = append(errors, "Unmatched parentheses: %d unclosed '('", openParens)
	}

	return errors
}
