// Package epub reads EPUB 2 and EPUB 3 files into a source.Book.
//
// An EPUB is a zip file. META-INF/container.xml points at the OPF package
// file, which lists the metadata, every file in the book (the manifest) and
// the reading order (the spine). The table of contents lives in a separate
// navigation document (EPUB 3) or NCX file (EPUB 2).
package epub

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"path"
	"slices"
	"strings"

	"lingqprep/internal/prose"
	"lingqprep/internal/source"
)

type container struct {
	Rootfiles []struct {
		FullPath string `xml:"full-path,attr"`
	} `xml:"rootfiles>rootfile"`
}

type opfPackage struct {
	Metadata struct {
		Titles    []string `xml:"title"`
		Creators  []string `xml:"creator"`
		Languages []string `xml:"language"`
	} `xml:"metadata"`
	Manifest []struct {
		ID         string `xml:"id,attr"`
		Href       string `xml:"href,attr"`
		MediaType  string `xml:"media-type,attr"`
		Properties string `xml:"properties,attr"`
	} `xml:"manifest>item"`
	Spine struct {
		Toc   string `xml:"toc,attr"`
		Items []struct {
			IDRef  string `xml:"idref,attr"`
			Linear string `xml:"linear,attr"`
		} `xml:"itemref"`
	} `xml:"spine"`
	Guide []struct {
		Type string `xml:"type,attr"`
		Href string `xml:"href,attr"`
	} `xml:"guide>reference"`
}

type ncxNavPoint struct {
	Label   string `xml:"navLabel>text"`
	Content struct {
		Src string `xml:"src,attr"`
	} `xml:"content"`
	Children []ncxNavPoint `xml:"navPoint"`
}

type ncxDoc struct {
	NavPoints []ncxNavPoint `xml:"navMap>navPoint"`
}

// link is a resolved reference into the book: a file in the zip plus an
// optional fragment (element id) inside it.
type link struct {
	doc, fragment string
}

type tocEntry struct {
	title string
	link
}

// reader holds the open zip while a book is being read.
type reader struct {
	files map[string]*zip.File
}

// Open reads the EPUB at path.
func Open(path string) (*source.Book, error) {
	z, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open epub: %w", err)
	}
	defer z.Close()

	r := &reader{files: make(map[string]*zip.File, len(z.File))}
	for _, f := range z.File {
		r.files[f.Name] = f
	}
	b, err := r.read()
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return b, nil
}

func (r *reader) read() (*source.Book, error) {
	var c container
	if err := r.decodeXML("META-INF/container.xml", &c); err != nil {
		return nil, err
	}
	if len(c.Rootfiles) == 0 {
		return nil, fmt.Errorf("container.xml lists no package file")
	}
	opfPath := c.Rootfiles[0].FullPath
	var opf opfPackage
	if err := r.decodeXML(opfPath, &opf); err != nil {
		return nil, err
	}

	b := &source.Book{
		Title:    first(opf.Metadata.Titles),
		Author:   first(opf.Metadata.Creators),
		Language: first(opf.Metadata.Languages),
	}

	manifest := map[string]string{} // id -> zip path
	var navPath string
	for _, it := range opf.Manifest {
		p := resolve(opfPath, it.Href).doc
		manifest[it.ID] = p
		if slices.Contains(strings.Fields(it.Properties), "nav") {
			navPath = p
		}
	}

	// Semantic hints per link: guide (EPUB 2) and landmarks (EPUB 3).
	hints := map[link][]string{}
	for _, g := range opf.Guide {
		l := resolve(opfPath, g.Href)
		hints[l] = append(hints[l], g.Type)
	}

	var toc []tocEntry
	if navPath != "" {
		var landmarks map[link][]string
		toc, landmarks = r.readNav(navPath)
		for l, t := range landmarks {
			hints[l] = append(hints[l], t...)
		}
	}
	if len(toc) == 0 && manifest[opf.Spine.Toc] != "" {
		toc = r.readNCX(manifest[opf.Spine.Toc])
	}

	for _, it := range opf.Spine.Items {
		if it.Linear == "no" || manifest[it.IDRef] == "" {
			continue
		}
		if err := r.addDocument(b, manifest[it.IDRef], toc, hints); err != nil {
			return nil, err
		}
	}

	// Drop sections without text, such as a cover that is only an image.
	b.Sections = slices.DeleteFunc(b.Sections, func(s source.Section) bool {
		return len(s.Blocks) == 0
	})
	return b, nil
}

// addDocument appends the content of one spine document to the book.
//
// A table of contents entry starts a new section, either at the start of a
// document or at an anchor inside it. A document without an entry continues
// the previous section (long chapters are often split over several files),
// unless it opens with a heading or carries semantic hints, which suggests
// it's a separate unit such as an untitled title page or "about the author".
func (r *reader) addDocument(b *source.Book, docPath string, toc []tocEntry, hints map[link][]string) error {
	f, err := r.open(docPath)
	if err != nil {
		return err
	}
	defer f.Close()
	doc, err := parseXHTML(f)
	if err != nil {
		return fmt.Errorf("parse %s: %w", docPath, err)
	}

	entries := map[string]string{} // fragment -> title; "" is the document itself
	for _, e := range toc {
		if _, seen := entries[e.fragment]; e.doc == docPath && !seen {
			entries[e.fragment] = e.title
		}
	}
	anchors := map[string]bool{}
	for frag := range entries {
		if frag != "" {
			anchors[frag] = true
		}
	}
	items := extract(doc, anchors)
	types := append(docTypes(doc), hints[link{doc: docPath}]...)

	title, hasEntry := entries[""]
	startsWithHeading := false
	for _, it := range items {
		if it.anchor == "" {
			startsWithHeading = it.block.Kind == prose.Heading
			break
		}
	}

	if hasEntry || len(b.Sections) == 0 || startsWithHeading || len(types) > 0 {
		b.Sections = append(b.Sections, source.Section{Title: title, Types: types})
	}
	for _, it := range items {
		cur := &b.Sections[len(b.Sections)-1]
		if it.anchor == "" {
			cur.Blocks = append(cur.Blocks, it.block)
			continue
		}
		title := entries[it.anchor]
		if len(cur.Blocks) == 0 {
			// The anchor sits at the top of a section we just started.
			if cur.Title == "" {
				cur.Title = title
			}
			continue
		}
		b.Sections = append(b.Sections, source.Section{
			Title: title,
			Types: hints[link{doc: docPath, fragment: it.anchor}],
		})
	}
	return nil
}

// readNav reads the EPUB 3 navigation document: the toc nav for chapter
// titles and the landmarks nav for semantic hints.
func (r *reader) readNav(navPath string) ([]tocEntry, map[link][]string) {
	f, err := r.open(navPath)
	if err != nil {
		return nil, nil
	}
	defer f.Close()
	doc, err := parseXHTML(f)
	if err != nil {
		return nil, nil
	}

	var toc []tocEntry
	landmarks := map[link][]string{}
	for _, nav := range doc.findAll(func(n *node) bool { return n.tag == "nav" }) {
		links := nav.findAll(func(n *node) bool { return n.tag == "a" && n.attr("href") != "" })
		switch {
		case nav.hasEpubType("toc"):
			for _, a := range links {
				toc = append(toc, tocEntry{
					title: prose.Normalize(a.textContent()),
					link:  resolve(navPath, a.attr("href")),
				})
			}
		case nav.hasEpubType("landmarks"):
			for _, a := range links {
				l := resolve(navPath, a.attr("href"))
				landmarks[l] = append(landmarks[l], a.epubTypes()...)
			}
		}
	}
	return toc, landmarks
}

// readNCX reads the EPUB 2 table of contents, flattening nested entries.
func (r *reader) readNCX(ncxPath string) []tocEntry {
	var n ncxDoc
	if err := r.decodeXML(ncxPath, &n); err != nil {
		return nil
	}
	var toc []tocEntry
	var walk func([]ncxNavPoint)
	walk = func(points []ncxNavPoint) {
		for _, p := range points {
			toc = append(toc, tocEntry{
				title: prose.Normalize(p.Label),
				link:  resolve(ncxPath, p.Content.Src),
			})
			walk(p.Children)
		}
	}
	walk(n.NavPoints)
	return toc
}

func (r *reader) open(name string) (io.ReadCloser, error) {
	f, ok := r.files[name]
	if !ok {
		return nil, fmt.Errorf("missing file %s", name)
	}
	return f.Open()
}

func (r *reader) decodeXML(name string, v any) error {
	f, err := r.open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	d := xml.NewDecoder(f)
	d.Strict = false
	d.Entity = xml.HTMLEntity
	if err := d.Decode(v); err != nil {
		return fmt.Errorf("parse %s: %w", name, err)
	}
	return nil
}

// resolve turns an href found in the file at base into a zip path plus
// fragment, e.g. ("OEBPS/nav.xhtml", "text/ch1.xhtml#s2") becomes
// {"OEBPS/text/ch1.xhtml", "s2"}.
func resolve(base, href string) link {
	href, frag, _ := strings.Cut(href, "#")
	if u, err := url.PathUnescape(href); err == nil {
		href = u
	}
	doc := base
	if href != "" {
		doc = path.Join(path.Dir(base), href)
	}
	return link{doc: doc, fragment: frag}
}

func first(s []string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.TrimSpace(s[0])
}
