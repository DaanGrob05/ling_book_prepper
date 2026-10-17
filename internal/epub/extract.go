package epub

import (
	"strings"
	"unicode/utf8"

	"lingqprep/internal/prose"
)

// item is one piece of a content document in reading order: either a text
// fragment or an anchor (an element id the table of contents points at).
type item struct {
	anchor string
	block  prose.Block
}

var blockTags = map[string]bool{
	"p": true, "div": true, "section": true, "article": true, "header": true,
	"footer": true, "main": true, "body": true, "blockquote": true, "li": true,
	"ul": true, "ol": true, "dl": true, "dt": true, "dd": true, "figure": true,
	"figcaption": true, "table": true, "tr": true, "td": true, "th": true,
	"pre": true, "address": true, "aside": true, "hr": true, "center": true,
	"nav": true,
}

var headingTags = map[string]bool{
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
}

// extract walks a content document and returns its fragments. Every block
// element and every <br> ends a fragment; prose.Repair glues them back
// together later. anchors holds the ids the table of contents refers to.
func extract(doc *node, anchors map[string]bool) []item {
	var items []item
	var buf strings.Builder

	flush := func() {
		if strings.TrimSpace(buf.String()) != "" {
			items = append(items, item{block: prose.Block{Kind: prose.Paragraph, Text: buf.String()}})
		}
		buf.Reset()
	}

	var walk func(*node)
	walk = func(n *node) {
		if n.tag == "" {
			buf.WriteString(n.text)
			return
		}
		if id := n.attr("id"); id != "" && anchors[id] {
			flush()
			items = append(items, item{anchor: id})
		}
		switch {
		case skip(n):
			return
		case headingTags[n.tag], isBoldLine(n):
			flush()
			items = append(items, item{block: prose.Block{Kind: prose.Heading, Text: n.textContent()}})
			return
		case n.tag == "br":
			flush()
			return
		}
		isBlock := blockTags[n.tag]
		if isBlock {
			flush()
		}
		for _, k := range n.kids {
			walk(k)
		}
		if isBlock {
			flush()
		}
	}
	walk(doc)
	flush()
	return items
}

// skip reports whether n holds no reading text: metadata, scripts, and
// footnote markers or bodies (which would otherwise end up mid-sentence).
func skip(n *node) bool {
	switch n.tag {
	case "head", "script", "style", "rt", "rp":
		return true
	case "sup":
		return n.find(func(k *node) bool { return k.tag == "a" }) != nil
	}
	return n.hasEpubType("noteref", "footnote", "endnote", "rearnote", "pagebreak")
}

// isBoldLine reports whether n is a short paragraph that is bold from start
// to end, like <p><b>Chapitre 5</b></p>. Books converted from PDF mark
// headings this way instead of using <h1>.
func isBoldLine(n *node) bool {
	if n.tag != "p" && n.tag != "div" {
		return false
	}
	text := prose.Normalize(n.textContent())
	if text == "" || utf8.RuneCountInString(text) > prose.MaxHeadingRunes {
		return false
	}
	var allBold func(n *node, bold bool) bool
	allBold = func(k *node, bold bool) bool {
		switch {
		case k.tag == "":
			return bold || strings.TrimSpace(k.text) == ""
		case k != n && blockTags[k.tag]:
			return false
		case k.tag == "b" || k.tag == "strong":
			bold = true
		}
		for _, kid := range k.kids {
			if !allBold(kid, bold) {
				return false
			}
		}
		return true
	}
	return allBold(n, false)
}

// docTypes collects epub:type hints from <body> and its top-level sections.
func docTypes(doc *node) []string {
	body := doc.find(func(n *node) bool { return n.tag == "body" })
	if body == nil {
		return nil
	}
	types := body.epubTypes()
	for _, k := range body.kids {
		if k.tag == "section" {
			types = append(types, k.epubTypes()...)
		}
	}
	return types
}
