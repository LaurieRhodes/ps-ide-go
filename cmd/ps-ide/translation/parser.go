package translation

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// OutputParser handles CLIXML deserialization from PowerShell
type OutputParser struct {
	ansiRegex *regexp.Regexp
}

// NewOutputParser creates a new output parser
func NewOutputParser() *OutputParser {
	DebugLog("OutputParser created")
	return &OutputParser{
		ansiRegex: regexp.MustCompile(`\x1b\[([0-9;]+)m`),
	}
}

// CLIXML structure definitions
type CLIXMLObjs struct {
	XMLName xml.Name    `xml:"Objs"`
	Version string      `xml:"Version,attr"`
	Objects []CLIXMLObj `xml:"Obj"`
}

type CLIXMLObj struct {
	RefID    string       `xml:"RefId,attr"`
	TypeName []string     `xml:"TN>T"`
	ToString string       `xml:"ToString"`
	Props    []CLIXMLProp `xml:"Props>*"`
	MS       []CLIXMLProp `xml:"MS>*"`
	S        string       `xml:"S,attr"` // Stream attribute
}

type CLIXMLProp struct {
	XMLName xml.Name
	Name    string `xml:"N,attr"`
	Value   string `xml:",chardata"`
}

// Parse parses CLIXML data and returns structured output
func (op *OutputParser) Parse(xmlData []byte) ([]PSOutput, error) {
	DebugLog("Parser.Parse called with %d bytes of data", len(xmlData))

	// Handle empty input
	if len(xmlData) == 0 {
		DebugLog("Parser.Parse: empty input, returning empty result")
		return []PSOutput{}, nil
	}

	DebugLogRaw("XML INPUT", string(xmlData))

	// Try to parse as CLIXML
	var objs CLIXMLObjs
	if err := xml.Unmarshal(xmlData, &objs); err != nil {
		DebugLog("Parser.Parse: XML parsing failed: %v, falling back to plain text", err)
		// If not valid XML, treat as plain text
		results := op.parsePlainText(string(xmlData))
		DebugLog("Parser.Parse: parsed as plain text, %d output objects", len(results))
		return results, nil
	}

	DebugLog("Parser.Parse: successfully parsed XML with %d objects", len(objs.Objects))

	// Convert CLIXML objects to PSOutput
	var results []PSOutput
	for i, obj := range objs.Objects {
		DebugLog("Parser.Parse: converting object %d, Stream=%q, ToString=%q", i, obj.S, obj.ToString)
		output := op.convertObject(obj)
		results = append(results, output)
	}

	DebugLog("Parser.Parse: complete, returning %d PSOutput objects", len(results))
	return results, nil
}

// convertObject converts a CLIXML object to PSOutput
func (op *OutputParser) convertObject(obj CLIXMLObj) PSOutput {
	stream := op.determineStream(obj)
	DebugLog("convertObject: determined stream type: %v for S attribute: %q", stream, obj.S)

	output := PSOutput{
		Stream:       stream,
		Content:      obj.ToString,
		ANSISegments: []ANSISegment{},
		ObjectData:   nil,
		IsFormatted:  false,
		Timestamp:    time.Now(),
	}

	// Parse ANSI codes if present in ToString
	if obj.ToString != "" {
		output.ANSISegments = op.ParseANSI(obj.ToString)
		output.IsFormatted = len(output.ANSISegments) > 0
		if output.IsFormatted {
			DebugLog("convertObject: found %d ANSI segments in content", len(output.ANSISegments))
		}
	}

	return output
}

// determineStream determines which stream the output belongs to
func (op *OutputParser) determineStream(obj CLIXMLObj) StreamType {
	// Check the S attribute (stream indicator)
	switch strings.ToLower(obj.S) {
	case "error":
		return ErrorStream
	case "warning":
		return WarningStream
	case "verbose":
		return VerboseStream
	case "debug":
		return DebugStream
	case "progress":
		return ProgressStream
	case "information":
		return InformationStream
	default:
		return OutputStream
	}
}

// ParseANSI parses ANSI escape sequences into segments
func (op *OutputParser) ParseANSI(text string) []ANSISegment {
	if !strings.Contains(text, "\x1b[") {
		// No ANSI codes, return single segment
		return []ANSISegment{{
			Text:      text,
			FGColor:   37, // Default white
			BGColor:   40, // Default black
			Bold:      false,
			Underline: false,
			Italic:    false,
		}}
	}

	DebugLog("ParseANSI: parsing text with ANSI codes, length=%d", len(text))

	var segments []ANSISegment
	currentSegment := ANSISegment{
		FGColor:   37,
		BGColor:   40,
		Bold:      false,
		Underline: false,
		Italic:    false,
	}

	// Split by ANSI codes
	matches := op.ansiRegex.FindAllStringSubmatchIndex(text, -1)
	DebugLog("ParseANSI: found %d ANSI code matches", len(matches))
	lastEnd := 0

	for i, match := range matches {
		// Add text before this code
		if match[0] > lastEnd {
			currentSegment.Text = text[lastEnd:match[0]]
			if currentSegment.Text != "" {
				segments = append(segments, currentSegment)
				DebugLog("ParseANSI: added segment %d with text: %q", len(segments)-1, currentSegment.Text)
			}
		}

		// Parse the ANSI code
		codeStr := text[match[2]:match[3]]
		codes := strings.Split(codeStr, ";")
		DebugLog("ParseANSI: match %d, codes=%v", i, codes)
		currentSegment = op.applyANSICodes(currentSegment, codes)

		lastEnd = match[1]
	}

	// Add remaining text
	if lastEnd < len(text) {
		currentSegment.Text = text[lastEnd:]
		if currentSegment.Text != "" {
			segments = append(segments, currentSegment)
			DebugLog("ParseANSI: added final segment with text: %q", currentSegment.Text)
		}
	}

	DebugLog("ParseANSI: complete, %d segments total", len(segments))
	return segments
}

// applyANSICodes applies ANSI codes to a segment
func (op *OutputParser) applyANSICodes(segment ANSISegment, codes []string) ANSISegment {
	for _, code := range codes {
		switch code {
		case "0":
			// Reset
			segment.FGColor = 37
			segment.BGColor = 40
			segment.Bold = false
			segment.Underline = false
			segment.Italic = false
		case "1":
			segment.Bold = true
		case "4":
			segment.Underline = true
		case "3":
			segment.Italic = true
		case "22":
			segment.Bold = false
		case "24":
			segment.Underline = false
		case "23":
			segment.Italic = false
		// Foreground colors (30-37, 90-97)
		case "30":
			segment.FGColor = 30
		case "31":
			segment.FGColor = 31
		case "32":
			segment.FGColor = 32
		case "33":
			segment.FGColor = 33
		case "34":
			segment.FGColor = 34
		case "35":
			segment.FGColor = 35
		case "36":
			segment.FGColor = 36
		case "37":
			segment.FGColor = 37
		case "90":
			segment.FGColor = 90
		case "91":
			segment.FGColor = 91
		case "92":
			segment.FGColor = 92
		case "93":
			segment.FGColor = 93
		case "94":
			segment.FGColor = 94
		case "95":
			segment.FGColor = 95
		case "96":
			segment.FGColor = 96
		case "97":
			segment.FGColor = 97
		// Background colors (40-47, 100-107)
		case "40":
			segment.BGColor = 40
		case "41":
			segment.BGColor = 41
		case "42":
			segment.BGColor = 42
		case "43":
			segment.BGColor = 43
		case "44":
			segment.BGColor = 44
		case "45":
			segment.BGColor = 45
		case "46":
			segment.BGColor = 46
		case "47":
			segment.BGColor = 47
		case "100":
			segment.BGColor = 100
		case "101":
			segment.BGColor = 101
		case "102":
			segment.BGColor = 102
		case "103":
			segment.BGColor = 103
		case "104":
			segment.BGColor = 104
		case "105":
			segment.BGColor = 105
		case "106":
			segment.BGColor = 106
		case "107":
			segment.BGColor = 107
		}
	}
	return segment
}

// parsePlainText handles non-XML output as plain text
func (op *OutputParser) parsePlainText(text string) []PSOutput {
	DebugLog("parsePlainText: parsing %d bytes as plain text", len(text))

	// Split by lines and create output for each
	lines := strings.Split(text, "\n")
	var results []PSOutput

	for i, line := range lines {
		if line == "" {
			continue
		}

		output := PSOutput{
			Stream:       OutputStream,
			Content:      line,
			ANSISegments: op.ParseANSI(line),
			ObjectData:   nil,
			IsFormatted:  strings.Contains(line, "\x1b["),
			Timestamp:    time.Now(),
		}

		results = append(results, output)
		DebugLog("parsePlainText: added line %d: %q", i, line)
	}

	DebugLog("parsePlainText: complete, %d output objects", len(results))
	return results
}

// FormatOutput formats PSOutput for display (removes ANSI codes but preserves meaning)
func (op *OutputParser) FormatOutput(output PSOutput) string {
	if !output.IsFormatted {
		return output.Content
	}

	// Build plain text from segments
	var builder strings.Builder
	for _, segment := range output.ANSISegments {
		builder.WriteString(segment.Text)
	}

	result := builder.String()
	DebugLog("FormatOutput: formatted content (length=%d): %q", len(result), result)
	return result
}

// FormatWithANSI formats PSOutput with ANSI codes for terminal display
func (op *OutputParser) FormatWithANSI(output PSOutput) string {
	if !output.IsFormatted {
		return output.Content
	}

	var builder strings.Builder

	for i, segment := range output.ANSISegments {
		// Build ANSI code
		var codes []string

		if segment.Bold {
			codes = append(codes, "1")
		}
		if segment.Underline {
			codes = append(codes, "4")
		}
		if segment.Italic {
			codes = append(codes, "3")
		}

		// Add color codes
		codes = append(codes, fmt.Sprintf("%d", segment.FGColor))
		if segment.BGColor != 40 { // Only add background if not default
			codes = append(codes, fmt.Sprintf("%d", segment.BGColor))
		}

		// Write ANSI sequence and text
		formatted := fmt.Sprintf("\x1b[%sm%s\x1b[0m", strings.Join(codes, ";"), segment.Text)
		builder.WriteString(formatted)

		if IsDebugEnabled() {
			DebugLog("FormatWithANSI: segment %d: codes=%v, text=%q", i, codes, segment.Text)
		}
	}

	return builder.String()
}

// ExtractErrorMessage extracts error message from output
func (op *OutputParser) ExtractErrorMessage(output PSOutput) string {
	if output.Stream == ErrorStream {
		DebugLog("ExtractErrorMessage: extracted error: %q", output.Content)
		return output.Content
	}
	return ""
}

// IsProgressRecord checks if output is a progress record
func (op *OutputParser) IsProgressRecord(output PSOutput) bool {
	isProgress := output.Stream == ProgressStream
	if isProgress {
		DebugLog("IsProgressRecord: detected progress record")
	}
	return isProgress
}

// GetStreamColor returns the appropriate color for a stream type
func (op *OutputParser) GetStreamColor(stream StreamType) (fg int, bg int) {
	switch stream {
	case ErrorStream:
		return 91, 40 // Bright red on black
	case WarningStream:
		return 93, 40 // Bright yellow on black
	case VerboseStream:
		return 96, 40 // Bright cyan on black
	case DebugStream:
		return 95, 40 // Bright magenta on black
	case InformationStream:
		return 92, 40 // Bright green on black
	default:
		return 97, 40 // Bright white on black (default)
	}
}

// StripANSI removes all ANSI escape codes from text
func (op *OutputParser) StripANSI(text string) string {
	result := op.ansiRegex.ReplaceAllString(text, "")
	if result != text {
		DebugLog("StripANSI: stripped ANSI codes, original length=%d, result length=%d", len(text), len(result))
	}
	return result
}

// HasANSICodes checks if text contains ANSI escape codes
func (op *OutputParser) HasANSICodes(text string) bool {
	hasCodes := strings.Contains(text, "\x1b[")
	if hasCodes {
		DebugLog("HasANSICodes: text contains ANSI codes")
	}
	return hasCodes
}
