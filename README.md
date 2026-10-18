# lingqprep

Prepares ebooks for import into [LingQ](https://www.lingq.com):

- reads EPUB files directly (EPUB 2 and 3), no conversion to text needed
- glues sentences back together that a bad conversion cut over several lines
- writes one or more lessons per chapter, each at most 6000 words, always
  ending on a full sentence and never crossing a chapter boundary
- long chapters are split evenly (6500 words → ~3250 + ~3250, not 6000 + 500)
- collects front matter (title page, copyright, ...) in `000 - Front matter.txt`
  and everything after the story (about the author, previews, ...) in a final
  `NNN - Back matter.txt`

Currently supports French.

## Building

The GUI uses [Fyne](https://fyne.io), which needs cgo, so building needs a C
compiler and the graphics headers. On Debian/Ubuntu:

```sh
sudo apt install gcc libgl1-mesa-dev xorg-dev libwayland-dev libxkbcommon-dev wayland-protocols
go build -o lingqprep ./cmd/lingqprep
```

The first build takes a few minutes; later builds are cached. For a
command-line-only binary that needs no C compiler, build with
`CGO_ENABLED=0 go build -tags nogui -o lingqprep ./cmd/lingqprep`.

## Packaging and installing (Linux)

The installable app is made with Fyne's packaging tool. Its settings (name,
app ID, version, icon) are in `cmd/lingqprep/FyneApp.toml`; `Icon.png` next
to it is a placeholder until there is a real logo. `fyne package` rewrites
that file on every run (it raises `Build` by one), so don't keep comments in
it. Its `Version` is only used for local packages: releases take the version
from the git tag.

```sh
go install fyne.io/tools/cmd/fyne@latest
cd cmd/lingqprep
fyne package --os linux --release   # writes "LingQ Prep.tar.xz"
```

To install, extract the archive and run, inside the `lingqprep` folder:

```sh
make user-install      # for yourself, in ~/.local (no sudo needed)
sudo make install      # or for all users, in /usr/local
```

That adds "LingQ Prep" to the app menu and puts the `lingqprep` command on
the PATH, so the command line works too. `make user-uninstall` or
`sudo make uninstall` removes it again.

The binary only runs on systems with the same or a newer C library (glibc)
as the one it was built on, so build releases on an older distribution (or
with [fyne-cross](https://github.com/fyne-io/fyne-cross)) to support more
systems.

## Usage

Started from the app menu or by double-clicking, `lingqprep` opens its
window. In a terminal:

```sh
# Open the graphical interface (optionally with a book)
./lingqprep gui
./lingqprep gui book.epub

# See how the book was split into front matter, chapters and back matter
./lingqprep list book.epub

# Write the lessons (to ./book/ in the current directory by default)
./lingqprep book.epub
./lingqprep -o lessons/ book.epub
```

If the classification is wrong, look up the chapter numbers with `list` and
correct them:

```sh
./lingqprep -body-from 4 -back-from 31 book.epub
```

`lingqprep -version` shows the version: the git tag the binary was built
from, such as `v1.2.0`, or a pseudo-version for an untagged commit.

Other flags: `-lang fr`, `-max-words 6000`, `-force` (write into a
non-empty directory). Run `lingqprep -h` for the full list.

## How it works

| Package            | Job                                                                 |
|--------------------|---------------------------------------------------------------------|
| `internal/epub`    | Reads the zip, OPF, table of contents and XHTML into raw text blocks |
| `internal/source`  | Format-independent book/section types every reader produces          |
| `internal/prose`   | Repairs broken paragraphs, splits sentences, counts words            |
| `internal/lang`    | Language rules: abbreviations, quotes, front/back matter titles      |
| `internal/book`    | Ties it together: load, repair, classify front/body/back             |
| `internal/lesson`  | Splits chapters into lessons and writes the files                    |
| `internal/prep`    | The API the front ends share: options in, lessons out, write to disk |
| `internal/gui`     | The graphical interface (Fyne)                                       |
| `cmd/lingqprep`    | The command-line interface; `lingqprep gui` starts the GUI           |

**Repairing sentences.** Each paragraph is glued to the previous one unless the
previous one clearly ended a sentence (`.`, `!`, `?`, `…`, closing quote), the
next one starts a line of dialogue (`—`), or one of them is a heading or scene
break. A paragraph starting with a lowercase letter is always a continuation.

**Word counts** follow LingQ (`l'homme` is one word, `par-dessus` two) but also
count numbers, so they come out a few words above LingQ's count, never below.

## Extending

- **New language:** add `internal/lang/<code>.go` modelled on `fr.go`.
- **Plain-text input:** add a reader that returns a `source.Book` and hook it
  up in `book.read`.
- **GUI:** the CLI only calls `prep.Load` and `Result.Write`; a GUI calls the
  same functions, so both front ends behave the same.
