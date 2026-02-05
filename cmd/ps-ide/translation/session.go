package translation

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// SessionStateManager tracks PowerShell session state
type SessionStateManager struct {
	state      *SessionState
	mutex      sync.RWMutex
	updateChan chan StateUpdate
}

// StateUpdate represents a state update request
type StateUpdate struct {
	Type UpdateType
	Data interface{}
}

// UpdateType represents different types of state updates
type UpdateType int

const (
	DirectoryUpdate UpdateType = iota
	VariablesUpdate
	FunctionsUpdate
	ModulesUpdate
	ExitCodeUpdate
)

// NewSessionStateManager creates a new session state manager
func NewSessionStateManager() *SessionStateManager {
	currentDir, _ := os.Getwd()

	return &SessionStateManager{
		state: &SessionState{
			CurrentDirectory: currentDir,
			Variables:        make(map[string]VariableInfo),
			Functions:        make(map[string]FunctionInfo),
			Modules:          make([]ModuleInfo, 0),
			PSVersion:        "",
			LastExitCode:     0,
			ErrorActionPref:  "Continue",
			LastSync:         time.Now(),
		},
		updateChan: make(chan StateUpdate, 10),
	}
}

// GetCurrentDirectory returns the current working directory
func (ssm *SessionStateManager) GetCurrentDirectory() string {
	ssm.mutex.RLock()
	defer ssm.mutex.RUnlock()

	return ssm.state.CurrentDirectory
}

// SetCurrentDirectory updates the current directory
func (ssm *SessionStateManager) SetCurrentDirectory(dir string) {
	ssm.mutex.Lock()
	defer ssm.mutex.Unlock()

	ssm.state.CurrentDirectory = dir
	ssm.state.LastSync = time.Now()
}

// GetVariable retrieves a variable by name
func (ssm *SessionStateManager) GetVariable(name string) (VariableInfo, bool) {
	ssm.mutex.RLock()
	defer ssm.mutex.RUnlock()

	variable, exists := ssm.state.Variables[name]
	return variable, exists
}

// SetVariable adds or updates a variable
func (ssm *SessionStateManager) SetVariable(name string, info VariableInfo) {
	ssm.mutex.Lock()
	defer ssm.mutex.Unlock()

	ssm.state.Variables[name] = info
}

// GetAllVariables returns all tracked variables
func (ssm *SessionStateManager) GetAllVariables() map[string]VariableInfo {
	ssm.mutex.RLock()
	defer ssm.mutex.RUnlock()

	// Return a copy
	result := make(map[string]VariableInfo)
	for k, v := range ssm.state.Variables {
		result[k] = v
	}
	return result
}

// GetFunction retrieves a function by name
func (ssm *SessionStateManager) GetFunction(name string) (FunctionInfo, bool) {
	ssm.mutex.RLock()
	defer ssm.mutex.RUnlock()

	function, exists := ssm.state.Functions[name]
	return function, exists
}

// SetFunction adds or updates a function
func (ssm *SessionStateManager) SetFunction(name string, info FunctionInfo) {
	ssm.mutex.Lock()
	defer ssm.mutex.Unlock()

	ssm.state.Functions[name] = info
}

// GetAllFunctions returns all tracked functions
func (ssm *SessionStateManager) GetAllFunctions() map[string]FunctionInfo {
	ssm.mutex.RLock()
	defer ssm.mutex.RUnlock()

	// Return a copy
	result := make(map[string]FunctionInfo)
	for k, v := range ssm.state.Functions {
		result[k] = v
	}
	return result
}

// GetModules returns all loaded modules
func (ssm *SessionStateManager) GetModules() []ModuleInfo {
	ssm.mutex.RLock()
	defer ssm.mutex.RUnlock()

	// Return a copy
	result := make([]ModuleInfo, len(ssm.state.Modules))
	copy(result, ssm.state.Modules)
	return result
}

// SetModules updates the list of loaded modules
func (ssm *SessionStateManager) SetModules(modules []ModuleInfo) {
	ssm.mutex.Lock()
	defer ssm.mutex.Unlock()

	ssm.state.Modules = modules
}

// GetPSVersion returns the PowerShell version
func (ssm *SessionStateManager) GetPSVersion() string {
	ssm.mutex.RLock()
	defer ssm.mutex.RUnlock()

	return ssm.state.PSVersion
}

// SetPSVersion sets the PowerShell version
func (ssm *SessionStateManager) SetPSVersion(version string) {
	ssm.mutex.Lock()
	defer ssm.mutex.Unlock()

	ssm.state.PSVersion = version
}

// GetLastExitCode returns the last exit code
func (ssm *SessionStateManager) GetLastExitCode() int {
	ssm.mutex.RLock()
	defer ssm.mutex.RUnlock()

	return ssm.state.LastExitCode
}

// SetLastExitCode updates the last exit code
func (ssm *SessionStateManager) SetLastExitCode(code int) {
	ssm.mutex.Lock()
	defer ssm.mutex.Unlock()

	ssm.state.LastExitCode = code
}

// GetCompletions returns completion items for a given prefix
func (ssm *SessionStateManager) GetCompletions(prefix string) []string {
	ssm.mutex.RLock()
	defer ssm.mutex.RUnlock()

	var completions []string

	// Variable completions (start with $)
	if len(prefix) > 0 && prefix[0] == '$' {
		varPrefix := prefix[1:]
		for varName := range ssm.state.Variables {
			if matchesPrefix(varName, varPrefix) {
				completions = append(completions, "$"+varName)
			}
		}
	}

	// Function completions
	for funcName := range ssm.state.Functions {
		if matchesPrefix(funcName, prefix) {
			completions = append(completions, funcName)
		}
	}

	return completions
}

// matchesPrefix checks if a string starts with prefix (case-insensitive)
func matchesPrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}

	for i := 0; i < len(prefix); i++ {
		c1, c2 := s[i], prefix[i]
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

// SyncFromJSON updates state from JSON response
func (ssm *SessionStateManager) SyncFromJSON(data []byte, updateType UpdateType) error {
	ssm.mutex.Lock()
	defer ssm.mutex.Unlock()

	switch updateType {
	case DirectoryUpdate:
		var dir string
		if err := json.Unmarshal(data, &dir); err != nil {
			return err
		}
		ssm.state.CurrentDirectory = dir

	case VariablesUpdate:
		var variables []VariableInfo
		if err := json.Unmarshal(data, &variables); err != nil {
			return err
		}
		for _, v := range variables {
			ssm.state.Variables[v.Name] = v
		}

	case FunctionsUpdate:
		var functions []FunctionInfo
		if err := json.Unmarshal(data, &functions); err != nil {
			return err
		}
		for _, f := range functions {
			ssm.state.Functions[f.Name] = f
		}

	case ModulesUpdate:
		var modules []ModuleInfo
		if err := json.Unmarshal(data, &modules); err != nil {
			return err
		}
		ssm.state.Modules = modules
	}

	ssm.state.LastSync = time.Now()
	return nil
}

// GetQueryCommand returns the PowerShell command to query state
func (ssm *SessionStateManager) GetQueryCommand(updateType UpdateType) string {
	switch updateType {
	case DirectoryUpdate:
		return "(Get-Location).Path"
	case VariablesUpdate:
		return "Get-Variable | Select-Object Name, @{Name='Type';Expression={$_.Value.GetType().Name}}, @{Name='Value';Expression={$_.Value.ToString()}} | ConvertTo-Json"
	case FunctionsUpdate:
		return "Get-Command -CommandType Function | Select-Object Name | ConvertTo-Json"
	case ModulesUpdate:
		return "Get-Module | Select-Object Name, Version, Path | ConvertTo-Json"
	default:
		return ""
	}
}

// GetLastSyncTime returns when state was last synchronized
func (ssm *SessionStateManager) GetLastSyncTime() time.Time {
	ssm.mutex.RLock()
	defer ssm.mutex.RUnlock()

	return ssm.state.LastSync
}

// NeedsSync returns true if state should be synchronized
func (ssm *SessionStateManager) NeedsSync(maxAge time.Duration) bool {
	ssm.mutex.RLock()
	defer ssm.mutex.RUnlock()

	return time.Since(ssm.state.LastSync) > maxAge
}

// GetState returns a copy of the current state
func (ssm *SessionStateManager) GetState() SessionState {
	ssm.mutex.RLock()
	defer ssm.mutex.RUnlock()

	// Deep copy
	state := SessionState{
		CurrentDirectory: ssm.state.CurrentDirectory,
		Variables:        make(map[string]VariableInfo),
		Functions:        make(map[string]FunctionInfo),
		Modules:          make([]ModuleInfo, len(ssm.state.Modules)),
		PSVersion:        ssm.state.PSVersion,
		LastExitCode:     ssm.state.LastExitCode,
		ErrorActionPref:  ssm.state.ErrorActionPref,
		LastSync:         ssm.state.LastSync,
	}

	for k, v := range ssm.state.Variables {
		state.Variables[k] = v
	}
	for k, v := range ssm.state.Functions {
		state.Functions[k] = v
	}
	copy(state.Modules, ssm.state.Modules)

	return state
}

// ClearVariables removes all tracked variables
func (ssm *SessionStateManager) ClearVariables() {
	ssm.mutex.Lock()
	defer ssm.mutex.Unlock()

	ssm.state.Variables = make(map[string]VariableInfo)
}

// ClearFunctions removes all tracked functions
func (ssm *SessionStateManager) ClearFunctions() {
	ssm.mutex.Lock()
	defer ssm.mutex.Unlock()

	ssm.state.Functions = make(map[string]FunctionInfo)
}

// Reset resets the session state to defaults
func (ssm *SessionStateManager) Reset() {
	ssm.mutex.Lock()
	defer ssm.mutex.Unlock()

	currentDir, _ := os.Getwd()
	ssm.state = &SessionState{
		CurrentDirectory: currentDir,
		Variables:        make(map[string]VariableInfo),
		Functions:        make(map[string]FunctionInfo),
		Modules:          make([]ModuleInfo, 0),
		PSVersion:        "",
		LastExitCode:     0,
		ErrorActionPref:  "Continue",
		LastSync:         time.Now(),
	}
}
