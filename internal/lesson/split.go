// Package lesson cuts a book into LingQ lessons: one or more files per
// chapter, never more than a word limit, never cutting a sentence in half.
package lesson

import (
	"strings"

	"lingqprep/internal/lang"
	"lingqprep/internal/prose"
)

// unit is the smallest piece a lesson boundary may fall between: a
// sentence, or a whole heading.
type unit struct {
	text     string
	words    int
	newBlock bool // starts a new paragraph or heading
}

func units(blocks []prose.Block, l *lang.Language) []unit {
	var out []unit
	for _, b := range blocks {
		parts := []string{b.Text}
		if b.Kind == prose.Paragraph {
			parts = prose.Sentences(b.Text, l)
		}
		for i, s := range parts {
			out = append(out, unit{text: s, words: prose.CountWords(s), newBlock: i == 0})
		}
	}
	return out
}

// split divides us into as few parts as possible that each stay within
// maxWords, and makes those parts as even as possible: 6500 words become
// two parts of about 3250 rather than 6000 + 500.
//
// It first finds the minimum number of parts by filling greedily up to
// maxWords, then binary-searches the smallest capacity that still needs
// no more parts than that. A single sentence longer than maxWords gets a
// part of its own; there is no way to avoid going over then.
func split(us []unit, maxWords int) [][]unit {
	total, largest := 0, 0
	for _, u := range us {
		total += u.words
		largest = max(largest, u.words)
	}
	if total <= maxWords {
		return [][]unit{us}
	}

	parts := len(fill(us, maxWords))
	lo := max(largest, (total+parts-1)/parts)
	hi := max(maxWords, largest)
	for lo < hi {
		mid := (lo + hi) / 2
		if len(fill(us, mid)) <= parts {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return fill(us, lo)
}

// fill packs units greedily into parts of at most capacity words.
func fill(us []unit, capacity int) [][]unit {
	var parts [][]unit
	start, words := 0, 0
	for i, u := range us {
		if words+u.words > capacity && i > start {
			parts = append(parts, us[start:i])
			start, words = i, 0
		}
		words += u.words
	}
	return append(parts, us[start:])
}

// render turns units back into text: paragraphs separated by blank lines,
// sentences within a paragraph by a space.
func render(us []unit) string {
	var b strings.Builder
	for i, u := range us {
		switch {
		case i == 0:
		case u.newBlock:
			b.WriteString("\n\n")
		default:
			b.WriteByte(' ')
		}
		b.WriteString(u.text)
	}
	b.WriteByte('\n')
	return b.String()
}
