func runSelection() {
	if editorBuf == nil || termWidget == nil {
		return
	}

	// Get selected text - GetSelectionBounds returns (hasSelection bool, start *TextIter, end *TextIter)
	hasSelection, start, end := editorBuf.GetSelectionBounds()
	if !hasSelection {
		statusLabel.SetText("No selection")
		return
	}

	selection, _ := editorBuf.GetText(start, end, false)
	if selection == "" {
		return
	}

	statusLabel.SetText("Executing selection...")

	// Send selection
	lines := strings.Split(selection, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		sendToTerminal(line + "\r")
	}
	sendToTerminal("\r")

	statusLabel.SetText("Completed")
}

func clearConsole() {
	sendToTerminal("Clear-Host\r")
	statusLabel.SetText("Console cleared")
}

func updateCursorPosition(buf *gtk.TextBuffer) {
	if cursorPosLabel == nil {
		return
	}

	mark := buf.GetInsert()
	iter := buf.GetIterAtMark(mark)
	line := iter.GetLine() + 1
	col := iter.GetLineOffset() + 1

	cursorPosLabel.SetText(fmt.Sprintf("Ln %d, Col %d", line, col))
}

func showAbout(win *gtk.Window) {
	dialog := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_OK,
		"PS-IDE-Go v0.1.0\n\nA PowerShell ISE clone for Linux\nBuilt with Go and GTK3\n\n© 2025")
	dialog.Run()
	dialog.Destroy()
}
