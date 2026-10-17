package lesson

import "testing"

func sizes(parts [][]unit) []int {
	var out []int
	for _, p := range parts {
		n := 0
		for _, u := range p {
			n += u.words
		}
		out = append(out, n)
	}
	return out
}

func repeat(n, words int) []unit {
	us := make([]unit, n)
	for i := range us {
		us[i] = unit{words: words}
	}
	return us
}

func TestSplitBalanced(t *testing.T) {
	// 650 sentences of 10 words: greedy would give 6000 + 500.
	got := sizes(split(repeat(650, 10), 6000))
	if len(got) != 2 || got[0] != 3250 || got[1] != 3250 {
		t.Errorf("got %v, want [3250 3250]", got)
	}
}

func TestSplitNeverOverLimit(t *testing.T) {
	us := []unit{}
	for i := range 2000 {
		us = append(us, unit{words: 1 + (i*37)%29})
	}
	parts := split(us, 6000)
	for _, n := range sizes(parts) {
		if n > 6000 {
			t.Errorf("part of %d words exceeds the limit", n)
		}
	}
	// Filling greedily up to the limit gives the fewest parts possible.
	if minParts := len(fill(us, 6000)); len(parts) != minParts {
		t.Errorf("got %d parts, want %d", len(parts), minParts)
	}
}

func TestSplitOversizedSentence(t *testing.T) {
	us := []unit{{words: 100}, {words: 7000}, {words: 100}}
	got := sizes(split(us, 6000))
	if len(got) != 3 || got[1] != 7000 {
		t.Errorf("got %v, want the long sentence alone", got)
	}
}

func TestFileName(t *testing.T) {
	tests := []struct {
		l    Lesson
		want string
	}{
		{Lesson{Number: 1, Title: "Chapitre 1", Parts: 1}, "001 - Chapitre 1.txt"},
		{Lesson{Number: 12, Title: "Chapitre 12", Part: 2, Parts: 3}, "012 - Chapitre 12 (2).txt"},
		{Lesson{Number: 3, Title: `Qui ? Quoi : "rien"...`, Parts: 1}, "003 - Qui Quoi rien.txt"},
	}
	for _, tt := range tests {
		if got := tt.l.FileName(); got != tt.want {
			t.Errorf("got %q, want %q", got, tt.want)
		}
	}
}
