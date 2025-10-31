func updateTabLabelText(pageNum int) {
	if pageNum < 0 || pageNum >= len(openTabs) {
		return
	}
	
	tab := openTabs[pageNum]
	page, _ := notebook.GetNthPage(pageNum)
	if page == nil {
		return
	}
	
	tabLabel, _ := notebook.GetTabLabel(page)
	if tabLabel == nil {
		return
	}
	
	// Get the box and find the label
	obj := glib.Take(unsafe.Pointer(tabLabel.Native()))
	box := &gtk.Box{gtk.Container{gtk.Widget{glib.InitiallyUnowned{obj}}}}
	children := box.GetChildren()
	if children != nil {
		labelWidget := children.Data().(*gtk.Widget)
		obj2 := glib.Take(unsafe.Pointer(labelWidget.Native()))
		label := &gtk.Label{gtk.Widget{glib.InitiallyUnowned{obj2}}}
		
		title := fmt.Sprintf("Untitled%d.ps1", pageNum+1)
		if tab.filename != "" {
			title = getBaseName(tab.filename)
		}
		if tab.modified {
			title = "* " + title
		}
		label.SetText(title)
	}
}
