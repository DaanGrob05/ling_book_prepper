//go:build !nogui

package gui

import (
	"testing"

	"lingqprep/internal/book"
	"lingqprep/internal/lesson"
	"lingqprep/internal/prep"
)

func TestChapterCell(t *testing.T) {
	r := &prep.Result{Book: &book.Book{Chapters: []book.Chapter{
		{Title: "Copyright", Section: book.Front, Words: 52, Reason: "copyright notice"},
		{Section: book.Body, Words: 4821},
	}}}
	want := [][]string{
		{"1", "front", "52", "Copyright", "copyright notice"},
		{"2", "body", "4821", "–", ""},
	}
	for row := range want {
		for col := range want[row] {
			if got := chapterCell(r, row, col); got != want[row][col] {
				t.Errorf("chapterCell(%d, %d) = %q, want %q", row, col, got, want[row][col])
			}
		}
	}
}

func TestLessonCell(t *testing.T) {
	r := &prep.Result{MaxWords: 100, Lessons: []lesson.Lesson{
		{Number: 1, Title: "Chapitre 1", Words: 100},
		{Number: 2, Title: "Chapitre 2", Words: 120},
	}}
	want := [][]string{
		{"001 - Chapitre 1.txt", "100", ""},
		{"002 - Chapitre 2.txt", "120", "has a sentence over the limit"},
	}
	for row := range want {
		for col := range want[row] {
			if got := lessonCell(r, row, col); got != want[row][col] {
				t.Errorf("lessonCell(%d, %d) = %q, want %q", row, col, got, want[row][col])
			}
		}
	}
}
