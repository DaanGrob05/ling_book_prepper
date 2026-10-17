package prose

import "unicode"

// CountWords counts words close to the way LingQ does: a run of letters or
// digits is a word, and an apostrophe between letters doesn't split it, so
// "l'homme" is one word while "par-dessus" is two. Unlike LingQ it also
// counts numbers, which keeps the result a little on the safe (high) side.
func CountWords(s string) int {
	rs := []rune(s)
	isWordRune := func(i int) bool {
		return i >= 0 && i < len(rs) && (unicode.IsLetter(rs[i]) || unicode.IsDigit(rs[i]) || unicode.IsMark(rs[i]))
	}
	n := 0
	for i := range rs {
		if !isWordRune(i) {
			continue
		}
		// Count a word at its first rune; a rune after an apostrophe that
		// itself follows a word rune continues that word.
		continues := isWordRune(i-1) || (i >= 2 && isApostrophe(rs[i-1]) && isWordRune(i-2))
		if !continues {
			n++
		}
	}
	return n
}

func isApostrophe(r rune) bool { return r == '\'' || r == '’' }
