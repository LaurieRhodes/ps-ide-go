package main

import (
	"fmt"
	"log"
	"unsafe"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

var (
	statusLabel    *gtk.Label
	cursorPosLabel *gtk.Label
	notebook       *gtk.Notebook
	tabCounter     int
	runButton      *gtk.ToolButton
	cutButton      *gtk.ToolButton
	copyButton     *gtk.ToolButton
	pasteButton    *gtk.ToolButton
	undoButton     *gtk.ToolButton
	redoButton     *gtk.ToolButton
	stopButton     *gtk.ToolButton
	isExecuting    bool
	mainWindow     *gtk.Window
	zoomScale      *gtk.Scale
	zoomLabel      *gtk.Label
	currentZoom    float64 = 100.0
)

type ScriptTab struct {
	textView    *gtk.TextView
	buffer      *gtk.TextBuffer
	undoStack   *UndoStack
	filename    string
	modified    bool
	lineNumView *LineNumberView
}

var openTabs []*ScriptTab

func main() {
	gtk.Init(nil)
	tabCounter = 1

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	mainWindow = win
	win.SetTitle("PS-IDE-Go - PowerShell ISE")
	win.SetDefaultSize(1000, 750)
	
	// Set PowerShell icon
	setWindowIcon(win)
	
	// Save session on window close
	win.Connect("destroy", func() {
		saveSession()
		shutdownTranslationLayer()
		gtk.MainQuit()
	})

	// Connect keyboard shortcuts
	win.Connect("key-press-event", onKeyPress)

	// Apply ISE-like CSS styling
	applyCss()

	mainVBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	
	// Menu bar
	menuBar := createMenuBar(win)
	menuBar.SetVExpand(false)  // Don't expand vertically
	mainVBox.PackStart(menuBar, false, false, 0)

	// Toolbar
	toolbar := createToolbar()
	toolbar.SetVExpand(false)  // Don't expand vertically
	mainVBox.PackStart(toolbar, false, false, 0)
	
	fmt.Println("DEBUG: Menu and toolbar added")

	// Notebook for script tabs
	notebook, _ = gtk.NotebookNew()
	notebook.SetScrollable(true)
	notebook.SetShowTabs(true)  // Explicitly show tabs
	notebook.SetShowBorder(true) // Show border for visibility
	notebook.SetVExpand(true)    // Allow vertical expansion
	notebook.SetHExpand(true)    // Allow horizontal expansion
	notebook.Connect("switch-page", func() {
		onTabSwitch()
	})
	fmt.Println("DEBUG: Notebook created and configured")

	// Try to load previous session, if fails create default tab
	if !loadSession() {
		createNewTab()
	}
	
	// Debug: Check notebook state
	numPages := notebook.GetNPages()
	showTabs := notebook.GetShowTabs()
	fmt.Printf("DEBUG: Notebook has %d pages, showTabs=%v\n", numPages, showTabs)

	// Create PowerShell console using Translation Layer
	consoleScroll, consoleErr := createConsoleUI()
	if consoleErr != nil {
		log.Fatal("Unable to create console:", consoleErr)
	}

	// Initialize Translation Layer
	if err := initTranslationLayer(); err != nil {
		log.Printf("Warning: Translation Layer failed to initialize: %v", err)
		log.Println("PowerShell functionality will be limited")
	}

	// Split pane layout (editor top, console bottom)
	paned, _ := gtk.PanedNew(gtk.ORIENTATION_VERTICAL)
	paned.Pack1(notebook, true, false)  // notebook resizable but NOT shrinkable
	paned.Pack2(consoleScroll, true, true)  // console resizable and shrinkable
	paned.SetWideHandle(true)  // Make divider easier to grab
	
	// Ensure notebook has minimum size
	notebook.SetSizeRequest(-1, 250)  // Minimum 250px height to ensure tabs visible
	
	fmt.Printf("DEBUG: Paned created\n")
	
	mainVBox.PackStart(paned, true, true, 0)
	
	// Set paned position AFTER adding to mainVBox (important!)
	// Use GLib idle to set position after window is realized
	paned.SetPosition(400)  // Initial position: 400px for editor
	
	fmt.Printf("DEBUG: Paned position set to 400\n")

	// Status bar
	statusBar := createStatusBar()
	mainVBox.PackStart(statusBar, false, false, 0)

	win.Add(mainVBox)
	win.ShowAll()

	gtk.Main()
}

func setWindowIcon(win *gtk.Window) {
	// Create a PowerShell-style icon (blue background with > symbol)
	pixbuf, err := gdk.PixbufNew(gdk.COLORSPACE_RGB, true, 8, 48, 48)
	if err != nil {
		return
	}

	// Fill with PowerShell blue background
	pixels := pixbuf.GetPixels()
	rowstride := pixbuf.GetRowstride()
	nChannels := pixbuf.GetNChannels()

	for y := 0; y < 48; y++ {
		for x := 0; x < 48; x++ {
			offset := y*rowstride + x*nChannels
			// PowerShell blue color
			pixels[offset] = 0x01     // R
			pixels[offset+1] = 0x24   // G
			pixels[offset+2] = 0x56   // B
			pixels[offset+3] = 0xFF   // A
		}
	}

	// Draw white > symbol and underscore
	for i := 0; i < 20; i++ {
		y := 14 + i
		x := 12 + i/2
		if x < 48 && y < 48 {
			offset := y*rowstride + x*nChannels
			pixels[offset] = 0xFF
			pixels[offset+1] = 0xFF
			pixels[offset+2] = 0xFF
			pixels[offset+3] = 0xFF
		}
	}
	for x := 24; x < 40; x++ {
		y := 35
		offset := y*rowstride + x*nChannels
		pixels[offset] = 0xFF
		pixels[offset+1] = 0xFF
		pixels[offset+2] = 0xFF
		pixels[offset+3] = 0xFF
	}

	win.SetIcon(pixbuf)
}

func onKeyPress(_ interface{}, event *gdk.Event) bool {
	keyEvent := gdk.EventKeyNewFromEvent(event)
	keyval := keyEvent.KeyVal()
	state := keyEvent.State()

	ctrl := (state & gdk.CONTROL_MASK) != 0

	if ctrl && keyval == gdk.KEY_n {
		newScript()
		return true
	}
	if ctrl && keyval == gdk.KEY_o {
		openScript(mainWindow)
		return true
	}
	if ctrl && keyval == gdk.KEY_s {
		saveScript(mainWindow)
		return true
	}
	if ctrl && keyval == gdk.KEY_x {
		cutText()
		return true
	}
	if ctrl && keyval == gdk.KEY_c {
		copyText()
		return true
	}
	if ctrl && keyval == gdk.KEY_v {
		pasteText()
		return true
	}
	if ctrl && keyval == gdk.KEY_z {
		undoText()
		return true
	}
	if ctrl && keyval == gdk.KEY_y {
		redoText()
		return true
	}
	if keyval == gdk.KEY_F5 {
		runScript()
		return true
	}
	if keyval == gdk.KEY_F8 {
		runSelection()
		return true
	}

	return false
}

func applyCss() {
	cssProvider, _ := gtk.CssProviderNew()
	css := `
		notebook {
			background-color: #F5F0E8;
		}
		notebook > header {
			background-color: #F5F0E8;
			border-bottom: 1px solid #CCCCCC;
		}
		notebook > header > tabs > tab {
			background-color: #CCCCCC;
			border: 1px solid #999999;
			border-bottom: none;
			padding: 6px 12px;
			margin-right: 2px;
		}
		notebook > header > tabs > tab:checked {
			background-color: #FFFFFF;
			border-bottom: 1px solid #FFFFFF;
		}
		textview {
			font-family: "Lucida Console", "Courier New", monospace;
			font-size: 9pt;
			background-color: #FFFFFF;
			padding: 3px;
		}
		#statusbar {
			background-color: #F0F0F0;
			border-top: 1px solid #CCCCCC;
			padding: 3px 8px;
		}
		toolbar {
			background: linear-gradient(to bottom, #FCFDFE 0%, #E8F3FB 50%, #DCE9F7 100%);
			border-bottom: 1px solid #C5D7E8;
			padding: 4px;
		}
		menubar {
			background: linear-gradient(to bottom, #FCFDFE 0%, #E8F3FB 50%, #DCE9F7 100%);
			border-bottom: 1px solid #C5D7E8;
		}
		menu {
			background-color: #F5F5F5;
		}
	`
	cssProvider.LoadFromData(css)
	screen, _ := gdk.ScreenGetDefault()
	gtk.AddProviderForScreen(screen, cssProvider, gtk.STYLE_PROVIDER_PRIORITY_USER)
}

func onTabSwitch() {
	tab := getCurrentTab()
	if tab != nil && tab.buffer != nil {
		updateCursorPosition(tab.buffer)
		updateToolbarButtons()
	}
}

func getCurrentTab() *ScriptTab {
	pageNum := notebook.GetCurrentPage()
	if pageNum == -1 || pageNum >= len(openTabs) {
		return nil
	}
	return openTabs[pageNum]
}

func createStatusBar() *gtk.Box {
	statusBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 8)
	statusBox.SetName("statusbar")
	statusBox.SetMarginTop(2)
	statusBox.SetMarginBottom(2)
	statusBox.SetMarginStart(8)
	statusBox.SetMarginEnd(8)

	statusLabel, _ = gtk.LabelNew("Ready")
	statusBox.PackStart(statusLabel, false, false, 0)

	spacer, _ := gtk.LabelNew("")
	statusBox.PackStart(spacer, true, true, 0)

	zoomLabel, _ = gtk.LabelNew("100%")
	statusBox.PackEnd(zoomLabel, false, false, 3)

	zoomScale, _ = gtk.ScaleNewWithRange(gtk.ORIENTATION_HORIZONTAL, 50, 400, 10)
	zoomScale.SetValue(100)
	zoomScale.SetSizeRequest(120, -1)
	zoomScale.SetDrawValue(false)
	zoomScale.Connect("value-changed", func() {
		value := zoomScale.GetValue()
		zoomLabel.SetText(fmt.Sprintf("%.0f%%", value))
		applyZoom(value)
	})
	statusBox.PackEnd(zoomScale, false, false, 3)

	zoomTextLabel, _ := gtk.LabelNew("Zoom:")
	statusBox.PackEnd(zoomTextLabel, false, false, 8)

	cursorPosLabel, _ = gtk.LabelNew("Ln 1, Col 1")
	statusBox.PackEnd(cursorPosLabel, false, false, 12)

	return statusBox
}

func applyZoom(zoomPercent float64) {
	currentZoom = zoomPercent
	baseFontSize := 9.0
	newFontSize := baseFontSize * (zoomPercent / 100.0)

	for _, tab := range openTabs {
		if tab.textView != nil {
			provider, _ := gtk.CssProviderNew()
			css := fmt.Sprintf(`textview { 
				font-family: "Lucida Console", "Courier New", monospace; 
				font-size: %.1fpt; 
				background-color: #FFFFFF;
				padding: 3px;
			}`, newFontSize)
			provider.LoadFromData(css)
			styleContext, _ := tab.textView.GetStyleContext()
			styleContext.AddProvider(provider, gtk.STYLE_PROVIDER_PRIORITY_USER)
			tab.textView.QueueDraw()
		}
		
		// Also update line numbers when zooming
		if tab.lineNumView != nil && tab.lineNumView.textView != nil {
			provider, _ := gtk.CssProviderNew()
			css := fmt.Sprintf(`textview { 
				font-family: "Lucida Console", "Courier New", monospace; 
				font-size: %.1fpt; 
				background-color: #F0F0F0;
				color: #808080;
				padding: 3px;
			}`, newFontSize)
			provider.LoadFromData(css)
			styleContext, _ := tab.lineNumView.textView.GetStyleContext()
			styleContext.AddProvider(provider, gtk.STYLE_PROVIDER_PRIORITY_USER)
			tab.lineNumView.textView.QueueDraw()
		}
	}

	if consoleTextView != nil {
		consoleFontSize := 10.0 * (zoomPercent / 100.0)
		provider, _ := gtk.CssProviderNew()
		css := fmt.Sprintf(`textview {
			background-color: #012456;
			color: #FFFFFF;
			font-family: "Courier New", "Lucida Console", monospace;
			font-size: %.1fpt;
			padding: 5px;
		}`, consoleFontSize)
		provider.LoadFromData(css)
		styleContext, _ := consoleTextView.GetStyleContext()
		styleContext.AddProvider(provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
		consoleTextView.QueueDraw()
	}
}

func updateCursorPosition(buf *gtk.TextBuffer) {
	if cursorPosLabel == nil {
		return
	}
	mark := buf.GetInsert()
	iter := buf.GetIterAtMark(mark)
	line := iter.GetLine() + 1
	col := iter.GetLineOffset() + 1
	cursorPosLabel.SetText(fmt.Sprintf("Ln %d, Col %d", line, col))
}

func updateToolbarButtons() {
	tab := getCurrentTab()
	if tab == nil {
		return
	}

	_, _, hasSelection := tab.buffer.GetSelectionBounds()
	if cutButton != nil {
		cutButton.SetSensitive(hasSelection)
	}
	if copyButton != nil {
		copyButton.SetSensitive(hasSelection)
	}

	if undoButton != nil {
		undoButton.SetSensitive(tab.undoStack != nil && tab.undoStack.CanUndo())
	}
	if redoButton != nil {
		redoButton.SetSensitive(tab.undoStack != nil && tab.undoStack.CanRedo())
	}
}

func setExecuting(executing bool) {
	isExecuting = executing
	if runButton != nil {
		runButton.SetSensitive(!executing)
	}
	if stopButton != nil {
		stopButton.SetSensitive(executing)
	}
	if executing {
		statusLabel.SetText("Executing...")
	} else {
		statusLabel.SetText("Ready")
	}
}

func getBaseName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}
	return path
}

func updateTabLabelText(pageNum int) {
	if pageNum < 0 || pageNum >= len(openTabs) {
		return
	}

	tab := openTabs[pageNum]
	page, _ := notebook.GetNthPage(pageNum)
	if page == nil {
		return
	}

	tabLabelWidget, _ := notebook.GetTabLabel(page)
	if tabLabelWidget == nil {
		return
	}

	widget := tabLabelWidget.ToWidget()
	if widget == nil {
		return
	}

	obj := glib.Take(unsafe.Pointer(widget.Native()))
	box := &gtk.Box{gtk.Container{gtk.Widget{glib.InitiallyUnowned{obj}}}}
	children := box.GetChildren()
	if children != nil {
		labelWidget := children.Data().(*gtk.Widget)
		obj2 := glib.Take(unsafe.Pointer(labelWidget.Native()))
		label := &gtk.Label{gtk.Widget{glib.InitiallyUnowned{obj2}}}

		title := fmt.Sprintf("Untitled%d.ps1", pageNum+1)
		if tab.filename != "" {
			title = getBaseName(tab.filename)
		}
		if tab.modified {
			title = "* " + title
		}
		label.SetText(title)
	}
}
