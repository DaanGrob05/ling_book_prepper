package book

import (
	"slices"
	"testing"

	"lingqprep/internal/lang"
)

func TestClassify(t *testing.T) {
	fr, err := lang.Get("fr")
	if err != nil {
		t.Fatal(err)
	}
	b := &Book{Language: fr, Chapters: []Chapter{
		{Title: "", Words: 12},
		{Title: "Chapitre 1", Words: 4000},
		{Title: "Chapitre 2", Words: 4000},
		{Title: "À propos de l'auteur", Words: 200},
	}}
	sections := func() []Section {
		var s []Section
		for _, c := range b.Chapters {
			s = append(s, c.Section)
		}
		return s
	}

	if err := b.Classify(0, 0); err != nil {
		t.Fatal(err)
	}
	if got, want := sections(), []Section{Front, Body, Body, Back}; !slices.Equal(got, want) {
		t.Errorf("detected: got %v, want %v", got, want)
	}

	if err := b.Classify(3, 0); err != nil {
		t.Fatal(err)
	}
	if got, want := sections(), []Section{Front, Front, Body, Back}; !slices.Equal(got, want) {
		t.Errorf("body from 3: got %v, want %v", got, want)
	}
	if r := b.Chapters[2].Reason; r != "set manually" {
		t.Errorf("reason = %q, want %q", r, "set manually")
	}

	before := sections()
	if err := b.Classify(3, 2); err == nil {
		t.Error("back matter before the body accepted")
	}
	if got := sections(); !slices.Equal(got, before) {
		t.Errorf("failed Classify changed the chapters: %v, was %v", got, before)
	}
	if err := b.Classify(5, 0); err == nil {
		t.Error("chapter 5 of 4 accepted")
	}
}
