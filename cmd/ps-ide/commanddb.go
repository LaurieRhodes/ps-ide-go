package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// CommandDatabase stores all available commands and their help
type CommandDatabase struct {
	Commands map[string]*CommandHelp // Name -> Help
	Modules  map[string][]string     // Module -> Command names
	mutex    sync.RWMutex
	loaded   bool
}

// CommandInfo represents basic command information from Get-Command
type CommandInfo struct {
	Name        string      `json:"Name"`
	CommandType interface{} `json:"CommandType"` // Can be string or number
	Module      string      `json:"ModuleName"`
	Version     interface{} `json:"Version"` // Can be string or object
}

// NewCommandDatabase creates a new command database
func NewCommandDatabase() *CommandDatabase {
	return &CommandDatabase{
		Commands: make(map[string]*CommandHelp),
		Modules:  make(map[string][]string),
		loaded:   false,
	}
}

// LoadHelp loads help files from PowerShell help directories
func (db *CommandDatabase) LoadHelp() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Search paths for help files
	searchPaths := []string{
		filepath.Join(homeDir, ".local/share/powershell/Modules"),
		"/opt/microsoft/powershell/7/Modules",
		"/usr/local/share/powershell/Modules",
	}

	// Load help from all paths
	for _, basePath := range searchPaths {
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Continue on errors
			}

			// Look for help XML files
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), "-help.xml") {
				log.Printf("Loading help from: %s", path)
				commands, err := parseMAMLFile(path)
				if err != nil {
					log.Printf("Error parsing %s: %v", path, err)
					return nil // Continue on parse errors
				}

				// Extract module name from path
				moduleName := extractModuleName(path)

				// Add commands to database
				for i := range commands {
					cmd := &commands[i]
					cmd.Module = moduleName
					db.Commands[cmd.Name] = cmd

					// Add to module index
					if _, exists := db.Modules[moduleName]; !exists {
						db.Modules[moduleName] = []string{}
					}
					db.Modules[moduleName] = append(db.Modules[moduleName], cmd.Name)
				}
			}
			return nil
		})

		if err != nil {
			log.Printf("Error walking path %s: %v", basePath, err)
		}
	}

	db.loaded = true
	log.Printf("Loaded %d commands from %d modules", len(db.Commands), len(db.Modules))
	return nil
}

// LoadFromPowerShell loads command list directly from PowerShell
func (db *CommandDatabase) LoadFromPowerShell() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	// Execute Get-Command and convert to JSON - select only string fields
	cmd := exec.Command("pwsh", "-NoProfile", "-Command",
		`Get-Command | Select-Object @{N='Name';E={$_.Name}}, @{N='CommandType';E={$_.CommandType.ToString()}}, @{N='ModuleName';E={$_.ModuleName}} | ConvertTo-Json -Compress`)

	output, err := cmd.Output()
	if err != nil {
		return err
	}

	// Handle single item vs array
	var commands []CommandInfoSimple

	// Try to unmarshal as array first
	err = json.Unmarshal(output, &commands)
	if err != nil {
		// Try as single object
		var singleCmd CommandInfoSimple
		err = json.Unmarshal(output, &singleCmd)
		if err != nil {
			return err
		}
		commands = []CommandInfoSimple{singleCmd}
	}

	// Add commands to database (without full help)
	for _, cmdInfo := range commands {
		if cmdInfo.Name == "" {
			continue
		}
		if _, exists := db.Commands[cmdInfo.Name]; !exists {
			// Create basic help entry
			help := &CommandHelp{
				Name:   cmdInfo.Name,
				Module: cmdInfo.ModuleName,
			}
			db.Commands[cmdInfo.Name] = help

			// Add to module index
			if cmdInfo.ModuleName != "" {
				if _, exists := db.Modules[cmdInfo.ModuleName]; !exists {
					db.Modules[cmdInfo.ModuleName] = []string{}
				}
				db.Modules[cmdInfo.ModuleName] = append(db.Modules[cmdInfo.ModuleName], cmdInfo.Name)
			}
		}
	}

	db.loaded = true
	log.Printf("Loaded %d commands from PowerShell", len(commands))
	return nil
}

// CommandInfoSimple is a simplified struct for JSON parsing
type CommandInfoSimple struct {
	Name        string `json:"Name"`
	CommandType string `json:"CommandType"`
	ModuleName  string `json:"ModuleName"`
}

// Search finds commands matching the query and optional module filter
func (db *CommandDatabase) Search(query string, moduleName string) []*CommandHelp {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	var results []*CommandHelp
	query = strings.ToLower(query)

	for _, cmd := range db.Commands {
		// Filter by module if specified
		if moduleName != "" && moduleName != "All" && cmd.Module != moduleName {
			continue
		}

		// Filter by query
		if query == "" || strings.Contains(strings.ToLower(cmd.Name), query) {
			results = append(results, cmd)
		}
	}

	return results
}

// GetCommandHelp retrieves detailed help for a specific command from PowerShell
func (db *CommandDatabase) GetCommandHelp(cmdName string) (*CommandHelp, error) {
	// First check if we have cached help with synopsis
	db.mutex.RLock()
	if cmd, exists := db.Commands[cmdName]; exists && cmd.Synopsis != "" {
		db.mutex.RUnlock()
		return cmd, nil
	}
	db.mutex.RUnlock()

	// Sanitize command name to prevent injection (only allow alphanumeric, -, _)
	for _, c := range cmdName {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return nil, fmt.Errorf("invalid command name: %s", cmdName)
		}
	}

	// Get help from PowerShell - embed command name directly in script
	psScript := fmt.Sprintf(`
$cmdName = '%s'
$cmd = Get-Command $cmdName -ErrorAction SilentlyContinue
if (-not $cmd) { exit 1 }

$help = Get-Help $cmdName -Full -ErrorAction SilentlyContinue
$result = @{
    Name = $cmd.Name
    Module = $cmd.ModuleName
    Synopsis = ''
    Description = ''
    ParameterSets = @()
    Parameters = @()
}

if ($help.Synopsis) { $result.Synopsis = $help.Synopsis.Trim() }
if ($help.Description) {
    $result.Description = ($help.Description | ForEach-Object { $_.Text }) -join [char]10
}

# Get parameter sets from Get-Command (has proper names)
if ($cmd.ParameterSets) {
    foreach ($paramSet in $cmd.ParameterSets) {
        $setParams = @()
        foreach ($p in $paramSet.Parameters) {
            # Skip common parameters
            $commonParams = @('Verbose','Debug','ErrorAction','WarningAction','InformationAction',
                'ProgressAction','ErrorVariable','WarningVariable','InformationVariable',
                'OutVariable','OutBuffer','PipelineVariable','WhatIf','Confirm')
            if ($p.Name -notin $commonParams) {
                $setParams += $p.Name
            }
        }
        $result.ParameterSets += @{
            Name = $paramSet.Name
            Parameters = $setParams
            IsDefault = $paramSet.IsDefault
        }
    }
}

# Get parameter details
if ($cmd.Parameters) {
    foreach ($key in $cmd.Parameters.Keys) {
        $param = $cmd.Parameters[$key]
        $helpParam = $help.parameters.parameter | Where-Object { $_.name -eq $key }
        
        $paramInfo = @{
            Name = $key
            Type = $param.ParameterType.Name
            Required = $false
            Position = -1
            Pipeline = $false
            Description = ''
            IsSwitchParameter = ($param.ParameterType.Name -eq 'SwitchParameter')
            ParameterSetName = '__AllParameterSets'
        }
        
        # Check parameter attributes
        foreach ($attr in $param.Attributes) {
            if ($attr -is [System.Management.Automation.ParameterAttribute]) {
                if ($attr.Mandatory) { $paramInfo.Required = $true }
                if ($attr.Position -ge 0) { $paramInfo.Position = $attr.Position }
                if ($attr.ValueFromPipeline -or $attr.ValueFromPipelineByPropertyName) { $paramInfo.Pipeline = $true }
                if ($attr.ParameterSetName) { $paramInfo.ParameterSetName = $attr.ParameterSetName }
            }
        }
        
        # Get description from help
        if ($helpParam -and $helpParam.description) {
            $paramInfo.Description = ($helpParam.description | ForEach-Object { $_.Text }) -join ' '
        }
        
        $result.Parameters += $paramInfo
    }
}

$result | ConvertTo-Json -Depth 5 -Compress
`, cmdName)

	cmd := exec.Command("pwsh", "-NoProfile", "-Command", psScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("PowerShell error for %s: %v, output: %s", cmdName, err, string(output))
		return nil, fmt.Errorf("PowerShell error for %s: %v", cmdName, err)
	}

	// Check for empty output
	if len(output) == 0 {
		return nil, fmt.Errorf("no help found for %s", cmdName)
	}

	// Check if output is valid JSON (starts with {)
	trimmedOutput := strings.TrimSpace(string(output))
	if len(trimmedOutput) == 0 {
		return nil, fmt.Errorf("no help found for %s", cmdName)
	}
	if !strings.HasPrefix(trimmedOutput, "{") {
		maxLen := len(trimmedOutput)
		if maxLen > 200 {
			maxLen = 200
		}
		log.Printf("Invalid JSON response for %s: %s", cmdName, trimmedOutput[:maxLen])
		return nil, fmt.Errorf("invalid help response for %s", cmdName)
	}

	// Parse JSON result
	var helpData struct {
		Name          string `json:"Name"`
		Module        string `json:"Module"`
		Synopsis      string `json:"Synopsis"`
		Description   string `json:"Description"`
		ParameterSets []struct {
			Name       string   `json:"Name"`
			Parameters []string `json:"Parameters"`
			IsDefault  bool     `json:"IsDefault"`
		} `json:"ParameterSets"`
		Parameters []struct {
			Name              string `json:"Name"`
			Type              string `json:"Type"`
			Required          bool   `json:"Required"`
			Position          int    `json:"Position"`
			Pipeline          bool   `json:"Pipeline"`
			Description       string `json:"Description"`
			IsSwitchParameter bool   `json:"IsSwitchParameter"`
			ParameterSetName  string `json:"ParameterSetName"`
		} `json:"Parameters"`
	}

	err = json.Unmarshal([]byte(trimmedOutput), &helpData)
	if err != nil {
		log.Printf("JSON parse error for %s: %v", cmdName, err)
		return nil, fmt.Errorf("failed to parse help for %s: %v", cmdName, err)
	}

	// Convert to CommandHelp
	help := &CommandHelp{
		Name:        helpData.Name,
		Module:      helpData.Module,
		Synopsis:    helpData.Synopsis,
		Description: helpData.Description,
	}

	// Convert parameter sets
	for _, ps := range helpData.ParameterSets {
		help.Syntax = append(help.Syntax, ParameterSet{
			Name:       ps.Name,
			Parameters: ps.Parameters,
			IsDefault:  ps.IsDefault,
		})
	}

	// Convert parameters
	for _, p := range helpData.Parameters {
		help.Parameters = append(help.Parameters, Parameter{
			Name:              p.Name,
			Type:              p.Type,
			Required:          p.Required,
			Position:          p.Position,
			Pipeline:          p.Pipeline,
			Description:       p.Description,
			IsSwitchParameter: p.IsSwitchParameter,
			ParameterSetName:  p.ParameterSetName,
		})
	}

	// Cache the result
	db.mutex.Lock()
	db.Commands[cmdName] = help
	db.mutex.Unlock()

	return help, nil
}

// GetModules returns list of all module names
func (db *CommandDatabase) GetModules() []string {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	modules := make([]string, 0, len(db.Modules))
	for module := range db.Modules {
		if module != "" {
			modules = append(modules, module)
		}
	}
	return modules
}

// GetCommand returns basic command info from cache (without fetching detailed help)
func (db *CommandDatabase) GetCommand(cmdName string) *CommandHelp {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	if cmd, exists := db.Commands[cmdName]; exists {
		return cmd
	}
	return nil
}

// IsLoaded returns whether the database has been loaded
func (db *CommandDatabase) IsLoaded() bool {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	return db.loaded
}

// extractModuleName extracts the module name from a help file path
func extractModuleName(path string) string {
	// Path typically looks like: .../Modules/ModuleName/en-US/ModuleName-help.xml
	parts := strings.Split(filepath.ToSlash(path), "/")

	// Find "Modules" in path
	for i, part := range parts {
		if part == "Modules" && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	// Fallback: try to extract from filename
	filename := filepath.Base(path)
	filename = strings.TrimSuffix(filename, "-help.xml")
	filename = strings.TrimSuffix(filename, ".dll-Help.xml")

	return filename
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
