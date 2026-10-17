package lesson

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"lingqprep/internal/book"
	"lingqprep/internal/prose"
)

// DefaultMaxWords is the largest lesson LingQ accepts.
const DefaultMaxWords = 6000

// Titles for the files that collect all front and back matter.
const (
	FrontTitle = "Front matter"
	BackTitle  = "Back matter"
)

// Lesson is one output file.
type Lesson struct {
	Number int    // position in reading order; 0 is the front matter
	Title  string // chapter title
	Part   int    // 1-based part within the chapter
	Parts  int    // number of parts the chapter was split into
	Words  int
	Text   string
}

// FileName returns e.g. "001 - Chapitre 1.txt", or "002 - Chapitre 2 (1).txt"
// and "003 - Chapitre 2 (2).txt" when a chapter is split over several files.
func (l Lesson) FileName() string {
	name := fmt.Sprintf("%03d - %s", l.Number, sanitize(l.Title))
	if l.Parts > 1 {
		name += fmt.Sprintf(" (%d)", l.Part)
	}
	return name + ".txt"
}

// group is a run of blocks that becomes one or more lessons.
type group struct {
	title  string
	blocks []prose.Block
}

// Build cuts the book into lessons of at most maxWords words. Every lesson
// gets its own number so the files sort in reading order; numbering starts
// at 0 for the front matter and at 1 otherwise.
func Build(b *book.Book, maxWords int) []Lesson {
	var lessons []Lesson
	gs := groups(b)
	first := 1
	if len(gs) > 0 && gs[0].title == FrontTitle {
		first = 0
	}
	for _, g := range gs {
		parts := split(units(g.blocks, b.Language), maxWords)
		for i, p := range parts {
			l := Lesson{Number: first + len(lessons), Title: g.title, Part: i + 1, Parts: len(parts), Text: render(p)}
			for _, u := range p {
				l.Words += u.words
			}
			lessons = append(lessons, l)
		}
	}
	return lessons
}

// groups merges all front matter into one group and all back matter into
// another, and folds body chapters that contain only headings (like a
// "Première partie" divider page) into the chapter after them.
func groups(b *book.Book) []group {
	var front, back []prose.Block
	var body []group
	var pending []prose.Block // headings waiting for the next body chapter
	for i := range b.Chapters {
		c := &b.Chapters[i]
		switch c.Section {
		case book.Front:
			front = append(front, c.Blocks...)
		case book.Back:
			back = append(back, c.Blocks...)
		case book.Body:
			if onlyHeadings(c.Blocks) {
				pending = append(pending, c.Blocks...)
				continue
			}
			title := c.Heading()
			if title == "" {
				title = fmt.Sprintf("Chapter %d", len(body)+1)
			}
			body = append(body, group{title: title, blocks: append(pending, c.Blocks...)})
			pending = nil
		}
	}
	if len(pending) > 0 {
		body = append(body, group{title: pending[0].Text, blocks: pending})
	}

	var out []group
	if len(front) > 0 {
		out = append(out, group{title: FrontTitle, blocks: front})
	}
	out = append(out, body...)
	if len(back) > 0 {
		out = append(out, group{title: BackTitle, blocks: back})
	}
	return out
}

func onlyHeadings(blocks []prose.Block) bool {
	for _, b := range blocks {
		if b.Kind != prose.Heading {
			return false
		}
	}
	return true
}

// Write saves the lessons as text files in dir. It refuses to write into a
// non-empty directory unless overwrite is set, so old lessons from another
// run don't get mixed in by accident.
func Write(dir string, lessons []Lesson, overwrite bool) error {
	entries, err := os.ReadDir(dir)
	switch {
	case errors.Is(err, os.ErrNotExist):
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	case err != nil:
		return err
	case len(entries) > 0 && !overwrite:
		return fmt.Errorf("output directory %s is not empty (use -force to overwrite)", dir)
	}
	for _, l := range lessons {
		if err := os.WriteFile(filepath.Join(dir, l.FileName()), []byte(l.Text), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// sanitize makes a title safe to use in a file name on every OS.
func sanitize(s string) string {
	s = strings.Map(func(r rune) rune {
		switch {
		case r < 32, strings.ContainsRune(`/\:*?"<>|`, r):
			return ' '
		}
		return r
	}, s)
	s = strings.Join(strings.Fields(s), " ")
	if utf8.RuneCountInString(s) > 80 {
		s = string([]rune(s)[:80])
	}
	s = strings.TrimRight(s, ". ") // Windows dislikes trailing dots and spaces
	if s == "" {
		return "Untitled"
	}
	return s
}
