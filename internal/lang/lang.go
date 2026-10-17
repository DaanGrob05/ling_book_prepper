// Package lang holds the language-specific knowledge the rest of the program
// needs: which abbreviations don't end a sentence, which headings mark front
// or back matter, which characters open or close a quote, and so on.
//
// Adding a language means adding a file like fr.go that calls register in an
// init function. Nothing else in the program should hard-code language rules.
package lang

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode"
)

// Language describes the rules for one language. All fields are plain data
// so a new language is mostly a matter of filling in lists.
type Language struct {
	Code string // ISO 639-1 code, e.g. "fr"
	Name string

	// SentenceEnders are the characters that can end a sentence.
	SentenceEnders string
	// Closers may follow a sentence ender and still belong to the same
	// sentence, like the » in « Non ! ».
	Closers string
	// Openers are non-letter characters that may start a sentence, such as
	// an opening quote or a dialogue dash.
	Openers string
	// DialogueDashes start a new line of dialogue. A fragment that begins
	// with one is never glued to the fragment before it.
	DialogueDashes string

	// Abbreviations are words (lowercase, without the trailing period) that
	// are followed by a period but do not end a sentence.
	Abbreviations map[string]bool
	// HyphenWords are words that may follow a line-final hyphen and form a
	// real hyphenated word ("par-" + "dessus"). For any other word the
	// hyphen is treated as a line-break hyphen and removed.
	HyphenWords map[string]bool

	// The title patterns are matched against a normalized, lowercase title
	// (see NormalizeTitle).
	FrontMatterTitle *regexp.Regexp
	BackMatterTitle  *regexp.Regexp
	ChapterTitle     *regexp.Regexp

	// CopyrightMarkers identify legal/colophon pages (lowercase substrings).
	CopyrightMarkers []string
}

var registry = map[string]*Language{}

func register(l *Language) { registry[l.Code] = l }

// Get returns the language with the given code. It also accepts regional
// codes such as "fr-FR" or "fr_CA".
func Get(code string) (*Language, error) {
	base, _, _ := strings.Cut(strings.ToLower(code), "-")
	base, _, _ = strings.Cut(base, "_")
	if l, ok := registry[base]; ok {
		return l, nil
	}
	return nil, fmt.Errorf("unsupported language %q (supported: %s)", code, strings.Join(Codes(), ", "))
}

// Codes lists the supported language codes in sorted order.
func Codes() []string {
	codes := make([]string, 0, len(registry))
	for c := range registry {
		codes = append(codes, c)
	}
	slices.Sort(codes)
	return codes
}

// IsEnder, IsCloser, IsOpener and IsDialogueDash report whether r belongs to
// the corresponding character class.
func (l *Language) IsEnder(r rune) bool        { return strings.ContainsRune(l.SentenceEnders, r) }
func (l *Language) IsCloser(r rune) bool       { return strings.ContainsRune(l.Closers, r) }
func (l *Language) IsOpener(r rune) bool       { return strings.ContainsRune(l.Openers, r) }
func (l *Language) IsDialogueDash(r rune) bool { return strings.ContainsRune(l.DialogueDashes, r) }

// IsAbbreviation reports whether word, the text directly before a period,
// is an abbreviation rather than the end of a sentence.
func (l *Language) IsAbbreviation(word string) bool {
	if word == "" {
		return false
	}
	// Initials like "J." or "M." and dotted forms like "J.-C" or "U.S.A".
	rs := []rune(word)
	if len(rs) == 1 && unicode.IsUpper(rs[0]) {
		return true
	}
	if strings.Contains(word, ".") {
		return true
	}
	return l.Abbreviations[strings.ToLower(word)]
}

// HasCopyrightMarker reports whether text looks like a legal or colophon page.
func (l *Language) HasCopyrightMarker(text string) bool {
	text = strings.ToLower(text)
	for _, m := range l.CopyrightMarkers {
		if strings.Contains(text, m) {
			return true
		}
	}
	return false
}

// NormalizeTitle lowercases a title, unifies apostrophes and collapses
// whitespace so the title patterns can stay simple.
func NormalizeTitle(s string) string {
	s = strings.ToLower(s)
	s = strings.NewReplacer("’", "'", "‘", "'", "\u00a0", " ", "\u202f", " ").Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

func set(words ...string) map[string]bool {
	m := make(map[string]bool, len(words))
	for _, w := range words {
		m[w] = true
	}
	return m
}
