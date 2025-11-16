package translation

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	debugLogger *log.Logger
	debugFile   *os.File
	debugMutex  sync.Mutex
	debugEnabled = false
)

// EnableDebugLogging enables debug logging to a file
func EnableDebugLogging() error {
	debugMutex.Lock()
	defer debugMutex.Unlock()
	
	if debugEnabled {
		return nil
	}
	
	// Create log directory
	homeDir, _ := os.UserHomeDir()
	logDir := filepath.Join(homeDir, ".ps-ide", "logs")
	os.MkdirAll(logDir, 0755)
	
	// Create log file with timestamp
	timestamp := time.Now().Format("20060102-150405")
	logPath := filepath.Join(logDir, fmt.Sprintf("ps-ide-debug-%s.log", timestamp))
	
	var err error
	debugFile, err = os.Create(logPath)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	
	debugLogger = log.New(debugFile, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	debugEnabled = true
	
	debugLogger.Printf("=== PS-IDE Debug Logging Started ===")
	debugLogger.Printf("Log file: %s", logPath)
	fmt.Printf("Debug logging enabled: %s\n", logPath)
	
	return nil
}

// DisableDebugLogging disables debug logging and closes the log file
func DisableDebugLogging() {
	debugMutex.Lock()
	defer debugMutex.Unlock()
	
	if !debugEnabled {
		return
	}
	
	if debugLogger != nil {
		debugLogger.Printf("=== PS-IDE Debug Logging Ended ===")
	}
	
	if debugFile != nil {
		debugFile.Close()
	}
	
	debugEnabled = false
	debugLogger = nil
	debugFile = nil
}

// DebugLog logs a debug message if debug logging is enabled
func DebugLog(format string, args ...interface{}) {
	debugMutex.Lock()
	defer debugMutex.Unlock()
	
	if debugEnabled && debugLogger != nil {
		debugLogger.Printf(format, args...)
	}
}

// DebugLogRaw logs raw data with a label
func DebugLogRaw(label string, data string) {
	debugMutex.Lock()
	defer debugMutex.Unlock()
	
	if debugEnabled && debugLogger != nil {
		// Show the raw data with escape sequences visible
		debugLogger.Printf("=== %s ===", label)
		debugLogger.Printf("Length: %d bytes", len(data))
		debugLogger.Printf("Raw: %q", data)
		debugLogger.Printf("=== End %s ===", label)
	}
}

// IsDebugEnabled returns true if debug logging is enabled
func IsDebugEnabled() bool {
	debugMutex.Lock()
	defer debugMutex.Unlock()
	return debugEnabled
}
