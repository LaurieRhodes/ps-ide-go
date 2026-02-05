package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// CommandAddOn represents the command browser add-on
type CommandAddOn struct {
	container *gtk.Box
	visible   bool

	// Search UI
	nameEntry   *gtk.SearchEntry
	moduleCombo *gtk.ComboBoxText

	// Command list
	commandList *gtk.TreeView
	listStore   *gtk.ListStore
	selection   *gtk.TreeSelection

	// Details area
	detailsBox        *gtk.Box
	nameLabel         *gtk.Label
	moduleLabel       *gtk.Label
	synopsisLabel     *gtk.Label
	paramsHeaderLabel *gtk.Label

	// Parameter UI with wrapping tabs (FlowBox + Stack)
	paramTabBox          *gtk.FlowBox                // Wrapping tab buttons
	paramStack           *gtk.Stack                  // Content pages
	paramWidgets         map[string]*ParameterWidget // parameter name -> widget
	commonParamWidgets   map[string]*ParameterWidget // common parameter widgets
	currentParamSet      int                         // Currently selected parameter set index
	paramSetButtons      []*gtk.ToggleButton         // Tab buttons for styling
	commonParamsExpander *gtk.Expander

	// Action buttons
	runButton    *gtk.Button
	insertButton *gtk.Button
	copyButton   *gtk.Button

	// Data
	database    *CommandDatabase
	selectedCmd *CommandHelp
}

// ParameterWidget holds the UI widget for a parameter
type ParameterWidget struct {
	Name        string
	Type        string
	IsSwitch    bool
	Entry       *gtk.Entry       // For string/value parameters
	CheckButton *gtk.CheckButton // For switch parameters
	Label       *gtk.Label       // Parameter label
}

var commandAddOn *CommandAddOn

// createCommandAddOn creates the command add-on UI
func createCommandAddOn(db *CommandDatabase) *CommandAddOn {
	addon := &CommandAddOn{
		database: db,
		visible:  false,
	}

	// Main container - this will be added to the paned
	addon.container, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	addon.container.SetHExpand(false) // Don't expand horizontally to allow window resize
	addon.container.SetVExpand(true)  // Fill full height

	// Title bar with close button (outside scrolled area)
	titleBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	titleBox.SetMarginStart(5)
	titleBox.SetMarginEnd(5)
	titleBox.SetMarginTop(5)

	titleLabel, _ := gtk.LabelNew("")
	titleLabel.SetMarkup("<b>Commands</b>")
	titleLabel.SetXAlign(0.0)
	titleLabel.SetHExpand(true)
	titleBox.PackStart(titleLabel, true, true, 0)

	// Close button
	closeButton, _ := gtk.ButtonNew()
	closeButton.SetLabel("×")
	closeButton.SetSizeRequest(24, 24)
	closeButton.SetTooltipText("Close Command Add-On")
	closeButton.Connect("clicked", func() {
		toggleCommandAddOn()
	})
	titleBox.PackEnd(closeButton, false, false, 0)

	addon.container.PackStart(titleBox, false, false, 0)

	// Scrolled window for the entire content area (vertical only)
	mainScroll, _ := gtk.ScrolledWindowNew(nil, nil)
	mainScroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC) // Only vertical scrolling
	mainScroll.SetVExpand(true)
	mainScroll.SetHExpand(true)

	// Inner content box (goes inside scroll)
	innerBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)
	innerBox.SetMarginStart(5)
	innerBox.SetMarginEnd(5)
	innerBox.SetMarginTop(5)
	innerBox.SetMarginBottom(5)

	// Search section
	searchBox := addon.createSearchSection()
	innerBox.PackStart(searchBox, false, false, 0)

	// Command list
	listScroll := addon.createCommandList()
	innerBox.PackStart(listScroll, false, false, 0) // Fixed height from SetSizeRequest

	// Details section
	detailsFrame := addon.createDetailsSection()
	innerBox.PackStart(detailsFrame, true, true, 0) // Expands to fill remaining space

	// Action buttons
	buttonBox := addon.createActionButtons()
	innerBox.PackStart(buttonBox, false, false, 5)

	mainScroll.Add(innerBox)
	addon.container.PackStart(mainScroll, true, true, 0)

	return addon
}

// createSearchSection creates the search and filter UI
func (addon *CommandAddOn) createSearchSection() *gtk.Box {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)

	// Module filter with Refresh button
	moduleBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	moduleLabel, _ := gtk.LabelNew("Modules:")
	moduleLabel.SetXAlign(0.0)
	moduleBox.PackStart(moduleLabel, false, false, 0)

	addon.moduleCombo, _ = gtk.ComboBoxTextNew()
	addon.moduleCombo.AppendText("All")
	addon.moduleCombo.SetActive(0)
	addon.moduleCombo.SetHExpand(true)
	addon.moduleCombo.Connect("changed", func() {
		addon.updateCommandList()
	})
	moduleBox.PackStart(addon.moduleCombo, true, true, 0)

	// Refresh button - use short label
	refreshButton, _ := gtk.ButtonNewWithLabel("↻")
	refreshButton.SetTooltipText("Reload commands from PowerShell")
	refreshButton.Connect("clicked", func() {
		addon.refreshCommands()
	})
	moduleBox.PackEnd(refreshButton, false, false, 0)

	box.PackStart(moduleBox, false, false, 0)

	// Name search
	nameBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	nameLabel, _ := gtk.LabelNew("Name:")
	nameLabel.SetXAlign(0.0)
	nameBox.PackStart(nameLabel, false, false, 0)

	addon.nameEntry, _ = gtk.SearchEntryNew()
	addon.nameEntry.SetPlaceholderText("Search...")
	addon.nameEntry.SetHExpand(true)
	addon.nameEntry.Connect("search-changed", func() {
		addon.updateCommandList()
	})
	nameBox.PackStart(addon.nameEntry, true, true, 0)

	box.PackStart(nameBox, false, false, 0)

	return box
}

// createCommandList creates the command list view
func (addon *CommandAddOn) createCommandList() *gtk.ScrolledWindow {
	scroll, _ := gtk.ScrolledWindowNew(nil, nil)
	scroll.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	scroll.SetSizeRequest(-1, 120) // Minimum height for command list
	scroll.SetVExpand(false)       // Don't expand - use fixed size

	// Create list store: Name, Module
	addon.listStore, _ = gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING)

	// Create tree view
	addon.commandList, _ = gtk.TreeViewNew()
	addon.commandList.SetModel(addon.listStore)
	addon.commandList.SetHeadersVisible(true)

	// Name column - ellipsize when narrow
	nameRenderer, _ := gtk.CellRendererTextNew()
	nameRenderer.Set("ellipsize", 3) // PANGO_ELLIPSIZE_END
	nameColumn, _ := gtk.TreeViewColumnNewWithAttribute("Command", nameRenderer, "text", 0)
	nameColumn.SetResizable(true)
	nameColumn.SetExpand(true) // Take available space
	nameColumn.SetMinWidth(80)
	nameColumn.SetSortColumnID(0)
	addon.commandList.AppendColumn(nameColumn)

	// Module column - ellipsize when narrow
	moduleRenderer, _ := gtk.CellRendererTextNew()
	moduleRenderer.Set("ellipsize", 3) // PANGO_ELLIPSIZE_END
	moduleColumn, _ := gtk.TreeViewColumnNewWithAttribute("Module", moduleRenderer, "text", 1)
	moduleColumn.SetResizable(true)
	moduleColumn.SetExpand(true) // Take available space
	moduleColumn.SetMinWidth(60)
	moduleColumn.SetSortColumnID(1)
	addon.commandList.AppendColumn(moduleColumn)

	// Selection handler
	addon.selection, _ = addon.commandList.GetSelection()
	addon.selection.Connect("changed", func() {
		addon.onCommandSelected()
	})

	scroll.Add(addon.commandList)
	return scroll
}

// createDetailsSection creates the command details display
func (addon *CommandAddOn) createDetailsSection() *gtk.Frame {
	frame, _ := gtk.FrameNew("Details")

	addon.detailsBox, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)
	addon.detailsBox.SetMarginStart(10)
	addon.detailsBox.SetMarginEnd(10)
	addon.detailsBox.SetMarginTop(10)
	addon.detailsBox.SetMarginBottom(10)

	// Command name
	addon.nameLabel, _ = gtk.LabelNew("")
	addon.nameLabel.SetXAlign(0.0)
	addon.nameLabel.SetLineWrap(true)
	addon.nameLabel.SetLineWrapMode(2)   // PANGO_WRAP_WORD_CHAR
	addon.nameLabel.SetWidthChars(10)    // Minimum width before wrapping
	addon.nameLabel.SetMaxWidthChars(40) // Maximum width
	addon.nameLabel.SetHExpand(true)
	addon.nameLabel.SetSizeRequest(50, -1) // Minimum width
	addon.detailsBox.PackStart(addon.nameLabel, false, false, 0)

	// Module
	addon.moduleLabel, _ = gtk.LabelNew("")
	addon.moduleLabel.SetXAlign(0.0)
	addon.moduleLabel.SetLineWrap(true)
	addon.moduleLabel.SetLineWrapMode(2) // PANGO_WRAP_WORD_CHAR
	addon.moduleLabel.SetWidthChars(10)
	addon.moduleLabel.SetMaxWidthChars(40)
	addon.moduleLabel.SetHExpand(true)
	addon.moduleLabel.SetSizeRequest(50, -1)
	addon.detailsBox.PackStart(addon.moduleLabel, false, false, 0)

	// Synopsis (hidden when command selected, replaced by parameters)
	addon.synopsisLabel, _ = gtk.LabelNew("")
	addon.synopsisLabel.SetXAlign(0.0)
	addon.synopsisLabel.SetLineWrap(true)
	addon.synopsisLabel.SetLineWrapMode(2) // PANGO_WRAP_WORD_CHAR
	addon.synopsisLabel.SetWidthChars(10)
	addon.synopsisLabel.SetMaxWidthChars(40)
	addon.synopsisLabel.SetHExpand(true)
	addon.synopsisLabel.SetMarginTop(5)
	addon.synopsisLabel.SetSizeRequest(50, -1)
	addon.detailsBox.PackStart(addon.synopsisLabel, false, false, 0)

	// Parameters header with help link
	paramsHeaderBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	addon.paramsHeaderLabel, _ = gtk.LabelNew("")
	addon.paramsHeaderLabel.SetMarkup("<b>Parameters:</b>")
	addon.paramsHeaderLabel.SetXAlign(0.0)
	addon.paramsHeaderLabel.SetMarginTop(10)
	addon.paramsHeaderLabel.SetLineWrap(true)
	addon.paramsHeaderLabel.SetHExpand(true)
	paramsHeaderBox.PackStart(addon.paramsHeaderLabel, true, true, 0)

	// Help link (clickable)
	helpLinkLabel, _ := gtk.LabelNew("")
	helpLinkLabel.SetMarkup("<a href='#'>?</a>")
	helpLinkLabel.SetTooltipText("Click to view full help for this command")
	helpLinkLabel.SetMarginTop(10)
	helpLinkLabel.Connect("activate-link", func(_ *gtk.Label, uri string) bool {
		addon.onShowHelp()
		return true // We handled the link
	})
	paramsHeaderBox.PackEnd(helpLinkLabel, false, false, 0)
	addon.detailsBox.PackStart(paramsHeaderBox, false, false, 0)

	// Parameter set tabs using FlowBox (allows wrapping/stacking)
	addon.paramTabBox, _ = gtk.FlowBoxNew()
	addon.paramTabBox.SetSelectionMode(gtk.SELECTION_NONE) // We handle selection manually
	addon.paramTabBox.SetHomogeneous(false)
	addon.paramTabBox.SetRowSpacing(2)
	addon.paramTabBox.SetColumnSpacing(2)
	addon.paramTabBox.SetMaxChildrenPerLine(20) // Allow many tabs per row
	addon.paramTabBox.SetMinChildrenPerLine(1)  // But allow wrapping to 1 if needed
	addon.detailsBox.PackStart(addon.paramTabBox, false, false, 5)

	// Parameter content stack (shows one parameter set at a time)
	addon.paramStack, _ = gtk.StackNew()
	addon.paramStack.SetTransitionType(gtk.STACK_TRANSITION_TYPE_NONE)
	addon.paramStack.SetVExpand(true)
	addon.paramStack.SetHExpand(true)
	addon.paramStack.SetSizeRequest(-1, 100) // Minimum height

	// Add stack directly (each grid page has its own scrolling)
	addon.detailsBox.PackStart(addon.paramStack, true, true, 0)

	// Common Parameters expander with grid
	addon.commonParamsExpander, _ = gtk.ExpanderNew("Common Parameters")
	addon.commonParamsExpander.SetExpanded(false)

	// Initialize common parameter widgets map
	addon.commonParamWidgets = make(map[string]*ParameterWidget)

	// Create grid for common parameters
	commonGrid, _ := gtk.GridNew()
	commonGrid.SetRowSpacing(5)
	commonGrid.SetColumnSpacing(10)
	commonGrid.SetMarginStart(15)
	commonGrid.SetMarginEnd(5)
	commonGrid.SetMarginTop(5)
	commonGrid.SetMarginBottom(5)

	// Common parameter names and their types
	commonParams := []struct {
		name string
		typ  string
		desc string
	}{
		{"Verbose", "SwitchParameter", "Writes detailed information about the operation"},
		{"Debug", "SwitchParameter", "Displays debugging information"},
		{"ErrorAction", "ActionPreference", "Controls how the cmdlet responds to errors"},
		{"WarningAction", "ActionPreference", "Controls how the cmdlet responds to warnings"},
		{"InformationAction", "ActionPreference", "Controls how the cmdlet responds to information"},
		{"ErrorVariable", "String", "Stores errors in the specified variable"},
		{"WarningVariable", "String", "Stores warnings in the specified variable"},
		{"InformationVariable", "String", "Stores information in the specified variable"},
		{"OutVariable", "String", "Stores output in the specified variable"},
		{"OutBuffer", "Int32", "Determines the number of objects to buffer before calling the next cmdlet"},
		{"PipelineVariable", "String", "Stores the value of the current pipeline element as a variable"},
	}

	row := 0
	for _, param := range commonParams {
		widget := &ParameterWidget{
			Name:     param.name,
			Type:     param.typ,
			IsSwitch: param.typ == "SwitchParameter",
		}

		if param.typ == "SwitchParameter" {
			// Switch parameter - checkbox
			widget.CheckButton, _ = gtk.CheckButtonNewWithLabel(param.name)
			widget.CheckButton.SetTooltipText(param.desc)
			widget.CheckButton.SetHExpand(false)
			commonGrid.Attach(widget.CheckButton, 0, row, 2, 1)
		} else {
			// Value parameter - label and entry
			label, _ := gtk.LabelNew(param.name + ":")
			label.SetXAlign(1.0)
			label.SetTooltipText(param.desc)
			label.SetHExpand(false)
			label.SetSizeRequest(120, -1)
			widget.Label = label

			widget.Entry, _ = gtk.EntryNew()
			widget.Entry.SetPlaceholderText("<" + param.typ + ">")
			widget.Entry.SetTooltipText(param.desc)
			widget.Entry.SetHExpand(false)
			widget.Entry.SetSizeRequest(150, -1)

			commonGrid.Attach(label, 0, row, 1, 1)
			commonGrid.Attach(widget.Entry, 1, row, 1, 1)
		}

		addon.commonParamWidgets[param.name] = widget
		row++
	}

	addon.commonParamsExpander.Add(commonGrid)
	addon.detailsBox.PackStart(addon.commonParamsExpander, false, false, 0)

	frame.Add(addon.detailsBox)
	return frame
}

// createActionButtons creates the action button bar
func (addon *CommandAddOn) createActionButtons() *gtk.Box {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 3)
	box.SetHAlign(gtk.ALIGN_END)
	box.SetHomogeneous(false)

	// Run button
	addon.runButton, _ = gtk.ButtonNewWithLabel("Run")
	addon.runButton.SetSensitive(false)
	addon.runButton.Connect("clicked", func() {
		addon.onRunCommand()
	})
	box.PackStart(addon.runButton, false, false, 0)

	// Insert button
	addon.insertButton, _ = gtk.ButtonNewWithLabel("Insert")
	addon.insertButton.SetSensitive(false)
	addon.insertButton.Connect("clicked", func() {
		addon.onInsertCommand()
	})
	box.PackStart(addon.insertButton, false, false, 0)

	// Copy button
	addon.copyButton, _ = gtk.ButtonNewWithLabel("Copy")
	addon.copyButton.SetSensitive(false)
	addon.copyButton.Connect("clicked", func() {
		addon.onCopyCommand()
	})
	box.PackStart(addon.copyButton, false, false, 0)

	return box
}

// loadModules populates the module dropdown
func (addon *CommandAddOn) loadModules() {
	if addon.database == nil {
		return
	}

	modules := addon.database.GetModules()
	sort.Strings(modules)

	// Clear existing (except "All")
	addon.moduleCombo.RemoveAll()
	addon.moduleCombo.AppendText("All")

	for _, module := range modules {
		addon.moduleCombo.AppendText(module)
	}

	addon.moduleCombo.SetActive(0)
}

// refreshCommands reloads commands from PowerShell
func (addon *CommandAddOn) refreshCommands() {
	if addon.database == nil {
		return
	}

	log.Println("Refreshing commands from PowerShell...")

	go func() {
		err := addon.database.LoadFromPowerShell()
		if err != nil {
			log.Printf("Error refreshing commands: %v", err)
		}

		// Update UI on main thread
		glib.IdleAdd(func() {
			addon.loadModules()
			addon.updateCommandList()
			log.Println("Commands refreshed")
		})
	}()
}

// updateCommandList refreshes the command list based on search criteria
func (addon *CommandAddOn) updateCommandList() {
	if addon.database == nil {
		return
	}

	query, _ := addon.nameEntry.GetText()
	module := addon.moduleCombo.GetActiveText()

	addon.listStore.Clear()

	commands := addon.database.Search(query, module)

	// Sort by name
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})

	for _, cmd := range commands {
		iter := addon.listStore.Append()
		addon.listStore.SetValue(iter, 0, cmd.Name)
		addon.listStore.SetValue(iter, 1, cmd.Module)
	}
}

// onCommandSelected handles command selection in the list
func (addon *CommandAddOn) onCommandSelected() {
	model, iter, ok := addon.selection.GetSelected()
	if !ok {
		addon.clearDetails()
		return
	}

	// Get command name
	value, _ := model.(*gtk.TreeModel).GetValue(iter, 0)
	cmdName, _ := value.GetString()

	// First show basic info from cache
	addon.selectedCmd = addon.database.GetCommand(cmdName)
	if addon.selectedCmd == nil {
		addon.clearDetails()
		return
	}

	// Update basic details immediately
	addon.nameLabel.SetMarkup(fmt.Sprintf("<b>Name: %s</b>", addon.selectedCmd.Name))
	addon.moduleLabel.SetText("Module: " + addon.selectedCmd.Module)
	addon.synopsisLabel.SetText("Loading details...")

	// Enable buttons
	addon.runButton.SetSensitive(true)
	addon.insertButton.SetSensitive(true)
	addon.copyButton.SetSensitive(true)

	// Fetch detailed help in background
	go func() {
		help, err := addon.database.GetCommandHelp(cmdName)
		if err != nil {
			log.Printf("Error getting help for %s: %v", cmdName, err)
			glib.IdleAdd(func() {
				addon.synopsisLabel.SetText("No synopsis available")
				addon.clearParameterUI()
			})
			return
		}

		// Update UI on main thread
		glib.IdleAdd(func() {
			addon.selectedCmd = help
			addon.updateDetails()
		})
	}()
}

// updateDetails updates the details section with selected command info
func (addon *CommandAddOn) updateDetails() {
	if addon.selectedCmd == nil {
		addon.clearDetails()
		return
	}

	// Update labels
	addon.nameLabel.SetMarkup(fmt.Sprintf("<b>Name: %s</b>", addon.selectedCmd.Name))

	// Format module like Windows ISE: "Module: ModuleName (Imported)"
	moduleText := "Module: " + addon.selectedCmd.Module
	if addon.selectedCmd.Module != "" {
		moduleText += " (Imported)"
	}
	addon.moduleLabel.SetText(moduleText)

	// Show synopsis
	synopsis := addon.selectedCmd.Synopsis
	if synopsis == "" {
		synopsis = "No synopsis available"
	}
	addon.synopsisLabel.SetText(synopsis)

	// Update parameters header to show command name
	addon.paramsHeaderLabel.SetMarkup(fmt.Sprintf("<b>Parameters for \"%s\":</b>", addon.selectedCmd.Name))

	// Enable buttons
	addon.runButton.SetSensitive(true)
	addon.insertButton.SetSensitive(true)
	addon.copyButton.SetSensitive(true)

	// Show parameters
	addon.updateParameterUI()
}

// clearDetails clears the details section
func (addon *CommandAddOn) clearDetails() {
	addon.selectedCmd = nil
	addon.nameLabel.SetText("")
	addon.moduleLabel.SetText("")
	addon.synopsisLabel.SetText("")
	addon.paramsHeaderLabel.SetMarkup("<b>Parameters:</b>")

	addon.runButton.SetSensitive(false)
	addon.insertButton.SetSensitive(false)
	addon.copyButton.SetSensitive(false)

	addon.clearParameterUI()
}

// updateParameterUI updates the parameter input area with wrapping tabs for parameter sets
func (addon *CommandAddOn) updateParameterUI() {
	addon.clearParameterUI()

	if addon.selectedCmd == nil {
		return
	}

	addon.paramWidgets = make(map[string]*ParameterWidget)
	addon.paramSetButtons = make([]*gtk.ToggleButton, 0)
	addon.currentParamSet = 0

	defaultTabIndex := 0

	// If there are parameter sets, create tabs for each
	if len(addon.selectedCmd.Syntax) > 0 {
		for i, paramSet := range addon.selectedCmd.Syntax {
			addon.createParameterSetTab(paramSet, i)
			if paramSet.IsDefault {
				defaultTabIndex = i
			}
		}
	} else if len(addon.selectedCmd.Parameters) > 0 {
		// No parameter sets defined, show all parameters in one tab
		allParams := ParameterSet{
			Name:       addon.selectedCmd.Name,
			Parameters: []string{},
			IsDefault:  true,
		}
		for _, p := range addon.selectedCmd.Parameters {
			allParams.Parameters = append(allParams.Parameters, p.Name)
		}
		addon.createParameterSetTab(allParams, 0)
	}

	addon.paramTabBox.ShowAll()
	addon.paramStack.ShowAll()

	// Select the default tab
	if len(addon.paramSetButtons) > 0 && defaultTabIndex < len(addon.paramSetButtons) {
		addon.selectParameterSet(defaultTabIndex)
	}
}

// clearParameterUI removes all tabs and widgets from the parameter area
func (addon *CommandAddOn) clearParameterUI() {
	// Remove all children from FlowBox
	addon.paramTabBox.GetChildren().Foreach(func(item interface{}) {
		if widget, ok := item.(*gtk.Widget); ok {
			addon.paramTabBox.Remove(widget)
		}
	})

	// Remove all children from Stack
	addon.paramStack.GetChildren().Foreach(func(item interface{}) {
		if widget, ok := item.(*gtk.Widget); ok {
			addon.paramStack.Remove(widget)
		}
	})

	addon.paramWidgets = nil
	addon.paramSetButtons = nil
	addon.currentParamSet = 0
}

// createParameterSetTab creates a tab button and content page for a parameter set
func (addon *CommandAddOn) createParameterSetTab(paramSet ParameterSet, index int) {
	// Create tab button for FlowBox
	tabButton, _ := gtk.ToggleButtonNew()
	tabName := paramSet.Name
	if len(tabName) > 25 {
		tabName = tabName[:22] + "..."
	}
	tabButton.SetLabel(tabName)
	tabButton.SetTooltipText(paramSet.Name)
	tabButton.SetCanFocus(false)

	// Style the button to look like a tab
	buttonIdx := index // Capture for closure
	tabButton.Connect("toggled", func(btn *gtk.ToggleButton) {
		if btn.GetActive() {
			addon.selectParameterSet(buttonIdx)
		} else {
			// Prevent deselection by clicking the same button
			if addon.currentParamSet == buttonIdx {
				btn.SetActive(true)
			}
		}
	})

	addon.paramSetButtons = append(addon.paramSetButtons, tabButton)
	addon.paramTabBox.Add(tabButton)

	// Create content page for Stack
	pageName := fmt.Sprintf("paramset-%d", index)

	// Create grid for parameters
	grid, _ := gtk.GridNew()
	grid.SetRowSpacing(5)
	grid.SetColumnSpacing(10)
	grid.SetMarginStart(5)
	grid.SetMarginEnd(5)
	grid.SetMarginTop(5)
	grid.SetMarginBottom(5)

	row := 0

	// Create widgets for each parameter in this set
	for _, paramName := range paramSet.Parameters {
		// Find the parameter details
		var param *Parameter
		for i := range addon.selectedCmd.Parameters {
			if addon.selectedCmd.Parameters[i].Name == paramName {
				param = &addon.selectedCmd.Parameters[i]
				break
			}
		}

		if param == nil {
			continue
		}

		// Skip common parameters unless checkbox is checked
		if isCommonParameter(param.Name) {
			continue
		}

		widget := addon.createParameterWidget(param)
		if widget == nil {
			continue
		}

		addon.paramWidgets[param.Name] = widget

		if param.IsSwitchParameter {
			// Switch parameters: just a checkbox spanning both columns
			widget.CheckButton.SetHExpand(false) // Don't expand
			grid.Attach(widget.CheckButton, 0, row, 2, 1)
		} else {
			// Value parameters: label and entry
			// Create label with name and required marker
			labelText := param.Name
			if param.Required {
				labelText += " *"
			}
			label, _ := gtk.LabelNew(labelText + ":")
			label.SetXAlign(1.0)
			label.SetTooltipText(param.Description)
			label.SetHExpand(false)       // Don't expand
			label.SetSizeRequest(120, -1) // Fixed width for labels
			widget.Label = label

			// Entry should not expand beyond reasonable size
			widget.Entry.SetHExpand(false)
			widget.Entry.SetSizeRequest(150, -1) // Fixed width for entries

			grid.Attach(label, 0, row, 1, 1)
			grid.Attach(widget.Entry, 1, row, 1, 1)
		}

		row++
	}

	// Wrap grid in a scrolled window with horizontal scrolling
	gridScroll, _ := gtk.ScrolledWindowNew(nil, nil)
	gridScroll.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_NEVER) // Auto horizontal, no vertical
	gridScroll.SetPropagateNaturalWidth(false)                   // Don't let grid expand beyond viewport
	gridScroll.SetPropagateNaturalHeight(false)
	gridScroll.Add(grid)

	// Add scrolled window to stack
	addon.paramStack.AddNamed(gridScroll, pageName)
}

// selectParameterSet selects a parameter set tab
func (addon *CommandAddOn) selectParameterSet(index int) {
	if index < 0 || index >= len(addon.paramSetButtons) {
		return
	}

	addon.currentParamSet = index

	// Update button states
	for i, btn := range addon.paramSetButtons {
		btn.SetActive(i == index)
	}

	// Show the corresponding stack page
	pageName := fmt.Sprintf("paramset-%d", index)
	addon.paramStack.SetVisibleChildName(pageName)
}

// createParameterWidget creates the appropriate widget for a parameter
func (addon *CommandAddOn) createParameterWidget(param *Parameter) *ParameterWidget {
	widget := &ParameterWidget{
		Name:     param.Name,
		Type:     param.Type,
		IsSwitch: param.IsSwitchParameter,
	}

	if param.IsSwitchParameter {
		// Create checkbox for switch parameters
		checkText := param.Name
		if param.Required {
			checkText += " *"
		}
		widget.CheckButton, _ = gtk.CheckButtonNewWithLabel(checkText)
		widget.CheckButton.SetTooltipText(param.Description)
	} else {
		// Create entry for value parameters
		widget.Entry, _ = gtk.EntryNew()
		widget.Entry.SetHExpand(true)

		// Set placeholder text with type hint
		placeholder := fmt.Sprintf("<%s>", param.Type)
		widget.Entry.SetPlaceholderText(placeholder)
		widget.Entry.SetTooltipText(param.Description)
	}

	return widget
}

// isCommonParameter checks if a parameter is a common PowerShell parameter
func isCommonParameter(name string) bool {
	commonParams := map[string]bool{
		"Verbose":             true,
		"Debug":               true,
		"ErrorAction":         true,
		"WarningAction":       true,
		"InformationAction":   true,
		"ProgressAction":      true,
		"ErrorVariable":       true,
		"WarningVariable":     true,
		"InformationVariable": true,
		"OutVariable":         true,
		"OutBuffer":           true,
		"PipelineVariable":    true,
		"WhatIf":              true,
		"Confirm":             true,
	}
	return commonParams[name]
}

// Action handlers

// buildCommandString builds the command string with parameters from the current tab
func (addon *CommandAddOn) buildCommandString() string {
	if addon.selectedCmd == nil {
		return ""
	}

	cmd := addon.selectedCmd.Name

	if addon.paramWidgets == nil {
		return cmd
	}

	// Get the current parameter set
	currentPage := addon.currentParamSet
	if currentPage < 0 || currentPage >= len(addon.selectedCmd.Syntax) {
		// No tabs or invalid page, just use all filled widgets
		for name, widget := range addon.paramWidgets {
			cmd += addon.formatParameter(name, widget)
		}
		return cmd
	}

	// Get parameters for the current parameter set
	currentSet := addon.selectedCmd.Syntax[currentPage]

	// Only add parameters that belong to this set
	for _, paramName := range currentSet.Parameters {
		if widget, exists := addon.paramWidgets[paramName]; exists {
			cmd += addon.formatParameter(paramName, widget)
		}
	}

	// Add common parameters if the expander is expanded
	if addon.commonParamsExpander != nil && addon.commonParamsExpander.GetExpanded() {
		for name, widget := range addon.commonParamWidgets {
			cmd += addon.formatParameter(name, widget)
		}
	}

	return cmd
}

// formatParameter formats a single parameter for the command string
func (addon *CommandAddOn) formatParameter(name string, widget *ParameterWidget) string {
	if widget.IsSwitch {
		if widget.CheckButton != nil && widget.CheckButton.GetActive() {
			return fmt.Sprintf(" -%s", name)
		}
	} else {
		if widget.Entry != nil {
			value, _ := widget.Entry.GetText()
			if value != "" {
				// Quote the value if it contains spaces
				if strings.Contains(value, " ") {
					return fmt.Sprintf(" -%s \"%s\"", name, value)
				}
				return fmt.Sprintf(" -%s %s", name, value)
			}
		}
	}
	return ""
}

func (addon *CommandAddOn) onRunCommand() {
	if addon.selectedCmd == nil {
		return
	}

	cmd := addon.buildCommandString()
	log.Printf("Running command: %s", cmd)

	// Execute in PowerShell console
	executePowerShellCommand(cmd)
}

func (addon *CommandAddOn) onInsertCommand() {
	if addon.selectedCmd == nil {
		return
	}

	cmd := addon.buildCommandString()

	// Insert at cursor in active script
	tab := getCurrentTab()
	if tab != nil {
		tab.buffer.InsertAtCursor(cmd)
	}
}

func (addon *CommandAddOn) onCopyCommand() {
	if addon.selectedCmd == nil {
		return
	}

	cmd := addon.buildCommandString()

	// Copy to clipboard
	clipboard, _ := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
	clipboard.SetText(cmd)

	log.Printf("Copied to clipboard: %s", cmd)
}

func (addon *CommandAddOn) onShowHelp() {
	if addon.selectedCmd == nil {
		return
	}

	// Show help in console
	helpCmd := fmt.Sprintf("Get-Help %s -Full", addon.selectedCmd.Name)
	executePowerShellCommand(helpCmd)
}

// executePowerShellCommand executes a command in the PowerShell console
func executePowerShellCommand(cmd string) {
	// Add command to console and execute
	if consoleTextView != nil {
		buffer, _ := consoleTextView.GetBuffer()
		endIter := buffer.GetEndIter()
		buffer.Insert(endIter, "\n"+cmd+"\n")
	}

	// Execute via console's executeCommand
	go executeCommand(cmd)
}
