//go:build !nogui

package main

import (
	_ "embed"
	"fmt"

	"lingqprep/internal/gui"
)

// icon is the app icon. "fyne package" also puts it in the menu entry; it
// is embedded here as well so plain "go build" binaries show it too.
//
//go:embed Icon.png
var icon []byte

// hasGUI tells main whether this binary includes the GUI.
const hasGUI = true

func runGUI(args []string) error {
	fs := newFlagSet("gui")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 1 {
		return fmt.Errorf("expected at most one book file, got %d", fs.NArg())
	}
	gui.Run(fs.Arg(0), icon) // Arg returns "" when there is no file
	return nil
}
