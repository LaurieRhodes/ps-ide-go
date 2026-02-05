package main

import (
	"log"
	"os/exec"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// updatePowerShellHelp executes Update-Help command
func updatePowerShellHelp() error {
	cmd := exec.Command("pwsh", "-NoProfile", "-Command",
		"Update-Help -Force -ErrorAction SilentlyContinue")
	return cmd.Run()
}

// showUpdateHelpDialog displays a progress dialog while updating help
func showUpdateHelpDialog(parent *gtk.Window, db *CommandDatabase) {
	dialog, _ := gtk.DialogNew()
	dialog.SetTitle("Updating PowerShell Help")
	dialog.SetTransientFor(parent)
	dialog.SetModal(true)
	dialog.SetDefaultSize(450, 150)

	contentArea, _ := dialog.GetContentArea()
	contentArea.SetMarginStart(20)
	contentArea.SetMarginEnd(20)
	contentArea.SetMarginTop(20)
	contentArea.SetMarginBottom(20)

	label, _ := gtk.LabelNew("Downloading PowerShell help files from Microsoft...")
	label.SetLineWrap(true)
	contentArea.PackStart(label, false, false, 10)

	sublabel, _ := gtk.LabelNew("This may take several minutes depending on your connection.")
	sublabel.SetLineWrap(true)
	contentArea.PackStart(sublabel, false, false, 5)

	progress, _ := gtk.ProgressBarNew()
	progress.SetShowText(false)
	contentArea.PackStart(progress, false, false, 10)

	dialog.ShowAll()

	// Pulse progress bar while working
	ticker := glib.TimeoutAdd(100, func() bool {
		progress.Pulse()
		return true
	})

	// Run Update-Help in goroutine
	go func() {
		log.Println("Starting Update-Help...")
		err := updatePowerShellHelp()

		// Update UI on main thread
		glib.IdleAdd(func() {
			glib.SourceRemove(ticker)
			dialog.Destroy()

			if err != nil {
				log.Printf("Update-Help failed: %v", err)
				showErrorDialog(parent, "Update Help Failed",
					"Failed to update PowerShell help files.\n\n"+
						"Error: "+err.Error()+"\n\n"+
						"Some modules may not have updatable help available.")
			} else {
				log.Println("Update-Help completed successfully")

				// Reload command database
				if db != nil {
					log.Println("Reloading command database...")
					go func() {
						err := db.LoadHelp()
						if err != nil {
							log.Printf("Error reloading help: %v", err)
						}

						// Also load from PowerShell to get all commands
						err = db.LoadFromPowerShell()
						if err != nil {
							log.Printf("Error loading from PowerShell: %v", err)
						}
					}()
				}

				showInfoDialog(parent, "Update Complete",
					"PowerShell help files have been updated successfully!\n\n"+
						"Help information is now available for commands in the Command Add-On.")
			}
		})
	}()
}

// showErrorDialog displays an error message
func showErrorDialog(parent *gtk.Window, title, message string) {
	dialog := gtk.MessageDialogNew(parent, gtk.DIALOG_MODAL,
		gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "%s", message)
	dialog.SetTitle(title)
	dialog.Run()
	dialog.Destroy()
}

// showInfoDialog displays an information message
func showInfoDialog(parent *gtk.Window, title, message string) {
	dialog := gtk.MessageDialogNew(parent, gtk.DIALOG_MODAL,
		gtk.MESSAGE_INFO, gtk.BUTTONS_OK, "%s", message)
	dialog.SetTitle(title)
	dialog.Run()
	dialog.Destroy()
}
