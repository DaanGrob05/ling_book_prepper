package prose

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"lingqprep/internal/lang"
)

// MaxHeadingRunes is the longest text we are willing to guess is a heading
// when the source didn't mark it as one.
const MaxHeadingRunes = 80

// Repair turns the raw fragments of one chapter into whole paragraphs.
//
// Badly converted ebooks often put every printed line in its own paragraph,
// cutting sentences in half. Repair glues a fragment to the one before it
// unless the previous one clearly ended a sentence. title is the chapter
// title from the table of contents (may be empty); it helps recognise
// headings the source didn't mark as such. vocab holds the words of the
// whole book and decides about words hyphenated across lines; it may be nil.
func Repair(frags []Block, title string, l *lang.Language, vocab Vocabulary) []Block {
	clean := make([]Block, 0, len(frags))
	for _, f := range frags {
		f.Text = Normalize(f.Text)
		if f.Text != "" {
			clean = append(clean, f)
		}
	}
	markLeadingHeadings(clean, title)

	out := make([]Block, 0, len(clean))
	for _, f := range clean {
		if n := len(out); n > 0 && f.Kind == Paragraph && out[n-1].Kind == Paragraph {
			if joined, ok := join(out[n-1].Text, f.Text, l, vocab); ok {
				out[n-1].Text = joined
				continue
			}
		}
		out = append(out, f)
	}
	return out
}

// markLeadingHeadings guesses which fragments at the very start of a chapter
// are unmarked headings, such as "Chapitre 1" followed by "Le survivant".
func markLeadingHeadings(frags []Block, title string) {
	normTitle := lang.NormalizeTitle(title)
	for i := range frags {
		f := &frags[i]
		if f.Kind == Heading {
			continue
		}
		text := f.Text
		if normTitle != "" && strings.Contains(normTitle, lang.NormalizeTitle(text)) {
			f.Kind = Heading
			continue
		}
		last, _ := utf8.DecodeLastRuneInString(text)
		looksLikeHeading := utf8.RuneCountInString(text) <= MaxHeadingRunes &&
			!strings.ContainsRune(".,;:!?…»”", last) &&
			i+1 < len(frags) && !unicode.IsLower(firstRune(frags[i+1].Text))
		if !looksLikeHeading {
			return
		}
		f.Kind = Heading
	}
}

// join decides whether next continues the paragraph prev, and if so returns
// the two glued together.
func join(prev, next string, l *lang.Language, vocab Vocabulary) (string, bool) {
	// Scene breaks like "* * *" stand on their own.
	if !hasWord(prev) || !hasWord(next) {
		return "", false
	}
	first := firstRune(next)

	// A word split over two lines: "au-" + "dessus", "pré-" + "nommé" or
	// "Grande-" + "Bretagne".
	if base, ok := strings.CutSuffix(prev, "-"); ok && unicode.IsLetter(first) {
		if last, _ := utf8.DecodeLastRuneInString(base); unicode.IsLetter(last) {
			if unicode.IsUpper(first) || keepHyphen(trailingWord(base), leadingWord(next), vocab, l.HyphenWords) {
				return prev + next, true
			}
			return base + next, true
		}
	}

	// Paragraphs practically never start with a lowercase letter, so this is
	// a continuation even when prev ends in "?" or "!" (« Qui ? » demanda-t-il).
	if unicode.IsLower(first) {
		return prev + " " + next, true
	}
	if l.IsDialogueDash(first) {
		return "", false
	}

	trimmed := strings.TrimRightFunc(prev, isSpace)
	last, _ := utf8.DecodeLastRuneInString(trimmed)
	switch {
	case l.IsCloser(last):
		// « C'est fini » followed by a capital: most likely a real break.
		return "", false
	case l.IsEnder(last):
		if last == '.' && l.IsAbbreviation(wordBefore(trimmed[:len(trimmed)-1])) {
			return prev + " " + next, true // "M." + "Dupont"
		}
		return "", false
	case last == ':' && l.IsOpener(first):
		return "", false // "il dit :" followed by quoted speech
	}
	// No sentence-ending punctuation, so the sentence carries on:
	// "un petit garçon prénommé" + "Dudley".
	return prev + " " + next, true
}

// wordBefore returns the word at the end of s, keeping internal periods and
// hyphens so abbreviations like "J.-C" survive.
func wordBefore(s string) string {
	start := strings.LastIndexFunc(s, func(r rune) bool {
		return !(unicode.IsLetter(r) || r == '.' || r == '-')
	})
	return s[start+1:]
}

// leadingWord returns the letters at the start of s.
func leadingWord(s string) string {
	end := strings.IndexFunc(s, func(r rune) bool { return !unicode.IsLetter(r) })
	if end < 0 {
		end = len(s)
	}
	return s[:end]
}

// trailingWord returns the letters at the end of s.
func trailingWord(s string) string {
	return s[strings.LastIndexFunc(s, func(r rune) bool { return !unicode.IsLetter(r) })+1:]
}
