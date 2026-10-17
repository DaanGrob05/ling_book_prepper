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
// empty that book is opened straight away. icon is a PNG image for the
// window and the app.
func Run(path string, icon []byte) {
	a := app.NewWithID(appID)
	a.SetIcon(fyne.NewStaticResource("Icon.png", icon))
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
	win      fyne.Window
	file     *widget.Label   // path of the open book
	info     *widget.Label   // title, author, counts
	main     *fyne.Container // the placeholder or the tabs
	tabs     *container.AppTabs
	chapters *widget.Table
	lessons  *widget.Table
	settings fyne.CanvasObject

	// settings widgets, see settings.go
	language *widget.Select
	maxWords *widget.Entry
	outDir   *widget.Entry
	write    *widget.Button

	// classification widgets, see classify.go
	bodyHere *widget.Button
	backHere *widget.Button
	detect   *widget.Button
	selected int // selected chapter row, -1 for none

	path   string       // the open book
	opts   prep.Options // options it was loaded with
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
	u.chapters = newTable(chapterColumns,
		func() int {
			if u.result == nil {
				return 0
			}
			return len(u.result.Book.Chapters)
		},
		func(row, col int) string { return chapterCell(u.result, row, col) })
	u.lessons = newTable(lessonColumns,
		func() int {
			if u.result == nil {
				return 0
			}
			return len(u.result.Lessons)
		},
		func(row, col int) string { return lessonCell(u.result, row, col) })
	u.tabs = container.NewAppTabs(
		container.NewTabItem("Chapters", container.NewBorder(u.newClassifyBar(), nil, nil, nil, u.chapters)),
		container.NewTabItem("Lessons", u.lessons),
	)

	placeholder := widget.NewLabel("Open an EPUB, or drop one on this window.")
	placeholder.Alignment = fyne.TextAlignCenter
	u.main = container.NewStack(container.NewCenter(placeholder))

	u.settings = u.newSettings()
	u.settings.Hide()

	openButton := widget.NewButtonWithIcon("Open EPUB…", theme.FolderOpenIcon(), u.chooseFile)
	top := container.NewVBox(
		container.NewBorder(nil, nil, openButton, nil, u.file),
		u.info,
	)
	win.SetContent(container.NewBorder(top, u.settings, nil, nil, u.main))
	win.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		if len(uris) > 0 {
			u.open(uris[0].Path())
		}
	})
	win.Resize(fyne.NewSize(900, 700))
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
// responsive, then shows it. It is also used to reload the open book when
// an option that affects loading changes.
func (u *ui) open(path string) {
	u.loads++
	load := u.loads
	opts := u.opts
	if path != u.path {
		// Chapter numbers of one book mean nothing for another.
		opts.BodyFrom, opts.BackFrom = 0, 0
	}
	u.file.SetText(path)
	u.info.SetText("Loading…")
	u.write.Disable()
	go func() {
		r, err := prep.Load(path, opts)
		// Widgets may only be changed on the main goroutine.
		fyne.Do(func() {
			if load != u.loads {
				return // another book was opened in the meantime
			}
			if err != nil {
				u.info.SetText("")
				if u.result != nil {
					u.file.SetText(u.path) // keep showing the previous book
					u.show()
				}
				dialog.ShowError(err, u.win)
				return
			}
			if path != u.path {
				u.outDir.SetText(prep.DirNextTo(path))
				u.tabs.SelectIndex(0)
				u.chapters.UnselectAll()
				u.chapters.ScrollToTop()
				u.lessons.ScrollToTop()
			}
			if u.opts.MaxWords != 0 && u.opts.MaxWords != r.MaxWords {
				r.SetMaxWords(u.opts.MaxWords) // changed while loading
			}
			u.opts.BodyFrom, u.opts.BackFrom = opts.BodyFrom, opts.BackFrom
			u.path, u.result = path, r
			u.show()
		})
	}()
}

// show displays the loaded book. It is called again whenever the lessons
// change.
func (u *ui) show() {
	b := u.result.Book
	u.win.SetTitle(orDash(b.Title) + " — " + appName)
	u.info.SetText(fmt.Sprintf("%s — %s · %s · %d chapters → %d lessons",
		orDash(b.Title), orDash(b.Author), b.Language.Name, len(b.Chapters), len(u.result.Lessons)))
	u.tabs.Items[1].Text = fmt.Sprintf("Lessons (%d)", len(u.result.Lessons))
	u.tabs.Refresh()
	u.chapters.Refresh()
	u.lessons.Refresh()
	u.main.Objects = []fyne.CanvasObject{u.tabs}
	u.main.Refresh()
	u.settings.Show()
	u.updateWrite()
	u.updateClassifyBar()
}

func orDash(s string) string {
	if s == "" {
		return "–"
	}
	return s
}
