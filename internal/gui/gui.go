//go:build !nogui

// Package gui is the graphical front end, built with Fyne. Like the command
// line it only collects options and shows results; the work is done by
// package prep.
//
// Fyne needs cgo and a C compiler, so the package is left out of builds made
// with -tags nogui, which produce a CLI-only binary.
package gui

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"lingqprep/internal/prep"
)

// appID identifies the app to the OS and names its preferences folder.
const appID = "io.github.daangrob05.lingqprep"

const appName = "LingQ Prep"

// Run opens the main window and blocks until it is closed. If path is not
// empty that book is opened straight away.
func Run(path string) {
	a := app.NewWithID(appID)
	u := newUI(a.NewWindow(appName))
	if path != "" {
		u.open(path)
	}
	u.win.ShowAndRun()
}

// ui holds the main window's widgets and the book shown in it. Fyne calls
// every callback on its main goroutine, so the fields need no locking as
// long as background work hands its results back with fyne.Do.
type ui struct {
	win    fyne.Window
	file   *widget.Label   // path of the open book
	info   *widget.Label   // title, author, counts
	main   *fyne.Container // the placeholder or the chapter table
	table  *widget.Table
	path   string
	result *prep.Result // nil until a book is loaded
	loads  int          // counts open calls, to drop results of stale loads
}

func newUI(win fyne.Window) *ui {
	u := &ui{
		win:  win,
		file: widget.NewLabel(""),
		info: widget.NewLabel(""),
	}
	u.file.Truncation = fyne.TextTruncateEllipsis
	u.table = newChapterTable(func() *prep.Result { return u.result })

	placeholder := widget.NewLabel("Open an EPUB, or drop one on this window.")
	placeholder.Alignment = fyne.TextAlignCenter
	u.main = container.NewStack(container.NewCenter(placeholder))

	openButton := widget.NewButtonWithIcon("Open EPUB…", theme.FolderOpenIcon(), u.chooseFile)
	top := container.NewVBox(
		container.NewBorder(nil, nil, openButton, nil, u.file),
		u.info,
	)
	win.SetContent(container.NewBorder(top, nil, nil, nil, u.main))
	win.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		if len(uris) > 0 {
			u.open(uris[0].Path())
		}
	})
	win.Resize(fyne.NewSize(900, 650))
	return u
}

// chooseFile shows a file dialog for EPUBs, starting in the folder of the
// book that is open.
func (u *ui) chooseFile() {
	d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, u.win)
			return
		}
		if r == nil {
			return // cancelled
		}
		r.Close() // we only need the path; prep reads the file itself
		u.open(r.URI().Path())
	}, u.win)
	d.SetFilter(storage.NewExtensionFileFilter([]string{".epub"}))
	if u.path != "" {
		if dir, err := storage.ListerForURI(storage.NewFileURI(filepath.Dir(u.path))); err == nil {
			d.SetLocation(dir)
		}
	}
	d.Resize(fyne.NewSize(800, 550))
	d.Show()
}

// open loads the book at path in the background so the window stays
// responsive, then shows it.
func (u *ui) open(path string) {
	u.loads++
	load := u.loads
	u.file.SetText(path)
	u.info.SetText("Loading…")
	go func() {
		r, err := prep.Load(path, prep.Options{})
		// Widgets may only be changed on the main goroutine.
		fyne.Do(func() {
			if load != u.loads {
				return // another book was opened in the meantime
			}
			if err != nil {
				u.info.SetText("")
				if u.path != "" {
					u.file.SetText(u.path) // keep showing the previous book
					u.show()
				}
				dialog.ShowError(err, u.win)
				return
			}
			u.path, u.result = path, r
			u.show()
		})
	}()
}

// show displays the loaded book.
func (u *ui) show() {
	b := u.result.Book
	u.win.SetTitle(orDash(b.Title) + " — " + appName)
	u.info.SetText(fmt.Sprintf("%s — %s · %s · %d chapters → %d lessons",
		orDash(b.Title), orDash(b.Author), b.Language.Name, len(b.Chapters), len(u.result.Lessons)))
	u.table.Refresh()
	u.table.ScrollToTop()
	u.main.Objects = []fyne.CanvasObject{u.table}
	u.main.Refresh()
}

func orDash(s string) string {
	if s == "" {
		return "–"
	}
	return s
}
