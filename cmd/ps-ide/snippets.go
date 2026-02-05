package main

import (
	"strings"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// Snippet represents a code template
type Snippet struct {
	Name        string
	Description string
	Code        string
}

// Common PowerShell snippets matching Windows PowerShell ISE
var powerShellSnippets = []Snippet{
	{
		Name:        "Cmdlet (advanced function)",
		Description: "Advanced function with CmdletBinding",
		Code: `function Verb-Noun {
    <#
    .SYNOPSIS
        Brief description
    .DESCRIPTION
        Detailed description
    .PARAMETER Name
        Parameter description
    .EXAMPLE
        Verb-Noun -Name "value"
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$true)]
        [string]$Name
    )
    
    begin {
    }
    
    process {
    }
    
    end {
    }
}`,
	},
	{
		Name:        "Cmdlet (advanced function) - Complete",
		Description: "Advanced function with full parameter attributes",
		Code: `function Verb-Noun {
    <#
    .SYNOPSIS
        Brief description
    .DESCRIPTION
        Detailed description
    .PARAMETER Name
        Parameter description
    .EXAMPLE
        Verb-Noun -Name "value"
    .NOTES
        Author: Your Name
        Date: ` + "`" + `(Get-Date -Format "yyyy-MM-dd")` + "`" + `
    #>
    [CmdletBinding(SupportsShouldProcess=$true)]
    param(
        [Parameter(
            Mandatory=$true,
            ValueFromPipeline=$true,
            ValueFromPipelineByPropertyName=$true,
            Position=0
        )]
        [ValidateNotNullOrEmpty()]
        [string]$Name
    )
    
    begin {
        Write-Verbose "Starting $($MyInvocation.MyCommand)"
    }
    
    process {
        if ($PSCmdlet.ShouldProcess($Name, "Process item")) {
            # Process logic here
        }
    }
    
    end {
        Write-Verbose "Completed $($MyInvocation.MyCommand)"
    }
}`,
	},
	{
		Name:        "class - simple",
		Description: "Simple PowerShell class definition",
		Code: `class MyClass {
    [string]$Property1
    [int]$Property2
    
    MyClass([string]$prop1, [int]$prop2) {
        $this.Property1 = $prop1
        $this.Property2 = $prop2
    }
    
    [string] ToString() {
        return "$($this.Property1): $($this.Property2)"
    }
}`,
	},
	{
		Name:        "class - with methods",
		Description: "Class with properties and methods",
		Code: `class MyClass {
    # Properties
    [string]$Name
    [int]$Value
    
    # Constructor
    MyClass([string]$name, [int]$value) {
        $this.Name = $name
        $this.Value = $value
    }
    
    # Method
    [void] DoSomething() {
        Write-Host "Processing $($this.Name)"
    }
    
    # Method with return
    [string] GetInfo() {
        return "$($this.Name) = $($this.Value)"
    }
}`,
	},
	{
		Name:        "Comment block",
		Description: "Multi-line comment block",
		Code: `<#
    Description:
    
    Author:
    Date:
#>`,
	},
	{
		Name:        "do-until",
		Description: "Do-until loop",
		Code: `do {
    # Loop body
} until ($condition)`,
	},
	{
		Name:        "do-while",
		Description: "Do-while loop",
		Code: `do {
    # Loop body
} while ($condition)`,
	},
	{
		Name:        "for",
		Description: "For loop",
		Code: `for ($i = 0; $i -lt 10; $i++) {
    # Loop body
}`,
	},
	{
		Name:        "foreach",
		Description: "ForEach loop",
		Code: `foreach ($item in $collection) {
    # Loop body
}`,
	},
	{
		Name:        "function - simple",
		Description: "Simple function definition",
		Code: `function FunctionName {
    param(
        [string]$Parameter1,
        [int]$Parameter2
    )
    
    # Function body
}`,
	},
	{
		Name:        "if-else",
		Description: "If-else statement",
		Code: `if ($condition) {
    # True block
}
else {
    # False block
}`,
	},
	{
		Name:        "if-elseif-else",
		Description: "If-elseif-else statement",
		Code: `if ($condition1) {
    # Condition 1 block
}
elseif ($condition2) {
    # Condition 2 block
}
else {
    # Default block
}`,
	},
	{
		Name:        "switch",
		Description: "Switch statement",
		Code: `switch ($variable) {
    "value1" {
        # Case 1
    }
    "value2" {
        # Case 2
    }
    default {
        # Default case
    }
}`,
	},
	{
		Name:        "try-catch",
		Description: "Try-catch error handling",
		Code: `try {
    # Code that might throw
}
catch {
    Write-Error "An error occurred: $_"
}`,
	},
	{
		Name:        "try-catch-finally",
		Description: "Try-catch-finally error handling",
		Code: `try {
    # Code that might throw
}
catch {
    Write-Error "An error occurred: $_"
}
finally {
    # Cleanup code
}`,
	},
	{
		Name:        "while",
		Description: "While loop",
		Code: `while ($condition) {
    # Loop body
}`,
	},
	{
		Name:        "parameter validation",
		Description: "Parameter with validation attributes",
		Code: `[Parameter(Mandatory=$true)]
[ValidateNotNullOrEmpty()]
[ValidateLength(1, 50)]
[ValidatePattern("^[a-zA-Z]+$")]
[string]$ParameterName`,
	},
	{
		Name:        "enum",
		Description: "Enumeration definition",
		Code: `enum MyEnum {
    Value1
    Value2
    Value3
}`,
	},
	{
		Name:        "region",
		Description: "Collapsible region",
		Code: `#region RegionName

# Code here

#endregion`,
	},
}

// showSnippetsDialog displays a dialog to select and insert a snippet
func showSnippetsDialog() {
	dialog, _ := gtk.DialogNew()
	dialog.SetTitle("Insert Snippet")
	dialog.SetTransientFor(mainWindow)
	dialog.SetModal(true)
	dialog.SetDefaultSize(600, 400)

	// Create list store for snippets
	listStore, _ := gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING)
	for _, snippet := range powerShellSnippets {
		iter := listStore.Append()
		listStore.Set(iter, []int{0, 1}, []interface{}{snippet.Name, snippet.Description})
	}

	// Create tree view
	treeView, _ := gtk.TreeViewNew()
	treeView.SetModel(listStore)
	treeView.SetHeadersVisible(true)

	// Name column
	nameRenderer, _ := gtk.CellRendererTextNew()
	nameColumn, _ := gtk.TreeViewColumnNewWithAttribute("Snippet", nameRenderer, "text", 0)
	nameColumn.SetExpand(false)
	nameColumn.SetMinWidth(250)
	treeView.AppendColumn(nameColumn)

	// Description column
	descRenderer, _ := gtk.CellRendererTextNew()
	descColumn, _ := gtk.TreeViewColumnNewWithAttribute("Description", descRenderer, "text", 1)
	descColumn.SetExpand(true)
	treeView.AppendColumn(descColumn)

	// Preview text view
	previewView, _ := gtk.TextViewNew()
	previewView.SetEditable(false)
	previewView.SetWrapMode(gtk.WRAP_NONE)
	previewView.SetMonospace(true)
	previewBuffer, _ := previewView.GetBuffer()

	// Selection changed handler
	selection, _ := treeView.GetSelection()
	selection.Connect("changed", func() {
		model, iter, ok := selection.GetSelected()
		if !ok {
			return
		}

		value, _ := model.(*gtk.TreeModel).GetValue(iter, 0)
		snippetName, _ := value.GetString()

		// Find snippet and show preview
		for _, snippet := range powerShellSnippets {
			if snippet.Name == snippetName {
				previewBuffer.SetText(snippet.Code)
				break
			}
		}
	})

	// Double-click to insert
	treeView.Connect("row-activated", func() {
		dialog.Response(gtk.RESPONSE_OK)
	})

	// Layout
	contentBox, _ := dialog.GetContentArea()

	// Top pane - list
	scrolledList, _ := gtk.ScrolledWindowNew(nil, nil)
	scrolledList.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	scrolledList.Add(treeView)

	// Bottom pane - preview
	previewLabel, _ := gtk.LabelNew("Preview:")
	previewLabel.SetXAlign(0)

	scrolledPreview, _ := gtk.ScrolledWindowNew(nil, nil)
	scrolledPreview.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	scrolledPreview.Add(previewView)
	scrolledPreview.SetSizeRequest(-1, 150)

	// Paned layout
	paned, _ := gtk.PanedNew(gtk.ORIENTATION_VERTICAL)
	paned.Pack1(scrolledList, true, true)

	previewBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)
	previewBox.PackStart(previewLabel, false, false, 5)
	previewBox.PackStart(scrolledPreview, true, true, 0)
	paned.Pack2(previewBox, false, true)

	contentBox.PackStart(paned, true, true, 10)

	// Buttons
	dialog.AddButton("Cancel", gtk.RESPONSE_CANCEL)
	dialog.AddButton("Insert", gtk.RESPONSE_OK)
	dialog.SetDefaultResponse(gtk.RESPONSE_OK)

	// Select first item
	if path, _ := gtk.TreePathNewFromString("0"); path != nil {
		selection.SelectPath(path)
	}

	dialog.ShowAll()
	response := dialog.Run()

	if response == gtk.RESPONSE_OK {
		// Get selected snippet
		model, iter, ok := selection.GetSelected()
		if ok {
			value, _ := model.(*gtk.TreeModel).GetValue(iter, 0)
			snippetName, _ := value.GetString()

			// Find and insert snippet
			for _, snippet := range powerShellSnippets {
				if snippet.Name == snippetName {
					insertSnippetAtCursor(snippet.Code)
					break
				}
			}
		}
	}

	dialog.Destroy()
}

// insertSnippetAtCursor inserts snippet code at the current cursor position
func insertSnippetAtCursor(code string) {
	tab := getCurrentTab()
	if tab == nil || tab.buffer == nil {
		return
	}

	// Get cursor position
	insertMark := tab.buffer.GetInsert()
	iter := tab.buffer.GetIterAtMark(insertMark)

	// Get current line's indentation
	lineStart := iter
	lineStart.SetLineOffset(0)
	lineEnd := tab.buffer.GetIterAtMark(insertMark)

	lineText, _ := tab.buffer.GetText(lineStart, lineEnd, false)
	indentation := getIndentation(lineText)

	// Indent all lines of the snippet
	lines := strings.Split(code, "\n")
	indentedCode := ""
	for i, line := range lines {
		if i > 0 {
			indentedCode += "\n"
		}
		if strings.TrimSpace(line) != "" {
			indentedCode += indentation + line
		} else {
			indentedCode += line
		}
	}

	// Insert the snippet
	tab.buffer.InsertAtCursor(indentedCode)

	// Scroll to cursor
	tab.textView.ScrollMarkOnscreen(insertMark)
}

// getIndentation extracts leading whitespace from a line
func getIndentation(line string) string {
	for i, char := range line {
		if char != ' ' && char != '\t' {
			return line[:i]
		}
	}
	return line
}
