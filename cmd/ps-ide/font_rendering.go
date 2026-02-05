package main

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"
)

// Font rendering configuration for crisp, clear text
const (
	// Font family - prioritize fonts with good hinting
	DefaultFontFamily = "Consolas, 'Liberation Mono', 'DejaVu Sans Mono', 'Courier New', monospace"

	// Base font size (larger for better clarity)
	DefaultFontSize = 11.0 // Increased from 9.0

	// Line height for better readability
	DefaultLineHeight = 1.3

	// Font weight
	DefaultFontWeight = "normal"

	// Background colors
	EditorBackground  = "#FFFFFF" // Pure white like Windows ISE
	ConsoleBackground = "#012456" // Dark blue console
)

// FontRenderingCSS generates CSS for optimal text rendering
func FontRenderingCSS(fontSize float64) string {
	return fmt.Sprintf(`
		textview {
			background-color: %s;
			color: #000000;
			font-family: %s;
			font-size: %.1fpt;
			font-weight: 500;  /* Medium weight for better clarity */
			-gtk-font-feature-settings: "liga" 0;
		}
		textview text {
			background-color: %s;
			color: #000000;
		}
	`, EditorBackground, DefaultFontFamily, fontSize, EditorBackground)
}

// ConsoleFontCSS generates CSS for console - background and font only, no text color
func ConsoleFontCSS(fontSize float64) string {
	return fmt.Sprintf(`
		textview {
			background-color: %s;
			font-family: %s;
			font-size: %.1fpt;
			font-weight: normal;
			padding: 5px;
			caret-color: #FFFFFF;
		}
		textview text {
			background-color: %s;
		}
		textview:selected {
			background-color: #0066CC;
		}
	`, ConsoleBackground, DefaultFontFamily, fontSize, ConsoleBackground)
}

// ApplyEditorStyling applies enhanced font rendering to editor text view
func ApplyEditorStyling(textView *gtk.TextView, fontSize float64) error {
	provider, err := gtk.CssProviderNew()
	if err != nil {
		return err
	}

	css := FontRenderingCSS(fontSize)
	if err := provider.LoadFromData(css); err != nil {
		return err
	}

	// Apply to this specific textView with high priority
	styleContext, err := textView.GetStyleContext()
	if err != nil {
		return err
	}
	styleContext.AddProvider(provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	// Enable better font rendering hints
	textView.SetPixelsAboveLines(2) // Add spacing above lines
	textView.SetPixelsBelowLines(2) // Add spacing below lines

	return nil
}

// ApplyConsoleStyling applies enhanced font rendering to console text view
func ApplyConsoleStyling(textView *gtk.TextView, fontSize float64) error {
	provider, err := gtk.CssProviderNew()
	if err != nil {
		return err
	}

	css := ConsoleFontCSS(fontSize)
	if err := provider.LoadFromData(css); err != nil {
		return err
	}

	// Apply to this specific textView with USER priority (higher than APPLICATION)
	styleContext, err := textView.GetStyleContext()
	if err != nil {
		return err
	}
	styleContext.AddProvider(provider, gtk.STYLE_PROVIDER_PRIORITY_USER)

	return nil
}

// SetupFontRendering configures GTK for optimal font rendering
func SetupFontRendering() {
	// These settings can be set via GTK settings
	// For best results, also configure at system level:
	// - Enable font hinting: slight or full
	// - Enable antialiasing: rgba (for LCD screens)
	// - Set subpixel order: rgb (for most monitors)

	settings, err := gtk.SettingsGetDefault()
	if err != nil {
		return
	}

	// Enable font hinting for clearer text
	settings.SetProperty("gtk-xft-hintstyle", "hintslight")

	// Enable antialiasing
	settings.SetProperty("gtk-xft-antialias", 1)

	// Enable RGBA subpixel rendering for LCD screens
	settings.SetProperty("gtk-xft-rgba", "rgb")

	// DPI setting (96 is standard for most displays)
	settings.SetProperty("gtk-xft-dpi", 98304) // 96 * 1024
}
