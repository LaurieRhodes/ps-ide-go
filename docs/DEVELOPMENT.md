# Development Guide

## Setup Development Environment

### Prerequisites
```bash
# Install Go (if not already installed)
sudo apt update
sudo apt install golang-go

# Install PowerShell Core
sudo apt install powershell

# Install GUI development dependencies
sudo apt install gcc libgl1-mesa-dev xorg-dev
```

### Clone and Build
```bash
cd /media/laurie/Data/Github/ps-ide-go

# Initialize Go modules
go mod init github.com/laurie/ps-ide-go
go mod tidy

# Build the application
go build -o ps-ide ./cmd/ps-ide

# Run the application
./ps-ide
```

## Project Architecture

### Directory Structure

- **cmd/ps-ide**: Main application entry point
- **internal/**: Private application packages
  - **editor/**: Text editor logic and file operations
  - **executor/**: PowerShell script execution
  - **highlighter/**: Syntax highlighting using Chroma
  - **ui/**: GUI components and window management
- **pkg/**: Public reusable packages
  - **config/**: Configuration management
- **assets/**: Icons, themes, and resources
- **docs/**: Documentation files

### Key Components

#### Editor Package
Manages text editing state, file I/O operations, and editor settings.

#### Executor Package
Handles PowerShell script execution using `os/exec`. Supports:
- Direct script execution
- File-based script execution
- Timeout management
- Stdout/stderr capture

#### Highlighter Package
Provides syntax highlighting using the Chroma library. Features:
- PowerShell syntax support
- Multiple color themes
- Basic syntax validation

#### UI Package
Implements the graphical interface using Fyne toolkit:
- Main window management
- Menu and toolbar
- File dialogs
- Split pane layout

#### Config Package
Manages application settings:
- JSON-based configuration
- Recent files tracking
- Window size/position
- Editor preferences

## Building and Testing

### Build Commands
```bash
# Development build
go build -o ps-ide ./cmd/ps-ide

# Build with version info
go build -ldflags="-X 'main.Version=0.1.0'" -o ps-ide ./cmd/ps-ide

# Cross-compile for different platforms
GOOS=linux GOARCH=amd64 go build -o ps-ide-linux ./cmd/ps-ide
GOOS=windows GOARCH=amd64 go build -o ps-ide.exe ./cmd/ps-ide
```

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests verbosely
go test -v ./...
```

## Development Workflow

### Phase 1: MVP (Current)
- [x] Basic project structure
- [x] Simple text editor
- [x] PowerShell execution
- [x] File open/save
- [ ] Basic syntax highlighting integration
- [ ] Testing and bug fixes

### Phase 2: Enhanced Features
- [ ] Multiple file tabs
- [ ] Integrated PowerShell console
- [ ] Search and replace
- [ ] Keyboard shortcuts
- [ ] Recent files menu
- [ ] Better error handling

### Phase 3: Advanced Features
- [ ] IntelliSense/autocomplete
- [ ] Debugging support
- [ ] Code snippets
- [ ] Customizable themes
- [ ] Plugin system

## Adding New Features

### Adding a New Menu Item
1. Edit `internal/ui/mainwindow.go`
2. Add menu item in `setupMenu()` function
3. Create handler function
4. Test functionality

### Adding Keyboard Shortcuts
Use Fyne's keyboard shortcuts:
```go
ctrlS := &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: desktop.ControlModifier}
w.Canvas().AddShortcut(ctrlS, func(shortcut fyne.Shortcut) {
    mw.saveFile()
})
```

### Customizing Themes
1. Create new theme in `internal/ui/theme.go`
2. Implement `fyne.Theme` interface
3. Apply theme: `app.Settings().SetTheme(myTheme)`

## Debugging

### Enable Debug Logging
```bash
# Set Fyne debug environment variable
export FYNE_DEBUG=1
./ps-ide
```

### Common Issues

**Issue**: GUI doesn't appear
- Check that X11 forwarding is enabled (if using SSH)
- Verify OpenGL support: `glxinfo | grep OpenGL`

**Issue**: PowerShell not found
- Verify installation: `which pwsh`
- Check PATH environment variable

**Issue**: Compilation errors
- Run: `go mod tidy`
- Update dependencies: `go get -u ./...`

## Code Style

### Go Conventions
- Follow standard Go formatting: `gofmt -w .`
- Use `golint` for linting: `golint ./...`
- Keep functions small and focused
- Add comments for exported functions

### Commit Messages
```
type(scope): subject

- Added feature X
- Fixed bug Y
- Improved Z

Types: feat, fix, docs, style, refactor, test, chore
```

## Performance Considerations

- Use goroutines for long-running operations
- Implement proper context cancellation
- Cache syntax highlighting results
- Limit output buffer size
- Use efficient data structures

## Contributing

1. Fork the repository
2. Create feature branch: `git checkout -b feature/my-feature`
3. Make changes and test
4. Commit: `git commit -am 'Add my feature'`
5. Push: `git push origin feature/my-feature`
6. Create Pull Request

## Resources

- [Fyne Documentation](https://developer.fyne.io/)
- [Chroma Documentation](https://github.com/alecthomas/chroma)
- [PowerShell Documentation](https://docs.microsoft.com/en-us/powershell/)
- [Go Documentation](https://golang.org/doc/)
