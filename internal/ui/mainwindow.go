package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/laurie/ps-ide-go/internal/editor"
	"github.com/laurie/ps-ide-go/internal/executor"
	"github.com/laurie/ps-ide-go/pkg/config"
)

// MainWindow represents the main application window
type MainWindow struct {
	window       fyne.Window
	editor       *editor.Editor
	executor     *executor.Executor
	config       *config.Config
	editorWidget *widget.Entry
	outputWidget *widget.Entry
}

// NewMainWindow creates and configures the main window
func NewMainWindow(app fyne.App, cfg *config.Config) *MainWindow {
	mw := &MainWindow{
		window:   app.NewWindow("PS-IDE-Go - PowerShell IDE"),
		editor:   editor.New(),
		executor: executor.New(),
		config:   cfg,
	}

	mw.setupUI()
	return mw
}

// setupUI creates and arranges UI components
func (mw *MainWindow) setupUI() {
	// Create editor widget
	mw.editorWidget = widget.NewMultiLineEntry()
	mw.editorWidget.Wrapping = fyne.TextWrapOff
	mw.editorWidget.OnChanged = func(content string) {
		mw.editor.SetContent(content)
		mw.updateTitle()
	}
	// Set placeholder after widget is created
	mw.editorWidget.PlaceHolder = "# Enter your PowerShell code here...\n# Example:\nGet-Process | Select-Object -First 5"

	// Create output widget
	mw.outputWidget = widget.NewMultiLineEntry()
	mw.outputWidget.Wrapping = fyne.TextWrapOff
	mw.outputWidget.PlaceHolder = "Output will appear here..."

	// Create menu
	mw.setupMenu()

	// Create toolbar
	toolbar := mw.createToolbar()

	// Create split layout
	split := container.NewVSplit(
		container.NewScroll(mw.editorWidget),
		container.NewScroll(mw.outputWidget),
	)
	split.SetOffset(0.6)

	// Main layout
	content := container.NewBorder(
		toolbar,
		nil,
		nil,
		nil,
		split,
	)

	mw.window.SetContent(content)
	mw.window.Resize(fyne.NewSize(
		float32(mw.config.WindowWidth),
		float32(mw.config.WindowHeight),
	))
}

// createToolbar creates the application toolbar
func (mw *MainWindow) createToolbar() *fyne.Container {
	runBtn := widget.NewButton("Run", mw.runScript)
	runBtn.Importance = widget.HighImportance

	newBtn := widget.NewButton("New", mw.newFile)
	openBtn := widget.NewButton("Open", mw.openFile)
	saveBtn := widget.NewButton("Save", mw.saveFile)

	return container.NewHBox(
		newBtn,
		openBtn,
		saveBtn,
		widget.NewSeparator(),
		runBtn,
	)
}

// setupMenu creates the application menu
func (mw *MainWindow) setupMenu() {
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("New", mw.newFile),
		fyne.NewMenuItem("Open", mw.openFile),
		fyne.NewMenuItem("Save", mw.saveFile),
		fyne.NewMenuItem("Save As", mw.saveFileAs),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() { mw.window.Close() }),
	)

	editMenu := fyne.NewMenu("Edit",
		fyne.NewMenuItem("Clear Output", mw.clearOutput),
	)

	runMenu := fyne.NewMenu("Run",
		fyne.NewMenuItem("Run Script", mw.runScript),
	)

	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About", mw.showAbout),
	)

	mainMenu := fyne.NewMainMenu(fileMenu, editMenu, runMenu, helpMenu)
	mw.window.SetMainMenu(mainMenu)
}

// runScript executes the PowerShell script
func (mw *MainWindow) runScript() {
	script := mw.editor.GetContent()
	if script == "" {
		mw.outputWidget.SetText("No script to execute")
		return
	}

	mw.outputWidget.SetText("Executing...\n")

	stdout, stderr, err := mw.executor.Execute(script)

	output := ""
	if stdout != "" {
		output += stdout
	}
	if stderr != "" {
		output += "\n--- Errors ---\n" + stderr
	}
	if err != nil {
		output += fmt.Sprintf("\n--- Execution Error ---\n%v", err)
	}

	if output == "" {
		output = "Script executed successfully with no output"
	}

	mw.outputWidget.SetText(output)
}

// newFile creates a new file
func (mw *MainWindow) newFile() {
	if mw.editor.IsModified() {
		dialog.ShowConfirm("Unsaved Changes",
			"Current file has unsaved changes. Discard changes?",
			func(ok bool) {
				if ok {
					mw.editor.Clear()
					mw.editorWidget.SetText("")
					mw.updateTitle()
				}
			}, mw.window)
	} else {
		mw.editor.Clear()
		mw.editorWidget.SetText("")
		mw.updateTitle()
	}
}

// openFile opens a file dialog to load a script
func (mw *MainWindow) openFile() {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()

		path := reader.URI().Path()
		if err := mw.editor.LoadFile(path); err != nil {
			dialog.ShowError(err, mw.window)
			return
		}

		mw.editorWidget.SetText(mw.editor.GetContent())
		mw.config.AddRecentFile(path)
		mw.updateTitle()
	}, mw.window)

	fd.SetFilter(storage.NewExtensionFileFilter([]string{".ps1", ".psm1", ".psd1"}))
	fd.Show()
}

// saveFile saves the current file
func (mw *MainWindow) saveFile() {
	if mw.editor.GetFilePath() == "" {
		mw.saveFileAs()
		return
	}

	if err := mw.editor.SaveFile(); err != nil {
		dialog.ShowError(err, mw.window)
		return
	}

	mw.updateTitle()
	dialog.ShowInformation("Saved", "File saved successfully", mw.window)
}

// saveFileAs shows save dialog
func (mw *MainWindow) saveFileAs() {
	fd := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil || writer == nil {
			return
		}
		defer writer.Close()

		path := writer.URI().Path()
		if err := mw.editor.SaveFileAs(path); err != nil {
			dialog.ShowError(err, mw.window)
			return
		}

		mw.config.AddRecentFile(path)
		mw.updateTitle()
		dialog.ShowInformation("Saved", "File saved successfully", mw.window)
	}, mw.window)

	fd.SetFilter(storage.NewExtensionFileFilter([]string{".ps1", ".psm1", ".psd1"}))
	fd.Show()
}

// clearOutput clears the output pane
func (mw *MainWindow) clearOutput() {
	mw.outputWidget.SetText("")
}

// showAbout displays the about dialog
func (mw *MainWindow) showAbout() {
	about := widget.NewLabel(
		"PS-IDE-Go v0.1.0\n\n" +
			"A lightweight PowerShell IDE for Linux\n" +
			"Built with Go and Fyne\n\n" +
			"© 2025 Laurie\n" +
			"MIT License",
	)
	dialog.ShowCustom("About PS-IDE-Go", "Close", about, mw.window)
}

// updateTitle updates the window title with file name and modified status
func (mw *MainWindow) updateTitle() {
	title := "PS-IDE-Go - "
	if mw.editor.GetFilePath() != "" {
		title += mw.editor.GetFileName()
	} else {
		title += "Untitled"
	}
	if mw.editor.IsModified() {
		title += " *"
	}
	mw.window.SetTitle(title)
}

// Show displays the main window
func (mw *MainWindow) Show() {
	mw.window.ShowAndRun()
}

// Window returns the underlying Fyne window
func (mw *MainWindow) Window() fyne.Window {
	return mw.window
}
