package book

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"lingqprep/internal/lang"
)

// Semantic types (EPUB 3 epub:type values and EPUB 2 guide types) that mark
// front or back matter. Some, like "toc", can appear on either side.
var (
	frontTypes = []string{
		"cover", "titlepage", "title-page", "halftitlepage", "copyright-page",
		"toc", "dedication", "epigraph", "frontmatter", "seriespage", "imprint",
	}
	backTypes = []string{
		"backmatter", "appendix", "afterword", "acknowledgments", "colophon",
		"index", "glossary", "bibliography", "endnotes", "rearnotes",
		"contributors", "other-credits", "copyright-page", "toc",
	}
	chapterTypes = []string{"chapter", "part", "prologue", "epilogue", "bodymatter"}

	// converterTitle matches pages that conversion tools add, whatever the
	// book's language, such as the outline page pdftohtml generates.
	converterTitle = regexp.MustCompile(`^(document outline)$`)
)

const (
	// shortWords is the size below which an untitled chapter is assumed to
	// be a title page or similar, not story text.
	shortWords = 40
	// legalWords is the size below which a chapter mentioning ISBN or
	// copyright is assumed to be a legal page.
	legalWords = 300
)

// classify sets Section and Reason on every chapter. The book is split in
// three consecutive runs: front matter, body, back matter.
func classify(chs []Chapter, l *lang.Language, opts Options) error {
	n := len(chs)
	for _, v := range []int{opts.BodyFrom, opts.BackFrom} {
		if v < 0 || v > n {
			return fmt.Errorf("chapter %d does not exist (the book has %d)", v, n)
		}
	}

	bodyStart, bodyReason := findBodyStart(chs, l, opts)
	backStart, backReason := findBackStart(chs, l, opts, bodyStart)
	if backStart < bodyStart {
		return fmt.Errorf("back matter (chapter %d) starts before the body (chapter %d)", backStart+1, bodyStart+1)
	}

	for i := range chs {
		c := &chs[i]
		switch {
		case i < bodyStart:
			c.Section, c.Reason = Front, frontReason(c, l)
			if c.Reason == "" {
				c.Reason = "before the body"
			}
		case i < backStart:
			c.Section = Body
		default:
			c.Section = Back
		}
	}
	if bodyStart < n {
		chs[bodyStart].Reason = bodyReason
	}
	if backStart < n {
		chs[backStart].Reason = backReason
	}
	return nil
}

func findBodyStart(chs []Chapter, l *lang.Language, opts Options) (int, string) {
	if opts.BodyFrom > 0 {
		return opts.BodyFrom - 1, "set with -body-from"
	}
	for i, c := range chs {
		if slices.Contains(c.Types, "bodymatter") {
			return i, "marked as start of body"
		}
	}
	for i := range chs {
		if frontReason(&chs[i], l) == "" {
			return i, "first chapter that isn't front matter"
		}
	}
	return 0, "no story text recognised"
}

func findBackStart(chs []Chapter, l *lang.Language, opts Options, bodyStart int) (int, string) {
	if opts.BackFrom > 0 {
		return opts.BackFrom - 1, "set with -back-from"
	}
	seen := map[string]bool{}
	for i := bodyStart; i < len(chs); i++ {
		c := &chs[i]
		if i > bodyStart {
			if r := backReason(c, l); r != "" {
				return i, r
			}
		}
		// A second "Chapitre 1" is usually a preview of the next book.
		title := lang.NormalizeTitle(c.Heading())
		if title != "" && isChapterish(c, l) {
			if seen[title] {
				return i, fmt.Sprintf("repeats title %q (preview of another book?)", c.Heading())
			}
			seen[title] = true
		}
	}
	return len(chs), ""
}

// frontReason explains why c looks like front matter, or returns "".
func frontReason(c *Chapter, l *lang.Language) string {
	title := lang.NormalizeTitle(c.Heading())
	switch {
	case hasAny(c.Types, frontTypes):
		return "marked as " + strings.Join(c.Types, ", ")
	case title != "" && l.FrontMatterTitle.MatchString(title):
		return "title looks like front matter"
	case c.Words < legalWords && l.HasCopyrightMarker(c.text()):
		return "copyright/legal page"
	case c.Words < shortWords && !isChapterish(c, l):
		return "very short, no chapter heading"
	}
	return ""
}

// backReason explains why c looks like back matter, or returns "".
func backReason(c *Chapter, l *lang.Language) string {
	title := lang.NormalizeTitle(c.Heading())
	switch {
	case hasAny(c.Types, backTypes):
		return "marked as " + strings.Join(c.Types, ", ")
	case converterTitle.MatchString(title):
		return "added by a conversion tool"
	case title != "" && l.BackMatterTitle.MatchString(title):
		return "title looks like back matter"
	case c.Words < legalWords && l.HasCopyrightMarker(c.text()):
		return "copyright/legal page"
	}
	return ""
}

func isChapterish(c *Chapter, l *lang.Language) bool {
	return hasAny(c.Types, chapterTypes) || l.ChapterTitle.MatchString(lang.NormalizeTitle(c.Heading()))
}

func (c *Chapter) text() string {
	var b strings.Builder
	for _, bl := range c.Blocks {
		b.WriteString(bl.Text)
		b.WriteByte('\n')
	}
	return b.String()
}

func hasAny(have, want []string) bool {
	for _, h := range have {
		if slices.Contains(want, h) {
			return true
		}
	}
	return false
}
