func runSelection() {
	if editorBuf == nil || termWidget == nil {
		return
	}

	// Get selected text
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
