func createVTETerminal() (*gtk.Widget, error) {
	vte := C.vte_terminal_new()
	if vte == nil {
		return nil, fmt.Errorf("failed to create VTE terminal")
	}

	vteTerm := (*C.VteTerminal)(unsafe.Pointer(vte))

	// PowerShell ISE colors
	bgColor := C.GdkRGBA{red: 0.0, green: 0.0, blue: 0x66 / 255.0, alpha: 1.0}
	fgColor := C.GdkRGBA{red: 1.0, green: 1.0, blue: 1.0, alpha: 1.0}
	C.vte_terminal_set_color_background(vteTerm, &bgColor)
	C.vte_terminal_set_color_foreground(vteTerm, &fgColor)

	fontDesc := C.pango_font_description_from_string(C.CString("Monospace 10"))
	C.vte_terminal_set_font(vteTerm, fontDesc)
	C.pango_font_description_free(fontDesc)

	C.vte_terminal_set_scrollback_lines(vteTerm, 10000)
	C.vte_terminal_set_mouse_autohide(vteTerm, C.TRUE)

	// Enable text selection
	C.vte_terminal_set_allow_hyperlink(vteTerm, C.TRUE)

	// Spawn PowerShell
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

	// Add right-click context menu for copy
	widget.Connect("button-press-event", func(_ interface{}, event *gdk.Event) bool {
		btnEvent := gdk.EventButtonNewFromEvent(event)
		if btnEvent.Button() == gdk.BUTTON_SECONDARY { // Right click
			menu, _ := gtk.MenuNew()
			
			copyItem, _ := gtk.MenuItemNewWithLabel("Copy")
			copyItem.Connect("activate", func() {
				copyTerminalSelection()
			})
			menu.Append(copyItem)
			
			pasteItem, _ := gtk.MenuItemNewWithLabel("Paste")
			pasteItem.Connect("activate", func() {
				pasteToTerminal()
			})
			menu.Append(pasteItem)
			
			menu.ShowAll()
			menu.PopupAtPointer(event)
			return true
		}
		return false
	})

	return widget, nil
}
