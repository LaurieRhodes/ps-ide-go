package translation

import (
	"fmt"
	"time"
)

// ExecuteCommandWithOutput executes a user-typed command and returns output
func (tl *TranslationLayer) ExecuteCommandWithOutput(cmd string) (string, error) {
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

	// Update session state (query current directory)
	go tl.updateDirectory()

	if err != nil {
		return "", fmt.Errorf("command execution failed: %w", err)
	}

	return result, nil
}
