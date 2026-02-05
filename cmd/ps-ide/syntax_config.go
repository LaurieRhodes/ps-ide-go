package main

import (
	"github.com/gotk3/gotk3/gtk"
)

// Syntax highlighting engine selection
const (
	SyntaxEngineRegex  = "regex"  // Original regex-based highlighter
	SyntaxEngineChroma = "chroma" // Chroma-based highlighter (recommended)
)

// CurrentSyntaxEngine specifies which highlighting engine to use
// Change this to switch between engines:
//   - SyntaxEngineChroma: Uses Chroma library (professional, 150+ languages)
//   - SyntaxEngineRegex:  Uses original regex-based highlighter (lightweight)
var CurrentSyntaxEngine = SyntaxEngineChroma

// SyntaxHighlighterInterface defines the common interface for all highlighters
type SyntaxHighlighterInterface interface {
	Highlight()
	HighlightRange(startLine, endLine int)
	OnBufferChanged(buffer *gtk.TextBuffer)
	UpdateZoom()
}

// CreateSyntaxHighlighter creates the appropriate syntax highlighter based on CurrentSyntaxEngine
func CreateSyntaxHighlighter(buffer *gtk.TextBuffer) SyntaxHighlighterInterface {
	switch CurrentSyntaxEngine {
	case SyntaxEngineChroma:
		return NewChromaSyntaxHighlighter(buffer)
	case SyntaxEngineRegex:
		return NewSyntaxHighlighter(buffer)
	default:
		// Default to Chroma
		return NewChromaSyntaxHighlighter(buffer)
	}
}
