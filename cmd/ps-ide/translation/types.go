package translation

import (
	"time"
)

// CommandType represents the type of command being executed
type CommandType int

const (
	// Interactive represents a command typed in the console
	Interactive CommandType = iota
	// Script represents a script file execution
	Script
	// Selection represents selected text execution
	Selection
	// Internal represents internal commands (like prompt query)
	Internal
)

// CommandEntry represents a single command in history
type CommandEntry struct {
	Command    string
	Timestamp  time.Time
	Type       CommandType
	WorkingDir string
	Duration   time.Duration
	Success    bool
	ExitCode   int
}

// StreamType represents different PowerShell output streams
type StreamType int

const (
	OutputStream StreamType = iota
	ErrorStream
	WarningStream
	VerboseStream
	DebugStream
	ProgressStream
	InformationStream
)

// String returns the string representation of StreamType
func (st StreamType) String() string {
	switch st {
	case OutputStream:
		return "Output"
	case ErrorStream:
		return "Error"
	case WarningStream:
		return "Warning"
	case VerboseStream:
		return "Verbose"
	case DebugStream:
		return "Debug"
	case ProgressStream:
		return "Progress"
	case InformationStream:
		return "Information"
	default:
		return "Unknown"
	}
}

// PSOutput represents parsed PowerShell output
type PSOutput struct {
	Stream       StreamType
	Content      string
	ANSISegments []ANSISegment
	ObjectData   interface{}
	IsFormatted  bool
	Timestamp    time.Time
}

// ANSISegment represents a segment of text with ANSI formatting
type ANSISegment struct {
	Text      string
	FGColor   int  // ANSI color code (30-37, 90-97)
	BGColor   int  // ANSI background color code (40-47, 100-107)
	Bold      bool
	Underline bool
	Italic    bool
}

// PipeCommand represents a command sent via pipe
type PipeCommand struct {
	ID      string
	Command string
	Type    CommandType
}

// PipeResponse represents a response from PowerShell
type PipeResponse struct {
	ID       string
	Stream   StreamType
	Data     []byte
	Complete bool
}

// VariableInfo represents a PowerShell variable
type VariableInfo struct {
	Name        string
	Type        string
	Value       string
	Description string
}

// FunctionInfo represents a PowerShell function
type FunctionInfo struct {
	Name       string
	Parameters []string
	Synopsis   string
}

// ModuleInfo represents a PowerShell module
type ModuleInfo struct {
	Name    string
	Version string
	Path    string
}

// SessionState represents the current PowerShell session state
type SessionState struct {
	CurrentDirectory string
	Variables        map[string]VariableInfo
	Functions        map[string]FunctionInfo
	Modules          []ModuleInfo
	PSVersion        string
	LastExitCode     int
	ErrorActionPref  string
	LastSync         time.Time
}

// CompletionItem represents an IntelliSense completion item
type CompletionItem struct {
	Text        string
	Type        CompletionType
	Description string
}

// CompletionType represents the type of completion
type CompletionType int

const (
	CommandCompletion CompletionType = iota
	VariableCompletion
	ParameterCompletion
	PathCompletion
	MemberCompletion
)

// PromptStyle represents different prompt styles
type PromptStyle int

const (
	DefaultPrompt PromptStyle = iota
	RemotePrompt
	CustomPrompt
)
