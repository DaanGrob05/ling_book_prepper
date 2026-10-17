//go:build !nogui

package gui

import (
	"testing"

	"lingqprep/internal/book"
)

func TestCell(t *testing.T) {
	b := &book.Book{Chapters: []book.Chapter{
		{Title: "Copyright", Section: book.Front, Words: 52, Reason: "copyright notice"},
		{Section: book.Body, Words: 4821},
	}}
	want := [][]string{
		{"1", "front", "52", "Copyright", "copyright notice"},
		{"2", "body", "4821", "–", ""},
	}
	for row := range want {
		for col := range want[row] {
			if got := cell(b, row, col); got != want[row][col] {
				t.Errorf("cell(%d, %d) = %q, want %q", row, col, got, want[row][col])
			}
		}
	}
}
