# Quick Start Guide - PS-IDE-Go v1.0.0

**Platform:** Optimized for Linux | **Status:** Production Ready

---

## For End Users (Recommended)

### One-Line Install (Linux)

```bash
curl -sSL https://raw.githubusercontent.com/LaurieRhodes/ps-ide-go/main/install.sh | bash
```

**What this does:**
- Installs PowerShell and GTK3 runtime
- Downloads latest binary for your architecture
- Installs to `~/.local/bin`
- Ready to run in 2 minutes!

**Then run:**
```bash
ps-ide
```

### Manual Binary Install (Linux)

If you prefer manual control:

```bash
# 1. Install runtime dependencies
sudo apt install powershell libgtk-3-0  # Ubuntu/Debian
# or: sudo dnf install powershell gtk3   # Fedora
# or: sudo pacman -S powershell gtk3     # Arch

# 2. Download binary
wget https://github.com/LaurieRhodes/ps-ide-go/releases/latest/download/ps-ide-linux-amd64.tar.gz

# 3. Extract and install
tar xzf ps-ide-linux-amd64.tar.gz
chmod +x ps-ide
sudo mv ps-ide /usr/local/bin/
# or: mv ps-ide ~/.local/bin/

# 4. Run
ps-ide
```

---

## For Developers (Build from Source)

### Prerequisites
- Go 1.21 or higher
- PowerShell 7.x
- GTK3 development libraries
- C compiler (gcc)

### Quick Build (Linux)

```bash
# Clone repository
git clone https://github.com/LaurieRhodes/ps-ide-go.git
cd ps-ide-go

# Install development dependencies
./install-dev-deps.sh

# Build
make build

# Run
./ps-ide
```

### Manual Build

```bash
# Install dependencies
sudo apt install golang-go powershell \
    build-essential pkg-config \
    libgtk-3-dev libglib2.0-dev libcairo2-dev libpango1.0-dev

# Get Go modules
go mod download

# Build
go build -o ps-ide ./cmd/ps-ide

# Run
./ps-ide
```

---

## First Look - Interface Overview

### Main Window Layout

```
┌─────────────────────────────────────────────┐
│ File  Edit  Run  View  Help         [×]     │ ← Menu Bar
├─────────────────────────────────────────────┤
│ [New] [Open] [Save] [Run] [Stop]  [Find]    │ ← Toolbar
├─────────────────────────────────────────────┤
│ script1.ps1  script2.ps1  [+]               │ ← Tab Bar
├─────────────────────────────────────────────┤
│ 1  │ # PowerShell Script                    │
│ 2  │ Write-Host "Hello World"               │ ← Code Editor
│ 3  │ Get-Date                                │   (with line numbers)
│ 4  │                                         │
├─────────────────────────────────────────────┤
│ PS> Write-Host "Hello World"                │
│ Hello World                                 │ ← Integrated Console
│ PS> _                                       │   (PowerShell output)
└─────────────────────────────────────────────┘
```

### Key Features

- **Multi-Tab Editor** - Work with multiple scripts simultaneously
- **Integrated Console** - Full PowerShell console with command history
- **Syntax Highlighting** - Powered by Chroma (150+ languages)
- **Code Snippets** - 18 built-in PowerShell templates
- **Find & Replace** - Search within your scripts
- **Native GTK3 UI** - Fast, responsive Linux interface

---

## Your First Script

### Quick Test

1. **Launch PS-IDE-Go**
   ```bash
   ps-ide
   ```

2. **Type this in the editor:**
   ```powershell
   Write-Host "Hello from PS-IDE-Go!" -ForegroundColor Green
   Get-Date
   $PSVersionTable.PSVersion
   ```

3. **Run the script:**
   - Click the **Run** button (toolbar)
   - Or press `F5`
   - Or menu: **Run → Run Script**

4. **See the output** in the console pane below

### Try a Real Script

Create a system information script:

```powershell
# System Information Script
Write-Host "=== System Information ===" -ForegroundColor Cyan

# OS Information
Write-Host "`nOperating System:" -ForegroundColor Yellow
$PSVersionTable.OS

# PowerShell Version
Write-Host "`nPowerShell Version:" -ForegroundColor Yellow
$PSVersionTable.PSVersion

# Current User
Write-Host "`nCurrent User:" -ForegroundColor Yellow
whoami

# Current Directory
Write-Host "`nCurrent Directory:" -ForegroundColor Yellow
Get-Location

# Available Modules
Write-Host "`nAvailable PowerShell Modules:" -ForegroundColor Yellow
Get-Module -ListAvailable | Select-Object -First 5 Name, Version
```

**Save it:**
- `Ctrl+S` or click **Save**
- Name it `system-info.ps1`
- Run it with `F5`

---

## Essential Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+N` | New file |
| `Ctrl+O` | Open file |
| `Ctrl+S` | Save file |
| `Ctrl+Shift+S` | Save as |
| `Ctrl+W` | Close current tab |
| `Ctrl+Tab` | Switch between tabs |
| `F5` | Run script |
| `Ctrl+C` | Stop execution |
| `Ctrl+F` | Find |
| `Ctrl+H` | Replace |
| `Ctrl+J` | Insert code snippet |
| `Ctrl+Z` | Undo |
| `Ctrl+Y` | Redo |
| `Ctrl+/` | Toggle comment |
| `Ctrl++` | Zoom in |
| `Ctrl+-` | Zoom out |

---

## Working with Tabs

### Creating Tabs
- **New tab:** `Ctrl+N` or click **New** button
- **Open file in new tab:** `Ctrl+O`

### Switching Tabs
- **Next tab:** `Ctrl+Tab`
- **Click tab header** to switch

### Closing Tabs
- **Close current:** `Ctrl+W`
- **Middle-click** on tab header
- **Right-click → Close** on tab header

### Tab Context Menu (Right-Click)
- **Close** - Close this tab
- **Close Other Tabs** - Close all except this one
- **Close All Tabs** - Close all tabs
- **Copy Full Path** - Copy file path to clipboard

---

## Using Code Snippets

**Access snippets:** Press `Ctrl+J`

### Built-in PowerShell Snippets

| Snippet | Description |
|---------|-------------|
| cmdlet | Basic cmdlet structure |
| function | Simple function |
| advanced-function | Advanced function with parameters |
| param | Parameter block |
| if-else | If-else statement |
| switch | Switch statement |
| foreach | ForEach loop |
| for | For loop |
| while | While loop |
| do-while | Do-While loop |
| try-catch | Try-Catch block |
| class | PowerShell class |
| enum | PowerShell enum |
| region | Collapsible region |
| comment-help | Comment-based help |
| pipeline | Pipeline example |
| filter | Filter function |
| validate | Parameter validation |

**How to use:**
1. Press `Ctrl+J`
2. Select snippet from dialog
3. Snippet inserts at cursor position
4. Edit placeholder values

---

## Find & Replace

### Find in Current File

**Open:** `Ctrl+F`

**Features:**
- Case-sensitive search
- Whole word matching
- Regular expressions support
- Navigate results with **Next/Previous**

### Replace in Current File

**Open:** `Ctrl+H`

**Features:**
- Replace single occurrence
- Replace all occurrences
- Preview before replace
- Undo support

---

## Integrated Console

### Console Features

- **Full PowerShell console** - Not just output display
- **Command history** - Up/Down arrows to navigate
- **Multi-line commands** - Supports complex scripts
- **ANSI colors** - Full color support
- **Error handling** - Errors display in red
- **Progress indicators** - Shows script execution

### Console Commands

Type PowerShell commands directly in console:

```powershell
PS> Get-Process | Where-Object CPU -gt 100
PS> Get-ChildItem -Recurse *.ps1
PS> $env:PATH
```

### Clear Console

- **Menu:** Edit → Clear Console
- **Command:** `Clear-Host` or `cls` in console

---

## File Operations

### Opening Files

**Single file:**
- `Ctrl+O` → Select file
- Drag & drop file onto window
- Command line: `ps-ide script.ps1`

**Recent files:**
- Menu: **File → Recent Files**
- Shows last 10 opened files

### Saving Files

**Save current:**
- `Ctrl+S`
- Click **Save** button

**Save as new file:**
- `Ctrl+Shift+S`
- Choose location and name

**Save all open tabs:**
- Menu: **File → Save All**

### File Indicators

- **Modified:** Tab shows `*` after filename
- **Unsaved:** Tab background changes color
- **Close prompt:** Asked to save when closing modified files

---

## Configuration

### Config File Location

`~/.config/ps-ide/config.json`

### Default Settings

```json
{
  "fontSize": 11,
  "fontFamily": "Consolas",
  "theme": "monokai",
  "tabSize": 4,
  "wordWrap": false,
  "lineNumbers": true,
  "windowWidth": 1200,
  "windowHeight": 800,
  "executionTimeout": 120,
  "powerShellPath": "pwsh",
  "syntaxEngine": "chroma",
  "recentFiles": []
}
```

### Customization

**Edit manually:**
```bash
nano ~/.config/ps-ide/config.json
```

**Common tweaks:**
- `fontSize`: 9-16 (default: 11)
- `fontFamily`: "Consolas", "Liberation Mono", "DejaVu Sans Mono"
- `theme`: "monokai", "github", "vim", "vs"
- `tabSize`: 2, 4, 8
- `wordWrap`: true/false
- `executionTimeout`: seconds (default: 120)

**Apply changes:** Restart PS-IDE-Go

---

## Troubleshooting

### Linux

#### Application Won't Start

**Error:** "Command not found: ps-ide"
```bash
# Add to PATH
echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.bashrc
source ~/.bashrc
```

**Error:** "PowerShell (pwsh) not found"
```bash
# Install PowerShell
sudo apt install powershell  # Ubuntu/Debian
sudo dnf install powershell  # Fedora
sudo pacman -S powershell    # Arch
```

**Error:** "error while loading shared libraries: libgtk-3.so.0"
```bash
# Install GTK3 runtime
sudo apt install libgtk-3-0  # Ubuntu/Debian
sudo dnf install gtk3        # Fedora
sudo pacman -S gtk3          # Arch
```

#### Script Won't Execute

**Check PowerShell works:**
```bash
pwsh -c "Write-Host 'Test'"
```

**Check script permissions:**
```bash
ls -l script.ps1
# Should be readable
```

**Try simple script first:**
```powershell
Write-Host "Hello World"
```

#### Display Issues

**Blurry text:**
- Increase font size in config
- Check display scaling settings

**Wrong colors:**
- Try different theme in config
- Check terminal color scheme

### Build Issues

**Error:** "Cannot find package"
```bash
go mod download
go mod tidy
```

**Error:** "C compiler not found"
```bash
sudo apt install build-essential
```

**Error:** "GTK headers not found"
```bash
sudo apt install libgtk-3-dev libglib2.0-dev
```

---

## Platform-Specific Notes

### Linux (Primary Platform) ✅

**Status:** Fully supported, optimized, production-ready

**Best experience on:**
- Ubuntu 20.04+
- Fedora 35+
- Arch Linux
- Any modern Linux with GTK3

### macOS (Experimental) ⚠️

**Requirements:**
```bash
brew install powershell gtk+3
```

**Limitations:**
- Not native macOS look and feel
- GTK3 via Homebrew required
- May have rendering quirks

**Recommendation:** Consider VS Code for native macOS experience

### Windows (Experimental) ⚠️

**Requirements:**
- MSYS2 environment
- GTK3 runtime (~200MB)
- Complex setup

**Recommendation:** Use **Visual Studio Code** with PowerShell extension for better Windows experience. PS-IDE-Go is optimized for Linux.

---

## Tips & Best Practices

### Script Development

1. **Start simple** - Test with basic scripts first
2. **Use snippets** - `Ctrl+J` for quick templates
3. **Test incrementally** - Run frequently during development
4. **Check errors** - Console shows detailed error messages
5. **Use variables** - `$Error` for last error details

### Code Organization

```powershell
#region Configuration
# Put configuration variables here
$LogPath = "/var/log/myscript.log"
$MaxRetries = 3
#endregion

#region Functions
function Get-MyData {
    # Function code here
}
#endregion

#region Main Script
# Main execution logic
Get-MyData
#endregion
```

### Performance

- **Long scripts:** Increase `executionTimeout` in config
- **Large output:** Redirect to file if output > 10,000 lines
- **Background jobs:** Use `Start-Job` for long operations

### File Management

- **Naming:** Use descriptive names (`backup-database.ps1`)
- **Location:** Organize by project/function
- **Version control:** Use Git for script history
- **Templates:** Save common patterns as snippets

---

## Example Workflows

### Daily Admin Tasks

```powershell
# Morning System Check
Get-Process | Where-Object CPU -gt 50 | Select-Object Name, CPU, Id
Get-Service | Where-Object Status -eq 'Stopped' | Select-Object Name, Status
Get-Disk | Select-Object Number, FriendlyName, Size, PartitionStyle
```

### Log Analysis

```powershell
# Parse recent logs
Get-Content /var/log/syslog | Select-Object -Last 100 | Where-Object { $_ -match "error" }
```

### Backup Script

```powershell
# Simple backup
$Source = "/home/user/documents"
$Destination = "/backup/documents-$(Get-Date -Format 'yyyy-MM-dd')"
Copy-Item -Path $Source -Destination $Destination -Recurse
```

---

## Learning Resources

### PowerShell

- **Official Docs:** https://docs.microsoft.com/powershell
- **Learn PowerShell:** https://github.com/PowerShell/PowerShell/tree/master/docs/learning-powershell
- **Community:** r/PowerShell on Reddit

### PS-IDE-Go

- **Repository:** https://github.com/LaurieRhodes/ps-ide-go
- **Issues:** Report bugs and request features
- **Discussions:** Ask questions and share scripts

### GTK3 (for developers)

- **GTK Documentation:** https://www.gtk.org/docs/
- **gotk3 Library:** https://github.com/gotk3/gotk3

---

## What's Next?

### Explore Features

- [x] Try all keyboard shortcuts
- [x] Insert different code snippets
- [x] Open multiple tabs
- [x] Use find & replace
- [x] Customize configuration
- [x] Try console commands

### Advanced Usage

- Write complex scripts with functions
- Use PowerShell modules
- Integrate with system services
- Create automation workflows
- Share scripts with team

### Contribute

PS-IDE-Go is open source! Contributions welcome:

- Report bugs on GitHub
- Suggest features
- Submit pull requests
- Share your experience
- Help other users

---

## Version Information

**Current Version:** v1.0.0  
**Release Date:** February 2026  
**Platform:** Linux (primary), macOS/Windows (experimental)  
**Status:** Production Ready

**What's included in v1.0.0:**
- ✅ Multi-tab editor
- ✅ Integrated PowerShell console
- ✅ Syntax highlighting (Chroma)
- ✅ Code snippets (18 built-in)
- ✅ Find & replace
- ✅ Keyboard shortcuts
- ✅ Tab management
- ✅ Native GTK3 UI
- ✅ Zero configuration

---

**Happy scripting with PS-IDE-Go!** 🚀🐧

*Optimized for Linux. Built by Linux users, for Linux users.*
