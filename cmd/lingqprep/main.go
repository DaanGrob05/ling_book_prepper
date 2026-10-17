// Command lingqprep prepares ebooks for import into LingQ: it repairs
// sentences broken over several lines and cuts the book into lessons that
// fit LingQ's word limit, one or more files per chapter.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"lingqprep/internal/lang"
	"lingqprep/internal/lesson"
	"lingqprep/internal/prep"
)

const usage = `lingqprep prepares ebooks for LingQ.

Usage:
  lingqprep [flags] book.epub        write the lessons to ./<book>/
  lingqprep list [flags] book.epub   show the chapters and how they were classified
  lingqprep gui [book.epub]          open the graphical interface

Run "lingqprep -h" or "lingqprep list -h" to see the flags.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	// Preparing a book is the default; "list" and "gui" are the subcommands.
	var err error
	switch os.Args[1] {
	case "list":
		err = runList(os.Args[2:])
	case "gui":
		err = runGUI(os.Args[2:])
	default:
		err = runPrepare(os.Args[1:])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// loadFlags are shared by both commands.
type loadFlags struct {
	lang     string
	bodyFrom int
	backFrom int
	maxWords int
}

func (f *loadFlags) register(fs *flag.FlagSet) {
	fs.StringVar(&f.lang, "lang", "", "book language ("+strings.Join(lang.Codes(), ", ")+"); default: from the file")
	fs.IntVar(&f.bodyFrom, "body-from", 0, "number of the first story chapter (see list); 0 = detect")
	fs.IntVar(&f.backFrom, "back-from", 0, "number of the first back matter chapter (see list); 0 = detect")
	fs.IntVar(&f.maxWords, "max-words", lesson.DefaultMaxWords, "maximum words per lesson")
}

func (f *loadFlags) load(path string) (*prep.Result, error) {
	if f.maxWords < 1 {
		return nil, fmt.Errorf("-max-words must be positive")
	}
	return prep.Load(path, prep.Options{
		Language: f.lang,
		BodyFrom: f.bodyFrom,
		BackFrom: f.backFrom,
		MaxWords: f.maxWords,
	})
}

func runList(args []string) error {
	fs := newFlagSet("list")
	var lf loadFlags
	lf.register(fs)
	path, err := parse(fs, args)
	if err != nil {
		return err
	}
	r, err := lf.load(path)
	if err != nil {
		return err
	}

	b := r.Book
	fmt.Printf("%s — %s (language: %s)\n\n", orDash(b.Title), orDash(b.Author), b.Language.Code)
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "#\tSECTION\tWORDS\tTITLE\tNOTE")
	for i, c := range b.Chapters {
		fmt.Fprintf(tw, "%d\t%s\t%d\t%s\t%s\n", i+1, c.Section, c.Words, orDash(truncate(c.Heading(), 50)), c.Reason)
	}
	tw.Flush()

	fmt.Printf("\n%d lesson files would be written. Fix the classification with -body-from / -back-from if needed.\n", len(r.Lessons))
	return nil
}

func runPrepare(args []string) error {
	fs := newFlagSet("")
	var lf loadFlags
	lf.register(fs)
	out := fs.String("o", "", "output directory (default: the book's file name without extension)")
	force := fs.Bool("force", false, "write into a non-empty output directory, overwriting files with the same name")
	path, err := parse(fs, args)
	if err != nil {
		return err
	}
	r, err := lf.load(path)
	if err != nil {
		return err
	}

	dir := *out
	if dir == "" {
		dir = prep.DefaultDir(path)
	}
	if err := r.Write(dir, *force); err != nil {
		if errors.Is(err, lesson.ErrNotEmpty) {
			return fmt.Errorf("output directory %s is not empty (use -force to overwrite)", dir)
		}
		return err
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	for _, l := range r.Lessons {
		fmt.Fprintf(tw, "%d\t  %s\t\n", l.Words, l.FileName())
	}
	tw.Flush()
	fmt.Printf("\nWrote %d lessons to %s\n", len(r.Lessons), dir)
	for _, l := range r.Oversized() {
		fmt.Fprintf(os.Stderr, "warning: %s has %d words: it contains a single sentence longer than the limit\n", l.FileName(), l.Words)
	}
	return nil
}

// newFlagSet creates the flag set for a command ("" for the default one)
// whose -h output starts with the general usage text.
func newFlagSet(cmd string) *flag.FlagSet {
	fs := flag.NewFlagSet(strings.TrimSpace("lingqprep "+cmd), flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprint(fs.Output(), usage)
		fmt.Fprintf(fs.Output(), "\nFlags of %q:\n", fs.Name())
		fs.PrintDefaults()
	}
	return fs
}

// parse parses flags and expects exactly one file argument. Unlike plain
// flag.Parse it also accepts flags after the file name.
func parse(fs *flag.FlagSet, args []string) (string, error) {
	var files []string
	for {
		if err := fs.Parse(args); err != nil {
			return "", err
		}
		if fs.NArg() == 0 {
			break
		}
		files = append(files, fs.Arg(0))
		args = fs.Args()[1:]
	}
	if len(files) != 1 {
		return "", fmt.Errorf("expected one book file, got %d", len(files))
	}
	return files[0], nil
}

func orDash(s string) string {
	if s == "" {
		return "–"
	}
	return s
}

func truncate(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n-1]) + "…"
	}
	return s
}
