package prep

import (
	"errors"
	"testing"

	"lingqprep/internal/lesson"
)

func TestDefaultDir(t *testing.T) {
	tests := map[string]string{
		"book.epub":                  "book",
		"/home/me/Books/Tome 1.epub": "Tome 1",
		"relative/dir/Le.Livre.epub": "Le.Livre",
	}
	for in, want := range tests {
		if got := DefaultDir(in); got != want {
			t.Errorf("DefaultDir(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLoadRejectsBadOptions(t *testing.T) {
	if _, err := Load("book.epub", Options{MaxWords: -1}); err == nil {
		t.Error("negative MaxWords accepted")
	}
	if _, err := Load("book.epub", Options{Language: "xx"}); err == nil {
		t.Error("unknown language accepted")
	}
}

func TestOversized(t *testing.T) {
	r := &Result{MaxWords: 10, Lessons: []lesson.Lesson{
		{Number: 1, Words: 10}, {Number: 2, Words: 11}, {Number: 3, Words: 3},
	}}
	over := r.Oversized()
	if len(over) != 1 || over[0].Number != 2 {
		t.Errorf("Oversized() = %+v, want only lesson 2", over)
	}
}

func TestWriteNotEmpty(t *testing.T) {
	dir := t.TempDir()
	r := &Result{Lessons: []lesson.Lesson{{Number: 1, Title: "Un", Text: "Texte.\n"}}}
	if err := r.Write(dir, false); err != nil {
		t.Fatal(err)
	}
	if err := r.Write(dir, false); !errors.Is(err, lesson.ErrNotEmpty) {
		t.Errorf("second Write: got %v, want ErrNotEmpty", err)
	}
	if err := r.Write(dir, true); err != nil {
		t.Errorf("Write with overwrite: %v", err)
	}
}
