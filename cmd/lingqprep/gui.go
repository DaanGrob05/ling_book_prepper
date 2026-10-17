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
	if fs.NArg() > 0 {
		return fmt.Errorf("gui takes no arguments, got %q", fs.Args())
	}
	gui.Run()
	return nil
}
