# PowerShell IDE Translation Layer Architecture

## Document Version

- **Version**: 1.0
- **Date**: 2025-01-15
- **Status**: Design Approved - Implementation Required

## Executive Summary

This document defines the architecture for a **Translation Layer** that sits between the GTK UI and PowerShell processes in ps-ide-go. The current architecture (direct interactive PowerShell with stdout parsing) has fundamental flaws preventing reliable prompt display, echo suppression, and output control. The new architecture uses **Named Pipes with XML output** to achieve Windows PowerShell ISE-like behavior.

---

## Current Architecture Problems

### Issues with Current Implementation (`psconsole.go`)

1. **Unreliable Echo Suppression**
   - Interactive mode echoes commands unpredictably
   - Byte-by-byte reading creates race conditions
   - String matching for suppression is fragile
2. **Prompt Detection Failures**
   - Prompts detected by `>` pattern cause duplicates
   - Escape sequences (`[?1h`, `[?1l`) pollute output
   - No clear boundary between prompt and output
3. **Poor Separation of Concerns**
   - UI tightly coupled to PowerShell process
   - Command history mixed with PowerShell output
   - No independent state management
4. **Limited Extensibility**
   - Cannot support IntelliSense without parsing
   - Debugging features impossible
   - Remote sessions cannot be implemented

---

## New Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                       PS-IDE-Go GTK Application                 │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────────────────┐ │
│  │   Editor     │  │   Console    │  │    Toolbar/Menu       │ │
│  │   Pane       │  │   TextView   │  │    Controls           │ │
│  │ (sourceview) │  │  (custom)    │  │  (Run, Stop, etc)     │ │
│  └──────────────┘  └──────────────┘  └───────────────────────┘ │
└──────────────────────────┬──────────────────────────────────────┘
                           │
              ┌────────────▼──────────────┐
              │   Translation Layer API   │
              │    (Go Package)           │
              └────────────┬──────────────┘
                           │
       ┌───────────────────┼───────────────────────┐
       │                   │                       │
  ┌────▼─────┐     ┌──────▼───────┐      ┌───────▼────────┐
  │ Command  │     │   Session    │      │     Output     │
  │  Queue   │     │    State     │      │     Parser     │
  │ Manager  │     │   Manager    │      │   (CLIXML)     │
  └────┬─────┘     └──────┬───────┘      └───────┬────────┘
       │                  │                       │
       └──────────────────┼───────────────────────┘
                          │
              ┌───────────▼────────────┐
              │  Named Pipe Manager    │
              │   (IPC Controller)     │
              └───────────┬────────────┘
                          │
          ┌───────────────┼──────────────────┐
          │               │                  │
    ┌─────▼─────┐   ┌─────▼──────┐   ┌─────▼─────┐
    │   Input   │   │   Output   │   │  Control  │
    │   Pipe    │   │    Pipe    │   │   Pipe    │
    └─────┬─────┘   └─────┬──────┘   └─────┬─────┘
          │               │                 │
          └───────────────┼─────────────────┘
                          │
        ┌─────────────────▼──────────────────┐
        │   Hidden PowerShell Process        │
        │   pwsh -NoLogo -NoProfile          │
        │        -OutputFormat XML           │
        │   (Background, no console window)  │
        └────────────────────────────────────┘
```

---

## Core Design Decisions

### Decision 1: Use XML Output Format (Not Text)

**Rationale:**

- PowerShell's ISE and remoting use CLIXML serialization (see `ConsoleHost.cs:340-380`)
- Preserves object types, formatting, and stream metadata
- Built-in deserialization in `System.Management.Automation`
- **Eliminates echo problem**: Commands never appear in output

**Benefits:**

- Perfect echo suppression
- Stream separation (output/error/warning/verbose/debug)
- ANSI color codes preserved in serialized objects
- Progress bars as ProgressRecord objects
- Type information for IntelliSense

**PowerShell Command:**

bash

```bash
pwsh -NoLogo -NoProfile -OutputFormat XML -Command -
```

### Decision 2: Translation Layer Controls All Display

**Key Principle:** The Translation Layer is the single source of truth for what appears in the console.

**What This Means:**

- Translation Layer generates and displays prompts (not PowerShell)
- Command echoing is UI-controlled (user sees typing, not output)
- PowerShell process is hidden - only sends results
- Command history managed by Translation Layer

**Why ISE Does This:**

- ISE doesn't use `-Interactive` mode
- ISE generates its own prompts
- ISE manages command history separately
- PowerShell is background worker only

### Decision 3: Named Pipes for IPC

**Why Named Pipes:**

- Cross-platform (Windows, Linux, macOS)
- Bi-directional communication
- Can handle large data (scripts, objects)
- Standard Go support (`net.Pipe` or OS-specific)

**Pipe Structure:**

```
Input Pipe:   /tmp/ps-ide-input-{pid}   (Go → PowerShell)
Output Pipe:  /tmp/ps-ide-output-{pid}  (PowerShell → Go)
Control Pipe: /tmp/ps-ide-control-{pid} (Signals, state queries)
```

### Decision 4: Command Queue for History and Echo Suppression

**Purpose:**

1. Store all executed commands for Up/Down arrow navigation
2. Know exactly what was sent (no guessing on echoes)
3. Single-entry for scripts (not line-by-line)
4. Persist history across sessions

**Structure:**

go

```go
type CommandQueue struct {
    history      []CommandEntry
    currentIndex int
    maxSize      int
    persistPath  string
}

type CommandEntry struct {
    Command     string
    Timestamp   time.Time
    Type        CommandType  // Interactive, Script, Internal
    WorkingDir  string
    ExitCode    int
}
```

---

## Component Specifications

### Component 1: Translation Layer Core

**Location:** `cmd/ps-ide/translation/layer.go`

**Responsibilities:**

- Central coordinator for all components
- Exposes API to GTK UI
- Manages PowerShell process lifecycle
- Orchestrates command execution flow

**Public API:**

go

```go
type TranslationLayer struct {
    pipes       *PipeManager
    queue       *CommandQueue
    session     *SessionState
    parser      *OutputParser
    prompt      *PromptGenerator
    intellisense *IntelliSenseManager
}

// Initialize and start background PowerShell
func New() (*TranslationLayer, error)

// Execute user-typed command
func (tl *TranslationLayer) ExecuteCommand(cmd string) error

// Execute script file
func (tl *TranslationLayer) ExecuteScript(path string) error

// Execute selection from editor
func (tl *TranslationLayer) ExecuteSelection(code string) error

// Get current prompt string
func (tl *TranslationLayer) GetPrompt() string

// Navigate command history
func (tl *TranslationLayer) GetHistoryUp() string
func (tl *TranslationLayer) GetHistoryDown() string

// Stop execution (Ctrl+C)
func (tl *TranslationLayer) StopExecution() error

// Shutdown PowerShell process
func (tl *TranslationLayer) Shutdown() error
```

**Execution Flow:**

```
1. User types command → UI calls tl.ExecuteCommand()
2. Translation Layer adds to queue
3. Send command via Input Pipe
4. PowerShell executes, sends XML to Output Pipe
5. Parser deserializes XML
6. Translation Layer displays results
7. Update session state (PWD, variables)
8. Display new prompt
```

---

## Implementation Phases

### Phase 1: Foundation (Week 1)

- [ ] Create `translation` package structure
- [ ] Implement `PipeManager` (basic named pipes)
- [ ] Implement `CommandQueue` (history storage)
- [ ] Create `SessionState` (directory tracking only)
- [ ] Basic `TranslationLayer` API

**Deliverable:** Can execute simple commands, see XML output

### Phase 2: Output Parsing (Week 2)

- [ ] Implement `OutputParser` (CLIXML deserializer)
- [ ] Parse ANSI color codes
- [ ] Handle multiple output streams
- [ ] Test with complex PowerShell output

**Deliverable:** Properly formatted, colored output

### Phase 3: Prompt & History (Week 3)

- [ ] Implement `PromptGenerator`
- [ ] Command history Up/Down navigation
- [ ] History persistence to disk
- [ ] Multi-line command detection

**Deliverable:** Full console experience matching ISE

### Phase 4: IntelliSense (Week 4)

- [ ] Implement `IntelliSenseManager`
- [ ] Variable completion
- [ ] Cmdlet completion
- [ ] Member access completion (stretch goal)

**Deliverable:** Tab completion working

### Phase 5: Integration & Polish (Week 5)

- [ ] Replace old `psconsole.go` with new implementation
- [ ] Handle all special cases (cls, progress, etc)
- [ ] Error handling and recovery
- [ ] Performance optimization

**Deliverable:** Production-ready Translation Layer

---

## File Structure

```
cmd/ps-ide/
├── main.go                    # Entry point
├── psconsole.go              # GTK console UI (thin wrapper)
├── editor.go                 # Editor pane
├── toolbar.go                # Toolbar
├── menu.go                   # Menu bar
└── translation/              # Translation Layer package
    ├── layer.go              # Main Translation Layer
    ├── pipes.go              # Named Pipe Manager
    ├── queue.go              # Command Queue Manager
    ├── session.go            # Session State Manager
    ├── parser.go             # Output Parser (CLIXML)
    ├── prompt.go             # Prompt Generator
    ├── intellisense.go       # IntelliSense Manager
    ├── types.go              # Shared types/structs
    └── utils.go              # Helper functions
```

---

## Next Steps for AI Sessions

1. **Read this document first** before making changes
2. **Implement one component at a time** following phases
3. **Write tests** for each component
4. **Update this document** if architecture changes
5. **Reference PowerShell source** for behavior clarification

### Starting Implementation

Begin with **Phase 1, Component 1: Basic Package Structure**

bash

```bash
mkdir -p cmd/ps-ide/translation
cd cmd/ps-ide/translation
# Create files: types.go, layer.go, pipes.go
```

---

**END OF ARCHITECTURE DOCUMENT**
