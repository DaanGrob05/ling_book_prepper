package epub

import (
	"encoding/xml"
	"io"
	"strings"
)

// node is a minimal DOM node. Text nodes have an empty tag.
type node struct {
	tag   string // lowercase local name
	attrs []xml.Attr
	text  string
	kids  []*node
}

// parseXHTML reads an XHTML document into a tree. The decoder runs in
// non-strict mode with HTML entities so slightly broken files still load.
func parseXHTML(r io.Reader) (*node, error) {
	d := xml.NewDecoder(r)
	d.Strict = false
	d.AutoClose = xml.HTMLAutoClose
	d.Entity = xml.HTMLEntity

	root := &node{tag: "#document"}
	stack := []*node{root}
	for {
		tok, err := d.Token()
		if err == io.EOF {
			return root, nil
		}
		if err != nil {
			return nil, err
		}
		parent := stack[len(stack)-1]
		switch t := tok.(type) {
		case xml.StartElement:
			n := &node{tag: strings.ToLower(t.Name.Local), attrs: t.Attr}
			parent.kids = append(parent.kids, n)
			stack = append(stack, n)
		case xml.EndElement:
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			parent.kids = append(parent.kids, &node{text: string(t)})
		}
	}
}

// attr returns the value of the attribute with the given local name.
func (n *node) attr(name string) string {
	for _, a := range n.attrs {
		if a.Name.Local == name && a.Name.Space == "" {
			return a.Value
		}
	}
	return ""
}

// epubTypes returns the values of the epub:type attribute. Depending on
// whether the file declares the namespace, the decoder reports the prefix
// as the full namespace URL or just "epub", so any prefixed "type" counts.
func (n *node) epubTypes() []string {
	for _, a := range n.attrs {
		if a.Name.Local == "type" && a.Name.Space != "" {
			return strings.Fields(a.Value)
		}
	}
	return nil
}

func (n *node) hasEpubType(types ...string) bool {
	for _, t := range n.epubTypes() {
		for _, want := range types {
			if t == want {
				return true
			}
		}
	}
	return false
}

// textContent concatenates all text below n.
func (n *node) textContent() string {
	var b strings.Builder
	var walk func(*node)
	walk = func(n *node) {
		if n.tag == "" {
			b.WriteString(n.text)
			return
		}
		if n.tag == "br" {
			b.WriteByte(' ')
		}
		for _, k := range n.kids {
			walk(k)
		}
	}
	walk(n)
	return b.String()
}

// find returns the first node, in document order, for which match is true.
func (n *node) find(match func(*node) bool) *node {
	if match(n) {
		return n
	}
	for _, k := range n.kids {
		if f := k.find(match); f != nil {
			return f
		}
	}
	return nil
}

// findAll returns all nodes for which match is true, in document order.
func (n *node) findAll(match func(*node) bool) []*node {
	var out []*node
	var walk func(*node)
	walk = func(n *node) {
		if match(n) {
			out = append(out, n)
		}
		for _, k := range n.kids {
			walk(k)
		}
	}
	walk(n)
	return out
}
