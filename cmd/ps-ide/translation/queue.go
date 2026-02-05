package translation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CommandQueue manages command history for Up/Down arrow navigation
type CommandQueue struct {
	history      []CommandEntry
	currentIndex int
	maxSize      int
	persistPath  string
	mutex        sync.RWMutex
}

// NewCommandQueue creates a new command queue
func NewCommandQueue(maxSize int) *CommandQueue {
	homeDir, _ := os.UserHomeDir()
	persistPath := filepath.Join(homeDir, ".ps-ide", "history.json")

	cq := &CommandQueue{
		history:      make([]CommandEntry, 0, maxSize),
		currentIndex: 0,
		maxSize:      maxSize,
		persistPath:  persistPath,
	}

	// Try to load existing history
	cq.Load()

	return cq
}

// Add adds a command to the history
func (cq *CommandQueue) Add(command string, cmdType CommandType) error {
	cq.mutex.Lock()
	defer cq.mutex.Unlock()

	// Don't add empty commands
	if command == "" {
		return nil
	}

	// Don't add duplicate consecutive commands
	if len(cq.history) > 0 {
		lastEntry := cq.history[len(cq.history)-1]
		if lastEntry.Command == command && lastEntry.Type == cmdType {
			return nil
		}
	}

	// Get current working directory
	workingDir, _ := os.Getwd()

	entry := CommandEntry{
		Command:    command,
		Timestamp:  time.Now(),
		Type:       cmdType,
		WorkingDir: workingDir,
		Success:    true, // Will be updated after execution
	}

	cq.history = append(cq.history, entry)

	// Trim if exceeds max size
	if len(cq.history) > cq.maxSize {
		cq.history = cq.history[1:]
	}

	// Reset navigation index
	cq.currentIndex = len(cq.history)

	return nil
}

// GetPrevious navigates backward in history
func (cq *CommandQueue) GetPrevious() (string, bool) {
	cq.mutex.Lock()
	defer cq.mutex.Unlock()

	if len(cq.history) == 0 {
		return "", false
	}

	// Move back in history
	if cq.currentIndex > 0 {
		cq.currentIndex--
	}

	if cq.currentIndex < len(cq.history) {
		return cq.history[cq.currentIndex].Command, true
	}

	return "", false
}

// GetNext navigates forward in history
func (cq *CommandQueue) GetNext() (string, bool) {
	cq.mutex.Lock()
	defer cq.mutex.Unlock()

	if len(cq.history) == 0 {
		return "", false
	}

	// Move forward in history
	if cq.currentIndex < len(cq.history) {
		cq.currentIndex++
	}

	// If at the end, return empty (like PSReadLine)
	if cq.currentIndex >= len(cq.history) {
		return "", true
	}

	return cq.history[cq.currentIndex].Command, true
}

// GetAll returns all history entries
func (cq *CommandQueue) GetAll() []CommandEntry {
	cq.mutex.RLock()
	defer cq.mutex.RUnlock()

	// Return a copy to prevent external modification
	result := make([]CommandEntry, len(cq.history))
	copy(result, cq.history)
	return result
}

// GetRecent returns the N most recent entries
func (cq *CommandQueue) GetRecent(n int) []CommandEntry {
	cq.mutex.RLock()
	defer cq.mutex.RUnlock()

	if n <= 0 || len(cq.history) == 0 {
		return []CommandEntry{}
	}

	start := len(cq.history) - n
	if start < 0 {
		start = 0
	}

	result := make([]CommandEntry, len(cq.history)-start)
	copy(result, cq.history[start:])
	return result
}

// Clear removes all history
func (cq *CommandQueue) Clear() error {
	cq.mutex.Lock()
	defer cq.mutex.Unlock()

	cq.history = make([]CommandEntry, 0, cq.maxSize)
	cq.currentIndex = 0

	return cq.Save()
}

// ResetIndex resets the navigation index to the end
func (cq *CommandQueue) ResetIndex() {
	cq.mutex.Lock()
	defer cq.mutex.Unlock()

	cq.currentIndex = len(cq.history)
}

// UpdateLastEntry updates the last entry with execution results
func (cq *CommandQueue) UpdateLastEntry(duration time.Duration, success bool, exitCode int) {
	cq.mutex.Lock()
	defer cq.mutex.Unlock()

	if len(cq.history) == 0 {
		return
	}

	lastIndex := len(cq.history) - 1
	cq.history[lastIndex].Duration = duration
	cq.history[lastIndex].Success = success
	cq.history[lastIndex].ExitCode = exitCode
}

// Save persists history to disk
func (cq *CommandQueue) Save() error {
	cq.mutex.RLock()
	defer cq.mutex.RUnlock()

	// Create directory if it doesn't exist
	dir := filepath.Dir(cq.persistPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(cq.history, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(cq.persistPath, data, 0644)
}

// Load restores history from disk
func (cq *CommandQueue) Load() error {
	cq.mutex.Lock()
	defer cq.mutex.Unlock()

	// Check if file exists
	if _, err := os.Stat(cq.persistPath); os.IsNotExist(err) {
		return nil // Not an error, just no history yet
	}

	// Read file
	data, err := os.ReadFile(cq.persistPath)
	if err != nil {
		return err
	}

	// Unmarshal JSON
	var history []CommandEntry
	if err := json.Unmarshal(data, &history); err != nil {
		return err
	}

	// Apply max size limit
	if len(history) > cq.maxSize {
		history = history[len(history)-cq.maxSize:]
	}

	cq.history = history
	cq.currentIndex = len(cq.history)

	return nil
}

// GetSize returns the number of entries in history
func (cq *CommandQueue) GetSize() int {
	cq.mutex.RLock()
	defer cq.mutex.RUnlock()

	return len(cq.history)
}

// GetCurrentIndex returns the current navigation index
func (cq *CommandQueue) GetCurrentIndex() int {
	cq.mutex.RLock()
	defer cq.mutex.RUnlock()

	return cq.currentIndex
}

// Search returns entries matching the query
func (cq *CommandQueue) Search(query string) []CommandEntry {
	cq.mutex.RLock()
	defer cq.mutex.RUnlock()

	if query == "" {
		return []CommandEntry{}
	}

	var results []CommandEntry
	for _, entry := range cq.history {
		if contains(entry.Command, query) {
			results = append(results, entry)
		}
	}

	return results
}

// contains is a case-insensitive substring check
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				indexSubstring(s, substr) >= 0)
}

// indexSubstring finds substring index (case-insensitive)
func indexSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFold(s[i:i+len(substr)], substr) {
			return i
		}
	}
	return -1
}

// equalFold compares strings case-insensitively
func equalFold(s1, s2 string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		c1, c2 := s1[i], s2[i]
		if c1 >= 'A' && c1 <= 'Z' {
			c1 += 'a' - 'A'
		}
		if c2 >= 'A' && c2 <= 'Z' {
			c2 += 'a' - 'A'
		}
		if c1 != c2 {
			return false
		}
	}
	return true
}
