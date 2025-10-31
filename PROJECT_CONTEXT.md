# PS-IDE-Go Project Context

## Project Overview
PS-IDE-Go is a PowerShell ISE (Integrated Scripting Environment) clone for Linux, built using Go and GTK3. The goal is to replicate the functionality and appearance of Windows PowerShell ISE on Linux systems.

## Current Status
The application is functional with core features working:
- ✅ Tabbed script editor with multiple scripts
- ✅ PowerShell console integration (VTE terminal)
- ✅ Script execution (Run Script F5, Run Selection F8)
- ✅ File operations (New, Open, Save, Save As)
- ✅ Windows ISE-like UI (cream tab background #FFFFCC, white tabs)
- ✅ Right-click copy/paste in terminal
- ✅ Status bar with cursor position
- ✅ Toolbar with tooltips matching Windows ISE

## Technical Architecture

### Technology Stack
- **Language**: Go
- **GUI Framework**: GTK3 (gotk3 v0.6.3)
- **Terminal**: VTE 2.91 (libvte-2.91)
- **Shell**: PowerShell Core (pwsh)

### Project Structure
```
ps-ide-go/
├── cmd/ps-ide/
│   ├── main.go         # Main application setup, CSS, utilities
│   ├── terminal.go     # VTE terminal creation and operations
│   ├── tabs.go         # Tab management (create, close, labels)
│   ├── toolbar.go      # Toolbar with all buttons
│   ├── menu.go         # Menu bar
│   ├── fileops.go      # File operations (new, open, save)
│   └── actions.go      # Script execution (run, run selection, clear)
├── go.mod
└── go.sum
```

### Key Dependencies
```go
require github.com/gotk3/gotk3 v0.6.3
```

### CGO Requirements
```
#cgo pkg-config: vte-2.91
#include <vte/vte.h>
#include <pango/pango.h>
```

## Critical Implementation Details

### 1. GTK Signal Handler Type Conversion Issue
**CRITICAL BUG TO AVOID**: When connecting GTK signals that receive events, the first parameter MUST be `interface{}`, NOT a typed parameter like `*gtk.Widget`.

❌ **WRONG** (causes crash):
```go
widget.Connect("button-press-event", func(w *gtk.Widget, event *gdk.Event) bool {
```

✅ **CORRECT**:
```go
widget.Connect("button-press-event", func(_ interface{}, event *gdk.Event) bool {
```

This is due to type conversion issues in gotk3's closure handling. Using a typed first parameter causes:
```
panic: reflect.Value.Convert: value of type *glib.Object cannot be converted to type *gtk.Widget
```

### 2. Script Execution Method
PowerShell ISE sends scripts as a complete block, not line-by-line. Implementation uses clipboard paste:

```go
// Copy script to clipboard
clipboard, _ := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
clipboard.SetText(script)

// Paste into terminal (preserves multi-line format)
pasteToTerminal()

// Execute with Enter
time.AfterFunc(100*time.Millisecond, func() {
    sendToTerminal("\r")
})
```

This produces correct output where the script is shown, then executed as a block.

### 3. Tab Management
Each tab has its own:
- `*gtk.TextView` (editor widget)
- `*gtk.TextBuffer` (text content)
- `filename` (file path, empty for unsaved)
- `modified` flag (shows asterisk in tab title)

Tabs are stored in `openTabs []* ScriptTab` slice, synchronized with notebook pages.

### 4. VTE Terminal Configuration
```go
// PowerShell ISE colors
bgColor := C.GdkRGBA{red: 0.0, green: 0.0, blue: 0x66 / 255.0, alpha: 1.0}  // Dark blue #000066
fgColor := C.GdkRGBA{red: 1.0, green: 1.0, blue: 1.0, alpha: 1.0}            // White

// Spawn PowerShell
argv := []*C.char{
    C.CString("pwsh"),
    C.CString("-NoLogo"),
    nil,
}
```

### 5. CSS Styling
```css
notebook {
    background-color: #FFFFCC;  /* Cream/yellow background */
}
notebook > header > tabs > tab {
    background-color: #FFFFFF;  /* White tabs */
    border: 1px solid #CCCCCC;
}
```

## Known Issues & TODOs

### Not Yet Implemented
1. **Line numbers** - Need to add GtkSourceView properly or implement custom line number display
2. **Syntax highlighting** - PowerShell syntax coloring
3. **Cut/Copy/Paste toolbar buttons** - Connected to editor text operations
4. **Undo/Redo** - Text buffer undo/redo functionality
5. **Stop Operation button** - Needs to send Ctrl+C to terminal
6. **New Remote PowerShell Tab** - Remote session management
7. **Start PowerShell in separate window** - Launch external PowerShell window
8. **Keyboard shortcuts** - F5 (Run), F8 (Run Selection), Ctrl+N, Ctrl+O, Ctrl+S
9. **Debugger support** - Breakpoints, step through, variables panel
10. **IntelliSense/Auto-completion** - Code completion features
11. **Command Add-ons menu** - ISE add-on system

### Fixed Issues
- ✅ Tab switching now shows different content for each tab
- ✅ Right-click menu in terminal (Copy/Paste)
- ✅ Script execution sends as block (not line-by-line)
- ✅ Toolbar tooltips match Windows ISE exactly

## Build & Run

### Prerequisites
```bash
sudo apt install libgtk-3-dev libvte-2.91-dev
# Install PowerShell Core
```

### Build
```bash
cd /media/laurie/Data/Github/ps-ide-go
go build -o ps-ide ./cmd/ps-ide
```

### Run
```bash
./ps-ide
```

## Windows PowerShell ISE Reference

### Toolbar Buttons (Left to Right)
1. New Script (document-new icon)
2. Open Script (document-open icon)
3. Save Script (document-save icon)
4. Cut (edit-cut icon) - grayed unless text selected
5. Copy (edit-copy icon) - grayed unless text selected
6. Paste (edit-paste icon)
7. Clear Console Pane (edit-clear icon)
8. Undo (edit-undo icon)
9. Redo (edit-redo icon) - grayed until undo used
10. **Run Script (F5)** (media-playback-start icon, should be GREEN)
11. Run Selection (F8) (media-skip-forward icon) - grayed unless text selected
12. Stop Operation (Ctrl+Break) (process-stop icon, RED) - grayed unless running
13. New Remote PowerShell Tab
14. Start PowerShell in a separate window

### UI Colors
- Tab strip background: `#FFFFCC` (cream/yellow)
- Tabs: `#FFFFFF` (white)
- Console background: `#000066` (dark blue)
- Console text: `#FFFFFF` (white)

## Development Notes

### File Refactoring Strategy
The codebase was refactored into separate files to make it easier for AI to manage:
- Each file focuses on one specific responsibility
- Main.go kept minimal with just setup and utilities
- Terminal, tabs, toolbar, menu, file operations, and actions separated
- This prevents files from becoming too large for AI context windows

### Testing
Test scripts should verify:
1. Multi-line script execution displays correctly
2. Each tab maintains separate content
3. Modified indicator (*) shows on unsaved changes
4. File save/load preserves content
5. Run Selection only executes selected text

### Common Pitfalls
1. **Don't forget CGO headers** in files using VTE
2. **Always use `interface{}`** for GTK signal handler first parameter
3. **Tab index tracking** - Close button callbacks need to capture correct index
4. **Memory in closures** - Be careful with variable capture in Connect() callbacks

## Next Steps Priority
1. Implement keyboard shortcuts (F5, F8, Ctrl+N/O/S)
2. Add line numbers to editor (using GtkSourceView properly)
3. Wire up Cut/Copy/Paste toolbar buttons to editor
4. Implement Stop Operation button (send SIGINT to PowerShell)
5. Add PowerShell syntax highlighting
6. Implement Undo/Redo functionality

## Contact & Resources
- GTK3 Docs: https://docs.gtk.org/gtk3/
- gotk3 GitHub: https://github.com/gotk3/gotk3
- VTE Docs: https://gnome.pages.gitlab.gnome.org/vte/
- PowerShell Docs: https://docs.microsoft.com/en-us/powershell/
