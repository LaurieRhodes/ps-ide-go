package main

import (
	"fmt"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

// setupTabClickHandlers adds middle-click and right-click handlers to tab buttons
func setupTabClickHandlers() {
	if stackSwitcher == nil {
		return
	}

	// Get all children (tab buttons) of the stack switcher
	container := stackSwitcher.GetChildren()

	// Add event handlers to each button
	container.Foreach(func(item interface{}) {
		// Children are RadioButtons
		switch v := item.(type) {
		case *gtk.RadioButton:
			// Enable ALL button events
			v.Widget.AddEvents(int(gdk.BUTTON_PRESS_MASK | gdk.BUTTON_RELEASE_MASK))

			// Connect button-press-event
			v.Connect("button-press-event", func(btn *gtk.RadioButton, event *gdk.Event) bool {
				return onTabButtonPress(btn, event)
			})

			// Also connect button-release-event for middle-click compatibility
			v.Connect("button-release-event", func(btn *gtk.RadioButton, event *gdk.Event) bool {
				eventButton := gdk.EventButtonNewFromEvent(event)
				if eventButton.Button() == 2 {
					return onTabButtonPress(btn, event)
				}
				return false
			})

		case *gtk.Widget:
			// Fallback: try to cast to RadioButton
			obj, err := v.Cast()
			if err == nil {
				if radioBtn, ok := obj.(*gtk.RadioButton); ok {
					radioBtn.Widget.AddEvents(int(gdk.BUTTON_PRESS_MASK | gdk.BUTTON_RELEASE_MASK))
					radioBtn.Connect("button-press-event", func(btn *gtk.RadioButton, event *gdk.Event) bool {
						return onTabButtonPress(btn, event)
					})
					radioBtn.Connect("button-release-event", func(btn *gtk.RadioButton, event *gdk.Event) bool {
						eventButton := gdk.EventButtonNewFromEvent(event)
						if eventButton.Button() == 2 {
							return onTabButtonPress(btn, event)
						}
						return false
					})
				}
			}
		}
	})
}

// onTabButtonPress handles middle-click and right-click on tabs
func onTabButtonPress(button *gtk.RadioButton, event *gdk.Event) bool {
	eventButton := gdk.EventButtonNewFromEvent(event)
	buttonNum := eventButton.Button()

	// Get the tab label to find which tab was clicked
	tabLabel := getTabLabelFromButton(button)
	if tabLabel == "" {
		return false
	}

	tabIndex := findTabIndexByLabel(tabLabel)
	if tabIndex == -1 {
		return false
	}

	switch buttonNum {
	case 1: // Left click - let default behavior handle it
		return false

	case 2: // Middle click - close tab
		closeTab(tabIndex)
		return true

	case 3: // Right click - show context menu
		showTabContextMenu(event, tabIndex)
		return true
	}

	return false
}

// getTabLabelFromButton extracts the tab label text from a RadioButton
func getTabLabelFromButton(radioBtn *gtk.RadioButton) string {
	// StackSwitcher RadioButtons contain a Label widget as a child
	container := &radioBtn.Container
	children := container.GetChildren()

	var foundLabel string
	children.Foreach(func(item interface{}) {
		// Try direct Label
		if label, ok := item.(*gtk.Label); ok {
			text, err := label.GetText()
			if err == nil && text != "" {
				foundLabel = text
			}
		}

		// Try Widget then cast to Label
		if widget, ok := item.(*gtk.Widget); ok {
			if obj, err := widget.Cast(); err == nil {
				if label, ok := obj.(*gtk.Label); ok {
					text, err := label.GetText()
					if err == nil && text != "" {
						foundLabel = text
					}
				}
			}
		}
	})

	return foundLabel
}

// findTabIndexByLabel finds the tab index by matching the label text
func findTabIndexByLabel(label string) int {
	for i, tab := range openTabs {
		tabTitle := fmt.Sprintf("Untitled%d.ps1", tab.tabID)
		if tab.filename != "" {
			tabTitle = getBaseName(tab.filename)
		}
		if tab.modified {
			tabTitle = "* " + tabTitle
		}

		if tabTitle == label {
			return i
		}
	}
	return -1
}

// showTabContextMenu displays a right-click context menu for tabs
func showTabContextMenu(event *gdk.Event, tabIndex int) {
	if tabIndex < 0 || tabIndex >= len(openTabs) {
		return
	}

	menu, _ := gtk.MenuNew()

	// Close Tab
	closeItem, _ := gtk.MenuItemNewWithLabel("Close Tab")
	closeItem.Connect("activate", func() {
		closeTab(tabIndex)
	})
	menu.Append(closeItem)

	// Close Other Tabs
	if len(openTabs) > 1 {
		closeOthersItem, _ := gtk.MenuItemNewWithLabel("Close Other Tabs")
		closeOthersItem.Connect("activate", func() {
			closeOtherTabs(tabIndex)
		})
		menu.Append(closeOthersItem)
	}

	// Close All Tabs
	if len(openTabs) > 1 {
		closeAllItem, _ := gtk.MenuItemNewWithLabel("Close All Tabs")
		closeAllItem.Connect("activate", func() {
			closeAllTabs()
		})
		menu.Append(closeAllItem)
	}

	// Separator
	separator, _ := gtk.SeparatorMenuItemNew()
	menu.Append(separator)

	// Save
	tab := openTabs[tabIndex]
	if tab.modified {
		saveItem, _ := gtk.MenuItemNewWithLabel("Save")
		saveItem.Connect("activate", func() {
			// Switch to this tab first
			setCurrentTab(tabIndex)
			saveScript(mainWindow)
		})
		menu.Append(saveItem)
	}

	// Save As
	saveAsItem, _ := gtk.MenuItemNewWithLabel("Save As...")
	saveAsItem.Connect("activate", func() {
		// Switch to this tab first
		setCurrentTab(tabIndex)
		if showSaveAsDialog() {
			saveCurrentFile()
		}
	})
	menu.Append(saveAsItem)

	// Separator
	separator2, _ := gtk.SeparatorMenuItemNew()
	menu.Append(separator2)

	// Copy Full Path (if file has been saved)
	if tab.filename != "" {
		copyPathItem, _ := gtk.MenuItemNewWithLabel("Copy Full Path")
		copyPathItem.Connect("activate", func() {
			clipboard, _ := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
			clipboard.SetText(tab.filename)
		})
		menu.Append(copyPathItem)
	}

	menu.ShowAll()
	menu.PopupAtPointer(event)
}

// closeOtherTabs closes all tabs except the specified one
func closeOtherTabs(keepIndex int) {
	// Close tabs in reverse order to maintain indices
	for i := len(openTabs) - 1; i >= 0; i-- {
		if i != keepIndex {
			closeTab(i)
			// Adjust keepIndex if necessary
			if i < keepIndex {
				keepIndex--
			}
		}
	}
}

// closeAllTabs closes all tabs (with save prompts)
func closeAllTabs() {
	// Close tabs in reverse order
	for i := len(openTabs) - 1; i >= 0; i-- {
		closeTab(i)
		// If user cancelled, stop closing
		if len(openTabs) > i {
			return
		}
	}
}

// setCurrentTab switches to the specified tab
func setCurrentTab(index int) {
	if index < 0 || index >= len(openTabs) {
		return
	}

	tab := openTabs[index]
	pageName := fmt.Sprintf("tab-%d", tab.tabID)
	contentStack.SetVisibleChildName(pageName)
}

// updateTabClickHandlers should be called after adding/removing tabs
func updateTabClickHandlers() {
	// Re-setup handlers for all tabs
	setupTabClickHandlers()
}
