# PowerShell IDE Translation Layer Architecture

**Project**: ps-ide-go  
**Document Version**: 1.0  
**Last Updated**: 2024  
**Purpose**: Define architecture for PowerShell console integration using translation layer pattern

---

## Executive Summary

The PowerShell IDE requires a robust translation layer to bridge the GTK UI with a background PowerShell process. The current implementation using `-Interactive` mode with stdout parsing is fundamentally flawed due to command echo pollution and unpredictable output timing.

**Solution**: Implement a translation layer that:
- Uses named pipes for bidirectional communication
- Employs XML output format for structured data
- Manages prompt generation and command history independently
- Maintains session state for IntelliSense and UI updates

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    PS-IDE-Go (GTK UI)                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │   Editor     │  │   Console    │  │   Toolbar    │ │
│  │   Pane       │  │   TextView   │  │   Controls   │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
└───────────────────────────┬─────────────────────────────┘
                            │
                  ┌─────────▼─────────┐
                  │ Translation Layer │
                  │   (Go Package)    │
                  └─────────┬─────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
   ┌────▼────┐      ┌──────▼──────┐      ┌────▼────┐
   │ Command │      │   Output    │      │ Session │
   │  Queue  │      │   Parser    │      │  State  │
   └─────────┘      └─────────────┘      └─────────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            │
                    ┌───────▼───────┐
                    │  Named Pipes  │
                    │  Communicator │
                    └───────┬───────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
   ┌────▼────┐         ┌───▼───┐         ┌────▼────┐
   │ Input   │         │ Output│         │  Error  │
   │  Pipe   │         │  Pipe │         │  Pipe   │
   └────┬────┘         └───┬───┘         └────┬────┘
        │                  │                   │
        └──────────────────┼───────────────────┘
                           │
            ┌──────────────▼──────────────┐
            │   Hidden PowerShell Process │
            │   pwsh -NoLogo -NoProfile   │
            │   -OutputFormat XML         │
            └─────────────────────────────┘
```

---

## Core Components

### 1. Translation Layer (Main Orchestrator)

**Location**: `cmd/ps-ide/translation/layer.go`

**Responsibilities**:
- Coordinate all components
- Handle command execution workflow
- Manage UI updates
- Control prompt display

**Interface**:
```go
type TranslationLayer interface {
    // Execution
    ExecuteCommand(cmd string) error
    ExecuteScript(path string) error
    ExecuteSelection(text string) error
    
    // State
    GetCurrentDirectory() string
    GetPrompt() string
    IsExecuting() bool
    
    // History
    GetHistory() []CommandEntry
    NavigateHistory(direction int) string
    
    // Lifecycle
    Start() error
    Stop() error
}
```

**Key Methods**:
```go
func (tl *TranslationLayer) ExecuteCommand(cmd string) error {
    // 1. Validate input
    // 2. Add to history queue
    // 3. Send via named pipe
    // 4. Parse XML response
    // 5. Display output with formatting
    // 6. Update session state
    // 7. Display new prompt
}
```

---

### 2. Named Pipe Communicator

**Location**: `cmd/ps-ide/translation/pipes.go`

**Purpose**: Handle bidirectional communication with PowerShell process via named pipes.

**Structure**:
```go
type PipeCommunicator struct {
    inputPipe    *os.File
    outputPipe   *os.File
    psProcess    *exec.Cmd
    mutex        sync.Mutex
    commandQueue chan PipeCommand
}

type PipeCommand struct {
    ID      string
    Command string
    Type    CommandType
}

type PipeResponse struct {
    ID       string
    Stream   StreamType
    Data     []byte
    Complete bool
}
```

**Pipe Naming Convention**:
```
Input:  /tmp/ps-ide-input-{pid}   (or Windows equivalent)
Output: /tmp/ps-ide-output-{pid}
```

**PowerShell Startup Command**:
```bash
pwsh -NoLogo -NoProfile -OutputFormat XML -Command -
```

**Communication Protocol**:
1. Write command to input pipe with unique ID
2. Read CLIXML from output pipe until completion marker
3. Deserialize and return structured data

**Key Methods**:
```go
func (pc *PipeCommunicator) Start() error
func (pc *PipeCommunicator) SendCommand(cmd string) (string, error)
func (pc *PipeCommunicator) ReadOutput() ([]PSOutput, error)
func (pc *PipeCommunicator) Stop() error
```

---

### 3. Command Queue Manager

**Location**: `cmd/ps-ide/translation/queue.go`

**Purpose**: Maintain command history for Up/Down arrow navigation, like PSReadLine.

**Structure**:
```go
type CommandQueue struct {
    history      []CommandEntry
    currentIndex int
    maxSize      int
    persistence  *HistoryPersistence
    mutex        sync.RWMutex
}

type CommandEntry struct {
    Command      string
    Timestamp    time.Time
    Type         CommandType  // Interactive, Script, Selection
    WorkingDir   string
    Duration     time.Duration
    Success      bool
}

type CommandType int
const (
    Interactive CommandType = iota
    Script
    Selection
)
```

**Persistence**:
- Store in `~/.ps-ide/history.json`
- Load on startup
- Max 1000 entries (configurable)

**Key Methods**:
```go
func (cq *CommandQueue) Add(cmd string, cmdType CommandType) error
func (cq *CommandQueue) GetPrevious() (string, bool)
func (cq *CommandQueue) GetNext() (string, bool)
func (cq *CommandQueue) GetAll() []CommandEntry
func (cq *CommandQueue) Clear() error
func (cq *CommandQueue) Save() error
func (cq *CommandQueue) Load() error
```

**Special Handling**:
- Scripts stored as single entry: `& '/path/to/script.ps1'`
- Multi-line commands preserved with newlines
- Duplicate consecutive commands not added

---

### 4. Output Parser (CLIXML Deserializer)

**Location**: `cmd/ps-ide/translation/parser.go`

**Purpose**: Deserialize PowerShell CLIXML output into structured Go types.

**Structure**:
```go
type OutputParser struct {
    decoder *xml.Decoder
}

type PSOutput struct {
    Stream       StreamType
    Content      string
    ANSISegments []ANSISegment
    ObjectData   interface{}
    IsFormatted  bool
}

type StreamType int
const (
    OutputStream StreamType = iota
    ErrorStream
    WarningStream
    VerboseStream
    DebugStream
    ProgressStream
)

type ANSISegment struct {
    Text      string
    FGColor   int
    BGColor   int
    Bold      bool
    Underline bool
}
```

**CLIXML Format Reference**:
```xml
<Objs Version="1.1.0.1" xmlns="http://schemas.microsoft.com/powershell/2004/04">
  <Obj RefId="0">
    <TN RefId="0">
      <T>System.String</T>
    </TN>
    <ToString>Hello World</ToString>
  </Obj>
</Objs>
```

**Key Methods**:
```go
func (op *OutputParser) Parse(xmlData []byte) ([]PSOutput, error)
func (op *OutputParser) ParseANSI(text string) []ANSISegment
func (op *OutputParser) ExtractStream(obj CLIXMLObject) StreamType
```

**ANSI Code Support**:
- Parse SGR sequences: `\x1b[<code>m`
- Support foreground colors (30-37, 90-97)
- Support background colors (40-47, 100-107)
- Support bold (1), underline (4), reset (0)

---

### 5. Session State Manager

**Location**: `cmd/ps-ide/translation/session.go`

**Purpose**: Track PowerShell session state for prompt generation and IntelliSense.

**Structure**:
```go
type SessionState struct {
    CurrentDirectory string
    Variables        map[string]VariableInfo
    Functions        map[string]FunctionInfo
    Modules          []ModuleInfo
    PSVersion        string
    LastExitCode     int
    ErrorActionPref  string
    LastSync         time.Time
    mutex            sync.RWMutex
}

type VariableInfo struct {
    Name        string
    Type        string
    Value       string
    Description string
}

type FunctionInfo struct {
    Name       string
    Parameters []string
    Synopsis   string
}

type ModuleInfo struct {
    Name    string
    Version string
    Path    string
}
```

**Update Strategy**:
1. **After each command**: Query `$PWD` for directory
2. **Periodic sync (5s)**: Query variables, functions, modules
3. **On-demand**: When IntelliSense requests data

**Key Methods**:
```go
func (ss *SessionState) UpdateDirectory() error
func (ss *SessionState) SyncVariables() error
func (ss *SessionState) SyncFunctions() error
func (ss *SessionState) GetVariable(name string) (VariableInfo, bool)
func (ss *SessionState) GetCompletions(prefix string) []string
```

**PowerShell Queries**:
```powershell
# Directory
(Get-Location).Path

# Variables
Get-Variable | Select-Object Name, Value | ConvertTo-Json

# Functions
Get-Command -CommandType Function | Select-Object Name | ConvertTo-Json

# Modules
Get-Module | Select-Object Name, Version | ConvertTo-Json
```

---

### 6. Prompt Generator

**Location**: `cmd/ps-ide/translation/prompt.go`

**Purpose**: Generate PowerShell-style prompts independently of PowerShell process.

**Structure**:
```go
type PromptGenerator struct {
    template string
    state    *SessionState
}
```

**Default Format**:
```
PS /path/to/directory>
```

**Remote Session Format**:
```
[RemoteServer]: PS C:\Users\Username>
```

**Key Methods**:
```go
func (pg *PromptGenerator) Generate() string
func (pg *PromptGenerator) SetTemplate(template string)
func (pg *PromptGenerator) GetANSIColored() string  // Returns prompt with green ANSI codes
```

**Color Codes**:
- Prompt color: `\x1b[32m` (Green) like PowerShell default
- Path simplified: `/home/user` → `~`

---

### 7. IntelliSense Manager

**Location**: `cmd/ps-ide/translation/intellisense.go`

**Purpose**: Track objects and variables for auto-completion.

**Structure**:
```go
type IntelliSenseManager struct {
    sessionState *SessionState
    cache        *CompletionCache
    mutex        sync.RWMutex
}

type CompletionCache struct {
    Variables []string
    Commands  []string
    Paths     []string
    LastUpdate time.Time
}
```

**Key Methods**:
```go
func (ism *IntelliSenseManager) GetCompletions(partial string) []Completion
func (ism *IntelliSenseManager) UpdateCache() error
func (ism *IntelliSenseManager) AddVariable(name string)
```

**Future Enhancement**: Tab completion in console TextView

---

## Critical Design Decisions

### 1. XML vs Text Output Format

**Decision**: Use XML (`-OutputFormat XML`)

**Rationale**:
- **Perfect echo suppression**: Commands never in output stream
- **Stream separation**: Clear distinction between output/error/warning
- **Type preservation**: Objects maintain type info for IntelliSense
- **Color preservation**: ANSI codes embedded in ToString()
- **Progress support**: ProgressRecord objects captured

**Reference**: PowerShell source `ConsoleHost.cs` lines 340-380

### 2. Named Pipes vs Stdin/Stdout

**Decision**: Use Named Pipes

**Rationale**:
- **Bidirectional**: Clear separation of input/output channels
- **No buffering issues**: Each pipe is independent
- **Process isolation**: PowerShell runs in background
- **Scalability**: Can add more pipes (debug, verbose, etc.)

**Platform Support**:
- Linux: `/tmp/ps-ide-*`
- Windows: `\\.\pipe\ps-ide-*`

### 3. Independent Prompt Generation

**Decision**: Generate prompts in Go, not PowerShell

**Rationale**:
- **No query latency**: Instant prompt display
- **No echo pollution**: Prompt not in output stream
- **Full control**: Can customize format easily
- **ISE consistency**: Matches Windows ISE behavior

**Reference**: Windows ISE does not query PowerShell for prompts

### 4. Command Queue in Translation Layer

**Decision**: Manage history in Go, not PowerShell

**Rationale**:
- **Persistence**: Save across sessions
- **UI control**: Up/Down arrows handled in GTK
- **Script tracking**: Store entire scripts as single entries
- **No PSReadLine dependency**: Works without module

---

## Implementation Phases

### Phase 1: Foundation (Current Priority)
- [ ] Create `translation/` package structure
- [ ] Implement `PipeCommunicator` with basic send/receive
- [ ] Build `CommandQueue` with persistence
- [ ] Create `SessionState` with directory tracking
- [ ] Implement `PromptGenerator`

### Phase 2: Output Handling
- [ ] Implement CLIXML parser
- [ ] Parse ANSI color codes
- [ ] Handle multiple streams (error, warning, etc.)
- [ ] Display formatted output in GTK TextView

### Phase 3: UI Integration
- [ ] Replace current `psconsole.go` with translation layer
- [ ] Wire Up/Down arrows to command history
- [ ] Implement script execution via translation layer
- [ ] Add progress bar display

### Phase 4: Advanced Features
- [ ] IntelliSense manager
- [ ] Tab completion
- [ ] Variable tracking
- [ ] Debugging support (future)

---

## File Structure

```
cmd/ps-ide/
├── main.go
├── psconsole.go           # UI layer - will be refactored
├── translation/
│   ├── layer.go           # Main orchestrator (TranslationLayer)
│   ├── pipes.go           # Named pipe communication
│   ├── queue.go           # Command history queue
│   ├── session.go         # Session state tracking
│   ├── parser.go          # CLIXML deserializer
│   ├── prompt.go          # Prompt generator
│   ├── intellisense.go    # Auto-completion support
│   └── types.go           # Shared type definitions
├── actions.go
├── fileops.go
├── menu.go
├── tabs.go
└── toolbar.go
```

---

## PowerShell Source Code References

**Repository**: https://github.com/PowerShell/PowerShell

**Key Files**:
1. `src/Microsoft.PowerShell.ConsoleHost/host/msh/ConsoleHost.cs`
   - Lines 1805-1825: Prompt evaluation
   - Lines 892-946: PSReadLine integration
   - Lines 340-380: Output format handling
   - Lines 560-600: CLIXML serialization

2. `src/System.Management.Automation/engine/remoting/` 
   - Remote session communication via XML

**Key Insights**:
- ISE does NOT use `-Interactive` mode
- ISE generates its own prompts
- ISE uses XML serialization internally
- PSReadLine is optional module, not core

---

## Testing Strategy

### Unit Tests
- `pipes_test.go`: Named pipe creation, send/receive
- `queue_test.go`: History management, persistence
- `parser_test.go`: CLIXML deserialization
- `session_test.go`: State updates, queries

### Integration Tests
- End-to-end command execution
- Multi-line script handling
- Error stream separation
- Progress bar display

### Manual Testing Checklist
- [ ] Clean startup (single prompt)
- [ ] Command execution (no echo)
- [ ] Script execution (F5)
- [ ] Selection execution (F8)
- [ ] Clear console (cls)
- [ ] Up/Down arrow history
- [ ] Error display (red text)
- [ ] Multi-line output
- [ ] Progress bars
- [ ] Ctrl+C interrupt

---

## Known Issues from Current Implementation

1. **Command echoing**: Commands appear in output (FIXED with XML)
2. **Duplicate prompts**: Prompt detected twice (FIXED with independent generation)
3. **Escape sequence pollution**: `[?1h`, `[?1l` appear (FIXED with XML)
4. **Timing issues**: Race conditions in byte-by-byte reading (FIXED with structured parsing)
5. **No state tracking**: Can't determine current directory reliably (FIXED with SessionState)

---

## Future Enhancements

### Debugging Support
- Breakpoint management
- Step through execution
- Variable inspection
- Call stack display

### Remote Sessions
- Connect to remote PowerShell
- Display remote hostname in prompt
- Handle network latency gracefully

### Advanced IntelliSense
- Tab completion for cmdlets
- Parameter hints
- Type-ahead suggestions
- Snippet support

### Performance Optimization
- Command batching
- Output buffering strategies
- Lazy variable sync
- Cache completion data

---

## References

1. **PowerShell ISE Behavior**:
   - Windows PowerShell ISE v5.1 (closed source)
   - Observed behavior for prompt, colors, history

2. **PowerShell Documentation**:
   - https://docs.microsoft.com/en-us/powershell/
   - CLIXML format specification
   - Remoting protocol documentation

3. **Go Libraries**:
   - `encoding/xml`: CLIXML parsing
   - `os/exec`: Process management
   - `github.com/gotk3/gotk3`: GTK bindings

---

## Glossary

- **CLIXML**: Command Line Interface XML - PowerShell's serialization format
- **Translation Layer**: Intermediary between UI and PowerShell process
- **Named Pipes**: IPC mechanism for bidirectional communication
- **Stream**: PowerShell output channel (output, error, warning, etc.)
- **PSReadLine**: PowerShell module for enhanced command-line editing
- **IntelliSense**: Auto-completion and suggestion system

---

## Document Maintenance

**Update Triggers**:
- Architecture changes
- New components added
- Design decisions revised
- Implementation phase completion

**Review Schedule**: After each major milestone

**Stakeholders**: AI assistants implementing features, future maintainers

---

*End of Architecture Document*
