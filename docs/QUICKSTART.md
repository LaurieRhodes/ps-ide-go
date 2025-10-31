# Quick Start Guide

## First Time Setup

### 1. Install Dependencies
```bash
cd /media/laurie/Data/Github/ps-ide-go
chmod +x install-deps.sh build.sh
./install-deps.sh
```

This will install:
- Go programming language
- PowerShell Core (pwsh)
- GUI development libraries (gcc, OpenGL, X11)

### 2. Build the Application
```bash
# Option 1: Using the build script
./build.sh

# Option 2: Using Make
make build

# Option 3: Manual build
go mod download
go build -o ps-ide ./cmd/ps-ide
```

### 3. Run the Application
```bash
./ps-ide
```

## Quick Tour

### Main Interface
- **Top**: Toolbar with New, Open, Save, and Run buttons
- **Middle**: Code editor (top pane) - write your PowerShell scripts here
- **Bottom**: Output pane - see script execution results here

### Writing Your First Script

1. Type or paste this in the editor:
```powershell
Write-Host "Hello from PS-IDE-Go!"
Get-Date
$PSVersionTable.PSVersion
```

2. Click the **Run** button or use menu: Run → Run Script

3. View the output in the bottom pane

### File Operations

**Create New File**
- Click **New** button
- Or: File → New

**Open Existing File**
- Click **Open** button
- Or: File → Open
- Navigate to a .ps1 file

**Save File**
- Click **Save** button
- Or: File → Save
- First save will prompt for location

### Example Scripts

A sample script is included at `assets/sample.ps1`:
1. Click **Open**
2. Navigate to `/media/laurie/Data/Github/ps-ide-go/assets/sample.ps1`
3. Click **Run** to see it in action

## Common Tasks

### Execute Selected Text (Coming in Phase 2)
Currently, the entire script is executed. Future versions will support running selected portions.

### Clear Output
- Menu: Edit → Clear Output

### Check PowerShell Version
In the editor, type and run:
```powershell
$PSVersionTable
```

## Keyboard Shortcuts (Future)

Phase 2 will include:
- `Ctrl+N` - New file
- `Ctrl+O` - Open file
- `Ctrl+S` - Save file
- `F5` - Run script
- `Ctrl+F` - Find
- `Ctrl+H` - Replace

## Troubleshooting

### Application Won't Start
**Error**: "PowerShell (pwsh) not found"
- Install PowerShell: `sudo apt install powershell`
- Verify: `which pwsh`

**Error**: GUI doesn't appear
- Check X11: `echo $DISPLAY`
- Test OpenGL: `glxinfo | grep "OpenGL version"`
- Install missing packages: `sudo apt install libgl1-mesa-dev xorg-dev`

### Script Execution Issues

**Error**: "Script not executing"
- Verify PowerShell works in terminal: `pwsh -c "Write-Host 'test'"`
- Check timeout setting (default 30 seconds)

**Error**: "Permission denied"
- Ensure script has content
- Try a simple script first: `Write-Host "test"`

### Build Issues

**Error**: "Cannot find package"
- Run: `go mod download`
- Run: `go mod tidy`

**Error**: "gcc not found"
- Install: `sudo apt install gcc`

## Configuration

Configuration is stored at: `~/.config/ps-ide/config.json`

Default settings:
```json
{
  "fontSize": 12,
  "theme": "monokai",
  "tabSize": 4,
  "wordWrap": false,
  "lineNumbers": true,
  "windowWidth": 900,
  "windowHeight": 700,
  "executionTimeout": 30,
  "powerShellPath": "pwsh",
  "recentFiles": []
}
```

You can manually edit this file to customize settings.

## Tips & Best Practices

### Script Development
1. Start with simple scripts to test functionality
2. Use `Write-Host` for debugging output
3. Check `$Error` variable for error details
4. Test scripts in terminal first for complex operations

### File Management
1. Use `.ps1` extension for scripts
2. Organize scripts in folders
3. Use meaningful file names
4. Recent files will appear in config (Phase 2 will add UI)

### Performance
1. Default timeout is 30 seconds
2. For long-running scripts, increase timeout in config
3. Large output may slow the UI - consider redirecting to files

## Next Steps

### Explore PowerShell
```powershell
# Get help on any command
Get-Help Get-Process

# List available cmdlets
Get-Command

# Explore modules
Get-Module -ListAvailable
```

### Customize Your Experience
1. Edit config file for preferences
2. Create script templates in a dedicated folder
3. Bookmark frequently used scripts

### Learn More
- PowerShell docs: https://docs.microsoft.com/powershell
- Fyne GUI docs: https://developer.fyne.io/
- Project docs: `/media/laurie/Data/Github/ps-ide-go/docs/`

## Roadmap

### Phase 1 (Current - MVP)
- ✅ Basic editor
- ✅ Script execution
- ✅ File operations
- 🔲 Syntax highlighting integration
- 🔲 Polish and bug fixes

### Phase 2 (Coming Soon)
- Multiple file tabs
- Integrated console
- Search and replace
- Keyboard shortcuts
- Recent files menu

### Phase 3 (Future)
- IntelliSense/autocomplete
- Debugging support
- Code snippets
- Custom themes
- Extensions/plugins

## Getting Help

- Check docs: `docs/DEVELOPMENT.md`
- Review sample: `assets/sample.ps1`
- Test PowerShell: `pwsh --help`

## Contributing

Have ideas or found bugs? This is a personal learning project, but suggestions are welcome!

---

**Happy scripting with PS-IDE-Go!** 🚀
