// Package prose works on plain text: it repairs paragraphs that were broken
// into pieces, splits paragraphs into sentences and counts words. It knows
// nothing about file formats; readers (EPUB, later TXT) hand it Blocks.
package prose

import (
	"strings"
	"unicode"
)

// Kind tells headings apart from ordinary paragraphs.
type Kind int

const (
	Paragraph Kind = iota
	Heading
)

// Block is one paragraph or heading of text.
type Block struct {
	Kind Kind
	Text string
}

// Normalize cleans up whitespace and invisible characters: runs of
// whitespace become a single space (or a single no-break space when the run
// contained one, as French typography uses them before ; : ! ? and »),
// soft hyphens and zero-width characters are dropped, and the result is trimmed.
func Normalize(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	var space rune // pending whitespace to write before the next character
	for _, r := range s {
		switch {
		case r == '\u00ad', r == '\u200b', r == '\u200c', r == '\u200d', r == '\ufeff':
			continue
		case r == '\u00a0' || r == '\u202f':
			space = r
		case unicode.IsSpace(r):
			if space == 0 {
				space = ' '
			}
		default:
			if space != 0 && b.Len() > 0 {
				b.WriteRune(space)
			}
			space = 0
			b.WriteRune(r)
		}
	}
	return b.String()
}

// hasWord reports whether s contains a letter or digit; text without one,
// like "* * *", is a scene break.
func hasWord(s string) bool {
	return strings.IndexFunc(s, func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }) >= 0
}

func firstRune(s string) rune {
	for _, r := range s {
		return r
	}
	return 0
}

func isSpace(r rune) bool { return unicode.IsSpace(r) }
