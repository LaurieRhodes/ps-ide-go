package main

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

var (
	statusLabel    *gtk.Label
	cursorPosLabel *gtk.Label
	contentStack   *gtk.Stack         // Holds editor content pages
	stackSwitcher  *gtk.StackSwitcher // Tab bar
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

	// Command Add-On
	commandDatabase          *CommandDatabase
	commandAddOnPane         *gtk.Paned
	commandAddOnVisible      bool = false
	showCommandAddonMenuItem *gtk.CheckMenuItem
	updatingCommandAddonMenu bool = false // Flag to prevent signal loops
)

type ScriptTab struct {
	textView          *gtk.TextView
	buffer            *gtk.TextBuffer
	undoStack         *UndoStack
	filename          string
	modified          bool
	lineNumView       *LineNumberView
	syntaxHighlighter SyntaxHighlighterInterface
	tabID             int // Unique, stable ID for this tab
}

var openTabs []*ScriptTab

func main() {
	gtk.Init(nil)
	tabCounter = 1

	// Setup optimal font rendering for crisp, clear text
	SetupFontRendering()

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
	menuBar.SetVExpand(false) // Don't expand vertically
	mainVBox.PackStart(menuBar, false, false, 0)

	// Toolbar
	toolbar := createToolbar()
	toolbar.SetVExpand(false) // Don't expand vertically
	mainVBox.PackStart(toolbar, false, false, 0)

	// Create Stack for editor content pages
	contentStack, _ = gtk.StackNew()
	contentStack.SetTransitionType(gtk.STACK_TRANSITION_TYPE_NONE)
	contentStack.SetVExpand(true)
	contentStack.SetHExpand(true)

	// Force homogeneous sizing to prevent children from requesting different sizes
	contentStack.Set("hhomogeneous", true)
	contentStack.Set("vhomogeneous", true)

	// Create StackSwitcher for tabs
	stackSwitcher, _ = gtk.StackSwitcherNew()
	stackSwitcher.SetStack(contentStack)
	stackSwitcher.SetVExpand(false)
	stackSwitcher.SetHExpand(true)

	// Add tab bar (StackSwitcher) to main layout - stays fixed
	mainVBox.PackStart(stackSwitcher, false, false, 0)

	// Try to load previous session, if fails create default tab
	if !loadSession() {
		createNewTab()
	}

	// Setup tab click handlers for middle-click and right-click
	setupTabClickHandlers()

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

	// Split pane layout (editor content top, console bottom)
	paned, _ := gtk.PanedNew(gtk.ORIENTATION_VERTICAL)
	paned.Pack1(contentStack, true, true)  // Stack content resizable and shrinkable
	paned.Pack2(consoleScroll, true, true) // console resizable and shrinkable
	paned.SetWideHandle(true)              // Make divider easier to grab

	// Set minimum sizes
	contentStack.SetSizeRequest(-1, 150)
	consoleScroll.SetSizeRequest(-1, 150)

	// Create horizontal paned for command add-on (editor+console | command-addon)
	commandAddOnPane, _ = gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	commandAddOnPane.SetWideHandle(true)
	commandAddOnPane.Pack1(paned, true, false) // Main content: resize=true, shrink=false

	// Initialize command database first
	commandDatabase = NewCommandDatabase()

	// Create and add command add-on
	commandAddOn = createCommandAddOn(commandDatabase)
	commandAddOnPane.Pack2(commandAddOn.container, true, true) // Add-on: resize=true, shrink=true (allows width resize and shrinking)
	commandAddOn.container.SetSizeRequest(200, -1)             // Minimum width (smaller)
	commandAddOn.container.Hide()                              // Hidden by default

	mainVBox.PackStart(commandAddOnPane, true, true, 0)

	// Set paned position AFTER adding to mainVBox (important!)
	// Use GLib idle to set position after window is realized
	paned.SetPosition(400) // Initial position: 400px for editor

	// Initialize command database in background
	initializeCommandAddOn()

	// Status bar
	statusBar := createStatusBar()
	mainVBox.PackStart(statusBar, false, false, 0)

	win.Add(mainVBox)
	win.ShowAll()

	// Hide Command Add-On after ShowAll (ShowAll shows everything)
	if commandAddOn != nil && !pendingCommandAddOnShow {
		commandAddOn.container.Hide()
	}

	// Restore Command Add-On visibility and position after window is shown
	glib.IdleAdd(func() {
		if pendingCommandAddOnShow && commandAddOn != nil {
			commandAddOn.container.ShowAll()
			commandAddOnVisible = true
			// Restore paned position if saved
			if pendingCommandAddOnWidth > 0 && commandAddOnPane != nil {
				commandAddOnPane.SetPosition(pendingCommandAddOnWidth)
			}
		} else if commandAddOn != nil {
			// Ensure it's hidden and set a default paned position for when it's shown later
			commandAddOn.container.Hide()
			commandAddOnVisible = false
		}

		// Sync menu checkbox with actual state (without triggering signal)
		if showCommandAddonMenuItem != nil {
			updatingCommandAddonMenu = true
			showCommandAddonMenuItem.SetActive(commandAddOnVisible)
			updatingCommandAddonMenu = false
		}
	})

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
			pixels[offset] = 0x01   // R
			pixels[offset+1] = 0x24 // G
			pixels[offset+2] = 0x56 // B
			pixels[offset+3] = 0xFF // A
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
	if ctrl && keyval == gdk.KEY_w {
		closeCurrentTab()
		return true
	}
	if ctrl && keyval == gdk.KEY_F4 {
		closeCurrentTab()
		return true
	}
	if ctrl && keyval == gdk.KEY_f {
		showFindDialog()
		return true
	}
	if ctrl && keyval == gdk.KEY_j {
		showSnippetsDialog()
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
		/* Make scrollbars thinner like Windows ISE */
		scrollbar {
			min-width: 14px;
			min-height: 14px;
		}
		scrollbar slider {
			min-width: 12px;
			min-height: 12px;
		}
		
		stackswitcher {
			background-color: #F5F0E8;
			border-bottom: 1px solid #CCCCCC;
			padding: 2px;
		}
		stackswitcher > button {
			background-color: #CCCCCC;
			border: 1px solid #999999;
			border-bottom: none;
			padding: 3px 10px;
			margin-right: 1px;
			min-height: 22px;
			font-size: 8.5pt;
		}
		stackswitcher > button:checked {
			background-color: #FFFFFF;
			border-bottom: 1px solid #FFFFFF;
			font-weight: normal;
		}
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
		textview:not(.console-textview) {
			font-family: "Lucida Console", "Courier New", monospace;
			font-size: 9pt;
			background-color: #FFFFFF;
			padding: 3px;
		}
		textview.console-textview {
			background-color: #012456;
			color: #FFFFFF;
			font-family: "Consolas", "Liberation Mono", "Courier New", monospace;
			font-size: 11pt;
			padding: 5px;
			caret-color: #FFFFFF;
		}
		textview.console-textview text {
			background-color: #012456;
			color: #FFFFFF;
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
		/* Command Add-On parameter set tabs (FlowBox toggle buttons) */
		flowbox {
			background-color: transparent;
		}
		flowbox flowboxchild {
			padding: 1px;
		}
		flowbox togglebutton {
			background-color: #E0E0E0;
			border: 1px solid #999999;
			padding: 4px 8px;
			min-height: 20px;
			font-size: 8pt;
		}
		flowbox togglebutton:checked {
			background-color: #FFFFFF;
			border-color: #666666;
			font-weight: bold;
		}
		flowbox togglebutton:hover {
			background-color: #F0F0F0;
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
	// Get the visible child name from Stack
	visibleName := contentStack.GetVisibleChildName()
	if visibleName == "" {
		return nil
	}

	// Parse the tab ID from the name (format: "tab-ID")
	var tabID int
	_, err := fmt.Sscanf(visibleName, "tab-%d", &tabID)
	if err != nil {
		return nil
	}

	// Find tab with matching ID
	for _, tab := range openTabs {
		if tab.tabID == tabID {
			return tab
		}
	}

	return nil
}

func updateTabTitle(tab *ScriptTab) {
	if tab == nil {
		return
	}

	pageName := fmt.Sprintf("tab-%d", tab.tabID)
	title := fmt.Sprintf("Untitled%d.ps1", tab.tabID)
	if tab.filename != "" {
		title = getBaseName(tab.filename)
	}
	if tab.modified {
		title = "* " + title
	}

	// Update the Stack child's title
	child, err := contentStack.GetChildByName(pageName)
	if err == nil && child != nil {
		contentStack.ChildSetProperty(child.ToWidget(), "title", title)
	}
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
	fontSize := DefaultFontSize * (zoomPercent / 100.0)

	for _, tab := range openTabs {
		if tab.textView != nil {
			// Apply enhanced font rendering
			if err := ApplyEditorStyling(tab.textView, fontSize); err != nil {
				// Fallback to old method
				provider, _ := gtk.CssProviderNew()
				css := fmt.Sprintf(`textview { 
					font-family: "Lucida Console", "Courier New", monospace; 
					font-size: %.1fpt; 
					background-color: #FFFFFF;
					padding: 3px;
				}`, fontSize)
				provider.LoadFromData(css)
				styleContext, _ := tab.textView.GetStyleContext()
				styleContext.AddProvider(provider, gtk.STYLE_PROVIDER_PRIORITY_USER)
			}
			tab.textView.QueueDraw()
		}

		// Also update line numbers when zooming
		if tab.lineNumView != nil && tab.lineNumView.textView != nil {
			provider, _ := gtk.CssProviderNew()
			css := fmt.Sprintf(`textview { 
				font-family: %s; 
				font-size: %.1fpt; 
				background-color: #F0F0F0;
				padding: 3px;
				border-right: 1px solid #D0D0D0;
			}
			textview text {
				color: #2B91AF;
			}`, DefaultFontFamily, fontSize)
			provider.LoadFromData(css)
			styleContext, _ := tab.lineNumView.textView.GetStyleContext()
			styleContext.AddProvider(provider, gtk.STYLE_PROVIDER_PRIORITY_USER)
			tab.lineNumView.textView.QueueDraw()
		}

		// Update syntax highlighting after zoom
		if tab.syntaxHighlighter != nil {
			tab.syntaxHighlighter.UpdateZoom()
		}
	}

	if consoleTextView != nil {
		consoleFontSize := DefaultFontSize * (zoomPercent / 100.0) // Same as editor
		provider, _ := gtk.CssProviderNew()
		css := fmt.Sprintf(`textview.console-textview {
			background-color: #012456 !important;
			color: #FFFFFF !important;
			font-family: "Courier New", "Lucida Console", monospace;
			font-size: %.1fpt;
			font-weight: normal;
			padding: 5px;
			caret-color: #FFFFFF;
		}
		textview.console-textview text {
			background-color: #012456 !important;
			color: #FFFFFF !important;
		}
		textview.console-textview:selected {
			background-color: #0066CC;
		}`, consoleFontSize)
		provider.LoadFromData(css)
		styleContext, _ := consoleTextView.GetStyleContext()
		// Use priority 900 (higher than USER 800) to override global screen CSS
		styleContext.AddProvider(provider, 900)
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

// initializeCommandAddOn initializes the command database and add-on
func initializeCommandAddOn() {
	log.Println("Initializing Command Add-On...")

	// Load help files in background (database already created)
	go func() {
		log.Println("Loading PowerShell help files...")
		err := commandDatabase.LoadHelp()
		if err != nil {
			log.Printf("Error loading help: %v", err)
		}

		// Also load from PowerShell to get all commands
		log.Println("Loading commands from PowerShell...")
		err = commandDatabase.LoadFromPowerShell()
		if err != nil {
			log.Printf("Error loading from PowerShell: %v", err)
		}

		// Update UI on main thread
		glib.IdleAdd(func() {
			if commandAddOn != nil {
				commandAddOn.loadModules()
				commandAddOn.updateCommandList()
			}
		})
	}()
}

// toggleCommandAddOn shows/hides the command add-on pane
func toggleCommandAddOn() {
	if commandAddOn == nil {
		return
	}

	// Prevent recursive calls from menu signal
	if updatingCommandAddonMenu {
		return
	}

	if commandAddOnVisible {
		// Hide the command add-on
		commandAddOn.container.Hide()
		commandAddOnVisible = false
	} else {
		// Show the command add-on
		commandAddOn.container.ShowAll()
		commandAddOnVisible = true

		// Set default width to ~20% of window (position is left edge of addon)
		if commandAddOnPane != nil {
			winWidth, _ := mainWindow.GetSize()
			defaultAddonWidth := winWidth / 4 // 25% for addon, so position is 75%
			commandAddOnPane.SetPosition(winWidth - defaultAddonWidth)
		}

		// Load data if not loaded yet
		if commandDatabase != nil && !commandDatabase.IsLoaded() {
			commandAddOn.loadModules()
			commandAddOn.updateCommandList()
		}
	}

	// Sync menu checkbox state
	if showCommandAddonMenuItem != nil {
		updatingCommandAddonMenu = true
		showCommandAddonMenuItem.SetActive(commandAddOnVisible)
		updatingCommandAddonMenu = false
	}
}
