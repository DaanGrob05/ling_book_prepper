// Package book turns an input file into repaired, classified chapters. It is
// the entry point a CLI or GUI uses: Load a book, inspect its chapters, then
// hand it to the lesson package.
package book

import (
	"fmt"
	"path/filepath"
	"strings"

	"lingqprep/internal/epub"
	"lingqprep/internal/lang"
	"lingqprep/internal/prose"
	"lingqprep/internal/source"
)

// Section says where in the book a chapter belongs.
type Section int

const (
	Front Section = iota // title page, copyright, table of contents, ...
	Body                 // the story itself
	Back                 // about the author, previews, acknowledgements, ...
)

func (s Section) String() string {
	return [...]string{"front", "body", "back"}[s]
}

// Chapter is one unit of the book with its paragraphs repaired.
type Chapter struct {
	Title   string
	Types   []string // semantic hints from the file
	Blocks  []prose.Block
	Words   int
	Section Section
	Reason  string // why it got its Section, for display
}

// Heading returns the chapter's title. Without a table of contents entry
// it is built from the headings the chapter opens with, so "Chapitre 1"
// followed by "Le survivant" becomes "Chapitre 1 - Le survivant".
func (c *Chapter) Heading() string {
	if c.Title != "" {
		return c.Title
	}
	var parts []string
	for _, b := range c.Blocks {
		if b.Kind != prose.Heading || len(parts) == 2 {
			break
		}
		parts = append(parts, strings.TrimRight(b.Text, ". "))
	}
	return strings.Join(parts, " - ")
}

// Book is a loaded, repaired and classified book.
type Book struct {
	Title    string
	Author   string
	Language *lang.Language
	Chapters []Chapter
}

// Options control loading. The zero value detects everything automatically.
type Options struct {
	// Language overrides the language declared in the file.
	Language *lang.Language
	// BodyFrom and BackFrom force the first body and first back matter
	// chapter (1-based numbers as shown by the list command); 0 means auto.
	BodyFrom, BackFrom int
}

// DefaultLanguage is used when the file declares no supported language.
const DefaultLanguage = "fr"

// Load reads, repairs and classifies the book at path.
func Load(path string, opts Options) (*Book, error) {
	src, err := read(path)
	if err != nil {
		return nil, err
	}

	l := opts.Language
	if l == nil {
		if l, err = lang.Get(src.Language); err != nil {
			l, _ = lang.Get(DefaultLanguage)
		}
	}

	vocab := prose.Vocabulary{}
	for _, s := range src.Sections {
		for _, bl := range s.Blocks {
			vocab.Add(bl.Text)
		}
	}

	b := &Book{Title: src.Title, Author: src.Author, Language: l}
	for _, s := range src.Sections {
		blocks := prose.Repair(s.Blocks, s.Title, l, vocab)
		for i, part := range splitAtHeadings(blocks, s.Title, l) {
			ch := Chapter{Title: s.Title, Types: s.Types, Blocks: part}
			if i > 0 {
				ch.Title, ch.Types = "", nil
			}
			for _, bl := range part {
				ch.Words += prose.CountWords(bl.Text)
			}
			b.Chapters = append(b.Chapters, ch)
		}
	}
	if len(b.Chapters) == 0 {
		return nil, fmt.Errorf("%s: no text found", path)
	}
	if err := b.Classify(opts.BodyFrom, opts.BackFrom); err != nil {
		return nil, err
	}
	return b, nil
}

// Classify sorts the chapters into front matter, body and back matter
// again. bodyFrom and backFrom force the first body and first back matter
// chapter (1-based); 0 means detect. On error the chapters are unchanged.
func (b *Book) Classify(bodyFrom, backFrom int) error {
	return classify(b.Chapters, b.Language, bodyFrom, backFrom)
}

// read picks a reader based on the file extension. A plain-text reader can
// be added here later; it only has to produce a source.Book.
func read(path string) (*source.Book, error) {
	switch ext := strings.ToLower(filepath.Ext(path)); ext {
	case ".epub":
		return epub.Open(path)
	default:
		return nil, fmt.Errorf("%s: unsupported file type %q (supported: .epub)", path, ext)
	}
}

// splitAtHeadings splits a section where a new chapter starts that the
// table of contents didn't point at:
//
//   - a back matter heading like "À propos de l'auteur", which happens when
//     the last chapter and the back matter share one file;
//   - in a section without a title (no usable table of contents, typical
//     for books converted from PDF), any chapter heading like "Chapitre 5";
//   - in a titled section, a chapter heading of the same kind as the title,
//     e.g. "Chapitre 6" inside "Chapitre 5" when one file holds several
//     chapters. Other headings, like numbered scenes, stay put.
func splitAtHeadings(blocks []prose.Block, title string, l *lang.Language) [][]prose.Block {
	titleWord := firstWord(lang.NormalizeTitle(title))
	startsChapter := func(text string) bool {
		norm := lang.NormalizeTitle(text)
		switch {
		case l.BackMatterTitle.MatchString(norm):
			return true
		case !l.ChapterTitle.MatchString(norm):
			return false
		case title == "":
			return true
		default:
			return firstWord(norm) == titleWord && norm != lang.NormalizeTitle(title)
		}
	}

	var parts [][]prose.Block
	start := 0
	for i, bl := range blocks {
		// i > 0 keeps a chapter's own opening heading attached to it.
		if i > 0 && bl.Kind == prose.Heading && startsChapter(bl.Text) {
			parts = append(parts, blocks[start:i])
			start = i
		}
	}
	return append(parts, blocks[start:])
}

func firstWord(s string) string {
	w, _, _ := strings.Cut(s, " ")
	return w
}
