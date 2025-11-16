package main

import (
	"fmt"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/laurie/ps-ide-go/cmd/ps-ide/translation"
)

var (
	translationLayer *translation.TranslationLayer
	consoleTextView  *gtk.TextView
	consoleTextBuffer *gtk.TextBuffer
	promptMark       *gtk.TextMark
	consoleTags      map[string]*gtk.TextTag
)

func initTranslationLayer() error {
	tl, err := translation.New()
	if err != nil {
		return fmt.Errorf("failed to create translation layer: %w", err)
	}
	
	translationLayer = tl
	
	// Display initial prompt after a short delay
	glib.TimeoutAdd(500, func() bool {
		displayPrompt()
		return false
	})
	
	return nil
}

func shutdownTranslationLayer() {
	if translationLayer != nil {
		translationLayer.Shutdown()
	}
}

func createConsoleUI() (*gtk.ScrolledWindow, error) {
	textView, _ := gtk.TextViewNew()
	textView.SetEditable(true)
	textView.SetWrapMode(gtk.WRAP_WORD_CHAR)
	textView.SetMonospace(true)
	textView.SetLeftMargin(5)
	textView.SetRightMargin(5)
	textView.SetCursorVisible(true)
	
	buffer, _ := textView.GetBuffer()
	
	// Create text tags for different output streams
	createConsoleTags(buffer)
	
	scroll, _ := gtk.ScrolledWindowNew(nil, nil)
	scroll.Add(textView)
	scroll.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	
	applyConsoleColors(textView)
	
	consoleTextView = textView
	consoleTextBuffer = buffer
	
	textView.Connect("key-press-event", onConsoleKeyPress)
	
	textView.AddEvents(int(gdk.BUTTON_PRESS_MASK))
	textView.Connect("button-press-event", func(_ interface{}, event *gdk.Event) bool {
		if gdk.EventButtonNewFromEvent(event).Button() == 3 {
			showConsoleContextMenu(event)
			return true
		}
		return false
	})
	
	return scroll, nil
}

// createConsoleTags creates text tags for styling different output streams
func createConsoleTags(buffer *gtk.TextBuffer) {
	consoleTags = make(map[string]*gtk.TextTag)
	
	// Error stream - bright red
	errorTag := buffer.CreateTag("error", map[string]interface{}{
		"foreground": "#FF6B6B",
		"weight":     700,
	})
	consoleTags["error"] = errorTag
	
	// Warning stream - bright yellow
	warningTag := buffer.CreateTag("warning", map[string]interface{}{
		"foreground": "#FFD93D",
	})
	consoleTags["warning"] = warningTag
	
	// Verbose stream - cyan
	verboseTag := buffer.CreateTag("verbose", map[string]interface{}{
		"foreground": "#6BCF7F",
	})
	consoleTags["verbose"] = verboseTag
	
	// Debug stream - magenta
	debugTag := buffer.CreateTag("debug", map[string]interface{}{
		"foreground": "#C77DFF",
	})
	consoleTags["debug"] = debugTag
	
	// Information stream - green
	infoTag := buffer.CreateTag("information", map[string]interface{}{
		"foreground": "#95E1D3",
	})
	consoleTags["information"] = infoTag
	
	// Default output - white
	outputTag := buffer.CreateTag("output", map[string]interface{}{
		"foreground": "#FFFFFF",
	})
	consoleTags["output"] = outputTag
	
	// Prompt - green (like PowerShell)
	promptTag := buffer.CreateTag("prompt", map[string]interface{}{
		"foreground": "#95E1D3",
	})
	consoleTags["prompt"] = promptTag
}

func displayPrompt() {
	if translationLayer == nil || consoleTextBuffer == nil {
		return
	}
	
	prompt := translationLayer.GetPrompt()
	
	endIter := consoleTextBuffer.GetEndIter()
	startOffset := endIter.GetOffset()
	
	consoleTextBuffer.Insert(endIter, prompt)
	endIter = consoleTextBuffer.GetEndIter()
	
	// Apply prompt tag for green color
	if promptTag, ok := consoleTags["prompt"]; ok {
		startIter := consoleTextBuffer.GetIterAtOffset(startOffset)
		consoleTextBuffer.ApplyTag(promptTag, startIter, endIter)
	}
	
	if promptMark != nil {
		consoleTextBuffer.DeleteMark(promptMark)
	}
	promptMark = consoleTextBuffer.CreateMark("prompt", endIter, true)
	
	consoleTextBuffer.PlaceCursor(endIter)
	consoleTextView.ScrollToIter(endIter, 0.0, false, 0.0, 0.0)
}

func displayOutput(text string) {
	if consoleTextBuffer == nil || translationLayer == nil {
		return
	}
	
	// Skip empty output
	if strings.TrimSpace(text) == "" {
		return
	}
	
	// Parse output using the translation layer's parser
	parsedOutput, err := translationLayer.ParseOutput(text)
	if err != nil {
		// If parsing fails, display as plain text
		endIter := consoleTextBuffer.GetEndIter()
		consoleTextBuffer.Insert(endIter, text+"\n")
		consoleTextView.ScrollToIter(consoleTextBuffer.GetEndIter(), 0.0, false, 0.0, 0.0)
		return
	}
	
	// Display each parsed output with appropriate formatting
	for _, output := range parsedOutput {
		displayParsedOutput(output)
	}
	
	consoleTextView.ScrollToIter(consoleTextBuffer.GetEndIter(), 0.0, false, 0.0, 0.0)
}

func displayParsedOutput(output translation.PSOutput) {
	if consoleTextBuffer == nil {
		return
	}
	
	// If output has ANSI segments, display each segment with its color
	if output.IsFormatted && len(output.ANSISegments) > 0 {
		for _, segment := range output.ANSISegments {
			if segment.Text == "" {
				continue
			}
			
			endIter := consoleTextBuffer.GetEndIter()
			startOffset := endIter.GetOffset()
			
			consoleTextBuffer.Insert(endIter, segment.Text)
			
			// Apply color based on ANSI foreground color
			color := getColorFromANSI(segment.FGColor)
			if color != "" {
				// Create or get tag for this color
				tagName := fmt.Sprintf("ansi-%d", segment.FGColor)
				tag, exists := consoleTags[tagName]
				if !exists {
					tag = consoleTextBuffer.CreateTag(tagName, map[string]interface{}{
						"foreground": color,
						"weight":     getWeightFromSegment(segment),
					})
					consoleTags[tagName] = tag
				}
				
				startIter := consoleTextBuffer.GetIterAtOffset(startOffset)
				endIter = consoleTextBuffer.GetEndIter()
				consoleTextBuffer.ApplyTag(tag, startIter, endIter)
			}
		}
		// Add newline after all segments
		consoleTextBuffer.Insert(consoleTextBuffer.GetEndIter(), "\n")
		return
	}
	
	// No ANSI codes - use stream-based coloring
	var tag *gtk.TextTag
	switch output.Stream {
	case translation.ErrorStream:
		tag = consoleTags["error"]
	case translation.WarningStream:
		tag = consoleTags["warning"]
	case translation.VerboseStream:
		tag = consoleTags["verbose"]
	case translation.DebugStream:
		tag = consoleTags["debug"]
	case translation.InformationStream:
		tag = consoleTags["information"]
	default:
		tag = consoleTags["output"]
	}
	
	// Format the output
	formattedText := output.Content
	
	// Ensure newline at end if not present
	if formattedText != "" && !strings.HasSuffix(formattedText, "\n") {
		formattedText += "\n"
	}
	
	// Insert with appropriate tag
	endIter := consoleTextBuffer.GetEndIter()
	startOffset := endIter.GetOffset()
	
	consoleTextBuffer.Insert(endIter, formattedText)
	
	if tag != nil {
		startIter := consoleTextBuffer.GetIterAtOffset(startOffset)
		endIter = consoleTextBuffer.GetEndIter()
		consoleTextBuffer.ApplyTag(tag, startIter, endIter)
	}
}

// getColorFromANSI converts ANSI color codes to hex colors
func getColorFromANSI(ansiCode int) string {
	switch ansiCode {
	case 30, 90: // Black / Bright Black
		return "#808080"
	case 31: // Red
		return "#CD0000"
	case 91: // Bright Red
		return "#FF0000"
	case 32: // Green
		return "#00CD00"
	case 92: // Bright Green
		return "#00FF00"
	case 33: // Yellow
		return "#CDCD00"
	case 93: // Bright Yellow
		return "#FFFF00"
	case 34: // Blue
		return "#0000EE"
	case 94: // Bright Blue
		return "#5C5CFF"
	case 35: // Magenta
		return "#CD00CD"
	case 95: // Bright Magenta
		return "#FF00FF"
	case 36: // Cyan
		return "#00CDCD"
	case 96: // Bright Cyan
		return "#00FFFF"
	case 37: // White
		return "#E5E5E5"
	case 97: // Bright White
		return "#FFFFFF"
	default:
		return "#FFFFFF" // Default white
	}
}

// getWeightFromSegment returns the font weight based on segment attributes
func getWeightFromSegment(segment translation.ANSISegment) int {
	if segment.Bold {
		return 700 // Bold
	}
	return 400 // Normal
}

func displayRawOutput(text string, streamType translation.StreamType) {
	if consoleTextBuffer == nil {
		return
	}
	
	// Get the appropriate tag
	var tag *gtk.TextTag
	switch streamType {
	case translation.ErrorStream:
		tag = consoleTags["error"]
	case translation.WarningStream:
		tag = consoleTags["warning"]
	case translation.VerboseStream:
		tag = consoleTags["verbose"]
	case translation.DebugStream:
		tag = consoleTags["debug"]
	case translation.InformationStream:
		tag = consoleTags["information"]
	default:
		tag = consoleTags["output"]
	}
	
	endIter := consoleTextBuffer.GetEndIter()
	startOffset := endIter.GetOffset()
	
	consoleTextBuffer.Insert(endIter, text)
	
	if tag != nil {
		startIter := consoleTextBuffer.GetIterAtOffset(startOffset)
		endIter = consoleTextBuffer.GetEndIter()
		consoleTextBuffer.ApplyTag(tag, startIter, endIter)
	}
	
	consoleTextView.ScrollToIter(consoleTextBuffer.GetEndIter(), 0.0, false, 0.0, 0.0)
}

func getUserInput() string {
	if promptMark == nil {
		return ""
	}
	
	text, _ := consoleTextBuffer.GetText(
		consoleTextBuffer.GetIterAtMark(promptMark),
		consoleTextBuffer.GetEndIter(),
		false)
	
	return text
}

func clearUserInput() {
	if promptMark != nil {
		consoleTextBuffer.Delete(
			consoleTextBuffer.GetIterAtMark(promptMark),
			consoleTextBuffer.GetEndIter())
	}
}

func onConsoleKeyPress(_ interface{}, event *gdk.Event) bool {
	if translationLayer == nil {
		return false
	}
	
	keyEvent := gdk.EventKeyNewFromEvent(event)
	keyval := keyEvent.KeyVal()
	state := keyEvent.State()
	
	if translationLayer.IsExecuting() {
		if keyval == gdk.KEY_c && (state&uint(gdk.CONTROL_MASK)) != 0 {
			translationLayer.StopExecution()
			displayRawOutput("\n^C\n", translation.WarningStream)
			glib.IdleAdd(func() bool {
				displayPrompt()
				return false
			})
			return true
		}
		return true
	}
	
	if keyval == gdk.KEY_Up {
		cmd := translationLayer.GetHistoryUp()
		clearUserInput()
		if cmd != "" {
			consoleTextBuffer.Insert(consoleTextBuffer.GetEndIter(), cmd)
		}
		return true
	}
	
	if keyval == gdk.KEY_Down {
		cmd := translationLayer.GetHistoryDown()
		clearUserInput()
		if cmd != "" {
			consoleTextBuffer.Insert(consoleTextBuffer.GetEndIter(), cmd)
		}
		return true
	}
	
	if keyval == gdk.KEY_BackSpace && promptMark != nil {
		cursorIter := consoleTextBuffer.GetIterAtMark(consoleTextBuffer.GetInsert())
		promptIter := consoleTextBuffer.GetIterAtMark(promptMark)
		if cursorIter.Compare(promptIter) <= 0 {
			return true
		}
	}
	
	if keyval == gdk.KEY_Return || keyval == gdk.KEY_KP_Enter {
		input := getUserInput()
		consoleTextBuffer.Insert(consoleTextBuffer.GetEndIter(), "\n")
		
		go executeCommand(input)
		
		return true
	}
	
	if promptMark != nil {
		cursorIter := consoleTextBuffer.GetIterAtMark(consoleTextBuffer.GetInsert())
		promptIter := consoleTextBuffer.GetIterAtMark(promptMark)
		if cursorIter.Compare(promptIter) < 0 {
			consoleTextBuffer.PlaceCursor(consoleTextBuffer.GetEndIter())
		}
	}
	
	return false
}

func executeCommand(cmd string) {
	cmd = strings.TrimSpace(cmd)
	
	if cmd == "" {
		glib.IdleAdd(func() bool {
			displayPrompt()
			return false
		})
		return
	}
	
	if cmd == "clear" || cmd == "cls" {
		glib.IdleAdd(func() bool {
			clearConsole()
			return false
		})
		return
	}
	
	// Execute command and get output
	output, err := translationLayer.ExecuteCommand(cmd)
	
	glib.IdleAdd(func() bool {
		if err != nil {
			displayRawOutput(fmt.Sprintf("Error: %v\n", err), translation.ErrorStream)
		} else {
			// Display the output
			displayOutput(output)
		}
		displayPrompt()
		return false
	})
}

func clearConsole() {
	if consoleTextBuffer == nil {
		return
	}
	
	consoleTextBuffer.Delete(
		consoleTextBuffer.GetStartIter(),
		consoleTextBuffer.GetEndIter())
	promptMark = nil
	displayPrompt()
}

func applyConsoleColors(textView *gtk.TextView) {
	provider, _ := gtk.CssProviderNew()
	provider.LoadFromData(`textview {
		background-color: #012456;
		color: #FFFFFF;
		font-family: "Courier New", "Lucida Console", monospace;
		font-size: 10pt;
		padding: 5px;
		caret-color: #FFFFFF;
	}
	textview text {
		background-color: #012456;
		color: #FFFFFF;
	}
	textview:selected {
		background-color: #004080;
	}`)
	
	styleContext, _ := textView.GetStyleContext()
	styleContext.AddProvider(provider, gtk.STYLE_PROVIDER_PRIORITY_USER)
}

func showConsoleContextMenu(event *gdk.Event) {
	menu, _ := gtk.MenuNew()
	
	copyItem, _ := gtk.MenuItemNewWithLabel("Copy")
	copyItem.Connect("activate", func() { copyConsoleSelection() })
	menu.Append(copyItem)
	
	pasteItem, _ := gtk.MenuItemNewWithLabel("Paste")
	pasteItem.Connect("activate", func() { pasteToConsole() })
	menu.Append(pasteItem)
	
	clearItem, _ := gtk.MenuItemNewWithLabel("Clear")
	clearItem.Connect("activate", func() { clearConsole() })
	menu.Append(clearItem)
	
	menu.ShowAll()
	menu.PopupAtPointer(event)
}

func copyConsoleSelection() {
	if consoleTextBuffer != nil {
		if start, end, hasSelection := consoleTextBuffer.GetSelectionBounds(); hasSelection {
			text, _ := consoleTextBuffer.GetText(start, end, false)
			clipboard, _ := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
			clipboard.SetText(text)
		}
	}
}

func pasteToConsole() {
	if consoleTextBuffer != nil {
		clipboard, _ := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
		if text, _ := clipboard.WaitForText(); text != "" {
			consoleTextBuffer.Insert(consoleTextBuffer.GetEndIter(), text)
		}
	}
}
