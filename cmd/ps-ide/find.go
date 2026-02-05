package main

import (
	"regexp"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

// FindDialog represents the find dialog window
type FindDialog struct {
	window         *gtk.Window
	searchEntry    *gtk.Entry
	matchCaseCheck *gtk.CheckButton
	wholeWordCheck *gtk.CheckButton
	regexCheck     *gtk.CheckButton
	searchUpCheck  *gtk.CheckButton
	findNextButton *gtk.Button
	statusLabel    *gtk.Label
	lastSearchText string
	lastSearchPos  int
}

var (
	findDialog *FindDialog
	// Remember dialog position
	findDialogX int = -1
	findDialogY int = -1
)

// createFindDialog creates and initializes the find dialog
func createFindDialog() *FindDialog {
	fd := &FindDialog{}

	// Create window
	window, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		return nil
	}
	fd.window = window

	window.SetTitle("Find")
	window.SetDefaultSize(380, 180)
	window.SetResizable(false)
	window.SetTransientFor(mainWindow) // Stay above main window
	window.SetKeepAbove(true)          // Always on top
	window.SetTypeHint(gdk.WINDOW_TYPE_HINT_DIALOG)
	window.SetDecorated(true) // Has title bar (draggable)

	// Prevent destruction on close - just hide
	window.Connect("delete-event", func() bool {
		window.Hide()
		return true // Prevent destruction
	})

	// Remember position when moved
	window.Connect("configure-event", func() {
		x, y := window.GetPosition()
		findDialogX = x
		findDialogY = y
	})

	// Handle keyboard shortcuts
	window.Connect("key-press-event", func(_ *gtk.Window, event *gdk.Event) bool {
		keyEvent := gdk.EventKeyNewFromEvent(event)
		keyval := keyEvent.KeyVal()

		if keyval == gdk.KEY_Escape {
			window.Hide()
			return true
		}
		if keyval == gdk.KEY_Return || keyval == gdk.KEY_KP_Enter {
			fd.findNext()
			return true
		}
		return false
	})

	// Main container
	mainBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	mainBox.SetMarginTop(10)
	mainBox.SetMarginBottom(10)
	mainBox.SetMarginStart(10)
	mainBox.SetMarginEnd(10)
	window.Add(mainBox)

	// Search entry row
	searchBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	mainBox.PackStart(searchBox, false, false, 0)

	searchLabel, _ := gtk.LabelNew("Find what:")
	searchBox.PackStart(searchLabel, false, false, 0)

	fd.searchEntry, _ = gtk.EntryNew()
	fd.searchEntry.SetHExpand(true)
	searchBox.PackStart(fd.searchEntry, true, true, 0)

	// Trigger search on entry change
	fd.searchEntry.Connect("changed", func() {
		// Reset search position when search text changes
		fd.lastSearchPos = -1
	})

	// Options grid
	optionsBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)
	mainBox.PackStart(optionsBox, false, false, 0)

	// Row 1: Match case, Whole word
	row1, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 15)
	optionsBox.PackStart(row1, false, false, 0)

	fd.matchCaseCheck, _ = gtk.CheckButtonNewWithLabel("Match case")
	row1.PackStart(fd.matchCaseCheck, false, false, 0)

	fd.wholeWordCheck, _ = gtk.CheckButtonNewWithLabel("Whole word")
	row1.PackStart(fd.wholeWordCheck, false, false, 0)

	// Row 2: Regular expressions
	row2, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 15)
	optionsBox.PackStart(row2, false, false, 0)

	fd.regexCheck, _ = gtk.CheckButtonNewWithLabel("Regular expressions")
	row2.PackStart(fd.regexCheck, false, false, 0)

	// Row 3: Search up
	row3, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 15)
	optionsBox.PackStart(row3, false, false, 0)

	fd.searchUpCheck, _ = gtk.CheckButtonNewWithLabel("Search up")
	row3.PackStart(fd.searchUpCheck, false, false, 0)

	// Status label
	fd.statusLabel, _ = gtk.LabelNew("")
	fd.statusLabel.SetXAlign(0.0) // Left align
	mainBox.PackStart(fd.statusLabel, false, false, 0)

	// Buttons
	buttonBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	buttonBox.SetHAlign(gtk.ALIGN_END)
	mainBox.PackStart(buttonBox, false, false, 0)

	fd.findNextButton, _ = gtk.ButtonNewWithLabel("Find Next")
	fd.findNextButton.SetSizeRequest(100, -1)
	fd.findNextButton.Connect("clicked", func() {
		fd.findNext()
	})
	buttonBox.PackStart(fd.findNextButton, false, false, 0)

	cancelButton, _ := gtk.ButtonNewWithLabel("Cancel")
	cancelButton.SetSizeRequest(100, -1)
	cancelButton.Connect("clicked", func() {
		window.Hide()
	})
	buttonBox.PackStart(cancelButton, false, false, 0)

	return fd
}

// showFindDialog shows the find dialog
func showFindDialog() {
	if findDialog == nil {
		findDialog = createFindDialog()
		if findDialog == nil {
			return
		}
	}

	// Pre-fill with selected text
	tab := getCurrentTab()
	if tab != nil {
		start, end, hasSelection := tab.buffer.GetSelectionBounds()
		if hasSelection {
			text, _ := tab.buffer.GetText(start, end, false)
			// Only pre-fill if text is reasonable length (single line)
			if len(text) < 100 && !strings.Contains(text, "\n") {
				findDialog.searchEntry.SetText(text)
				findDialog.searchEntry.SelectRegion(0, -1)
			}
		}
	}

	// Clear status
	findDialog.statusLabel.SetText("")

	// Restore position if previously set
	if findDialogX >= 0 && findDialogY >= 0 {
		findDialog.window.Move(findDialogX, findDialogY)
	}

	findDialog.window.ShowAll()
	findDialog.searchEntry.GrabFocus()
}

// findNext performs the search
func (fd *FindDialog) findNext() {
	searchText, _ := fd.searchEntry.GetText()
	if searchText == "" {
		fd.statusLabel.SetText("Please enter search text")
		return
	}

	tab := getCurrentTab()
	if tab == nil {
		fd.statusLabel.SetText("No active document")
		return
	}

	matchCase := fd.matchCaseCheck.GetActive()
	wholeWord := fd.wholeWordCheck.GetActive()
	isRegex := fd.regexCheck.GetActive()
	searchUp := fd.searchUpCheck.GetActive()

	// Get current cursor position or last search position
	var startOffset int
	if fd.lastSearchText == searchText && fd.lastSearchPos >= 0 {
		// Continue from last position
		startOffset = fd.lastSearchPos
	} else {
		// Start from current cursor
		cursor := tab.buffer.GetInsert()
		cursorIter := tab.buffer.GetIterAtMark(cursor)
		startOffset = cursorIter.GetOffset()
		fd.lastSearchText = searchText
	}

	// Perform search
	matchStart, matchEnd, found, wrapped := findInBuffer(
		tab.buffer,
		searchText,
		matchCase,
		wholeWord,
		isRegex,
		searchUp,
		startOffset,
	)

	if found {
		// Highlight found text
		highlightFoundText(tab.buffer, tab.textView, matchStart, matchEnd)

		// Update last search position for next search
		if searchUp {
			fd.lastSearchPos = matchStart.GetOffset()
		} else {
			fd.lastSearchPos = matchEnd.GetOffset()
		}

		// Show wrapped message if applicable
		if wrapped {
			if searchUp {
				fd.statusLabel.SetText("Search wrapped to end")
			} else {
				fd.statusLabel.SetText("Search wrapped to beginning")
			}
		} else {
			fd.statusLabel.SetText("")
		}
	} else {
		fd.statusLabel.SetText("Text not found")
		fd.lastSearchPos = -1
	}
}

// findInBuffer searches for text in the buffer
func findInBuffer(buffer *gtk.TextBuffer, searchText string,
	matchCase, wholeWord, isRegex, searchUp bool,
	startOffset int) (*gtk.TextIter, *gtk.TextIter, bool, bool) {

	// Get all text
	bufStart := buffer.GetStartIter()
	bufEnd := buffer.GetEndIter()
	text, _ := buffer.GetText(bufStart, bufEnd, false)

	wrapped := false
	var matchStart, matchEnd *gtk.TextIter

	if isRegex {
		// Regular expression search
		var re *regexp.Regexp
		var err error

		if matchCase {
			re, err = regexp.Compile(searchText)
		} else {
			re, err = regexp.Compile("(?i)" + searchText)
		}

		if err != nil {
			return nil, nil, false, false
		}

		// Search from start offset
		searchText := text[startOffset:]
		if searchUp {
			// For searching up, search in text before cursor
			searchText = text[:startOffset]
			matches := re.FindAllStringIndex(searchText, -1)
			if len(matches) > 0 {
				// Get last match (closest to cursor)
				match := matches[len(matches)-1]
				matchStart = buffer.GetIterAtOffset(match[0])
				matchEnd = buffer.GetIterAtOffset(match[1])
				return matchStart, matchEnd, true, false
			}
		} else {
			// Search forward
			loc := re.FindStringIndex(searchText)
			if loc != nil {
				matchStart = buffer.GetIterAtOffset(startOffset + loc[0])
				matchEnd = buffer.GetIterAtOffset(startOffset + loc[1])
				return matchStart, matchEnd, true, false
			}
		}

		// Try wrapping
		wrapped = true
		if searchUp {
			matches := re.FindAllStringIndex(text, -1)
			if len(matches) > 0 {
				match := matches[len(matches)-1]
				matchStart = buffer.GetIterAtOffset(match[0])
				matchEnd = buffer.GetIterAtOffset(match[1])
				return matchStart, matchEnd, true, wrapped
			}
		} else {
			loc := re.FindStringIndex(text)
			if loc != nil {
				matchStart = buffer.GetIterAtOffset(loc[0])
				matchEnd = buffer.GetIterAtOffset(loc[1])
				return matchStart, matchEnd, true, wrapped
			}
		}

		return nil, nil, false, false
	}

	// Plain text search
	searchFor := searchText
	searchIn := text

	if !matchCase {
		searchFor = strings.ToLower(searchFor)
		searchIn = strings.ToLower(text)
	}

	var foundOffset int = -1

	if searchUp {
		// Search backwards from cursor
		foundOffset = strings.LastIndex(searchIn[:startOffset], searchFor)
	} else {
		// Search forward from cursor
		idx := strings.Index(searchIn[startOffset:], searchFor)
		if idx != -1 {
			foundOffset = startOffset + idx
		}
	}

	// If not found, try wrapping
	if foundOffset == -1 {
		wrapped = true
		if searchUp {
			foundOffset = strings.LastIndex(searchIn, searchFor)
		} else {
			foundOffset = strings.Index(searchIn, searchFor)
		}
	}

	if foundOffset == -1 {
		return nil, nil, false, false
	}

	// Check whole word match if needed
	if wholeWord {
		if !isWholeWord(text, foundOffset, len(searchText)) {
			// Not a whole word match, continue searching
			// This is simplified - ideally would continue searching
			return nil, nil, false, false
		}
	}

	// Create iters for found text
	matchStart = buffer.GetIterAtOffset(foundOffset)
	matchEnd = buffer.GetIterAtOffset(foundOffset + len(searchText))

	return matchStart, matchEnd, true, wrapped
}

// isWholeWord checks if the match is a whole word
func isWholeWord(text string, offset, length int) bool {
	// Check character before
	if offset > 0 {
		charBefore := text[offset-1]
		if isWordChar(charBefore) {
			return false
		}
	}

	// Check character after
	if offset+length < len(text) {
		charAfter := text[offset+length]
		if isWordChar(charAfter) {
			return false
		}
	}

	return true
}

// isWordChar checks if a character is a word character
func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '_'
}

// highlightFoundText selects and scrolls to the found text
func highlightFoundText(buffer *gtk.TextBuffer, textView *gtk.TextView,
	start, end *gtk.TextIter) {

	// Select found text
	buffer.SelectRange(start, end)

	// Scroll to make visible (centered at 30% from top)
	textView.ScrollToIter(start, 0.0, true, 0.0, 0.3)
}
