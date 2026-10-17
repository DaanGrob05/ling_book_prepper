package prose

import (
	"strings"
	"unicode"

	"lingqprep/internal/lang"
)

// Sentences splits a paragraph into sentences. Joining the result with
// single spaces gives back the (normalized) paragraph.
//
// A sentence ends at an ender (. ! ? …), optionally followed by more enders
// and closing quotes or brackets, but only when whitespace and something
// that can start a sentence follow: a capital, a digit or an opener such as
// « or —. A period after an abbreviation never ends a sentence.
func Sentences(p string, l *lang.Language) []string {
	rs := []rune(p)
	var out []string
	start := 0
	for i := 0; i < len(rs); i++ {
		if !l.IsEnder(rs[i]) {
			continue
		}
		if rs[i] == '.' && l.IsAbbreviation(wordBefore(string(rs[start:i]))) {
			continue
		}
		end := i + 1
		for end < len(rs) && l.IsEnder(rs[end]) {
			end++
		}
		// Take closers along, also when (no-break) spaces separate them as
		// in French « Non ! ». A character that is both opener and closer,
		// like ", only closes when it touches the punctuation.
		for {
			k := skipSpace(rs, end)
			if k < len(rs) && l.IsCloser(rs[k]) && (k == end || !l.IsOpener(rs[k])) {
				end = k + 1
				continue
			}
			break
		}
		next := skipSpace(rs, end)
		if next == end || next == len(rs) || !startsSentence(rs[next], l) {
			i = end - 1
			continue
		}
		out = append(out, strings.TrimSpace(string(rs[start:end])))
		start = next
		i = next - 1
	}
	if tail := strings.TrimSpace(string(rs[start:])); tail != "" {
		out = append(out, tail)
	}
	return out
}

func skipSpace(rs []rune, i int) int {
	for i < len(rs) && isSpace(rs[i]) {
		i++
	}
	return i
}

func startsSentence(r rune, l *lang.Language) bool {
	return unicode.IsUpper(r) || unicode.IsDigit(r) || l.IsOpener(r)
}
