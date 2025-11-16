package main

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

func cutText() {
	tab := getCurrentTab()
	if tab == nil || tab.buffer == nil {
		return
	}

	_, _, hasSelection := tab.buffer.GetSelectionBounds()
	if !hasSelection {
		return
	}

	clipboard, _ := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
	tab.buffer.CopyClipboard(clipboard)
	tab.buffer.DeleteSelection(true, true)

	statusLabel.SetText("Cut to clipboard")
	updateToolbarButtons()
}

func copyText() {
	tab := getCurrentTab()
	if tab == nil || tab.buffer == nil {
		return
	}

	_, _, hasSelection := tab.buffer.GetSelectionBounds()
	if !hasSelection {
		return
	}

	clipboard, _ := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
	tab.buffer.CopyClipboard(clipboard)

	statusLabel.SetText("Copied to clipboard")
}

func pasteText() {
	tab := getCurrentTab()
	if tab == nil || tab.buffer == nil {
		return
	}

	clipboard, _ := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
	tab.buffer.PasteClipboard(clipboard, nil, true)

	statusLabel.SetText("Pasted from clipboard")
}

func undoText() {
	tab := getCurrentTab()
	if tab == nil || tab.undoStack == nil {
		return
	}

	if tab.undoStack.CanUndo() {
		tab.undoStack.Undo()
		statusLabel.SetText("Undo")
		updateToolbarButtons()
	}
}

func redoText() {
	tab := getCurrentTab()
	if tab == nil || tab.undoStack == nil {
		return
	}

	if tab.undoStack.CanRedo() {
		tab.undoStack.Redo()
		statusLabel.SetText("Redo")
		updateToolbarButtons()
	}
}
