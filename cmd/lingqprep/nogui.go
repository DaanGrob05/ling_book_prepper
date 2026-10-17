//go:build nogui

package main

import "errors"

const hasGUI = false

// runGUI replaces the GUI in binaries built with -tags nogui.
func runGUI(args []string) error {
	return errors.New("this lingqprep was built without the GUI (-tags nogui)")
}
