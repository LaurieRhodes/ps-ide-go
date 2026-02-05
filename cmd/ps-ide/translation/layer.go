package translation

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// TranslationLayer is the main orchestrator for PowerShell interaction
type TranslationLayer struct {
	pipes       *PipeCommunicator
	queue       *CommandQueue
	session     *SessionStateManager
	prompt      *PromptGenerator
	parser      *OutputParser
	mutex       sync.Mutex
	isExecuting bool
	stopChan    chan bool
}

// New creates a new Translation Layer instance
func New() (*TranslationLayer, error) {
	tl := &TranslationLayer{
		pipes:       NewPipeCommunicator(),
		queue:       NewCommandQueue(1000), // Max 1000 history entries
		session:     NewSessionStateManager(),
		prompt:      NewPromptGenerator(),
		parser:      NewOutputParser(),
		isExecuting: false,
		stopChan:    make(chan bool, 1),
	}

	// Start the pipe communicator
	if err := tl.pipes.Start(); err != nil {
		return nil, fmt.Errorf("failed to start pipe communicator: %w", err)
	}

	// Initialize session with PowerShell version
	go tl.initializeSession()

	return tl, nil
}

// initializeSession queries initial PowerShell state
func (tl *TranslationLayer) initializeSession() {
	// Wait a moment for PowerShell to be ready
	time.Sleep(500 * time.Millisecond)

	// Query PowerShell version
	if result, err := tl.pipes.QueryState("$PSVersionTable.PSVersion.ToString()"); err == nil {
		tl.session.SetPSVersion(strings.TrimSpace(result))
	}

	// Query current directory
	if result, err := tl.pipes.QueryState("(Get-Location).Path"); err == nil {
		tl.session.SetCurrentDirectory(strings.TrimSpace(result))
	}
}

// ExecuteCommand executes a user-typed command and returns the output
func (tl *TranslationLayer) ExecuteCommand(cmd string) (string, error) {
	tl.mutex.Lock()
	if tl.isExecuting {
		tl.mutex.Unlock()
		return "", fmt.Errorf("another command is executing")
	}
	tl.isExecuting = true
	tl.mutex.Unlock()

	defer func() {
		tl.mutex.Lock()
		tl.isExecuting = false
		tl.mutex.Unlock()
	}()

	// Add to history
	if err := tl.queue.Add(cmd, Interactive); err != nil {
		return "", err
	}

	// Record start time
	startTime := time.Now()

	// Execute command
	result, err := tl.pipes.SendCommand(cmd, Interactive)

	// Record execution time and result
	duration := time.Since(startTime)
	success := err == nil
	exitCode := 0
	if !success {
		exitCode = 1
	}

	tl.queue.UpdateLastEntry(duration, success, exitCode)

	// Update session state (synchronous to ensure prompt shows correct directory)
	tl.updateDirectory()

	if err != nil {
		return "", fmt.Errorf("command execution failed: %w", err)
	}

	return result, nil
}

// ExecuteScript executes a script file and returns the output
func (tl *TranslationLayer) ExecuteScript(path string) (string, error) {
	tl.mutex.Lock()
	if tl.isExecuting {
		tl.mutex.Unlock()
		return "", fmt.Errorf("another command is executing")
	}
	tl.isExecuting = true
	tl.mutex.Unlock()

	defer func() {
		tl.mutex.Lock()
		tl.isExecuting = false
		tl.mutex.Unlock()
	}()

	// Add to history
	scriptCmd := fmt.Sprintf("& '%s'", path)
	if err := tl.queue.Add(scriptCmd, Script); err != nil {
		return "", err
	}

	// Record start time
	startTime := time.Now()

	// Execute script
	result, err := tl.pipes.ExecuteScript(path)

	// Record execution time
	duration := time.Since(startTime)
	success := err == nil
	exitCode := 0
	if !success {
		exitCode = 1
	}

	tl.queue.UpdateLastEntry(duration, success, exitCode)

	// Update session state (synchronous to ensure prompt shows correct directory)
	tl.updateDirectory()

	if err != nil {
		return "", fmt.Errorf("script execution failed: %w", err)
	}

	return result, nil
}

// ExecuteSelection executes selected text and returns the output
func (tl *TranslationLayer) ExecuteSelection(code string) (string, error) {
	tl.mutex.Lock()
	if tl.isExecuting {
		tl.mutex.Unlock()
		return "", fmt.Errorf("another command is executing")
	}
	tl.isExecuting = true
	tl.mutex.Unlock()

	defer func() {
		tl.mutex.Lock()
		tl.isExecuting = false
		tl.mutex.Unlock()
	}()

	// Add to history
	if err := tl.queue.Add(code, Selection); err != nil {
		return "", err
	}

	// Record start time
	startTime := time.Now()

	// Execute selection
	result, err := tl.pipes.ExecuteScriptText(code)

	// Record execution time
	duration := time.Since(startTime)
	success := err == nil
	exitCode := 0
	if !success {
		exitCode = 1
	}

	tl.queue.UpdateLastEntry(duration, success, exitCode)

	// Update session state (synchronous to ensure prompt shows correct directory)
	tl.updateDirectory()

	if err != nil {
		return "", fmt.Errorf("selection execution failed: %w", err)
	}

	return result, nil
}

// ParseOutput parses raw output using the parser
func (tl *TranslationLayer) ParseOutput(rawOutput string) ([]PSOutput, error) {
	return tl.parser.Parse([]byte(rawOutput))
}

// FormatOutput formats parsed output for display
func (tl *TranslationLayer) FormatOutput(output PSOutput) string {
	return tl.parser.FormatOutput(output)
}

// FormatOutputWithColors formats output with ANSI colors
func (tl *TranslationLayer) FormatOutputWithColors(output PSOutput) string {
	return tl.parser.FormatWithANSI(output)
}

// GetParser returns the output parser (for direct use if needed)
func (tl *TranslationLayer) GetParser() *OutputParser {
	return tl.parser
}

// GetPrompt returns the current prompt string
func (tl *TranslationLayer) GetPrompt() string {
	currentDir := tl.session.GetCurrentDirectory()
	return tl.prompt.Generate(currentDir)
}

// GetPromptANSI returns the prompt with ANSI color codes
func (tl *TranslationLayer) GetPromptANSI() string {
	currentDir := tl.session.GetCurrentDirectory()
	return tl.prompt.GenerateANSI(currentDir)
}

// GetHistoryUp navigates backward in command history
func (tl *TranslationLayer) GetHistoryUp() string {
	cmd, _ := tl.queue.GetPrevious()
	return cmd
}

// GetHistoryDown navigates forward in command history
func (tl *TranslationLayer) GetHistoryDown() string {
	cmd, _ := tl.queue.GetNext()
	return cmd
}

// ResetHistoryIndex resets history navigation to the end
func (tl *TranslationLayer) ResetHistoryIndex() {
	tl.queue.ResetIndex()
}

// StopExecution interrupts the current execution
func (tl *TranslationLayer) StopExecution() error {
	return tl.pipes.SendInterrupt()
}

// Shutdown cleanly stops the Translation Layer
func (tl *TranslationLayer) Shutdown() error {
	// Save history
	if err := tl.queue.Save(); err != nil {
		// Log error but don't fail shutdown
		fmt.Printf("Warning: failed to save history: %v\n", err)
	}

	// Stop pipe communicator
	return tl.pipes.Stop()
}

// IsExecuting returns true if a command is currently executing
func (tl *TranslationLayer) IsExecuting() bool {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()
	return tl.isExecuting
}

// GetCurrentDirectory returns the current working directory
func (tl *TranslationLayer) GetCurrentDirectory() string {
	return tl.session.GetCurrentDirectory()
}

// GetHistory returns all command history
func (tl *TranslationLayer) GetHistory() []CommandEntry {
	return tl.queue.GetAll()
}

// GetRecentHistory returns the N most recent commands
func (tl *TranslationLayer) GetRecentHistory(n int) []CommandEntry {
	return tl.queue.GetRecent(n)
}

// ClearHistory clears all command history
func (tl *TranslationLayer) ClearHistory() error {
	return tl.queue.Clear()
}

// SearchHistory searches history for matching commands
func (tl *TranslationLayer) SearchHistory(query string) []CommandEntry {
	return tl.queue.Search(query)
}

// GetCompletions returns IntelliSense completions for the given prefix
func (tl *TranslationLayer) GetCompletions(prefix string) []string {
	return tl.session.GetCompletions(prefix)
}

// GetPSVersion returns the PowerShell version string
func (tl *TranslationLayer) GetPSVersion() string {
	return tl.session.GetPSVersion()
}

// GetVariables returns all tracked variables
func (tl *TranslationLayer) GetVariables() map[string]VariableInfo {
	return tl.session.GetAllVariables()
}

// GetFunctions returns all tracked functions
func (tl *TranslationLayer) GetFunctions() map[string]FunctionInfo {
	return tl.session.GetAllFunctions()
}

// GetModules returns all loaded modules
func (tl *TranslationLayer) GetModules() []ModuleInfo {
	return tl.session.GetModules()
}

// SyncState synchronizes session state with PowerShell
func (tl *TranslationLayer) SyncState() error {
	// Query current directory
	if err := tl.updateDirectory(); err != nil {
		return err
	}

	// TODO: Query variables, functions, modules
	// This will be implemented in Phase 2

	return nil
}

// updateDirectory queries and updates the current directory
func (tl *TranslationLayer) updateDirectory() error {
	result, err := tl.pipes.QueryState(tl.session.GetQueryCommand(DirectoryUpdate))
	if err != nil {
		return err
	}

	// Clean up result (remove quotes, whitespace)
	cleanResult := strings.TrimSpace(result)
	if len(cleanResult) > 0 && cleanResult[0] == '"' {
		cleanResult = cleanResult[1:]
	}
	if len(cleanResult) > 0 && cleanResult[len(cleanResult)-1] == '"' {
		cleanResult = cleanResult[:len(cleanResult)-1]
	}

	tl.session.SetCurrentDirectory(cleanResult)
	return nil
}

// GetResponseChannel returns the pipe response channel for monitoring
func (tl *TranslationLayer) GetResponseChannel() <-chan PipeResponse {
	return tl.pipes.GetResponseChannel()
}

// SetPromptStyle changes the prompt style
func (tl *TranslationLayer) SetPromptStyle(style PromptStyle) {
	tl.prompt.SetStyle(style)
}

// SetRemoteHost sets the remote hostname (for remote sessions)
func (tl *TranslationLayer) SetRemoteHost(hostname string) {
	tl.prompt.SetRemoteHost(hostname)
}
