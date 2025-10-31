func createVTETerminal() (*gtk.Widget, error) {
	vte := C.vte_terminal_new()
	if vte == nil {
		return nil, fmt.Errorf("failed to create VTE terminal")
	}

	vteTerm := (*C.VteTerminal)(unsafe.Pointer(vte))

	bgColor := C.GdkRGBA{red: 0.0, green: 0.0, blue: 0x66 / 255.0, alpha: 1.0}
	fgColor := C.GdkRGBA{red: 1.0, green: 1.0, blue: 1.0, alpha: 1.0}
	C.vte_terminal_set_color_background(vteTerm, &bgColor)
	C.vte_terminal_set_color_foreground(vteTerm, &fgColor)

	fontDesc := C.pango_font_description_from_string(C.CString("Monospace 10"))
	C.vte_terminal_set_font(vteTerm, fontDesc)
	C.pango_font_description_free(fontDesc)

	C.vte_terminal_set_scrollback_lines(vteTerm, 10000)
	C.vte_terminal_set_mouse_autohide(vteTerm, C.TRUE)
	C.vte_terminal_set_allow_hyperlink(vteTerm, C.TRUE)

	cwd, _ := os.Getwd()
	cwdC := C.CString(cwd)
	defer C.free(unsafe.Pointer(cwdC))

	argv := []*C.char{
		C.CString("pwsh"),
		C.CString("-NoLogo"),
		nil,
	}
	defer C.free(unsafe.Pointer(argv[0]))
	defer C.free(unsafe.Pointer(argv[1]))

	C.vte_terminal_spawn_async(
		vteTerm,
		C.G_SPAWN_DEFAULT,
		cwdC,
		&argv[0],
		nil,
		C.G_SPAWN_DEFAULT,
		nil,
		nil,
		nil,
		-1,
		nil,
		nil,
		nil,
	)

	obj := glib.Take(unsafe.Pointer(vte))
	widget := &gtk.Widget{glib.InitiallyUnowned{obj}}

	// Create context menu
	termMenu, _ = gtk.MenuNew()
	
	copyItem, _ := gtk.MenuItemNewWithLabel("Copy")
	copyItem.Connect("activate", func() {
		copyTerminalSelection()
	})
	termMenu.Append(copyItem)
	
	pasteItem, _ := gtk.MenuItemNewWithLabel("Paste")
	pasteItem.Connect("activate", func() {
		pasteToTerminal()
	})
	termMenu.Append(pasteItem)
	
	termMenu.ShowAll()

	// Add right-click handler - use interface{} to avoid type conversion issues
	widget.AddEvents(int(gdk.BUTTON_PRESS_MASK))
	widget.Connect("button-press-event", func(_ interface{}, event *gdk.Event) bool {
		eventButton := gdk.EventButtonNewFromEvent(event)
		if eventButton.Button() == 3 { // Right click
			termMenu.PopupAtPointer(event)
			return true
		}
		return false
	})

	return widget, nil
}
