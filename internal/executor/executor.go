package executor

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

// Executor handles PowerShell script execution
type Executor struct {
	timeout time.Duration
}

// New creates a new Executor with default timeout
func New() *Executor {
	return &Executor{
		timeout: 30 * time.Second,
	}
}

// SetTimeout sets the execution timeout
func (e *Executor) SetTimeout(d time.Duration) {
	e.timeout = d
}

// Execute runs a PowerShell script and returns stdout, stderr, and error
func (e *Executor) Execute(script string) (stdout, stderr string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pwsh", "-NoProfile", "-Command", script)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()

	return
}

// ExecuteFile runs a PowerShell script file
func (e *Executor) ExecuteFile(filepath string) (stdout, stderr string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pwsh", "-NoProfile", "-File", filepath)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()

	return
}

// CheckPowerShellInstalled verifies PowerShell is available
func CheckPowerShellInstalled() bool {
	cmd := exec.Command("pwsh", "-Version")
	return cmd.Run() == nil
}
