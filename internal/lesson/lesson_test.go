package lesson

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lingqprep/internal/book"
)

// writeEPUB builds a minimal EPUB 3 file. docs maps file names to their
// <body> content and are put in the spine in the given order.
func writeEPUB(t *testing.T, docs [][2]string, nav string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.epub")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	z := zip.NewWriter(f)
	add := func(name, content string) {
		w, err := z.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(content))
	}
	add("mimetype", "application/epub+zip")
	add("META-INF/container.xml", `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
<rootfiles><rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/></rootfiles>
</container>`)

	var manifest, spine strings.Builder
	for i, d := range docs {
		fmt.Fprintf(&manifest, `<item id="d%d" href="text/%s" media-type="application/xhtml+xml"/>`, i, d[0])
		fmt.Fprintf(&spine, `<itemref idref="d%d"/>`, i)
		add("OEBPS/text/"+d[0], `<?xml version="1.0" encoding="utf-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head><title>x</title><link rel="stylesheet" href="s.css"/></head>
<body>`+d[1]+`</body></html>`)
	}
	add("OEBPS/content.opf", `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
<metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
<dc:title>Le Livre</dc:title><dc:creator>Jeanne Auteur</dc:creator><dc:language>fr-FR</dc:language>
</metadata>
<manifest><item id="nav" href="nav.xhtml" properties="nav" media-type="application/xhtml+xml"/>`+manifest.String()+`</manifest>
<spine>`+spine.String()+`</spine>
</package>`)
	add("OEBPS/nav.xhtml", `<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops"><body>`+nav+`</body></html>`)
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return path
}

// brokenParagraphs returns n sentences of 10 words each, with every
// sentence cut over two <p> elements the way bad conversions do.
func brokenParagraphs(n int) string {
	var b strings.Builder
	for i := range n {
		fmt.Fprintf(&b, "<p>Le chat numéro %d dort sur le</p>\n<p>tapis rouge du salon.</p>\n", i)
	}
	return b.String()
}

func TestEndToEnd(t *testing.T) {
	docs := [][2]string{
		{"cover.xhtml", `<section epub:type="cover"><img src="c.jpg" alt=""/></section>`},
		{"title.xhtml", `<p>Le Livre</p><p>Jeanne Auteur</p>`},
		{"copy.xhtml", `<p>© 2024 Éditions Exemple. ISBN 978-2-0000-0000-0</p>`},
		{"ch1.xhtml", `<h1>Chapitre 1</h1>` + brokenParagraphs(20)},
		{"ch1b.xhtml", `<p>Suite du premier chapitre</p><p>sans nouveau titre.</p>`},
		{"ch2.xhtml", `<h1>Chapitre 2</h1>` + brokenParagraphs(1300)},
		{"ch3.xhtml", `<h1>Chapitre 3</h1>` + brokenParagraphs(10) +
			`<h2>À propos de l'auteur</h2><p>Jeanne Auteur vit à Lyon.</p>`},
		{"preview.xhtml", `<h1>Chapitre 1</h1><p>Extrait du prochain livre.</p>`},
	}
	nav := `<nav epub:type="toc"><ol>
<li><a href="text/ch1.xhtml">Chapitre 1</a></li>
<li><a href="text/ch2.xhtml">Chapitre 2</a></li>
<li><a href="text/ch3.xhtml">Chapitre 3</a></li>
<li><a href="text/preview.xhtml">Chapitre 1</a></li>
</ol></nav>`
	b, err := book.Load(writeEPUB(t, docs, nav), book.Options{})
	if err != nil {
		t.Fatal(err)
	}

	var names []string
	lessons := Build(b, DefaultMaxWords)
	for _, l := range lessons {
		names = append(names, l.FileName())
		if l.Words > DefaultMaxWords {
			t.Errorf("%s has %d words", l.FileName(), l.Words)
		}
	}
	want := []string{
		"000 - Front matter.txt",
		"001 - Chapitre 1.txt",
		"002 - Chapitre 2 (1).txt",
		"003 - Chapitre 2 (2).txt",
		"004 - Chapitre 2 (3).txt",
		"005 - Chapitre 3.txt",
		"006 - Back matter.txt",
	}
	if strings.Join(names, "\n") != strings.Join(want, "\n") {
		t.Fatalf("files:\n%s\nwant:\n%s", strings.Join(names, "\n"), strings.Join(want, "\n"))
	}

	ch1 := lessons[1].Text
	if !strings.HasPrefix(ch1, "Chapitre 1\n\nLe chat numéro 0 dort sur le tapis rouge du salon.\n\nLe chat numéro 1") {
		t.Errorf("chapter 1 not repaired:\n%.200s", ch1)
	}
	if !strings.HasSuffix(ch1, "Suite du premier chapitre sans nouveau titre.\n") {
		t.Errorf("continuation file not appended to chapter 1:\n...%s", ch1[len(ch1)-100:])
	}
	for _, l := range lessons[2:5] {
		if !strings.HasSuffix(l.Text, "salon.\n") {
			t.Errorf("%s does not end on a full sentence", l.FileName())
		}
	}
	back := lessons[6].Text
	if !strings.Contains(back, "À propos de l'auteur") || !strings.Contains(back, "Extrait du prochain livre.") {
		t.Errorf("back matter incomplete:\n%s", back)
	}
}

// TestPDFConvertedBook mimics an EPUB made from a PDF: one big file, a table
// of contents with a dead link, chapters marked only by bold lines, every
// printed line in its own paragraph and the author's biography at the end.
func TestPDFConvertedBook(t *testing.T) {
	chapter := func(n int, subtitle string) string {
		return fmt.Sprintf(`<p><b>Chapitre %d</b></p><p> <i><b>%s</b></i></p>`, n, subtitle) + brokenParagraphs(5)
	}
	body := `<p><b>Le Livre</b></p><p><b>par Jeanne Auteur</b></p>` +
		chapter(1, "Le début") + chapter(2, "Le milieu.") + chapter(3, "La fin") +
		`<p> <i><b>JEANNE AUTEUR, L'Auteur</b></i></p><p>Jeanne est née à Lyon.</p>`
	docs := [][2]string{
		{"index_split_000.html", body},
		{"index_split_001.html", `<h1>Document Outline</h1><ul><li><a href="index_split_000.html#72">Le milieu</a></li></ul>`},
	}
	nav := `<nav epub:type="toc"><ol><li><a href="text/index_split_000.html#72">Le milieu</a></li></ol></nav>`
	b, err := book.Load(writeEPUB(t, docs, nav), book.Options{})
	if err != nil {
		t.Fatal(err)
	}

	var names []string
	lessons := Build(b, DefaultMaxWords)
	for _, l := range lessons {
		names = append(names, l.FileName())
	}
	want := []string{
		"000 - Front matter.txt",
		"001 - Chapitre 1 - Le début.txt",
		"002 - Chapitre 2 - Le milieu.txt",
		"003 - Chapitre 3 - La fin.txt",
		"004 - Back matter.txt",
	}
	if strings.Join(names, "\n") != strings.Join(want, "\n") {
		t.Fatalf("files:\n%s\nwant:\n%s", strings.Join(names, "\n"), strings.Join(want, "\n"))
	}
	if got := lessons[1].Text; !strings.HasPrefix(got, "Chapitre 1\n\nLe début\n\nLe chat numéro 0 dort sur le tapis rouge du salon.\n\n") {
		t.Errorf("chapter 1 starts wrong:\n%.150s", got)
	}
	if got := lessons[3].Text; strings.Contains(got, "Lyon") {
		t.Errorf("biography left in last chapter:\n%s", got)
	}
}
