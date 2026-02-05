package main

import (
	"sync"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/gotk3/gotk3/gtk"
)

// ChromaSyntaxHighlighter uses Chroma library for tokenization
type ChromaSyntaxHighlighter struct {
	buffer        *gtk.TextBuffer
	tags          map[string]*gtk.TextTag
	mutex         sync.Mutex
	lexer         chroma.Lexer
	fallbackLexer chroma.Lexer
}

// NewChromaSyntaxHighlighter creates a Chroma-based syntax highlighter
func NewChromaSyntaxHighlighter(buffer *gtk.TextBuffer) *ChromaSyntaxHighlighter {
	sh := &ChromaSyntaxHighlighter{
		buffer: buffer,
		tags:   make(map[string]*gtk.TextTag),
	}

	// Get PowerShell lexer
	sh.lexer = lexers.Get("powershell")
	if sh.lexer == nil {
		// Fallback to text lexer if PowerShell not available
		sh.lexer = lexers.Fallback
	}

	// Keep fallback lexer for error cases
	sh.fallbackLexer = lexers.Fallback

	sh.createTags()
	return sh
}

// createTags creates all the text tags for syntax highlighting
func (sh *ChromaSyntaxHighlighter) createTags() {
	tagTable, _ := sh.buffer.GetTagTable()

	// Helper function to create a tag with a foreground color
	createTag := func(name, color string) {
		tag := sh.buffer.CreateTag(name, map[string]interface{}{
			"foreground": color,
		})
		sh.tags[name] = tag
	}

	// Remove existing tags if they exist
	existingTags := []string{"keyword", "string", "comment", "variable", "operator", "number", "cmdlet", "type", "function", "builtin"}
	for _, tagName := range existingTags {
		if tag, err := tagTable.Lookup(tagName); err == nil && tag != nil {
			tagTable.Remove(tag)
		}
	}

	// Create tags matching Windows ISE colors
	createTag("keyword", ColorKeyword)   // #0000FF - Blue
	createTag("string", ColorString)     // #8B0000 - Dark red
	createTag("comment", ColorComment)   // #008000 - Green
	createTag("variable", ColorVariable) // #000000 - Black
	createTag("operator", ColorOperator) // #808080 - Gray
	createTag("number", ColorNumber)     // #800080 - Purple
	createTag("cmdlet", ColorCmdlet)     // #0000FF - Blue
	createTag("type", ColorType)         // #008080 - Teal
	createTag("function", ColorCmdlet)   // #0000FF - Blue (same as cmdlet)
	createTag("builtin", ColorCmdlet)    // #0000FF - Blue (same as cmdlet)
}

// Highlight performs syntax highlighting on the entire buffer using Chroma
func (sh *ChromaSyntaxHighlighter) Highlight() {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	start := sh.buffer.GetStartIter()
	end := sh.buffer.GetEndIter()

	// Remove all existing tags
	sh.buffer.RemoveAllTags(start, end)

	// Get text
	text, _ := sh.buffer.GetText(start, end, false)
	if text == "" {
		return
	}

	// Tokenize with Chroma
	iterator, err := sh.lexer.Tokenise(nil, text)
	if err != nil {
		// If Chroma fails, try fallback lexer
		iterator, err = sh.fallbackLexer.Tokenise(nil, text)
		if err != nil {
			return // Give up if even fallback fails
		}
	}

	// Apply highlighting
	offset := 0
	for token := iterator(); token != chroma.EOF; token = iterator() {
		if token.Value == "" {
			continue
		}

		// Map Chroma token type to GTK tag
		tagName := sh.mapChromaTokenToGTK(token.Type)
		if tagName != "" {
			startIter := sh.buffer.GetIterAtOffset(offset)
			endIter := sh.buffer.GetIterAtOffset(offset + len(token.Value))

			if tag, ok := sh.tags[tagName]; ok {
				sh.buffer.ApplyTag(tag, startIter, endIter)
			}
		}

		offset += len(token.Value)
	}
}

// HighlightRange performs syntax highlighting on a specific range
func (sh *ChromaSyntaxHighlighter) HighlightRange(startLine, endLine int) {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	// For Chroma, we need full context, so highlight entire buffer
	// but only apply to the visible range
	bufStart := sh.buffer.GetStartIter()
	bufEnd := sh.buffer.GetEndIter()
	fullText, _ := sh.buffer.GetText(bufStart, bufEnd, false)

	// Get the range to apply highlighting
	startIter := sh.buffer.GetIterAtLine(startLine)
	endIter := sh.buffer.GetIterAtLine(endLine)
	if !endIter.ForwardToLineEnd() {
		endIter = sh.buffer.GetEndIter()
	}

	// Remove tags in range
	sh.buffer.RemoveAllTags(startIter, endIter)

	// Tokenize entire text for context
	iterator, err := sh.lexer.Tokenise(nil, fullText)
	if err != nil {
		iterator, err = sh.fallbackLexer.Tokenise(nil, fullText)
		if err != nil {
			return
		}
	}

	rangeStart := startIter.GetOffset()
	rangeEnd := endIter.GetOffset()

	// Apply tokens that overlap with our range
	offset := 0
	for token := iterator(); token != chroma.EOF; token = iterator() {
		if token.Value == "" {
			continue
		}

		tokenEnd := offset + len(token.Value)

		// Only apply if token overlaps with visible range
		if tokenEnd > rangeStart && offset < rangeEnd {
			tagName := sh.mapChromaTokenToGTK(token.Type)
			if tagName != "" {
				startPos := offset
				endPos := tokenEnd

				// Clip to visible range
				if startPos < rangeStart {
					startPos = rangeStart
				}
				if endPos > rangeEnd {
					endPos = rangeEnd
				}

				startIter := sh.buffer.GetIterAtOffset(startPos)
				endIter := sh.buffer.GetIterAtOffset(endPos)

				if tag, ok := sh.tags[tagName]; ok {
					sh.buffer.ApplyTag(tag, startIter, endIter)
				}
			}
		}

		offset = tokenEnd
	}
}

// mapChromaTokenToGTK maps Chroma token types to GTK tag names
func (sh *ChromaSyntaxHighlighter) mapChromaTokenToGTK(tokenType chroma.TokenType) string {
	// Chroma token types are hierarchical
	// Check specific types first, then fall back to parent types

	switch tokenType {
	// Keywords
	case chroma.Keyword, chroma.KeywordConstant, chroma.KeywordDeclaration,
		chroma.KeywordNamespace, chroma.KeywordPseudo, chroma.KeywordReserved,
		chroma.KeywordType:
		return "keyword"

	// Strings
	case chroma.String, chroma.StringAffix, chroma.StringBacktick,
		chroma.StringChar, chroma.StringDelimiter, chroma.StringDoc,
		chroma.StringDouble, chroma.StringEscape, chroma.StringHeredoc,
		chroma.StringInterpol, chroma.StringOther, chroma.StringRegex,
		chroma.StringSingle, chroma.StringSymbol:
		return "string"

	// Comments
	case chroma.Comment, chroma.CommentHashbang, chroma.CommentMultiline,
		chroma.CommentSingle, chroma.CommentSpecial, chroma.CommentPreproc:
		return "comment"

	// Variables and Names (includes functions, builtins, cmdlets)
	case chroma.NameVariable, chroma.NameVariableClass,
		chroma.NameVariableGlobal, chroma.NameVariableInstance,
		chroma.NameVariableMagic:
		return "variable"

	// Functions and cmdlets (PowerShell commands)
	case chroma.NameFunction, chroma.NameFunctionMagic,
		chroma.NameBuiltin, chroma.NameBuiltinPseudo:
		return "cmdlet"

	// Types and classes
	case chroma.NameClass:
		return "type"

	// Other names (treat as cmdlets for PowerShell)
	case chroma.Name, chroma.NameAttribute, chroma.NameConstant,
		chroma.NameDecorator, chroma.NameEntity, chroma.NameException,
		chroma.NameLabel, chroma.NameNamespace, chroma.NameOther,
		chroma.NameProperty, chroma.NameTag:
		return "cmdlet"

	// Operators
	case chroma.Operator, chroma.OperatorWord:
		return "operator"

	// Numbers
	case chroma.Number, chroma.NumberBin, chroma.NumberFloat,
		chroma.NumberHex, chroma.NumberInteger, chroma.NumberIntegerLong,
		chroma.NumberOct:
		return "number"

	default:
		// No highlighting for other token types
		return ""
	}
}

// OnBufferChanged is called when the buffer changes (incremental highlighting)
func (sh *ChromaSyntaxHighlighter) OnBufferChanged(buffer *gtk.TextBuffer) {
	// Get the current cursor position to determine changed line
	insertMark := buffer.GetInsert()
	iter := buffer.GetIterAtMark(insertMark)
	lineNum := iter.GetLine()

	// Highlight a range around the changed line for context
	startLine := lineNum - 2
	if startLine < 0 {
		startLine = 0
	}

	endLine := lineNum + 2
	lineCount := buffer.GetLineCount()
	if endLine >= lineCount {
		endLine = lineCount - 1
	}

	sh.HighlightRange(startLine, endLine)
}

// UpdateZoom updates syntax highlighting after zoom changes
func (sh *ChromaSyntaxHighlighter) UpdateZoom() {
	// Recreate tags with new font size (if needed)
	// For now, just re-highlight everything
	sh.Highlight()
}
