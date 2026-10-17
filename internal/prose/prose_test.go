package prose

import (
	"slices"
	"testing"

	"lingqprep/internal/lang"
)

func french(t *testing.T) *lang.Language {
	t.Helper()
	l, err := lang.Get("fr")
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func TestSentences(t *testing.T) {
	l := french(t)
	tests := []struct {
		in   string
		want []string
	}{
		{"Il pleut. Elle sort.", []string{"Il pleut.", "Elle sort."}},
		{"M. Dupont arriva. Il salua.", []string{"M. Dupont arriva.", "Il salua."}},
		{"Il lut J. K. Rowling. Puis il dormit.", []string{"Il lut J. K. Rowling.", "Puis il dormit."}},
		{"« Qui est là ? » demanda-t-il.", []string{"« Qui est là ? » demanda-t-il."}},
		{"« Non ! » Il partit.", []string{"« Non ! »", "Il partit."}},
		{"Elle hésita… Puis non.", []string{"Elle hésita…", "Puis non."}},
		{"Quoi ?! — Rien.", []string{"Quoi ?!", "— Rien."}},
		{`Il dit "fin." Puis "Bonjour" encore.`, []string{`Il dit "fin."`, `Puis "Bonjour" encore.`}},
		{"Il vit 3.14 partout.", []string{"Il vit 3.14 partout."}},
		{"pommes, poires, etc. Ensuite rien.", []string{"pommes, poires, etc.", "Ensuite rien."}},
		{"Il pleut. »", []string{"Il pleut. »"}},
	}
	for _, tt := range tests {
		if got := Sentences(tt.in, l); !slices.Equal(got, tt.want) {
			t.Errorf("Sentences(%q)\n got %q\nwant %q", tt.in, got, tt.want)
		}
	}
}

func p(text string) Block { return Block{Kind: Paragraph, Text: text} }
func h(text string) Block { return Block{Kind: Heading, Text: text} }

func TestRepair(t *testing.T) {
	l := french(t)
	in := []Block{
		p("Chapitre 1"),
		p("Le survivant"),
		p("Mr et Mrs Dursley, qui habitaient au 4, Privet Drive,"),
		p("avaient toujours affirmé avec la plus"),
		p("grande fierté qu'ils étaient normaux."),
		p("Ils avaient un petit garçon prénommé"),
		p("Dudley et c'était un enfant pré-"),
		p("cieux. Il vivait par-"),
		p("dessus tout pour lui. Il rencontra M."),
		p("Dupont."),
		p("« Qui est là ? »"),
		p("demanda-t-il."),
		p("* * *"),
		p("Il dit :"),
		p("« Bonjour. »"),
		p("— Salut, répondit-elle."),
		p("Il vit en Grande-"),
		p("Bretagne au-"),
		p("dessus d'un pub, au numéro"),
		p("4."),
	}
	want := []Block{
		h("Chapitre 1"),
		h("Le survivant"),
		p("Mr et Mrs Dursley, qui habitaient au 4, Privet Drive, avaient toujours affirmé avec la plus grande fierté qu'ils étaient normaux."),
		p("Ils avaient un petit garçon prénommé Dudley et c'était un enfant précieux. Il vivait par-dessus tout pour lui. Il rencontra M. Dupont."),
		p("« Qui est là ? » demanda-t-il."),
		p("* * *"),
		p("Il dit :"),
		p("« Bonjour. »"),
		p("— Salut, répondit-elle."),
		p("Il vit en Grande-Bretagne au-dessus d'un pub, au numéro 4."),
	}
	vocab := Vocabulary{}
	vocab.Add("Il est précieux. Il saute par-dessus le mur.")
	got := Repair(in, "Chapitre 1", l, vocab)
	if !slices.Equal(got, want) {
		t.Errorf("Repair mismatch\n got:")
		for _, b := range got {
			t.Errorf("  %d %q", b.Kind, b.Text)
		}
	}
}

func TestNormalize(t *testing.T) {
	got := Normalize("  Il  dit\n\tbon­jour !  ")
	if want := "Il dit bonjour !"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCountWords(t *testing.T) {
	if got := CountWords("« L'homme est là — aujourd’hui par-dessus 4 murs. »"); got != 8 {
		t.Errorf("got %d, want 8", got) // l'homme, est, là, aujourd’hui, par, dessus, 4, murs
	}
}
