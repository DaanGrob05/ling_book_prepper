//go:build !nogui

package main

import (
	"fmt"

	"lingqprep/internal/gui"
)

func runGUI(args []string) error {
	fs := newFlagSet("gui")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 1 {
		return fmt.Errorf("expected at most one book file, got %d", fs.NArg())
	}
	gui.Run(fs.Arg(0)) // Arg returns "" when there is no file
	return nil
}
