package editor

import (
	"os"
	"path/filepath"
)

// Editor manages the text editing state and file operations
type Editor struct {
	content      string
	filepath     string
	modified     bool
	lineNumbers  bool
	wordWrap     bool
}

// New creates a new Editor instance
func New() *Editor {
	return &Editor{
		content:     "",
		filepath:    "",
		modified:    false,
		lineNumbers: true,
		wordWrap:    false,
	}
}

// GetContent returns the current editor content
func (e *Editor) GetContent() string {
	return e.content
}

// SetContent updates the editor content
func (e *Editor) SetContent(content string) {
	e.content = content
	e.modified = true
}

// GetFilePath returns the current file path
func (e *Editor) GetFilePath() string {
	return e.filepath
}

// IsModified returns whether content has been modified
func (e *Editor) IsModified() bool {
	return e.modified
}

// LoadFile loads content from a file
func (e *Editor) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	e.content = string(data)
	e.filepath = path
	e.modified = false

	return nil
}

// SaveFile saves content to the current file path
func (e *Editor) SaveFile() error {
	if e.filepath == "" {
		return os.ErrInvalid
	}

	return e.SaveFileAs(e.filepath)
}

// SaveFileAs saves content to a specified file path
func (e *Editor) SaveFileAs(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write file
	if err := os.WriteFile(path, []byte(e.content), 0644); err != nil {
		return err
	}

	e.filepath = path
	e.modified = false

	return nil
}

// GetFileName returns just the filename without path
func (e *Editor) GetFileName() string {
	if e.filepath == "" {
		return "Untitled.ps1"
	}
	return filepath.Base(e.filepath)
}

// ToggleLineNumbers toggles line number display
func (e *Editor) ToggleLineNumbers() {
	e.lineNumbers = !e.lineNumbers
}

// ShowLineNumbers returns whether line numbers should be shown
func (e *Editor) ShowLineNumbers() bool {
	return e.lineNumbers
}

// ToggleWordWrap toggles word wrapping
func (e *Editor) ToggleWordWrap() {
	e.wordWrap = !e.wordWrap
}

// IsWordWrap returns whether word wrap is enabled
func (e *Editor) IsWordWrap() bool {
	return e.wordWrap
}

// Clear clears the editor content
func (e *Editor) Clear() {
	e.content = ""
	e.filepath = ""
	e.modified = false
}
