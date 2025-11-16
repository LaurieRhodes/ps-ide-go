package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gotk3/gotk3/gtk"
)

type SessionData struct {
	Tabs []TabData `json:"tabs"`
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
	textView.SetRightMargin(5)
	
	buffer, _ := textView.GetBuffer()
	buffer.SetText("1")
	
	// Style the line number view
	baseFontSize := 9.0
	newFontSize := baseFontSize * (currentZoom / 100.0)
	provider, _ := gtk.CssProviderNew()
	css := fmt.Sprintf(`textview { 
		font-family: "Lucida Console", "Courier New", monospace; 
		font-size: %.1fpt; 
		background-color: #F0F0F0;
		color: #808080;
		padding: 3px;
	}`, newFontSize)
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
	fmt.Println("DEBUG: createNewTab() called")
	// Create main editor first
	textView, _ := gtk.TextViewNew()
	fmt.Println("DEBUG: TextView created")
	textView.SetWrapMode(gtk.WRAP_NONE)
	textView.SetMonospace(true)
	textView.SetLeftMargin(5)
	textView.SetRightMargin(5)

	buffer, _ := textView.GetBuffer()
	buffer.SetText("")
	fmt.Println("DEBUG: Buffer created")

	baseFontSize := 9.0
	newFontSize := baseFontSize * (currentZoom / 100.0)
	provider, _ := gtk.CssProviderNew()
	css := fmt.Sprintf(`textview { 
		font-family: "Lucida Console", "Courier New", monospace; 
		font-size: %.1fpt; 
		background-color: #FFFFFF;
		padding: 3px;
	}`, newFontSize)
	provider.LoadFromData(css)
	
	styleContext, _ := textView.GetStyleContext()
	styleContext.AddProvider(provider, gtk.STYLE_PROVIDER_PRIORITY_USER)

	// Create editor scroll window with scrollbars
	editorScroll, _ := gtk.ScrolledWindowNew(nil, nil)
	editorScroll.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_ALWAYS)  // Always show vertical scrollbar
	editorScroll.SetShadowType(gtk.SHADOW_IN)  // Add shadow for visibility
	// Ensure minimum content size is set for proper scrollbar function
	editorScroll.SetMinContentHeight(400)
	editorScroll.Add(textView)
	
	// Create line number view
	lineNumView := createLineNumberView()
	
	// Create line number scroll window - share vertical adjustment with editor
	editorVAdj := editorScroll.GetVAdjustment()
	lineNumScroll, _ := gtk.ScrolledWindowNew(nil, editorVAdj)
	lineNumScroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_NEVER)  // No scrollbars on line numbers
	lineNumScroll.Add(lineNumView.textView)
	
	// Create HBox to hold line numbers and editor
	hbox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	hbox.PackStart(lineNumScroll, false, false, 0)
	hbox.PackStart(editorScroll, true, true, 0)
	
	// Initialize line numbers
	updateLineNumbers(lineNumView, buffer)

	tabLabel := createTabLabel(fmt.Sprintf("Untitled%d.ps1", tabCounter), len(openTabs))

	tab := &ScriptTab{
		textView:    textView,
		buffer:      buffer,
		undoStack:   NewUndoStack(buffer, 100),
		filename:    "",
		modified:    false,
		lineNumView: lineNumView,
	}

	openTabs = append(openTabs, tab)
	tabCounter++

	pageNum := notebook.AppendPage(hbox, tabLabel)
	fmt.Printf("DEBUG: Tab appended to notebook, pageNum=%d\n", pageNum)
	notebook.SetCurrentPage(pageNum)
	notebook.SetTabReorderable(hbox, true)
	fmt.Println("DEBUG: Tab configuration complete")

	buffer.Connect("notify::cursor-position", func() {
		updateCursorPosition(buffer)
		updateToolbarButtons()
	})

	buffer.Connect("changed", func() {
		tab.modified = true
		updateTabLabelText(pageNum)
		updateLineNumbers(lineNumView, buffer)
	})

	buffer.Connect("mark-set", func() {
		updateToolbarButtons()
	})

	hbox.ShowAll()
	updateCursorPosition(buffer)
	updateToolbarButtons()

	return tab
}

func createTabLabel(title string, tabIndex int) *gtk.Box {
	fmt.Printf("DEBUG: Creating tab label '%s' for index %d\n", title, tabIndex)
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 4)

	label, _ := gtk.LabelNew(title)
	fmt.Printf("DEBUG: Tab label widget created\n")
	box.PackStart(label, false, false, 0)

	closeBtn, _ := gtk.ButtonNew()
	closeBtn.SetRelief(gtk.RELIEF_NONE)
	closeLabel, _ := gtk.LabelNew("×")
	closeBtn.Add(closeLabel)
	closeBtn.SetTooltipText("Close this script tab")
	
	closeBtnContext, _ := closeBtn.GetStyleContext()
	cssProvider, _ := gtk.CssProviderNew()
	cssProvider.LoadFromData(`
		button {
			min-width: 18px;
			min-height: 18px;
			padding: 0px;
			margin: 0px;
			border: none;
			background: transparent;
			font-size: 14px;
		}
		button:hover {
			background-color: rgba(0, 0, 0, 0.15);
		}
	`)
	closeBtnContext.AddProvider(cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	closeBtn.Connect("clicked", func() {
		closeTab(tabIndex)
	})
	box.PackStart(closeBtn, false, false, 0)

	box.ShowAll()
	return box
}

func closeTab(tabIndex int) {
	pageNum := notebook.GetCurrentPage()
	if pageNum == -1 {
		return
	}

	if pageNum < len(openTabs) && openTabs[pageNum].modified {
		tab := openTabs[pageNum]
		
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

	if notebook.GetNPages() == 1 {
		if len(openTabs) > 0 {
			openTabs[0].buffer.SetText("")
			openTabs[0].filename = ""
			openTabs[0].modified = false
			updateTabLabelText(0)
		}
		return
	}

	notebook.RemovePage(pageNum)
	if pageNum < len(openTabs) {
		openTabs = append(openTabs[:pageNum], openTabs[pageNum+1:]...)
	}
}

func closeCurrentTab() {
	pageNum := notebook.GetCurrentPage()
	closeTab(pageNum)
}

func getSessionFilePath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "ps-ide-go", "session.json")
}

func saveSession() {
	sessionData := SessionData{
		Tabs: make([]TabData, 0),
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

		pageNum := len(openTabs) - 1
		updateTabLabelText(pageNum)
	}

	return true
}
