package main

import (
	"os"
	"path/filepath"

	"github.com/gotk3/gotk3/gtk"
	"github.com/laurie/ps-ide-go/cmd/ps-ide/translation"
)

var debugLoggingEnabled bool = false

func createMenuBar(win *gtk.Window) *gtk.MenuBar {
	menuBar, _ := gtk.MenuBarNew()

	// File Menu
	fileMenu, _ := gtk.MenuNew()
	fileMenuItem, _ := gtk.MenuItemNewWithLabel("File")
	fileMenuItem.SetSubmenu(fileMenu)

	newItem, _ := gtk.MenuItemNewWithLabel("New")
	openItem, _ := gtk.MenuItemNewWithLabel("Open...")
	saveItem, _ := gtk.MenuItemNewWithLabel("Save")
	saveAsItem, _ := gtk.MenuItemNewWithLabel("Save As...")
	fileSep1, _ := gtk.SeparatorMenuItemNew()
	runFileItem, _ := gtk.MenuItemNewWithLabel("Run")
	runFileSelectionItem, _ := gtk.MenuItemNewWithLabel("Run Selection")
	fileSep2, _ := gtk.SeparatorMenuItemNew()
	closeItem, _ := gtk.MenuItemNewWithLabel("Close")
	fileSep3, _ := gtk.SeparatorMenuItemNew()
	newPSTabItem, _ := gtk.MenuItemNewWithLabel("New PowerShell Tab")
	closePSTabItem, _ := gtk.MenuItemNewWithLabel("Close PowerShell Tab")
	fileSep4, _ := gtk.SeparatorMenuItemNew()
	newRemotePSTabItem, _ := gtk.MenuItemNewWithLabel("New Remote PowerShell Tab...")
	fileSep5, _ := gtk.SeparatorMenuItemNew()
	exitItem, _ := gtk.MenuItemNewWithLabel("Exit")

	fileMenu.Append(newItem)
	fileMenu.Append(openItem)
	fileMenu.Append(saveItem)
	fileMenu.Append(saveAsItem)
	fileMenu.Append(fileSep1)
	fileMenu.Append(runFileItem)
	fileMenu.Append(runFileSelectionItem)
	fileMenu.Append(fileSep2)
	fileMenu.Append(closeItem)
	fileMenu.Append(fileSep3)
	fileMenu.Append(newPSTabItem)
	fileMenu.Append(closePSTabItem)
	fileMenu.Append(fileSep4)
	fileMenu.Append(newRemotePSTabItem)
	fileMenu.Append(fileSep5)
	fileMenu.Append(exitItem)

	newItem.Connect("activate", func() { newScript() })
	openItem.Connect("activate", func() { openScript(win) })
	saveItem.Connect("activate", func() { saveScript(win) })
	saveAsItem.Connect("activate", func() { saveScriptAs(win) })
	runFileItem.Connect("activate", func() { runScript() })
	runFileSelectionItem.Connect("activate", func() { runSelection() })
	closeItem.Connect("activate", func() { closeCurrentTab() })
	newPSTabItem.Connect("activate", func() { newScript() })
	closePSTabItem.Connect("activate", func() { closeCurrentTab() })
	newRemotePSTabItem.SetSensitive(false) // Disabled - remote connectivity not implemented
	exitItem.Connect("activate", func() {
		saveSession()
		shutdownTranslationLayer()
		gtk.MainQuit()
	})

	// Edit Menu
	editMenu, _ := gtk.MenuNew()
	editMenuItem, _ := gtk.MenuItemNewWithLabel("Edit")
	editMenuItem.SetSubmenu(editMenu)

	undoItem, _ := gtk.MenuItemNewWithLabel("Undo")
	redoItem, _ := gtk.MenuItemNewWithLabel("Redo")
	cutItem, _ := gtk.MenuItemNewWithLabel("Cut")
	copyItem, _ := gtk.MenuItemNewWithLabel("Copy")
	pasteItem, _ := gtk.MenuItemNewWithLabel("Paste")
	findItem, _ := gtk.MenuItemNewWithLabel("Find in Script...")
	clearItem, _ := gtk.MenuItemNewWithLabel("Clear Console")

	editMenu.Append(undoItem)
	editMenu.Append(redoItem)
	sep2, _ := gtk.SeparatorMenuItemNew()
	editMenu.Append(sep2)
	editMenu.Append(cutItem)
	editMenu.Append(copyItem)
	editMenu.Append(pasteItem)
	sep3, _ := gtk.SeparatorMenuItemNew()
	editMenu.Append(sep3)
	editMenu.Append(findItem)
	sep3b, _ := gtk.SeparatorMenuItemNew()
	editMenu.Append(sep3b)
	editMenu.Append(clearItem)

	undoItem.Connect("activate", func() { undoText() })
	redoItem.Connect("activate", func() { redoText() })
	cutItem.Connect("activate", func() { cutText() })
	copyItem.Connect("activate", func() { copyText() })
	pasteItem.Connect("activate", func() { pasteText() })
	findItem.Connect("activate", func() { showFindDialog() })
	clearItem.Connect("activate", func() { clearConsole() })

	// View Menu
	viewMenu, _ := gtk.MenuNew()
	viewMenuItem, _ := gtk.MenuItemNewWithLabel("View")
	viewMenuItem.SetSubmenu(viewMenu)

	showScriptItem, _ := gtk.MenuItemNewWithLabel("Show Script Pane")
	showConsoleItem, _ := gtk.MenuItemNewWithLabel("Show Console Pane")
	showCommandAddonItem, _ := gtk.CheckMenuItemNewWithLabel("Show Command Add-On")
	showCommandAddonMenuItem = showCommandAddonItem // Store global reference
	viewMenu.Append(showScriptItem)
	viewMenu.Append(showConsoleItem)
	sep5, _ := gtk.SeparatorMenuItemNew()
	viewMenu.Append(sep5)
	viewMenu.Append(showCommandAddonItem)

	showCommandAddonItem.Connect("toggled", func() {
		toggleCommandAddOn()
	})

	// Tools Menu
	toolsMenu, _ := gtk.MenuNew()
	toolsMenuItem, _ := gtk.MenuItemNewWithLabel("Tools")
	toolsMenuItem.SetSubmenu(toolsMenu)

	// Debug logging toggle
	debugLoggingItem, _ := gtk.CheckMenuItemNewWithLabel("Enable Debug Logging")
	debugLoggingItem.SetActive(debugLoggingEnabled)
	debugLoggingItem.Connect("toggled", func() {
		toggleDebugLogging(debugLoggingItem, win)
	})
	toolsMenu.Append(debugLoggingItem)

	sep4, _ := gtk.SeparatorMenuItemNew()
	toolsMenu.Append(sep4)

	optionsItem, _ := gtk.MenuItemNewWithLabel("Options...")
	toolsMenu.Append(optionsItem)

	// Debug Menu
	debugMenu, _ := gtk.MenuNew()
	debugMenuItem, _ := gtk.MenuItemNewWithLabel("Debug")
	debugMenuItem.SetSubmenu(debugMenu)

	runItem, _ := gtk.MenuItemNewWithLabel("Run/Continue (F5)")
	runSelectionItem, _ := gtk.MenuItemNewWithLabel("Run Selection (F8)")
	stopItem, _ := gtk.MenuItemNewWithLabel("Stop Debugger")
	debugMenu.Append(runItem)
	debugMenu.Append(runSelectionItem)
	debugMenu.Append(stopItem)

	runItem.Connect("activate", func() { runScript() })
	runSelectionItem.Connect("activate", func() { runSelection() })
	stopItem.Connect("activate", func() { stopExecution() })

	// Add-ons Menu
	addonsMenu, _ := gtk.MenuNew()
	addonsMenuItem, _ := gtk.MenuItemNewWithLabel("Add-ons")
	addonsMenuItem.SetSubmenu(addonsMenu)

	// Help Menu
	helpMenu, _ := gtk.MenuNew()
	helpMenuItem, _ := gtk.MenuItemNewWithLabel("Help")
	helpMenuItem.SetSubmenu(helpMenu)

	updateHelpItem, _ := gtk.MenuItemNewWithLabel("Update Windows PowerShell Help")
	updateHelpItem.Connect("activate", func() {
		if commandDatabase != nil {
			showUpdateHelpDialog(win, commandDatabase)
		}
	})
	helpMenu.Append(updateHelpItem)

	sep6, _ := gtk.SeparatorMenuItemNew()
	helpMenu.Append(sep6)

	aboutItem, _ := gtk.MenuItemNewWithLabel("About")
	helpMenu.Append(aboutItem)
	aboutItem.Connect("activate", func() { showAbout(win) })

	// Add all to menu bar
	menuBar.Append(fileMenuItem)
	menuBar.Append(editMenuItem)
	menuBar.Append(viewMenuItem)
	menuBar.Append(toolsMenuItem)
	menuBar.Append(debugMenuItem)
	menuBar.Append(addonsMenuItem)
	menuBar.Append(helpMenuItem)

	return menuBar
}

func toggleDebugLogging(item *gtk.CheckMenuItem, win *gtk.Window) {
	active := item.GetActive()
	debugLoggingEnabled = active

	if active {
		err := translation.EnableDebugLogging()
		if err != nil {
			dialog := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK,
				"Failed to enable debug logging:\n%s", err.Error())
			dialog.Run()
			dialog.Destroy()
			item.SetActive(false)
			debugLoggingEnabled = false
		} else {
			// Show success message with log file location
			homeDir := os.Getenv("HOME")
			logDir := filepath.Join(homeDir, ".ps-ide", "logs")
			dialog := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_OK,
				"Debug logging enabled!\n\nLog files are being saved to:\n%s\n\nThis will help troubleshoot issues with the Translation Layer.", logDir)
			dialog.Run()
			dialog.Destroy()
		}
	} else {
		translation.DisableDebugLogging()
		dialog := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_OK,
			"Debug logging disabled.")
		dialog.Run()
		dialog.Destroy()
	}
}

func showAbout(win *gtk.Window) {
	dialog := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_OK,
		"PS-IDE-Go v0.2.0\n\nA PowerShell ISE clone for Linux\nBuilt with Go and GTK3\nWith Translation Layer Architecture\n\n© 2025")
	dialog.Run()
	dialog.Destroy()
}
