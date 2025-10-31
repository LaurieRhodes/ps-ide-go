# Next Session Checklist

## Before You Start

1. ✅ Read PROJECT_CONTEXT.md (mandatory - contains critical bug info)
2. ✅ Read README.md for quick overview
3. ✅ Verify build works: `go build -o ps-ide ./cmd/ps-ide`
4. ✅ Test run: `./ps-ide`

## Active Files (DO NOT MODIFY ARCHIVED FILES)

### Source Code (in cmd/ps-ide/)

- main.go - Application setup
- terminal.go - VTE terminal (contains critical `interface{}` fix)
- tabs.go - Tab management
- toolbar.go - Toolbar buttons  
- menu.go - Menu bar
- fileops.go - File operations
- actions.go - Script execution

### Documentation

- PROJECT_CONTEXT.md - **START HERE** - Complete project state
- README.md - Public-facing documentation
- CHANGELOG.md - Version history

## Immediate Priorities (Pick One)

### Option 1: Keyboard Shortcuts (High Impact, Medium Difficulty)

- F5 for Run Script
- F8 for Run Selection
- Ctrl+N/O/S for File operations
- See GTK accelerators documentation

### Option 2: Line Numbers with Script display (High Impact, High Difficulty)

- Need GtkSourceView integration (attempted before, has issues)
- Or implement custom line number widget
- See PROJECT_CONTEXT.md for past attempt details

### Option 3: Cut/Copy/Paste Buttons (Low Impact, Easy)

- Hex Color CCFFFF color change for toolbar background
- Hex Color FFE5CC color change for tab menu background
- Toolbar Button Save icon change to Floppy Disk
- Wire toolbar buttons to editor operations
- gtk.Clipboard operations
- Enable/disable based on text selection

### Option 4: Syntax Highlighting (Medium Impact, High Difficulty)

- GtkSourceView integration
- PowerShell language definition
- May conflict with line numbers

## Critical Warnings

⚠️ **GTK Signal Handler Type Issue** (causes crash)

```go
// WRONG - crashes on click
widget.Connect("button-press-event", func(w *gtk.Widget, event *gdk.Event) bool {

// RIGHT - always use interface{}
widget.Connect("button-press-event", func(_ interface{}, event *gdk.Event) bool {
```

⚠️ **Tab Index Tracking**

- Close button callbacks must capture correct tab index
- Current implementation has potential race condition
- See tabs.go createTabLabel()

⚠️ **Script Execution Method**

- Must use clipboard paste, not line-by-line
- See actions.go runScript() for working implementation

## Testing Checklist

Before committing changes:

- [ ] Application builds without errors
- [ ] Can create new tabs
- [ ] Can switch between tabs
- [ ] Each tab shows different content
- [ ] Run Script shows output correctly
- [ ] Run Selection works with selected text
- [ ] File open/save works
- [ ] Right-click in terminal shows menu
- [ ] No crashes on left-click in terminal

## Git Status

Current branch: (check with `git branch`)
Uncommitted changes: (check with `git status`)

## Questions to Ask User

1. Which feature should we prioritize next?
2. Are there any bugs or issues with current functionality?
3. Do you want to match Windows ISE exactly, or make improvements?

## Resources

- GTK3 Docs: https://docs.gtk.org/gtk3/
- gotk3 Examples: https://github.com/gotk3/gotk3-examples
- VTE Docs: https://gnome.pages.gitlab.gnome.org/vte/
- PowerShell: https://docs.microsoft.com/en-us/powershell/

---

**Session prepared and ready. Good luck!** 🚀
