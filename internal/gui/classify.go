//go:build !nogui

package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// newClassifyBar builds the row above the chapter table for correcting
// where the story and the back matter start: the GUI's -body-from and
// -back-from.
func (u *ui) newClassifyBar() fyne.CanvasObject {
	u.bodyHere = widget.NewButton("Story starts here", func() { u.setBodyFrom(u.selected + 1) })
	u.backHere = widget.NewButton("Back matter starts here", func() { u.setBackFrom(u.selected + 1) })
	u.detect = widget.NewButton("Detect automatically", func() { u.reclassify(0, 0) })

	u.chapters.OnSelected = func(id widget.TableCellID) {
		u.selected = id.Row
		u.updateClassifyBar()
	}
	u.chapters.OnUnselected = func(widget.TableCellID) {
		u.selected = -1
		u.updateClassifyBar()
	}
	u.selected = -1
	u.updateClassifyBar()

	hint := widget.NewLabel("Select a chapter, then:")
	hint.Truncation = fyne.TextTruncateEllipsis
	return container.NewBorder(nil, nil, nil, container.NewHBox(u.bodyHere, u.backHere, u.detect), hint)
}

// updateClassifyBar enables the buttons that make sense right now.
func (u *ui) updateClassifyBar() {
	enable := func(b *widget.Button, on bool) {
		if on {
			b.Enable()
		} else {
			b.Disable()
		}
	}
	enable(u.bodyHere, u.selected >= 0)
	enable(u.backHere, u.selected >= 0)
	enable(u.detect, u.opts.BodyFrom != 0 || u.opts.BackFrom != 0)
}

// setBodyFrom makes chapter n the first story chapter. A back matter start
// chosen earlier that would now fall before the story is dropped.
func (u *ui) setBodyFrom(n int) {
	back := u.opts.BackFrom
	if back != 0 && back <= n {
		back = 0
	}
	u.reclassify(n, back)
}

// setBackFrom makes chapter n the first back matter chapter.
func (u *ui) setBackFrom(n int) {
	u.reclassify(u.opts.BodyFrom, n)
}

// reclassify applies new body and back matter starts (0 = detect) and
// remembers them, so they survive reloading the book with another language.
func (u *ui) reclassify(bodyFrom, backFrom int) {
	if err := u.result.Reclassify(bodyFrom, backFrom); err != nil {
		dialog.ShowError(err, u.win)
		return
	}
	u.opts.BodyFrom, u.opts.BackFrom = bodyFrom, backFrom
	u.show()
}
