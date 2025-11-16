package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gotk3/gotk3/gtk"
)

var lastOpenDirectory string

func newScript() {
	createNewTab()
	statusLabel.SetText("New script created")
}

func openScript(win *gtk.Window) {
	dialog, _ := gtk.FileChooserDialogNewWith2Buttons(
		"Open PowerShell Script",
		win,
		gtk.FILE_CHOOSER_ACTION_OPEN,
		"Cancel", gtk.RESPONSE_CANCEL,
		"Open", gtk.RESPONSE_ACCEPT)

	filter, _ := gtk.FileFilterNew()
	filter.SetName("PowerShell Scripts")
	filter.AddPattern("*.ps1")
	filter.AddPattern("*.psm1")
	filter.AddPattern("*.psd1")
	dialog.AddFilter(filter)

	if lastOpenDirectory != "" {
		dialog.SetCurrentFolder(lastOpenDirectory)
	}

	if dialog.Run() == gtk.RESPONSE_ACCEPT {
		filename := dialog.GetFilename()
		lastOpenDirectory = filepath.Dir(filename)
		
		content, err := os.ReadFile(filename)
		if err == nil {
			currentPageNum := notebook.GetCurrentPage()
			currentTab := getCurrentTab()
			
			shouldReplaceCurrentTab := false
			if currentTab != nil && currentTab.filename == "" && !currentTab.modified {
				start := currentTab.buffer.GetStartIter()
				end := currentTab.buffer.GetEndIter()
				currentContent, _ := currentTab.buffer.GetText(start, end, false)
				if currentContent == "" {
					shouldReplaceCurrentTab = true
				}
			}

			if shouldReplaceCurrentTab {
				currentTab.buffer.SetText(string(content))
				currentTab.filename = filename
				currentTab.modified = false
				updateTabLabelText(currentPageNum)
			} else {
				tab := createNewTab()
				tab.buffer.SetText(string(content))
				tab.filename = filename
				tab.modified = false
				pageNum := notebook.GetCurrentPage()
				updateTabLabelText(pageNum)
			}
			
			statusLabel.SetText("Opened: " + filename)
		} else {
			statusLabel.SetText("Error opening file")
		}
	}

	dialog.Destroy()
}

func saveScript(win *gtk.Window) {
	tab := getCurrentTab()
	if tab == nil {
		return
	}

	if tab.filename == "" {
		saveScriptAs(win)
		return
	}

	start, end := tab.buffer.GetBounds()
	content, _ := tab.buffer.GetText(start, end, false)
	
	err := os.WriteFile(tab.filename, []byte(content), 0644)
	if err == nil {
		tab.modified = false
		pageNum := notebook.GetCurrentPage()
		updateTabLabelText(pageNum)
		statusLabel.SetText("Saved: " + tab.filename)
	} else {
		statusLabel.SetText("Error saving file")
	}
}

func saveScriptAs(win *gtk.Window) {
	tab := getCurrentTab()
	if tab == nil {
		return
	}

	dialog, _ := gtk.FileChooserDialogNewWith2Buttons(
		"Save PowerShell Script",
		win,
		gtk.FILE_CHOOSER_ACTION_SAVE,
		"Cancel", gtk.RESPONSE_CANCEL,
		"Save", gtk.RESPONSE_ACCEPT)

	filter, _ := gtk.FileFilterNew()
	filter.SetName("PowerShell Scripts")
	filter.AddPattern("*.ps1")
	dialog.AddFilter(filter)

	if lastOpenDirectory != "" {
		dialog.SetCurrentFolder(lastOpenDirectory)
	}

	if tab.filename != "" {
		dialog.SetCurrentName(getBaseName(tab.filename))
	} else {
		pageNum := notebook.GetCurrentPage()
		dialog.SetCurrentName(fmt.Sprintf("Untitled%d.ps1", pageNum+1))
	}

	if dialog.Run() == gtk.RESPONSE_ACCEPT {
		filename := dialog.GetFilename()
		if !strings.HasSuffix(filename, ".ps1") {
			filename += ".ps1"
		}

		lastOpenDirectory = filepath.Dir(filename)

		start, end := tab.buffer.GetBounds()
		content, _ := tab.buffer.GetText(start, end, false)
		
		err := os.WriteFile(filename, []byte(content), 0644)
		if err == nil {
			tab.filename = filename
			tab.modified = false
			pageNum := notebook.GetCurrentPage()
			updateTabLabelText(pageNum)
			statusLabel.SetText("Saved: " + filename)
		} else {
			statusLabel.SetText("Error saving file")
		}
	}

	dialog.Destroy()
}

func saveCurrentFile() {
	tab := getCurrentTab()
	if tab == nil || tab.filename == "" {
		return
	}

	start, end := tab.buffer.GetBounds()
	content, _ := tab.buffer.GetText(start, end, false)
	
	err := os.WriteFile(tab.filename, []byte(content), 0644)
	if err == nil {
		tab.modified = false
		pageNum := notebook.GetCurrentPage()
		updateTabLabelText(pageNum)
		statusLabel.SetText("Saved: " + tab.filename)
	} else {
		statusLabel.SetText("Error saving file")
	}
}

func showSaveAsDialog() bool {
	tab := getCurrentTab()
	if tab == nil {
		return false
	}

	dialog, _ := gtk.FileChooserDialogNewWith2Buttons(
		"Save PowerShell Script",
		mainWindow,
		gtk.FILE_CHOOSER_ACTION_SAVE,
		"Cancel", gtk.RESPONSE_CANCEL,
		"Save", gtk.RESPONSE_ACCEPT)

	filter, _ := gtk.FileFilterNew()
	filter.SetName("PowerShell Scripts")
	filter.AddPattern("*.ps1")
	dialog.AddFilter(filter)

	if lastOpenDirectory != "" {
		dialog.SetCurrentFolder(lastOpenDirectory)
	}

	if tab.filename != "" {
		dialog.SetCurrentName(getBaseName(tab.filename))
	} else {
		pageNum := notebook.GetCurrentPage()
		dialog.SetCurrentName(fmt.Sprintf("Untitled%d.ps1", pageNum+1))
	}

	response := dialog.Run()
	saved := false

	if response == gtk.RESPONSE_ACCEPT {
		filename := dialog.GetFilename()
		if !strings.HasSuffix(filename, ".ps1") {
			filename += ".ps1"
		}

		lastOpenDirectory = filepath.Dir(filename)

		start, end := tab.buffer.GetBounds()
		content, _ := tab.buffer.GetText(start, end, false)
		
		err := os.WriteFile(filename, []byte(content), 0644)
		if err == nil {
			tab.filename = filename
			tab.modified = false
			pageNum := notebook.GetCurrentPage()
			updateTabLabelText(pageNum)
			statusLabel.SetText("Saved: " + filename)
			saved = true
		} else {
			statusLabel.SetText("Error saving file")
		}
	}

	dialog.Destroy()
	return saved
}
