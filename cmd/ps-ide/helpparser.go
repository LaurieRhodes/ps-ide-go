package main

import (
	"encoding/xml"
	"io/ioutil"
	"strings"
)

// MAML XML structure definitions
type HelpItems struct {
	XMLName  xml.Name  `xml:"helpItems"`
	Commands []Command `xml:"command"`
}

type Command struct {
	XMLName     xml.Name        `xml:"command"`
	Details     CommandDetails  `xml:"details"`
	Description Description     `xml:"description"`
	Syntax      CommandSyntax   `xml:"syntax"`
	Parameters  ParametersBlock `xml:"parameters"`
	Examples    Examples        `xml:"examples"`
}

type CommandDetails struct {
	Name        string      `xml:"name"`
	Verb        string      `xml:"verb"`
	Noun        string      `xml:"noun"`
	Description Description `xml:"description"`
}

type Description struct {
	Para []string `xml:"para"`
}

type CommandSyntax struct {
	SyntaxItems []SyntaxItem `xml:"syntaxItem"`
}

type SyntaxItem struct {
	Name       string              `xml:"name"`
	Parameters []ParameterInSyntax `xml:"parameter"`
}

type ParameterInSyntax struct {
	Name          string        `xml:"name,attr"`
	Required      string        `xml:"required,attr"`
	Position      string        `xml:"position,attr"`
	PipelineInput string        `xml:"pipelineInput,attr"`
	Type          ParameterType `xml:"parameterValue"`
}

type ParameterType struct {
	Required string `xml:"required,attr"`
	Value    string `xml:",chardata"`
}

type ParametersBlock struct {
	Parameters []ParameterDetail `xml:"parameter"`
}

type ParameterDetail struct {
	Name          string        `xml:"name,attr"`
	Required      string        `xml:"required,attr"`
	Position      string        `xml:"position,attr"`
	PipelineInput string        `xml:"pipelineInput,attr"`
	Description   Description   `xml:"description"`
	Type          ParameterType `xml:"parameterValue"`
	DefaultValue  string        `xml:"defaultValue"`
}

type Examples struct {
	Example []Example `xml:"example"`
}

type Example struct {
	Title       string      `xml:"title"`
	Code        string      `xml:"code"`
	Description Description `xml:"description"`
}

// Structured help data for internal use
type CommandHelp struct {
	Name        string
	Module      string
	Synopsis    string
	Description string
	Syntax      []ParameterSet
	Parameters  []Parameter
	Examples    []ExampleInfo
}

type ParameterSet struct {
	Name       string
	Parameters []string // Parameter names in this set
	IsDefault  bool     // Whether this is the default parameter set
}

type Parameter struct {
	Name              string
	Type              string
	Required          bool
	Position          int
	Pipeline          bool
	PipelineByName    bool
	DefaultValue      string
	Description       string
	ValidValues       []string
	IsSwitchParameter bool
	ParameterSetName  string
}

type ExampleInfo struct {
	Title       string
	Code        string
	Description string
}

// parseMAMLFile parses a MAML help XML file and returns command help data
func parseMAMLFile(filePath string) ([]CommandHelp, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var helpItems HelpItems
	err = xml.Unmarshal(data, &helpItems)
	if err != nil {
		return nil, err
	}

	var commands []CommandHelp
	for _, cmd := range helpItems.Commands {
		cmdHelp := convertCommandToHelp(cmd)
		commands = append(commands, cmdHelp)
	}

	return commands, nil
}

// convertCommandToHelp converts raw XML structure to our internal format
func convertCommandToHelp(cmd Command) CommandHelp {
	help := CommandHelp{
		Name:        cmd.Details.Name,
		Synopsis:    joinParas(cmd.Details.Description.Para),
		Description: joinParas(cmd.Description.Para),
	}

	// Convert syntax items to parameter sets
	for i, syntaxItem := range cmd.Syntax.SyntaxItems {
		paramSet := ParameterSet{
			Name:       syntaxItem.Name,
			Parameters: []string{},
		}

		for _, param := range syntaxItem.Parameters {
			paramSet.Parameters = append(paramSet.Parameters, param.Name)
		}

		if paramSet.Name == "" {
			paramSet.Name = "ParameterSet" + string(rune(i+1))
		}

		help.Syntax = append(help.Syntax, paramSet)
	}

	// Convert parameter details
	for _, param := range cmd.Parameters.Parameters {
		p := Parameter{
			Name:         param.Name,
			Type:         param.Type.Value,
			Required:     param.Required == "true",
			DefaultValue: param.DefaultValue,
			Description:  joinParas(param.Description.Para),
		}

		// Detect switch parameters
		if strings.EqualFold(param.Type.Value, "SwitchParameter") {
			p.IsSwitchParameter = true
		}

		// Parse position
		if param.Position != "named" && param.Position != "" {
			// Position is a number or "0", "1", etc.
			// We'll keep it simple and just mark named vs positional
			p.Position = 0
			if param.Position != "named" {
				p.Position = 1
			}
		}

		// Parse pipeline input
		if strings.Contains(strings.ToLower(param.PipelineInput), "true") {
			p.Pipeline = true
			if strings.Contains(strings.ToLower(param.PipelineInput), "bypropertyname") {
				p.PipelineByName = true
			}
		}

		help.Parameters = append(help.Parameters, p)
	}

	// Convert examples
	for _, ex := range cmd.Examples.Example {
		example := ExampleInfo{
			Title:       ex.Title,
			Code:        strings.TrimSpace(ex.Code),
			Description: joinParas(ex.Description.Para),
		}
		help.Examples = append(help.Examples, example)
	}

	return help
}

// joinParas joins multiple paragraph strings with newlines
func joinParas(paras []string) string {
	var result []string
	for _, p := range paras {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return strings.Join(result, "\n")
}
