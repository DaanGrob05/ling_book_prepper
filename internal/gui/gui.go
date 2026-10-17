//go:build !nogui

// Package gui is the graphical front end, built with Fyne. Like the command
// line it only collects options and shows results; the work is done by
// package prep.
//
// Fyne needs cgo and a C compiler, so the package is left out of builds made
// with -tags nogui, which produce a CLI-only binary.
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

// appID identifies the app to the OS and names its preferences folder.
const appID = "io.github.daangrob05.lingqprep"

// Run opens the main window and blocks until it is closed.
func Run() {
	a := app.NewWithID(appID)
	w := a.NewWindow("LingQ Prep")
	w.SetContent(widget.NewLabel("Choose an EPUB to prepare it for LingQ."))
	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}
