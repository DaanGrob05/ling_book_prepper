// Package prep is the API shared by the command line and the GUI: it turns
// a book file plus a few options into lessons and writes them to disk. The
// front ends only collect options and show results.
package prep

import (
	"fmt"
	"path/filepath"
	"strings"

	"lingqprep/internal/book"
	"lingqprep/internal/lang"
	"lingqprep/internal/lesson"
)

// Options are the user's choices for preparing a book. The zero value
// means "detect everything and use LingQ's limit".
type Options struct {
	// Language is a language code such as "fr"; "" uses the file's language.
	Language string
	// BodyFrom and BackFrom force the first body and first back matter
	// chapter (1-based, as numbered in Result.Book.Chapters); 0 means detect.
	BodyFrom, BackFrom int
	// MaxWords is the largest lesson; 0 means lesson.DefaultMaxWords.
	MaxWords int
}

// Result is a loaded book together with the lessons cut from it.
type Result struct {
	Book     *book.Book
	Lessons  []lesson.Lesson
	MaxWords int // the limit the lessons were built with
}

// Load reads the book at path and builds its lessons.
func Load(path string, opts Options) (*Result, error) {
	if opts.MaxWords == 0 {
		opts.MaxWords = lesson.DefaultMaxWords
	}
	if opts.MaxWords < 0 {
		return nil, fmt.Errorf("maximum words per lesson must be positive, got %d", opts.MaxWords)
	}
	bo := book.Options{BodyFrom: opts.BodyFrom, BackFrom: opts.BackFrom}
	if opts.Language != "" {
		l, err := lang.Get(opts.Language)
		if err != nil {
			return nil, err
		}
		bo.Language = l
	}
	b, err := book.Load(path, bo)
	if err != nil {
		return nil, err
	}
	return &Result{Book: b, Lessons: lesson.Build(b, opts.MaxWords), MaxWords: opts.MaxWords}, nil
}

// SetMaxWords rebuilds the lessons with a new word limit. That is much
// faster than Load because the book is not read and repaired again.
func (r *Result) SetMaxWords(maxWords int) error {
	if maxWords < 1 {
		return fmt.Errorf("maximum words per lesson must be positive, got %d", maxWords)
	}
	r.MaxWords = maxWords
	r.Lessons = lesson.Build(r.Book, maxWords)
	return nil
}

// Reclassify changes where the body and the back matter start (1-based
// chapter numbers, 0 = detect) and rebuilds the lessons. Like SetMaxWords
// it doesn't read the book again. On error nothing changes.
func (r *Result) Reclassify(bodyFrom, backFrom int) error {
	if err := r.Book.Classify(bodyFrom, backFrom); err != nil {
		return err
	}
	r.Lessons = lesson.Build(r.Book, r.MaxWords)
	return nil
}

// Oversized returns the lessons over the word limit. That only happens when
// a single sentence is longer than the limit, since lessons never end
// mid-sentence.
func (r *Result) Oversized() []lesson.Lesson {
	var over []lesson.Lesson
	for _, l := range r.Lessons {
		if l.Words > r.MaxWords {
			over = append(over, l)
		}
	}
	return over
}

// Write writes the lessons to dir; see lesson.Write.
func (r *Result) Write(dir string, overwrite bool) error {
	return lesson.Write(dir, r.Lessons, overwrite)
}

// DefaultDir is the command line's output directory for the book at path:
// its file name without extension, relative to the current directory.
func DefaultDir(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

// DirNextTo is the GUI's output directory for the book at path: a folder
// named after the book, next to it. A program started from a menu has no
// meaningful current directory, so DefaultDir would be a poor choice there.
func DirNextTo(path string) string {
	return filepath.Join(filepath.Dir(path), DefaultDir(path))
}
