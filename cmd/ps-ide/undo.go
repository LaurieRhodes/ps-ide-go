package main

import (
	"github.com/gotk3/gotk3/gtk"
)

// UndoStack manages undo/redo operations for a text buffer
type UndoStack struct {
	buffer    *gtk.TextBuffer
	undoStack []string
	redoStack []string
	maxLevels int
	isUndoing bool
	lastText  string
}

func NewUndoStack(buffer *gtk.TextBuffer, maxLevels int) *UndoStack {
	us := &UndoStack{
		buffer:    buffer,
		undoStack: make([]string, 0),
		redoStack: make([]string, 0),
		maxLevels: maxLevels,
		isUndoing: false,
		lastText:  "",
	}

	// Get initial text
	start := buffer.GetStartIter()
	end := buffer.GetEndIter()
	us.lastText, _ = buffer.GetText(start, end, false)

	// Connect to buffer changes
	buffer.Connect("changed", func() {
		if !us.isUndoing {
			us.recordChange()
		}
	})

	return us
}

func (us *UndoStack) recordChange() {
	start := us.buffer.GetStartIter()
	end := us.buffer.GetEndIter()
	currentText, _ := us.buffer.GetText(start, end, false)

	// Only record if text actually changed
	if currentText != us.lastText {
		// Add last text to undo stack
		us.undoStack = append(us.undoStack, us.lastText)

		// Limit stack size
		if len(us.undoStack) > us.maxLevels {
			us.undoStack = us.undoStack[1:]
		}

		// Clear redo stack on new change
		us.redoStack = make([]string, 0)

		us.lastText = currentText
	}
}

func (us *UndoStack) CanUndo() bool {
	return len(us.undoStack) > 0
}

func (us *UndoStack) CanRedo() bool {
	return len(us.redoStack) > 0
}

func (us *UndoStack) Undo() {
	if !us.CanUndo() {
		return
	}

	// Get current text before undo
	start := us.buffer.GetStartIter()
	end := us.buffer.GetEndIter()
	currentText, _ := us.buffer.GetText(start, end, false)

	// Pop from undo stack
	lastIndex := len(us.undoStack) - 1
	textToRestore := us.undoStack[lastIndex]
	us.undoStack = us.undoStack[:lastIndex]

	// Push current text to redo stack
	us.redoStack = append(us.redoStack, currentText)

	// Restore text
	us.isUndoing = true
	us.buffer.SetText(textToRestore)
	us.lastText = textToRestore
	us.isUndoing = false
}

func (us *UndoStack) Redo() {
	if !us.CanRedo() {
		return
	}

	// Get current text before redo
	start := us.buffer.GetStartIter()
	end := us.buffer.GetEndIter()
	currentText, _ := us.buffer.GetText(start, end, false)

	// Pop from redo stack
	lastIndex := len(us.redoStack) - 1
	textToRestore := us.redoStack[lastIndex]
	us.redoStack = us.redoStack[:lastIndex]

	// Push current text to undo stack
	us.undoStack = append(us.undoStack, currentText)

	// Restore text
	us.isUndoing = true
	us.buffer.SetText(textToRestore)
	us.lastText = textToRestore
	us.isUndoing = false
}
