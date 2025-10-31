# Translation Layer Implementation Roadmap

**Project**: ps-ide-go  
**Purpose**: Step-by-step implementation guide

See: `/media/laurie/Data/Github/ps-ide-go/docs/TRANSLATION_LAYER_ARCHITECTURE.md`

## Phase 1: Foundation (START HERE)

### 1. Create types.go - Shared type definitions
### 2. Create prompt.go - Independent prompt generation  
### 3. Create queue.go - Command history with persistence
### 4. Create session.go - PowerShell state tracking
### 5. Create pipes.go - Named pipe communication (basic)

## Phase 2: Output Handling

### 1. Implement CLIXML parser
### 2. Handle ANSI color codes
### 3. Separate output streams

## Phase 3: Integration

### 1. Refactor psconsole.go to use translation layer
### 2. Wire keyboard shortcuts
### 3. Test end-to-end

## Key Decision: Use XML Output Format

PowerShell with `-OutputFormat XML` provides:
- No command echo pollution
- Structured output with streams
- ANSI colors preserved
- Type information for IntelliSense

## File Structure

```
cmd/ps-ide/translation/
├── types.go       - Shared types
├── prompt.go      - PS prompt generation
├── queue.go       - Command history
├── session.go     - State tracking
├── parser.go      - CLIXML deserializer
├── pipes.go       - Named pipe comms
└── layer.go       - Main orchestrator
```

## Testing Checklist

- [ ] Simple command: `Get-Date`
- [ ] Script execution (F5)
- [ ] Up/Down arrow history
- [ ] Clear console works
- [ ] History persists across restarts

---

Full details in TRANSLATION_LAYER_ARCHITECTURE.md
