# Translation Layer Package

## Overview

The Translation Layer provides a clean, structured API for PowerShell interaction, eliminating the issues of direct interactive PowerShell integration.

## Status

✅ **Phase 1: Foundation** - Complete  
✅ **Phase 2A: CLIXML Parser** - Complete  
🔄 **Phase 2B: Advanced Features** - In Progress

## Files

1. **types.go** - Shared type definitions (streams, commands, output)
2. **prompt.go** - Independent PowerShell prompt generation  
3. **queue.go** - Command history with disk persistence
4. **session.go** - PowerShell session state tracking
5. **pipes.go** - Process communication via stdin/stdout
6. **layer.go** - Main Translation Layer orchestrator
7. **parser.go** - CLIXML and ANSI code parser (NEW in Phase 2A)

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "github.com/laurie/ps-ide-go/cmd/ps-ide/translation"
)

func main() {
    // Create Translation Layer
    tl, err := translation.New()
    if err != nil {
        log.Fatal(err)
    }
    defer tl.Shutdown()

    // Get prompt
    prompt := tl.GetPrompt()
    fmt.Println(prompt) // Output: PS /home/user>

    // Execute command
    err = tl.ExecuteCommand("Get-Date")
    if err != nil {
        log.Printf("Error: %v", err)
    }

    // Navigate history
    previousCmd := tl.GetHistoryUp()
    fmt.Println(previousCmd)
    
    // Parse output with colors
    output := "Some output text"
    parsed, _ := tl.ParseOutput(output)
    for _, p := range parsed {
        fmt.Printf("[%s] %s\n", p.Stream, p.Content)
    }
}
```

## Public API

### Creation & Lifecycle
- `New() (*TranslationLayer, error)` - Create new instance
- `Shutdown() error` - Clean shutdown with history save

### Command Execution
- `ExecuteCommand(cmd string) error` - Execute typed command
- `ExecuteScript(path string) error` - Execute .ps1 script file
- `ExecuteSelection(code string) error` - Execute selected text
- `StopExecution() error` - Send interrupt (Ctrl+C)

### Output Parsing (NEW)
- `ParseOutput(rawOutput string) ([]PSOutput, error)` - Parse CLIXML
- `FormatOutput(output PSOutput) string` - Format without ANSI
- `FormatOutputWithColors(output PSOutput) string` - Format with ANSI
- `GetParser() *OutputParser` - Get parser for advanced use

### Prompt Generation
- `GetPrompt() string` - Get plain text prompt
- `GetPromptANSI() string` - Get colored prompt (green)
- `SetPromptStyle(style PromptStyle)` - Change prompt style
- `SetRemoteHost(hostname string)` - Set for remote sessions

### History Navigation
- `GetHistoryUp() string` - Navigate backward
- `GetHistoryDown() string` - Navigate forward
- `ResetHistoryIndex()` - Reset to end
- `GetHistory() []CommandEntry` - Get all history
- `GetRecentHistory(n int) []CommandEntry` - Get recent N
- `ClearHistory() error` - Clear all history
- `SearchHistory(query string) []CommandEntry` - Search history

### Session State
- `GetCurrentDirectory() string` - Current working directory
- `GetPSVersion() string` - PowerShell version string
- `GetVariables() map[string]VariableInfo` - All tracked variables
- `GetFunctions() map[string]FunctionInfo` - All tracked functions
- `GetModules() []ModuleInfo` - All loaded modules
- `SyncState() error` - Force state synchronization

### IntelliSense
- `GetCompletions(prefix string) []string` - Basic completions

### Execution State
- `IsExecuting() bool` - Check if command is running

## Parser Features (Phase 2A)

### CLIXML Support
```go
parser := tl.GetParser()

// Parse PowerShell XML output
xmlData := []byte("<Objs>...</Objs>")
outputs, _ := parser.Parse(xmlData)

for _, output := range outputs {
    fmt.Printf("Stream: %s\n", output.Stream)
    fmt.Printf("Content: %s\n", output.Content)
    fmt.Printf("Has ANSI: %v\n", output.IsFormatted)
}
```

### ANSI Code Parsing
```go
// Parse ANSI color codes
text := "\x1b[31mRed Text\x1b[0m"
segments := parser.ParseANSI(text)

for _, seg := range segments {
    fmt.Printf("Text: %s, Color: %d, Bold: %v\n", 
        seg.Text, seg.FGColor, seg.Bold)
}
```

### Stream Detection
- Automatic detection of output streams
- Error (red), Warning (yellow), Verbose (cyan), Debug (magenta), etc.
- Stream-specific colors via `GetStreamColor(stream)`

### Helper Methods
```go
// Strip ANSI codes
clean := parser.StripANSI(text)

// Check for ANSI codes
hasColor := parser.HasANSICodes(text)

// Extract error messages
if output.Stream == ErrorStream {
    errMsg := parser.ExtractErrorMessage(output)
}
```

## Key Advantages

### Problems Solved ✅
- ✅ No command echo pollution
- ✅ No duplicate prompts
- ✅ No escape sequence junk
- ✅ Clean, structured communication
- ✅ Thread-safe operations
- ✅ Proper color support (Phase 2A)
- ✅ Stream separation (Phase 2A)

### New Capabilities ✅
- ✅ Persistent command history (`~/.ps-ide/history.json`)
- ✅ Session state tracking
- ✅ IntelliSense foundation
- ✅ Better error handling
- ✅ CLIXML parsing (Phase 2A)
- ✅ ANSI color interpretation (Phase 2A)

## Architecture

```
GTK UI → Translation Layer → Parser → Formatted Output
            ↓
          Pipes → PowerShell (XML) → Raw Output
```

### Data Flow
1. User types command in UI
2. Translation Layer adds to history
3. Pipes send to PowerShell (`pwsh -OutputFormat XML`)
4. PowerShell executes, returns CLIXML
5. Parser deserializes XML → PSOutput[]
6. UI displays with appropriate colors

## Configuration

### History
- Location: `~/.ps-ide/history.json`
- Max entries: 1000 (configurable)
- Persists across restarts
- Thread-safe access

### PowerShell Process
- Command: `pwsh -NoLogo -NoProfile -OutputFormat XML -Command -`
- Output format: XML (eliminates echo)
- Communication: stdin/stdout pipes
- Interrupt: SIGINT (Ctrl+C)

### Output Streams
Stream colors (via GTK TextTags in UI):
- **Error**: Bright Red (#FF6B6B) + Bold
- **Warning**: Bright Yellow (#FFD93D)
- **Verbose**: Cyan (#6BCF7F)
- **Debug**: Magenta (#C77DFF)
- **Information**: Green (#95E1D3)
- **Output**: White (#FFFFFF)
- **Prompt**: Green (#95E1D3)

## Type Reference

### PSOutput
```go
type PSOutput struct {
    Stream       StreamType    // Output/Error/Warning/etc
    Content      string        // Text content
    ANSISegments []ANSISegment // Parsed color segments
    ObjectData   interface{}   // For future object support
    IsFormatted  bool          // Has ANSI codes
    Timestamp    time.Time     // When received
}
```

### ANSISegment
```go
type ANSISegment struct {
    Text      string  // Text content
    FGColor   int     // Foreground color (30-37, 90-97)
    BGColor   int     // Background color (40-47, 100-107)
    Bold      bool    // Bold text
    Underline bool    // Underlined text
    Italic    bool    // Italic text
}
```

### StreamType
```go
const (
    OutputStream StreamType = iota
    ErrorStream
    WarningStream
    VerboseStream
    DebugStream
    ProgressStream
    InformationStream
)
```

## Testing

### Unit Test Example
```go
func TestParser() {
    parser := NewOutputParser()
    
    // Test plain text
    outputs := parser.parsePlainText("Hello World")
    assert.Equal(t, "Hello World", outputs[0].Content)
    
    // Test ANSI parsing
    segments := parser.ParseANSI("\x1b[31mRed\x1b[0m")
    assert.Equal(t, 31, segments[0].FGColor)
    
    // Test XML parsing
    xml := []byte("<Objs>...</Objs>")
    outputs, _ := parser.Parse(xml)
    assert.NotNil(t, outputs)
}
```

### Manual Testing
```bash
# Build
go build -o ps-ide ./cmd/ps-ide

# Run
./ps-ide

# In console, try:
Get-Date
Write-Host "Green" -ForegroundColor Green
Write-Warning "This is yellow"
Write-Error "This is red"
1..5 | ForEach-Object { "Item $_" }
```

## Next Steps

### Phase 2B: Console Enhancements
- [ ] Progress bar rendering
- [ ] Enhanced error display (stack traces)
- [ ] Tab completion in console
- [ ] Multi-line command support

### Phase 2C: Editor Enhancements  
- [ ] Line numbers in editor
- [ ] Enhanced PowerShell syntax highlighting
- [ ] IntelliSense dropdown
- [ ] Bracket/brace matching

### Phase 3: Advanced Features
- [ ] Debugging support (breakpoints, step)
- [ ] Remote PowerShell sessions
- [ ] Variable inspection panel
- [ ] Performance profiling

## Documentation

Project documentation:
- `/PHASE_2A_COMPLETE.md` - Phase 2A completion report
- `/BUILD_COMPLETE.md` - Full build status
- `/docs/TRANSLATION_LAYER_ARCHITECTURE.md` - Detailed architecture
- `/docs/IMPLEMENTATION_ROADMAP.md` - Development plan

## Troubleshooting

### PowerShell Not Found
```bash
sudo snap install powershell --classic
which pwsh  # Verify installation
```

### XML Not Parsing
- Check PowerShell version: `pwsh --version`
- Verify XML output: `pwsh -OutputFormat XML -Command 'Get-Date'`
- Look for `<Objs>` tags in output

### Colors Not Showing
- GTK TextTags may need initialization
- Check console tag creation in `createConsoleTags()`
- Verify stream type detection

### History Not Persisting
- Check `~/.ps-ide/history.json` exists
- Ensure write permissions
- Call `tl.Shutdown()` on exit

## Contributing

When modifying the Translation Layer:
1. Read `/docs/TRANSLATION_LAYER_ARCHITECTURE.md` first
2. Update types.go for new data structures
3. Add tests for new features
4. Update this README
5. Document in CHANGELOG.md

## License

Part of PS-IDE-Go project - MIT License

---

**Status**: Phase 2A Complete ✅  
**Version**: 0.2.0  
**Last Updated**: 2025-11-03
