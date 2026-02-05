package main

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

func createToolbar() *gtk.Toolbar {
	toolbar, _ := gtk.ToolbarNew()
	toolbar.SetStyle(gtk.TOOLBAR_ICONS)
	toolbar.SetIconSize(gtk.ICON_SIZE_SMALL_TOOLBAR)

	// New Script
	newBtn, _ := gtk.ToolButtonNew(nil, "")
	newBtn.SetIconName("document-new")
	newBtn.SetTooltipText("New Script (Ctrl+N)")
	newBtn.Connect("clicked", func() { newScript() })
	toolbar.Insert(newBtn, -1)

	// Open Script
	openBtn, _ := gtk.ToolButtonNew(nil, "")
	openBtn.SetIconName("document-open")
	openBtn.SetTooltipText("Open Script (Ctrl+O)")
	openBtn.Connect("clicked", func() { openScript(nil) })
	toolbar.Insert(openBtn, -1)

	// Save Script
	saveBtn, _ := gtk.ToolButtonNew(nil, "")
	saveBtn.SetIconName("document-save")
	saveBtn.SetTooltipText("Save Script (Ctrl+S)")
	saveBtn.Connect("clicked", func() { saveScript(nil) })
	toolbar.Insert(saveBtn, -1)

	sep1, _ := gtk.SeparatorToolItemNew()
	toolbar.Insert(sep1, -1)

	// Cut
	cutBtn, _ := gtk.ToolButtonNew(nil, "")
	cutBtn.SetIconName("edit-cut")
	cutBtn.SetTooltipText("Cut (Ctrl+X)")
	cutBtn.Connect("clicked", func() { cutText() })
	toolbar.Insert(cutBtn, -1)
	cutButton = cutBtn

	// Copy
	copyBtn, _ := gtk.ToolButtonNew(nil, "")
	copyBtn.SetIconName("edit-copy")
	copyBtn.SetTooltipText("Copy (Ctrl+C)")
	copyBtn.Connect("clicked", func() { copyText() })
	toolbar.Insert(copyBtn, -1)
	copyButton = copyBtn

	// Paste
	pasteBtn, _ := gtk.ToolButtonNew(nil, "")
	pasteBtn.SetIconName("edit-paste")
	pasteBtn.SetTooltipText("Paste (Ctrl+V)")
	pasteBtn.Connect("clicked", func() { pasteText() })
	toolbar.Insert(pasteBtn, -1)
	pasteButton = pasteBtn

	sep2, _ := gtk.SeparatorToolItemNew()
	toolbar.Insert(sep2, -1)

	// Undo
	undoBtn, _ := gtk.ToolButtonNew(nil, "")
	undoBtn.SetIconName("edit-undo")
	undoBtn.SetTooltipText("Undo (Ctrl+Z)")
	undoBtn.Connect("clicked", func() { undoText() })
	toolbar.Insert(undoBtn, -1)
	undoButton = undoBtn

	// Redo
	redoBtn, _ := gtk.ToolButtonNew(nil, "")
	redoBtn.SetIconName("edit-redo")
	redoBtn.SetTooltipText("Redo (Ctrl+Y)")
	redoBtn.Connect("clicked", func() { redoText() })
	toolbar.Insert(redoBtn, -1)
	redoButton = redoBtn

	sep3, _ := gtk.SeparatorToolItemNew()
	toolbar.Insert(sep3, -1)

	// Run Script (F5) - Green play arrow
	runBtn, _ := gtk.ToolButtonNew(nil, "")

	greenPlayIcon := createGreenPlayIcon()
	if greenPlayIcon != nil {
		runBtn.SetIconWidget(greenPlayIcon)
	} else {
		runBtn.SetIconName("media-playback-start")
	}

	runBtn.SetTooltipText("Run Script (F5)")
	runBtn.Connect("clicked", func() { runScript() })
	toolbar.Insert(runBtn, -1)
	runButton = runBtn

	// Run Selection (F8)
	runSelBtn, _ := gtk.ToolButtonNew(nil, "")
	runSelBtn.SetIconName("media-skip-forward")
	runSelBtn.SetTooltipText("Run Selection (F8)")
	runSelBtn.Connect("clicked", func() { runSelection() })
	toolbar.Insert(runSelBtn, -1)

	// Stop Operation (Ctrl+Break) - Red square
	stopBtn, _ := gtk.ToolButtonNew(nil, "")
	stopBtn.SetIconName("process-stop")
	stopBtn.SetTooltipText("Stop Operation (Ctrl+Break)")
	stopBtn.SetSensitive(false)
	stopBtn.Connect("clicked", func() { stopExecution() })
	toolbar.Insert(stopBtn, -1)
	stopButton = stopBtn

	return toolbar
}

func createGreenPlayIcon() *gtk.Image {
	pixbuf, err := gdk.PixbufNew(gdk.COLORSPACE_RGB, true, 8, 24, 24)
	if err != nil {
		return nil
	}

	pixbuf.Fill(0x00000000)

	pixels := pixbuf.GetPixels()
	rowstride := pixbuf.GetRowstride()
	nChannels := pixbuf.GetNChannels()

	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			if isInsideTriangle(x, y, 6, 4, 6, 20, 20, 12) {
				offset := y*rowstride + x*nChannels
				pixels[offset] = 0x00
				pixels[offset+1] = 0xAA
				pixels[offset+2] = 0x00
				pixels[offset+3] = 0xFF
			}
		}
	}

	image, err := gtk.ImageNewFromPixbuf(pixbuf)
	if err != nil {
		return nil
	}

	return image
}

func isInsideTriangle(px, py, x1, y1, x2, y2, x3, y3 int) bool {
	denominator := float64((y2-y3)*(x1-x3) + (x3-x2)*(y1-y3))
	if denominator == 0 {
		return false
	}

	a := float64((y2-y3)*(px-x3)+(x3-x2)*(py-y3)) / denominator
	b := float64((y3-y1)*(px-x3)+(x1-x3)*(py-y3)) / denominator
	c := 1 - a - b

	return a >= 0 && b >= 0 && c >= 0
}
