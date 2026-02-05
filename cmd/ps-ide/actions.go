package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func runScript() {
	tab := getCurrentTab()
	if tab == nil || translationLayer == nil {
		return
	}

	// Check if file needs to be saved
	if tab.modified {
		dialog := gtk.MessageDialogNew(
			mainWindow,
			gtk.DIALOG_MODAL,
			gtk.MESSAGE_QUESTION,
			gtk.BUTTONS_YES_NO,
			"The file has unsaved changes. Save before running?",
		)
		response := dialog.Run()
		dialog.Destroy()

		if response == gtk.RESPONSE_YES {
			if tab.filename == "" {
				if !showSaveAsDialog() {
					statusLabel.SetText("Run cancelled - file not saved")
					return
				}
			} else {
				saveCurrentFile()
				if tab.modified {
					statusLabel.SetText("Run cancelled - file save failed")
					return
				}
			}
		} else {
			statusLabel.SetText("Run cancelled")
			return
		}
	}

	if tab.filename == "" {
		dialog := gtk.MessageDialogNew(
			mainWindow,
			gtk.DIALOG_MODAL,
			gtk.MESSAGE_INFO,
			gtk.BUTTONS_OK,
			"Please save the file before running it.",
		)
		dialog.Run()
		dialog.Destroy()
		return
	}

	setExecuting(true)
	statusLabel.SetText("Running script. Press Ctrl+Break to stop.")

	go func() {
		output, err := translationLayer.ExecuteScript(tab.filename)

		glib.IdleAdd(func() bool {
			if err != nil {
				displayOutput(fmt.Sprintf("\nError: %v\n", err))
			} else {
				// Display the script output
				displayOutput(output)
			}
			displayPrompt()
			setExecuting(false)
			statusLabel.SetText("Ready")
			return false
		})
	}()
}

func runSelection() {
	tab := getCurrentTab()
	if tab == nil || translationLayer == nil {
		return
	}

	start, end, hasSelection := tab.buffer.GetSelectionBounds()
	if !hasSelection {
		statusLabel.SetText("No selection")
		time.AfterFunc(2*time.Second, func() {
			glib.IdleAdd(func() bool {
				statusLabel.SetText("Ready")
				return false
			})
		})
		return
	}

	selection, _ := tab.buffer.GetText(start, end, false)
	if selection == "" {
		return
	}

	setExecuting(true)
	statusLabel.SetText("Running selection. Press Ctrl+Break to stop.")

	go func() {
		output, err := translationLayer.ExecuteSelection(strings.TrimSpace(selection))

		glib.IdleAdd(func() bool {
			if err != nil {
				displayOutput(fmt.Sprintf("\nError: %v\n", err))
			} else {
				// Display the selection output
				displayOutput(output)
			}
			displayPrompt()
			setExecuting(false)
			statusLabel.SetText("Ready")
			return false
		})
	}()
}

func stopExecution() {
	if translationLayer == nil {
		return
	}

	translationLayer.StopExecution()
	statusLabel.SetText("Execution stopped")
	time.AfterFunc(2*time.Second, func() {
		glib.IdleAdd(func() bool {
			statusLabel.SetText("Ready")
			setExecuting(false)
			return false
		})
	})
}
