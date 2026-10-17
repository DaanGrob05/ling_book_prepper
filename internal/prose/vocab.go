package prose

import (
	"strings"
	"unicode"
)

// Vocabulary counts the words that occur in a book, lowercased. Hyphenated
// words are kept whole ("au-dessus"), so the book itself can tell whether a
// word split over two lines as "au-" + "dessus" really has a hyphen.
type Vocabulary map[string]int

// Add counts the words in text.
func (v Vocabulary) Add(text string) {
	rs := []rune(strings.ToLower(text))
	start := -1
	for i := 0; i <= len(rs); i++ {
		inWord := i < len(rs) && (unicode.IsLetter(rs[i]) ||
			// a hyphen counts only between two letters
			(rs[i] == '-' && start >= 0 && i+1 < len(rs) && unicode.IsLetter(rs[i+1])))
		switch {
		case inWord && start < 0:
			start = i
		case !inWord && start >= 0:
			v[string(rs[start:i])]++
			start = -1
		}
	}
}

// keepHyphen decides whether "head-" at the end of a line followed by
// "tail" on the next is a real hyphenated word. The book's own spelling
// decides; without evidence the hyphen stays, because in French
// compounds (peut-être, lui-même, dit-il) are far more common at line ends
// than words broken by hyphenation.
func keepHyphen(head, tail string, v Vocabulary, hyphenWords map[string]bool) bool {
	head, tail = strings.ToLower(head), strings.ToLower(tail)
	if hyphenWords[tail] {
		return true
	}
	joined, hyphenated := v[head+tail], v[head+"-"+tail]
	return hyphenated > 0 || joined == 0
}
