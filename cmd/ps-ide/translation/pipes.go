package translation

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

// PipeCommunicator handles bidirectional communication with PowerShell
type PipeCommunicator struct {
	psProcess    *exec.Cmd
	stdin        io.WriteCloser
	stdout       io.ReadCloser
	stderr       io.ReadCloser
	mutex        sync.Mutex
	responseChan chan string
	isRunning    bool
	stopChan     chan bool
	escapeRegex  *regexp.Regexp
	promptRegex  *regexp.Regexp
}

// NewPipeCommunicator creates a new pipe communicator
func NewPipeCommunicator() *PipeCommunicator {
	return &PipeCommunicator{
		responseChan: make(chan string, 100),
		stopChan:     make(chan bool, 1),
		isRunning:    false,
		// Match common terminal escape sequences we want to strip
		escapeRegex:  regexp.MustCompile(`\x1b\[\?[0-9]+[hl]|\x1b\[H|\x1b\[[0-9;]*J`),
		// Match PowerShell prompts - be more aggressive
		promptRegex:  regexp.MustCompile(`^(PS\s+.*?>|.*?>\s*)$`),
	}
}

// Start initializes PowerShell process
func (pc *PipeCommunicator) Start() error {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	
	if pc.isRunning {
		return fmt.Errorf("pipe communicator already running")
	}
	
	DebugLog("Starting PowerShell process...")
	
	// Start PowerShell in interactive mode
	pc.psProcess = exec.Command("pwsh", 
		"-NoLogo",
		"-NoProfile",
		"-Interactive")
	
	// Set environment variables to help with ANSI support
	// Note: Write-Host -ForegroundColor won't work in pipe mode
	// Users should use Write-Error, Write-Warning, Write-Verbose, Write-Debug
	// or $PSStyle for colored output
	pc.psProcess.Env = append(os.Environ(),
		"TERM=xterm-256color",
	)
	
	// Get stdin/stdout/stderr pipes
	stdin, err := pc.psProcess.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	pc.stdin = stdin
	
	stdout, err := pc.psProcess.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	pc.stdout = stdout
	
	stderr, err := pc.psProcess.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}
	pc.stderr = stderr
	
	// Start the process
	if err := pc.psProcess.Start(); err != nil {
		return fmt.Errorf("failed to start PowerShell: %w", err)
	}
	
	DebugLog("PowerShell process started, PID: %d", pc.psProcess.Process.Pid)
	
	pc.isRunning = true
	
	// Start goroutines for reading output
	go pc.readOutputLoop(stdout, false)
	go pc.readOutputLoop(stderr, true)
	
	// Give PowerShell time to initialize and consume initial prompt
	DebugLog("Waiting 500ms for PowerShell initialization...")
	time.Sleep(500 * time.Millisecond)
	pc.FlushOutput()
	
	// Set encoding to UTF8
	DebugLog("Setting PowerShell encoding...")
	pc.stdin.Write([]byte("[Console]::OutputEncoding = [System.Text.Encoding]::UTF8\n"))
	time.Sleep(100 * time.Millisecond)
	pc.FlushOutput()
	
	DebugLog("PowerShell initialized")
	
	return nil
}

// Stop terminates the PowerShell process
func (pc *PipeCommunicator) Stop() error {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	
	if !pc.isRunning {
		return nil
	}
	
	DebugLog("Stopping PowerShell process...")
	
	// Signal stop
	select {
	case pc.stopChan <- true:
	default:
	}
	
	// Close pipes
	if pc.stdin != nil {
		pc.stdin.Close()
	}
	if pc.stdout != nil {
		pc.stdout.Close()
	}
	if pc.stderr != nil {
		pc.stderr.Close()
	}
	
	// Terminate PowerShell process
	if pc.psProcess != nil && pc.psProcess.Process != nil {
		pc.psProcess.Process.Kill()
		pc.psProcess.Wait()
	}
	
	pc.isRunning = false
	DebugLog("PowerShell process stopped")
	return nil
}

// SendCommand sends a command to PowerShell and waits for output
func (pc *PipeCommunicator) SendCommand(command string, cmdType CommandType) (string, error) {
	pc.mutex.Lock()
	if !pc.isRunning {
		pc.mutex.Unlock()
		return "", fmt.Errorf("pipe communicator not running")
	}
	pc.mutex.Unlock()
	
	DebugLog("SendCommand called: type=%v", cmdType)
	DebugLogRaw("COMMAND SENT", command)
	
	// Clear any pending responses
	flushed := pc.FlushOutput()
	if flushed > 0 {
		DebugLog("Flushed %d pending responses", flushed)
	}
	
	// Write command to stdin
	_, err := pc.stdin.Write([]byte(command + "\n"))
	if err != nil {
		DebugLog("ERROR writing command: %v", err)
		return "", fmt.Errorf("failed to write command: %w", err)
	}
	
	DebugLog("Command written to stdin, waiting for output...")
	
	// Collect output until we see the prompt again or timeout
	var output strings.Builder
	timeout := time.After(5 * time.Second)
	
	// Wait a bit for output to start
	time.Sleep(150 * time.Millisecond)
	
	collecting := true
	noOutputCount := 0
	echoSkipped := false  // Track if we've skipped the command echo line
	lineCount := 0
	
	for collecting {
		select {
		case line := <-pc.responseChan:
			lineCount++
			DebugLogRaw(fmt.Sprintf("RAW LINE %d", lineCount), line)
			
			// Clean the line
			cleanLine := pc.cleanLine(line)
			DebugLog("Cleaned line %d: %q", lineCount, cleanLine)
			
			// Skip empty lines
			if cleanLine == "" {
				DebugLog("Skipping empty line")
				continue
			}
			
			// Skip the first line that contains prompt + command (echo)
			// This is the "PS /path> command" line
			if !echoSkipped && strings.HasPrefix(cleanLine, "PS ") && strings.Contains(cleanLine, ">") {
				DebugLog("Skipping command echo line (prompt + command)")
				echoSkipped = true
				continue
			}
			
			// Skip if it's a PowerShell prompt line (just "PS /path>")
			if pc.isPromptLine(cleanLine) {
				DebugLog("Found prompt line, stopping collection: %q", cleanLine)
				collecting = false
				break
			}
			
			DebugLog("Adding line to output: %q", cleanLine)
			output.WriteString(cleanLine)
			output.WriteString("\n")
			noOutputCount = 0
			
		case <-time.After(100 * time.Millisecond):
			noOutputCount++
			DebugLog("No output for 100ms (count: %d, output length: %d)", noOutputCount, output.Len())
			// If no output for 400ms after seeing some output, consider it done
			if noOutputCount >= 4 && output.Len() > 0 {
				DebugLog("Timeout waiting for more output, stopping collection")
				collecting = false
			}
			
		case <-timeout:
			DebugLog("5 second timeout reached, stopping collection")
			collecting = false
		}
	}
	
	result := output.String()
	result = strings.TrimSpace(result)
	
	DebugLogRaw("FINAL OUTPUT", result)
	DebugLog("SendCommand complete: %d bytes returned", len(result))
	
	return result, nil
}

// isPromptLine checks if a line is a PowerShell prompt
func (pc *PipeCommunicator) isPromptLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	
	// Match "PS path>" or just ">" at end (but not if it has text after the >)
	if strings.HasPrefix(trimmed, "PS ") && strings.HasSuffix(trimmed, ">") {
		DebugLog("isPromptLine: matched PS prefix: %q", trimmed)
		return true
	}
	
	// Match lines that are just "path>"
	if strings.HasSuffix(trimmed, ">") && !strings.Contains(trimmed, " ") {
		DebugLog("isPromptLine: matched simple prompt: %q", trimmed)
		return true
	}
	
	return false
}

// cleanLine removes unwanted escape sequences but keeps ANSI color codes
func (pc *PipeCommunicator) cleanLine(line string) string {
	// Remove cursor movement and screen clearing sequences
	cleaned := pc.escapeRegex.ReplaceAllString(line, "")
	
	// Trim whitespace
	cleaned = strings.TrimSpace(cleaned)
	
	return cleaned
}

// readOutputLoop continuously reads from output pipe
func (pc *PipeCommunicator) readOutputLoop(reader io.Reader, isError bool) {
	streamName := "STDOUT"
	if isError {
		streamName = "STDERR"
	}
	DebugLog("readOutputLoop started for %s", streamName)
	
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	
	for scanner.Scan() {
		select {
		case <-pc.stopChan:
			DebugLog("readOutputLoop %s received stop signal", streamName)
			return
		default:
			line := scanner.Text()
			DebugLog("%s received: %q", streamName, line)
			
			// Send to response channel (non-blocking)
			select {
			case pc.responseChan <- line:
				// Sent successfully
			default:
				// Channel full, drop oldest
				DebugLog("WARNING: response channel full, dropping oldest")
				select {
				case <-pc.responseChan:
					pc.responseChan <- line
				default:
				}
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		DebugLog("readOutputLoop %s scanner error: %v", streamName, err)
	}
	DebugLog("readOutputLoop %s ended", streamName)
}

// IsRunning returns true if the communicator is running
func (pc *PipeCommunicator) IsRunning() bool {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	return pc.isRunning
}

// GetPID returns the PowerShell process ID
func (pc *PipeCommunicator) GetPID() int {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	
	if pc.psProcess != nil && pc.psProcess.Process != nil {
		return pc.psProcess.Process.Pid
	}
	return -1
}

// SendInterrupt sends Ctrl+C to PowerShell
func (pc *PipeCommunicator) SendInterrupt() error {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	
	if !pc.isRunning || pc.psProcess == nil || pc.psProcess.Process == nil {
		return fmt.Errorf("process not running")
	}
	
	DebugLog("Sending interrupt signal to PowerShell")
	// Send interrupt signal (SIGINT on Unix)
	return pc.psProcess.Process.Signal(os.Interrupt)
}

// ExecuteScript executes a script file
func (pc *PipeCommunicator) ExecuteScript(scriptPath string) (string, error) {
	// Use PowerShell's script execution syntax
	command := fmt.Sprintf("& '%s'", scriptPath)
	return pc.SendCommand(command, Script)
}

// ExecuteScriptText executes script text
func (pc *PipeCommunicator) ExecuteScriptText(scriptText string) (string, error) {
	// Write to temp file and execute
	tmpFile, err := os.CreateTemp("", "ps-ide-*.ps1")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())
	
	if _, err := tmpFile.WriteString(scriptText); err != nil {
		return "", err
	}
	tmpFile.Close()
	
	return pc.ExecuteScript(tmpFile.Name())
}

// QueryState queries PowerShell state (silent, no output to console)
func (pc *PipeCommunicator) QueryState(query string) (string, error) {
	DebugLog("QueryState called: %s", query)
	// For state queries, we want to suppress output
	// Use Out-String to avoid displaying in console
	silentQuery := fmt.Sprintf("(%s) | Out-String -Stream", query)
	return pc.SendCommand(silentQuery, Internal)
}

// GetResponseChannel returns the response channel for external monitoring
func (pc *PipeCommunicator) GetResponseChannel() <-chan PipeResponse {
	// Create a dummy channel for compatibility
	ch := make(chan PipeResponse)
	close(ch)
	return ch
}

// FlushOutput clears any pending output and returns count of flushed items
func (pc *PipeCommunicator) FlushOutput() int {
	count := 0
	for {
		select {
		case line := <-pc.responseChan:
			count++
			if IsDebugEnabled() {
				DebugLog("Flushed: %q", line)
			}
		default:
			return count
		}
	}
}
