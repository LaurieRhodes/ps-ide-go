package translation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PromptGenerator generates PowerShell-style prompts independently
type PromptGenerator struct {
	template     string
	style        PromptStyle
	remoteHost   string
	customFormat string
}

// NewPromptGenerator creates a new prompt generator
func NewPromptGenerator() *PromptGenerator {
	return &PromptGenerator{
		template: "PS %s> ",
		style:    DefaultPrompt,
	}
}

// Generate creates a prompt string based on current directory
func (pg *PromptGenerator) Generate(currentDir string) string {
	// Simplify home directory to ~
	homeDir, _ := os.UserHomeDir()
	displayPath := currentDir

	if homeDir != "" && strings.HasPrefix(currentDir, homeDir) {
		displayPath = "~" + strings.TrimPrefix(currentDir, homeDir)
	}

	// Use forward slashes for consistency
	displayPath = filepath.ToSlash(displayPath)

	switch pg.style {
	case RemotePrompt:
		return fmt.Sprintf("[%s]: PS %s> ", pg.remoteHost, displayPath)
	case CustomPrompt:
		return pg.formatCustomPrompt(displayPath)
	default:
		return fmt.Sprintf(pg.template, displayPath)
	}
}

// GenerateANSI returns a prompt with ANSI color codes (green like PowerShell)
func (pg *PromptGenerator) GenerateANSI(currentDir string) string {
	prompt := pg.Generate(currentDir)
	// Wrap in green ANSI codes (\x1b[32m for green, \x1b[0m for reset)
	return fmt.Sprintf("\x1b[32m%s\x1b[0m", prompt)
}

// SetStyle changes the prompt style
func (pg *PromptGenerator) SetStyle(style PromptStyle) {
	pg.style = style
}

// SetRemoteHost sets the remote hostname for remote prompt style
func (pg *PromptGenerator) SetRemoteHost(hostname string) {
	pg.remoteHost = hostname
	pg.style = RemotePrompt
}

// SetTemplate changes the prompt template
// Template should contain %s for the directory path
func (pg *PromptGenerator) SetTemplate(template string) {
	if strings.Contains(template, "%s") {
		pg.template = template
	}
}

// SetCustomFormat sets a custom prompt format
func (pg *PromptGenerator) SetCustomFormat(format string) {
	pg.customFormat = format
	pg.style = CustomPrompt
}

// formatCustomPrompt applies custom formatting
func (pg *PromptGenerator) formatCustomPrompt(displayPath string) string {
	if pg.customFormat == "" {
		return pg.Generate(displayPath)
	}

	// Replace placeholders in custom format
	result := pg.customFormat
	result = strings.ReplaceAll(result, "{path}", displayPath)
	result = strings.ReplaceAll(result, "{dir}", filepath.Base(displayPath))

	// If custom format doesn't end with prompt indicator, add it
	if !strings.HasSuffix(result, "> ") && !strings.HasSuffix(result, ">") {
		result += "> "
	}

	return result
}

// GetPlainPrompt returns a prompt without any formatting
func (pg *PromptGenerator) GetPlainPrompt(currentDir string) string {
	return pg.Generate(currentDir)
}

// IsRemoteSession returns true if this is a remote session
func (pg *PromptGenerator) IsRemoteSession() bool {
	return pg.style == RemotePrompt
}

// GetRemoteHost returns the remote hostname (empty if not remote)
func (pg *PromptGenerator) GetRemoteHost() string {
	return pg.remoteHost
}
