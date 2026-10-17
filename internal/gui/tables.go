//go:build !nogui

package gui

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"lingqprep/internal/prep"
)

// column describes one column of a table.
type column struct {
	name  string
	width float32
	right bool // right-align, for numbers
}

// chapterColumns are the same as the CLI's list command.
var chapterColumns = []column{
	{"#", 45, true},
	{"Section", 75, false},
	{"Words", 70, true},
	{"Title", 400, false},
	{"Note", 260, false},
}

// chapterCell returns the text for one cell of the chapter table.
func chapterCell(r *prep.Result, row, col int) string {
	c := &r.Book.Chapters[row]
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

var lessonColumns = []column{
	{"File", 560, false},
	{"Words", 70, true},
	{"Note", 220, false},
}

// lessonCell returns the text for one cell of the lesson table.
func lessonCell(r *prep.Result, row, col int) string {
	l := r.Lessons[row]
	switch col {
	case 0:
		return l.FileName()
	case 1:
		return strconv.Itoa(l.Words)
	default:
		if l.Words > r.MaxWords {
			return "has a sentence over the limit"
		}
		return ""
	}
}

// newTable creates a table with a header row. rows and text are called
// whenever the table is drawn, so after the data changes a Refresh is
// enough to show it.
//
// Fyne tables are lazy: they ask for the size and then fill only the
// visible cells, reusing a few label widgets while scrolling.
func newTable(cols []column, rows func() int, text func(row, col int) string) *widget.Table {
	t := widget.NewTable(
		func() (int, int) { return rows(), len(cols) },
		func() fyne.CanvasObject {
			l := widget.NewLabel("")
			l.Truncation = fyne.TextTruncateEllipsis
			return l
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			l := o.(*widget.Label)
			l.Alignment = fyne.TextAlignLeading
			if cols[id.Col].right {
				l.Alignment = fyne.TextAlignTrailing
			}
			l.SetText(text(id.Row, id.Col))
		},
	)
	t.ShowHeaderRow = true
	t.CreateHeader = func() fyne.CanvasObject {
		return widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	}
	t.UpdateHeader = func(id widget.TableCellID, o fyne.CanvasObject) {
		o.(*widget.Label).SetText(cols[id.Col].name)
	}
	for i, c := range cols {
		t.SetColumnWidth(i, c.width)
	}
	return t
}
