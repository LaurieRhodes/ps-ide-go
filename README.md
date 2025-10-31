# PS-IDE-Go

A PowerShell ISE (Integrated Scripting Environment) clone for Linux, built with Go and GTK3.

![Version](https://img.shields.io/badge/version-0.1.0-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Go](https://img.shields.io/badge/go-1.21+-00ADD8)

## Features

✅ **Currently Working:**
- Multi-tabbed script editor (cream/yellow tab strip like Windows ISE)
- Integrated PowerShell console with ISE colors (dark blue background)
- Script execution (F5 - Run Script, F8 - Run Selection)
- File operations (New, Open, Save, Save As)
- Right-click Copy/Paste in console
- Toolbar matching Windows PowerShell ISE
- Status bar with cursor position
- Modified indicator (*) on unsaved tabs

## Quick Start

### Prerequisites

```bash
# Ubuntu/Debian/Mint
sudo apt install libgtk-3-dev libvte-2.91-dev

# Install PowerShell Core
# See: https://docs.microsoft.com/en-us/powershell/scripting/install/install-ubuntu
```

### Build & Run

```bash
cd /media/laurie/Data/Github/ps-ide-go
go build -o ps-ide ./cmd/ps-ide
./ps-ide
```

Or use the provided scripts:
```bash
chmod +x build.sh
./build.sh
./ps-ide
```

## Project Structure

```
ps-ide-go/
├── cmd/ps-ide/
│   ├── main.go       # Application setup and utilities
│   ├── terminal.go   # VTE terminal integration
│   ├── tabs.go       # Tab management
│   ├── toolbar.go    # Toolbar buttons
│   ├── menu.go       # Menu bar
│   ├── fileops.go    # File operations
│   └── actions.go    # Script execution
├── PROJECT_CONTEXT.md  # Complete project documentation
├── go.mod
└── go.sum
```

## Documentation

**📖 For the next development session or contributors:**
- Read **[PROJECT_CONTEXT.md](PROJECT_CONTEXT.md)** - Contains everything you need to know about the project, including critical bugs to avoid

## Key Architecture Notes

- **GTK3** for UI (gotk3 v0.6.3)
- **VTE 2.91** for embedded terminal
- **PowerShell Core** (pwsh) as the shell
- Refactored into small, manageable files for easier AI development

## Known Issues & TODOs

See [PROJECT_CONTEXT.md](PROJECT_CONTEXT.md) for the complete list.

**Top Priorities:**
1. Implement keyboard shortcuts (F5, F8, Ctrl+N/O/S)
2. Add line numbers to editor
3. Wire up Cut/Copy/Paste toolbar buttons
4. Add PowerShell syntax highlighting
5. Implement Stop Operation button

## Development

**CRITICAL:** When working with GTK signal handlers, always use `interface{}` as the first parameter type to avoid type conversion crashes. See PROJECT_CONTEXT.md for details.

```go
// ✅ CORRECT
widget.Connect("button-press-event", func(_ interface{}, event *gdk.Event) bool {

// ❌ WRONG - causes crash
widget.Connect("button-press-event", func(w *gtk.Widget, event *gdk.Event) bool {
```

## License

MIT License - See [LICENSE](LICENSE) file

## Contributing

This project is under active development. Check [PROJECT_CONTEXT.md](PROJECT_CONTEXT.md) for the current status and contribution guidelines.

## Acknowledgments

- Inspired by Windows PowerShell ISE
- Built with [gotk3](https://github.com/gotk3/gotk3)
- Terminal powered by [VTE](https://wiki.gnome.org/Apps/Terminal/VTE)
