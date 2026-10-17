package lang

import "regexp"

func init() {
	register(&Language{
		Code: "fr",
		Name: "Français",

		SentenceEnders: ".!?…",
		Closers:        "»”\"’)]",
		Openers:        "«“\"‘'—–-([",
		DialogueDashes: "—–",

		Abbreviations: set(
			"m", "mm", "mme", "mmes", "mlle", "mlles", "mr", "mrs", "ms",
			"dr", "pr", "me", "mgr", "st", "ste", "cf", "ex", "env", "vol",
			"chap", "fig", "no", "av", "apr", "éd", "coll", "p", "pp", "vs",
		),
		HyphenWords: set(
			"t", "ci", "là", "même", "mêmes", "dessus", "dessous",
			"il", "ils", "elle", "elles", "on", "je", "tu", "nous", "vous",
			"le", "la", "les", "lui", "leur", "moi", "toi", "y", "en",
		),

		FrontMatterTitle: regexp.MustCompile(`^(couverture|cover|page de titre|titre|title|copyright|mentions légales|crédits|table des matières|sommaire|contents|dédicace|épigraphe|du même auteur|de la même autrice|des mêmes auteurs|chez le même éditeur|dans la même collection)`),
		BackMatterTitle:  regexp.MustCompile(`^(à propos de l'auteur|à propos de l'autrice|à propos des auteurs|l'auteur|l'autrice|les auteurs|biographie|du même auteur|de la même autrice|des mêmes auteurs|dans la même collection|chez le même éditeur|remerciements|postface|notes|note de l'auteur|note de l'autrice|glossaire|index|bibliographie|annexe|extrait|découvrez|à paraître|à découvrir|lisez|achevé d'imprimer|table des matières|sommaire|copyright|crédits|mentions légales|about the author|acknowledg|.*[,:–—-] ?l'(auteur|autrice)$)`),
		ChapterTitle:     regexp.MustCompile(`^((chapitre|partie|livre|tome|prologue|prélude|épilogue|interlude|acte)\b|[ivxlcdm]+([\s.:–—-]|$)|\d+([\s.:–—-]|$)|\S+ partie\b)`),

		CopyrightMarkers: []string{
			"isbn", "dépôt légal", "achevé d'imprimer", "achevé d’imprimer",
			"tous droits réservés", "©", "propriété intellectuelle",
		},
	})
}
