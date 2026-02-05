package main

import (
	"regexp"
	"strings"
	"sync"

	"github.com/gotk3/gotk3/gtk"
)

// PowerShell syntax highlighting colors matching Windows ISE
const (
	ColorKeyword  = "#0000FF" // Blue - keywords and control structures
	ColorString   = "#8B0000" // Dark red/brown - strings
	ColorComment  = "#008000" // Green - comments
	ColorVariable = "#000000" // Black - variables
	ColorOperator = "#808080" // Gray - operators
	ColorNumber   = "#800080" // Purple - numbers
	ColorCmdlet   = "#0000FF" // Blue - cmdlets (Verb-Noun pattern)
	ColorType     = "#008080" // Teal - types [int], [string], etc.
	ColorDefault  = "#000000" // Black - default text
)

// PowerShell keywords
var psKeywords = []string{
	"begin", "break", "catch", "class", "continue", "data", "define", "do",
	"dynamicparam", "else", "elseif", "end", "exit", "filter", "finally",
	"for", "foreach", "from", "function", "if", "in", "param", "process",
	"return", "switch", "throw", "trap", "try", "until", "using", "var",
	"while", "workflow", "parallel", "sequence", "inlinescript",
	"hidden", "static", "enum", "clean", "default",
}

// Comparison and logical operators
var psOperators = []string{
	"-eq", "-ne", "-gt", "-ge", "-lt", "-le",
	"-like", "-notlike", "-match", "-notmatch",
	"-contains", "-notcontains", "-in", "-notin",
	"-replace", "-split", "-join",
	"-is", "-isnot", "-as",
	"-and", "-or", "-not", "-xor",
	"-band", "-bor", "-bnot", "-bxor",
	"-shl", "-shr",
	"-ceq", "-cne", "-cgt", "-cge", "-clt", "-cle",
	"-clike", "-cnotlike", "-cmatch", "-cnotmatch",
	"-ccontains", "-cnotcontains", "-cin", "-cnotin",
	"-creplace", "-csplit",
	"-ieq", "-ine", "-igt", "-ige", "-ilt", "-ile",
	"-ilike", "-inotlike", "-imatch", "-inotmatch",
	"-icontains", "-inotcontains", "-iin", "-inotin",
	"-ireplace", "-isplit",
	"-f", // Format operator
}

// Common PowerShell cmdlet verbs for Verb-Noun pattern recognition
var psVerbs = []string{
	"Add", "Approve", "Assert", "Backup", "Block", "Build", "Checkpoint",
	"Clear", "Close", "Compare", "Complete", "Compress", "Confirm",
	"Connect", "Convert", "ConvertFrom", "ConvertTo", "Copy", "Debug",
	"Deny", "Deploy", "Disable", "Disconnect", "Dismount", "Edit",
	"Enable", "Enter", "Exit", "Expand", "Export", "Find", "Format",
	"Get", "Grant", "Group", "Hide", "Import", "Initialize", "Install",
	"Invoke", "Join", "Limit", "Lock", "Measure", "Merge", "Mount",
	"Move", "New", "Open", "Optimize", "Out", "Ping", "Pop", "Protect",
	"Publish", "Push", "Read", "Receive", "Redo", "Register", "Remove",
	"Rename", "Repair", "Request", "Reset", "Resize", "Resolve",
	"Restart", "Restore", "Resume", "Revoke", "Save", "Search", "Select",
	"Send", "Set", "Show", "Skip", "Split", "Start", "Step", "Stop",
	"Submit", "Suspend", "Switch", "Sync", "Test", "Trace", "Unblock",
	"Undo", "Uninstall", "Unlock", "Unprotect", "Unpublish",
	"Unregister", "Update", "Use", "Wait", "Watch", "Write",
}

// SyntaxHighlighter handles PowerShell syntax highlighting
type SyntaxHighlighter struct {
	buffer *gtk.TextBuffer
	tags   map[string]*gtk.TextTag
	mutex  sync.Mutex

	// Compiled regular expressions for performance
	blockCommentRegex *regexp.Regexp
	lineCommentRegex  *regexp.Regexp
	doubleQuoteRegex  *regexp.Regexp
	singleQuoteRegex  *regexp.Regexp
	hereStringRegex   *regexp.Regexp
	variableRegex     *regexp.Regexp
	cmdletRegex       *regexp.Regexp
	numberRegex       *regexp.Regexp
	typeRegex         *regexp.Regexp
}

// NewSyntaxHighlighter creates a new syntax highlighter
func NewSyntaxHighlighter(buffer *gtk.TextBuffer) *SyntaxHighlighter {
	sh := &SyntaxHighlighter{
		buffer: buffer,
		tags:   make(map[string]*gtk.TextTag),
	}

	// Compile regular expressions once for performance
	sh.blockCommentRegex = regexp.MustCompile(`(?s)<#.*?#>`)
	sh.lineCommentRegex = regexp.MustCompile(`#[^\n]*`)
	sh.doubleQuoteRegex = regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)
	sh.singleQuoteRegex = regexp.MustCompile(`'[^']*'`)
	sh.hereStringRegex = regexp.MustCompile(`(?s)@["'].*?["']@`)
	sh.variableRegex = regexp.MustCompile(`\$(?:[\w]+(?::[\w]+)?|{[^}]+})`)
	sh.cmdletRegex = regexp.MustCompile(`\b(?i:` + strings.Join(psVerbs, "|") + `)-[\w]+\b`)
	sh.numberRegex = regexp.MustCompile(`\b(?:0x[0-9a-fA-F]+|\d+\.?\d*(?:[eE][+-]?\d+)?)\b`)
	sh.typeRegex = regexp.MustCompile(`\[[^\]]+\]`)

	sh.createTags()
	return sh
}

// createTags creates all the text tags for syntax highlighting
func (sh *SyntaxHighlighter) createTags() {
	tagTable, _ := sh.buffer.GetTagTable()

	// Helper function to create a tag with a foreground color
	createTag := func(name, color string) {
		tag := sh.buffer.CreateTag(name, map[string]interface{}{
			"foreground": color,
		})
		sh.tags[name] = tag
	}

	// Remove existing tags if they exist
	existingTags := []string{"keyword", "string", "comment", "variable", "operator", "number", "cmdlet", "type"}
	for _, tagName := range existingTags {
		if tag, err := tagTable.Lookup(tagName); err == nil && tag != nil {
			tagTable.Remove(tag)
		}
	}

	// Create new tags
	createTag("keyword", ColorKeyword)
	createTag("string", ColorString)
	createTag("comment", ColorComment)
	createTag("variable", ColorVariable)
	createTag("operator", ColorOperator)
	createTag("number", ColorNumber)
	createTag("cmdlet", ColorCmdlet)
	createTag("type", ColorType)
}

// TokenType represents the type of token
type TokenType int

const (
	TokenDefault TokenType = iota
	TokenKeyword
	TokenString
	TokenComment
	TokenVariable
	TokenOperator
	TokenNumber
	TokenCmdlet
	TokenTypeAnnotation // Renamed to avoid conflict with TokenType
)

// Token represents a syntax token
type Token struct {
	Type  TokenType
	Start int
	End   int
}

// Highlight performs syntax highlighting on the entire buffer
func (sh *SyntaxHighlighter) Highlight() {
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

	// Tokenize and apply highlighting
	tokens := sh.tokenize(text)
	for _, token := range tokens {
		sh.applyToken(token)
	}
}

// HighlightRange performs syntax highlighting on a specific range
func (sh *SyntaxHighlighter) HighlightRange(startLine, endLine int) {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	// Expand range to include complete tokens
	// Start from beginning of start line
	startIter := sh.buffer.GetIterAtLine(startLine)

	// End at end of end line (or buffer end)
	endIter := sh.buffer.GetIterAtLine(endLine)
	if !endIter.ForwardToLineEnd() {
		endIter = sh.buffer.GetEndIter()
	}

	// Remove tags in range
	sh.buffer.RemoveAllTags(startIter, endIter)

	// Get text for entire buffer (needed for context like multi-line comments)
	bufStart := sh.buffer.GetStartIter()
	bufEnd := sh.buffer.GetEndIter()
	fullText, _ := sh.buffer.GetText(bufStart, bufEnd, false)

	// Tokenize full text but only apply highlighting to visible range
	tokens := sh.tokenize(fullText)

	rangeStart := startIter.GetOffset()
	rangeEnd := endIter.GetOffset()

	for _, token := range tokens {
		// Only apply tokens that overlap with our range
		if token.End > rangeStart && token.Start < rangeEnd {
			sh.applyToken(token)
		}
	}
}

// tokenize breaks text into tokens with their types
func (sh *SyntaxHighlighter) tokenize(text string) []Token {
	var tokens []Token

	// Track regions that are already tokenized (comments, strings)
	occupied := make([]bool, len(text))

	// Helper to mark a region as occupied
	markOccupied := func(start, end int) {
		for i := start; i < end && i < len(occupied); i++ {
			occupied[i] = true
		}
	}

	// 1. Find block comments first (highest priority)
	matches := sh.blockCommentRegex.FindAllStringIndex(text, -1)
	for _, match := range matches {
		tokens = append(tokens, Token{TokenComment, match[0], match[1]})
		markOccupied(match[0], match[1])
	}

	// 2. Find here-strings
	matches = sh.hereStringRegex.FindAllStringIndex(text, -1)
	for _, match := range matches {
		if !occupied[match[0]] {
			tokens = append(tokens, Token{TokenString, match[0], match[1]})
			markOccupied(match[0], match[1])
		}
	}

	// 3. Find double-quoted strings
	matches = sh.doubleQuoteRegex.FindAllStringIndex(text, -1)
	for _, match := range matches {
		if !occupied[match[0]] {
			tokens = append(tokens, Token{TokenString, match[0], match[1]})
			markOccupied(match[0], match[1])
		}
	}

	// 4. Find single-quoted strings
	matches = sh.singleQuoteRegex.FindAllStringIndex(text, -1)
	for _, match := range matches {
		if !occupied[match[0]] {
			tokens = append(tokens, Token{TokenString, match[0], match[1]})
			markOccupied(match[0], match[1])
		}
	}

	// 5. Find line comments
	matches = sh.lineCommentRegex.FindAllStringIndex(text, -1)
	for _, match := range matches {
		if !occupied[match[0]] {
			tokens = append(tokens, Token{TokenComment, match[0], match[1]})
			markOccupied(match[0], match[1])
		}
	}

	// 6. Find types [Type]
	matches = sh.typeRegex.FindAllStringIndex(text, -1)
	for _, match := range matches {
		if !occupied[match[0]] {
			tokens = append(tokens, Token{TokenTypeAnnotation, match[0], match[1]})
			markOccupied(match[0], match[1])
		}
	}

	// 7. Find cmdlets (Verb-Noun pattern)
	matches = sh.cmdletRegex.FindAllStringIndex(text, -1)
	for _, match := range matches {
		if !occupied[match[0]] {
			tokens = append(tokens, Token{TokenCmdlet, match[0], match[1]})
			markOccupied(match[0], match[1])
		}
	}

	// 8. Find variables
	matches = sh.variableRegex.FindAllStringIndex(text, -1)
	for _, match := range matches {
		if !occupied[match[0]] {
			tokens = append(tokens, Token{TokenVariable, match[0], match[1]})
			markOccupied(match[0], match[1])
		}
	}

	// 9. Find numbers
	matches = sh.numberRegex.FindAllStringIndex(text, -1)
	for _, match := range matches {
		if !occupied[match[0]] {
			tokens = append(tokens, Token{TokenNumber, match[0], match[1]})
			markOccupied(match[0], match[1])
		}
	}

	// 10. Find keywords (must check word boundaries)
	for _, keyword := range psKeywords {
		// Case-insensitive word boundary search
		pattern := `(?i)\b` + regexp.QuoteMeta(keyword) + `\b`
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringIndex(text, -1)
		for _, match := range matches {
			if !occupied[match[0]] {
				tokens = append(tokens, Token{TokenKeyword, match[0], match[1]})
				markOccupied(match[0], match[1])
			}
		}
	}

	// 11. Find operators
	for _, operator := range psOperators {
		// Must be preceded by whitespace or start of string
		pattern := `(?:^|\s)` + regexp.QuoteMeta(operator) + `\b`
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringIndex(text, -1)
		for _, match := range matches {
			// Adjust start to exclude leading whitespace
			start := match[0]
			if start < len(text) && (text[start] == ' ' || text[start] == '\t' || text[start] == '\n') {
				start++
			}
			if !occupied[start] {
				tokens = append(tokens, Token{TokenOperator, start, match[1]})
				markOccupied(start, match[1])
			}
		}
	}

	return tokens
}

// applyToken applies a token's highlighting to the buffer
func (sh *SyntaxHighlighter) applyToken(token Token) {
	startIter := sh.buffer.GetIterAtOffset(token.Start)
	endIter := sh.buffer.GetIterAtOffset(token.End)

	var tagName string
	switch token.Type {
	case TokenKeyword:
		tagName = "keyword"
	case TokenString:
		tagName = "string"
	case TokenComment:
		tagName = "comment"
	case TokenVariable:
		tagName = "variable"
	case TokenOperator:
		tagName = "operator"
	case TokenNumber:
		tagName = "number"
	case TokenCmdlet:
		tagName = "cmdlet"
	case TokenTypeAnnotation:
		tagName = "type"
	default:
		return
	}

	if tag, ok := sh.tags[tagName]; ok {
		sh.buffer.ApplyTag(tag, startIter, endIter)
	}
}

// OnBufferChanged is called when the buffer changes
// It performs incremental highlighting of changed lines
func (sh *SyntaxHighlighter) OnBufferChanged(buffer *gtk.TextBuffer) {
	// Get the current cursor position to determine changed line
	insertMark := buffer.GetInsert()
	iter := buffer.GetIterAtMark(insertMark)
	lineNum := iter.GetLine()

	// Highlight a range around the changed line for context
	// (multi-line constructs like comments may span multiple lines)
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
func (sh *SyntaxHighlighter) UpdateZoom() {
	// Recreate tags with new font size (if needed)
	// For now, just re-highlight everything
	sh.Highlight()
}
