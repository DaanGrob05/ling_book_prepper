// Package source defines what every input reader (EPUB now, plain text
// later) produces: the book's metadata and its sections in reading order.
package source

import "lingqprep/internal/prose"

// Book is the format-independent result of reading an input file.
type Book struct {
	Title    string
	Author   string
	Language string // as declared by the file, e.g. "fr-FR"; may be empty
	Sections []Section
}

// Section is a stretch of the book that starts a new chapter or other unit.
// Its Blocks are raw: they may still contain broken paragraphs.
type Section struct {
	Title  string   // from the table of contents; empty if it had no entry
	Types  []string // semantic hints from the file, e.g. "cover", "backmatter"
	Blocks []prose.Block
}
