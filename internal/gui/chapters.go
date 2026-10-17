//go:build !nogui

package gui

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"lingqprep/internal/book"
	"lingqprep/internal/prep"
)

// columns of the chapter table, the same as the CLI's list command.
var columns = []struct {
	name  string
	width float32
}{
	{"#", 45},
	{"Section", 75},
	{"Words", 70},
	{"Title", 400},
	{"Note", 260},
}

// cell returns the text for one cell of the chapter table.
func cell(b *book.Book, row, col int) string {
	c := &b.Chapters[row]
	switch col {
	case 0:
		return strconv.Itoa(row + 1)
	case 1:
		return c.Section.String()
	case 2:
		return strconv.Itoa(c.Words)
	case 3:
		return orDash(c.Heading())
	default:
		return c.Reason
	}
}

// newChapterTable creates a table listing the chapters of the book that
// result returns. Fyne tables are lazy: they ask for the size and then
// fill only the visible cells, reusing a few label widgets while scrolling.
func newChapterTable(result func() *prep.Result) *widget.Table {
	t := widget.NewTable(
		func() (int, int) {
			if r := result(); r != nil {
				return len(r.Book.Chapters), len(columns)
			}
			return 0, len(columns)
		},
		func() fyne.CanvasObject {
			l := widget.NewLabel("")
			l.Truncation = fyne.TextTruncateEllipsis
			return l
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			l := o.(*widget.Label)
			l.Alignment = fyne.TextAlignLeading
			if id.Col == 0 || id.Col == 2 {
				l.Alignment = fyne.TextAlignTrailing
			}
			l.SetText(cell(result().Book, id.Row, id.Col))
		},
	)
	t.ShowHeaderRow = true
	t.CreateHeader = func() fyne.CanvasObject {
		return widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	}
	t.UpdateHeader = func(id widget.TableCellID, o fyne.CanvasObject) {
		o.(*widget.Label).SetText(columns[id.Col].name)
	}
	for i, c := range columns {
		t.SetColumnWidth(i, c.width)
	}
	return t
}
