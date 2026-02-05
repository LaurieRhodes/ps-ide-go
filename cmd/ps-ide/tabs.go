package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type SessionData struct {
	Tabs                []TabData `json:"tabs"`
	CommandAddOnVisible bool      `json:"commandAddOnVisible"`
	CommandAddOnWidth   int       `json:"commandAddOnWidth,omitempty"`
}

type TabData struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
	Modified bool   `json:"modified"`
}

// LineNumberView holds the line number TextView and related data
type LineNumberView struct {
	textView *gtk.TextView
	buffer   *gtk.TextBuffer
}

func createLineNumberView() *LineNumberView {
	textView, _ := gtk.TextViewNew()
	textView.SetEditable(false)
	textView.SetCursorVisible(false)
	textView.SetWrapMode(gtk.WRAP_NONE)
	textView.SetMonospace(true)
	textView.SetLeftMargin(5)
	textView.SetRightMargin(8) // More right margin for spacing

	buffer, _ := textView.GetBuffer()
	buffer.SetText("1")

	// Style like Windows ISE - gray background, gray text
	fontSize := DefaultFontSize * (currentZoom / 100.0)
	provider, _ := gtk.CssProviderNew()
	css := fmt.Sprintf(`textview { 
		font-family: %s; 
		font-size: %.1fpt; 
		background-color: #F0F0F0;
		padding: 3px;
	}
	textview text {
		color: #2B91AF;
	}`, DefaultFontFamily, fontSize)
	provider.LoadFromData(css)

	styleContext, _ := textView.GetStyleContext()
	styleContext.AddProvider(provider, gtk.STYLE_PROVIDER_PRIORITY_USER)

	// Prevent horizontal expansion
	textView.SetSizeRequest(50, -1)

	return &LineNumberView{
		textView: textView,
		buffer:   buffer,
	}
}

func updateLineNumbers(lineNumView *LineNumberView, editorBuffer *gtk.TextBuffer) {
	if lineNumView == nil {
		return
	}

	lineCount := editorBuffer.GetLineCount()

	// Build line number text
	var numbers strings.Builder
	for i := 1; i <= lineCount; i++ {
		if i > 1 {
			numbers.WriteString("\n")
		}
		numbers.WriteString(fmt.Sprintf("%d", i))
	}

	lineNumView.buffer.SetText(numbers.String())
}

func createNewTab() *ScriptTab {
	// Create main editor TextView
	textView, _ := gtk.TextViewNew()
	textView.SetWrapMode(gtk.WRAP_NONE)
	textView.SetMonospace(true)
	textView.SetLeftMargin(0) // No left margin - line numbers provide spacing
	textView.SetRightMargin(5)

	// Add right-click context menu
	textView.Connect("button-press-event", func(_ *gtk.TextView, event *gdk.Event) bool {
		eventButton := gdk.EventButtonNewFromEvent(event)
		if eventButton.Button() == 3 { // Right-click
			showEditorContextMenu(event)
			return true
		}
		return false
	})

	buffer, _ := textView.GetBuffer()
	buffer.SetText("")

	// Apply enhanced font rendering for crisp, clear text
	fontSize := DefaultFontSize * (currentZoom / 100.0)
	if err := ApplyEditorStyling(textView, fontSize); err != nil {
		// Fallback to old method if enhanced styling fails
		provider, _ := gtk.CssProviderNew()
		css := fmt.Sprintf(`textview { 
			font-family: "Lucida Console", "Courier New", monospace; 
			font-size: %.1fpt; 
			background-color: #FFFFFF;
			padding: 3px;
		}`, fontSize)
		provider.LoadFromData(css)
		styleContext, _ := textView.GetStyleContext()
		styleContext.AddProvider(provider, gtk.STYLE_PROVIDER_PRIORITY_USER)
	}

	// Create editor ScrolledWindow
	editorScroll, _ := gtk.ScrolledWindowNew(nil, nil)
	editorScroll.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_ALWAYS)
	editorScroll.SetShadowType(gtk.SHADOW_IN)
	editorScroll.Add(textView)

	// CRITICAL: Prevent ScrolledWindow from requesting child's natural height
	// Must be set AFTER adding the child
	editorScroll.Set("propagate-natural-height", false)
	editorScroll.Set("propagate-natural-width", false)

	// Set maximum content height to prevent growing beyond allocated space
	editorScroll.Set("max-content-height", 1)

	// Make ScrolledWindow expand to fill available space
	editorScroll.SetVExpand(true)
	editorScroll.SetHExpand(true)

	// Create line number view
	lineNumView := createLineNumberView()

	// Create a separate ScrolledWindow JUST for line numbers
	// This has its own scrolling but we'll sync it with the editor
	lineNumScroll, _ := gtk.ScrolledWindowNew(nil, editorScroll.GetVAdjustment())
	lineNumScroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_EXTERNAL)
	lineNumScroll.Add(lineNumView.textView)
	lineNumScroll.SetShadowType(gtk.SHADOW_NONE) // No border
	lineNumScroll.SetVExpand(true)
	lineNumScroll.SetHExpand(false)
	lineNumScroll.SetSizeRequest(50, -1)

	// Use HBox but DON'T let line numbers affect height
	hbox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	hbox.SetVExpand(true)
	hbox.SetHExpand(true)

	// CRITICAL: Set these properties to prevent HBox from requesting natural height
	hbox.Set("baseline-position", 0)

	// Pack line numbers without expand - fixed width
	hbox.PackStart(lineNumScroll, false, false, 0)
	// Pack editor with expand - takes remaining space
	hbox.PackStart(editorScroll, true, true, 0)

	// Wrap in another container that enforces size constraints
	wrapper, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	wrapper.SetVExpand(true)
	wrapper.SetHExpand(true)
	wrapper.PackStart(hbox, true, true, 0)

	container := wrapper

	// Initialize line numbers
	updateLineNumbers(lineNumView, buffer)

	// Create syntax highlighter (using configured engine)
	syntaxHighlighter := CreateSyntaxHighlighter(buffer)

	// Generate unique stable tab ID
	tabID := tabCounter
	tabCounter++

	tab := &ScriptTab{
		textView:          textView,
		buffer:            buffer,
		undoStack:         NewUndoStack(buffer, 100),
		filename:          "",
		modified:          false,
		lineNumView:       lineNumView,
		syntaxHighlighter: syntaxHighlighter,
		tabID:             tabID,
	}

	openTabs = append(openTabs, tab)

	// Use stable tabID for page name (not array index!)
	pageName := fmt.Sprintf("tab-%d", tabID)
	tabTitle := fmt.Sprintf("Untitled%d.ps1", tabID)

	// Add to Stack
	contentStack.AddTitled(container, pageName, tabTitle)
	contentStack.SetVisibleChildName(pageName)

	// Connect buffer signals
	buffer.Connect("notify::cursor-position", func() {
		updateCursorPosition(buffer)
		updateToolbarButtons()
	})

	buffer.Connect("changed", func() {
		tab.modified = true
		updateTabTitle(tab)
		updateLineNumbers(lineNumView, buffer)

		// Perform incremental syntax highlighting
		if tab.syntaxHighlighter != nil {
			tab.syntaxHighlighter.OnBufferChanged(buffer)
		}
	})

	buffer.Connect("mark-set", func(buf *gtk.TextBuffer, location *gtk.TextIter, mark *gtk.TextMark) {
		updateToolbarButtons()
		// Only auto-scroll for the insert mark (cursor), not for other marks
		insertMark := buffer.GetInsert()
		if mark.Native() == insertMark.Native() {
			// Use ScrollMarkOnscreen instead of ScrollToMark to avoid recursion
			textView.ScrollMarkOnscreen(mark)
		}
	})

	// Connect Stack visibility signal for tab switching
	contentStack.Connect("notify::visible-child-name", func() {
		onTabSwitch()
	})

	container.ShowAll()
	updateCursorPosition(buffer)
	updateToolbarButtons()

	// Update tab click handlers for middle-click and right-click
	updateTabClickHandlers()

	return tab
}

func closeTab(tabIndex int) {
	// Get current tab
	tab := getCurrentTab()
	if tab == nil {
		return
	}

	if tab.modified {
		tabName := "Untitled"
		if tab.filename != "" {
			tabName = getBaseName(tab.filename)
		}

		dialog := gtk.MessageDialogNew(
			mainWindow,
			gtk.DIALOG_MODAL,
			gtk.MESSAGE_WARNING,
			gtk.BUTTONS_NONE,
			"Do you want to save changes to '%s'?", tabName,
		)
		dialog.AddButton("Don't Save", gtk.RESPONSE_NO)
		dialog.AddButton("Cancel", gtk.RESPONSE_CANCEL)
		dialog.AddButton("Save", gtk.RESPONSE_YES)
		dialog.SetDefaultResponse(gtk.RESPONSE_YES)

		response := dialog.Run()
		dialog.Destroy()

		switch response {
		case gtk.RESPONSE_YES:
			if tab.filename == "" {
				if !showSaveAsDialog() {
					return
				}
			} else {
				saveCurrentFile()
			}
		case gtk.RESPONSE_CANCEL:
			return
		}
	}

	// If only one tab, clear it instead of removing
	if len(openTabs) == 1 {
		if len(openTabs) > 0 {
			openTabs[0].buffer.SetText("")
			openTabs[0].filename = ""
			openTabs[0].modified = false
			updateTabTitle(openTabs[0])
		}
		return
	}

	// Remove from Stack using the tab's stable ID
	pageName := fmt.Sprintf("tab-%d", tab.tabID)
	child, err := contentStack.GetChildByName(pageName)
	if err == nil && child != nil {
		contentStack.Remove(child.ToWidget())
	}

	// Remove from openTabs array by finding the tab
	for i, t := range openTabs {
		if t.tabID == tab.tabID {
			openTabs = append(openTabs[:i], openTabs[i+1:]...)
			break
		}
	}

	// Update tab click handlers after removing tab
	updateTabClickHandlers()
}

func closeCurrentTab() {
	currentTabIndex := getCurrentTabIndex()
	closeTab(currentTabIndex)
}

func getSessionFilePath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "ps-ide-go", "session.json")
}

func saveSession() {
	sessionData := SessionData{
		Tabs:                make([]TabData, 0),
		CommandAddOnVisible: commandAddOnVisible,
	}

	// Save Command Add-On paned position (represents width allocation)
	if commandAddOnPane != nil {
		sessionData.CommandAddOnWidth = commandAddOnPane.GetPosition()
	}

	for _, tab := range openTabs {
		start := tab.buffer.GetStartIter()
		end := tab.buffer.GetEndIter()
		content, _ := tab.buffer.GetText(start, end, false)

		if content != "" || tab.filename != "" {
			sessionData.Tabs = append(sessionData.Tabs, TabData{
				Filename: tab.filename,
				Content:  content,
				Modified: tab.modified,
			})
		}
	}

	sessionPath := getSessionFilePath()
	os.MkdirAll(filepath.Dir(sessionPath), 0755)

	data, err := json.MarshalIndent(sessionData, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(sessionPath, data, 0644)
}

func loadSession() bool {
	sessionPath := getSessionFilePath()

	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return false
	}

	var sessionData SessionData
	if err := json.Unmarshal(data, &sessionData); err != nil {
		return false
	}

	// Restore Command Add-On visibility state
	if sessionData.CommandAddOnVisible {
		// Will be shown after main window is realized
		pendingCommandAddOnShow = true
		pendingCommandAddOnWidth = sessionData.CommandAddOnWidth
	}

	if len(sessionData.Tabs) == 0 {
		return false
	}

	for _, tabData := range sessionData.Tabs {
		tab := createNewTab()
		tab.buffer.SetText(tabData.Content)
		tab.filename = tabData.Filename
		tab.modified = tabData.Modified

		// Force scroll to the beginning of the document
		startIter := tab.buffer.GetStartIter()
		tab.buffer.PlaceCursor(startIter)

		// Get the mark at the start and scroll to it
		startMark := tab.buffer.CreateMark("start", startIter, true)
		tab.textView.ScrollToMark(startMark, 0.0, true, 0.0, 0.0)
		tab.buffer.DeleteMark(startMark)

		updateTabTitle(tab)

		// Perform initial syntax highlighting
		if tab.syntaxHighlighter != nil {
			tab.syntaxHighlighter.Highlight()
		}
	}

	return true
}

// showEditorContextMenu displays a right-click context menu for the editor
func showEditorContextMenu(event *gdk.Event) {
	menu, _ := gtk.MenuNew()

	// Cut
	cutItem, _ := gtk.MenuItemNewWithLabel("Cut")
	cutItem.Connect("activate", func() {
		cutText()
	})
	menu.Append(cutItem)

	// Copy
	copyItem, _ := gtk.MenuItemNewWithLabel("Copy")
	copyItem.Connect("activate", func() {
		copyText()
	})
	menu.Append(copyItem)

	// Paste
	pasteItem, _ := gtk.MenuItemNewWithLabel("Paste")
	pasteItem.Connect("activate", func() {
		pasteText()
	})
	menu.Append(pasteItem)

	// Separator
	separator1, _ := gtk.SeparatorMenuItemNew()
	menu.Append(separator1)

	// Undo
	undoItem, _ := gtk.MenuItemNewWithLabel("Undo")
	undoItem.Connect("activate", func() {
		undoText()
	})
	menu.Append(undoItem)

	// Redo
	redoItem, _ := gtk.MenuItemNewWithLabel("Redo")
	redoItem.Connect("activate", func() {
		redoText()
	})
	menu.Append(redoItem)

	// Separator
	separator2, _ := gtk.SeparatorMenuItemNew()
	menu.Append(separator2)

	// Insert Snippet (Ctrl+J)
	snippetItem, _ := gtk.MenuItemNewWithLabel("Insert Snippet...                    Ctrl+J")
	snippetItem.Connect("activate", func() {
		showSnippetsDialog()
	})
	menu.Append(snippetItem)

	menu.ShowAll()
	menu.PopupAtPointer(event)
}

// Pending Command Add-On state for deferred show
var pendingCommandAddOnShow bool
var pendingCommandAddOnWidth int
