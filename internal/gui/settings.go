//go:build !nogui

package gui

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"lingqprep/internal/lang"
	"lingqprep/internal/lesson"
	"lingqprep/internal/prep"
)

// autoLanguage is the language choice that uses the file's own language.
const autoLanguage = "From the book"

// languageChoices returns the options of the language selector, e.g.
// "From the book", "Français (fr)", and a map back to language codes.
func languageChoices() ([]string, map[string]string) {
	choices := []string{autoLanguage}
	codes := map[string]string{autoLanguage: ""}
	for _, code := range lang.Codes() {
		l, _ := lang.Get(code)
		choice := fmt.Sprintf("%s (%s)", l.Name, code)
		choices = append(choices, choice)
		codes[choice] = code
	}
	return choices, codes
}

// parseMaxWords validates the max words field.
func parseMaxWords(s string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < 1 {
		return 0, errors.New("enter a whole number above 0")
	}
	return n, nil
}

// newSettings builds the panel below the tables: the options and the write
// button.
func (u *ui) newSettings() fyne.CanvasObject {
	choices, codes := languageChoices()
	u.language = widget.NewSelect(choices, func(choice string) {
		u.opts.Language = codes[choice]
		if u.path != "" {
			u.open(u.path) // the language changes how the text is repaired
		}
	})
	u.language.SetSelected(autoLanguage)

	u.maxWords = widget.NewEntry()
	u.maxWords.SetText(strconv.Itoa(lesson.DefaultMaxWords))
	u.maxWords.Validator = func(s string) error {
		_, err := parseMaxWords(s)
		return err
	}
	u.maxWords.OnChanged = func(s string) {
		n, err := parseMaxWords(s)
		if err != nil {
			u.write.Disable()
			return
		}
		u.opts.MaxWords = n
		if u.result != nil {
			u.result.SetMaxWords(n)
			u.show()
		}
	}

	u.outDir = widget.NewEntry()
	u.outDir.SetPlaceHolder("Folder for the lesson files")
	u.outDir.OnChanged = func(string) { u.updateWrite() }
	choose := widget.NewButtonWithIcon("Choose…", theme.FolderIcon(), u.chooseOutDir)

	u.write = widget.NewButtonWithIcon("Write lessons", theme.DocumentSaveIcon(), u.writeLessons)
	u.write.Importance = widget.HighImportance
	u.write.Disable()

	form := widget.NewForm(
		widget.NewFormItem("Language", u.language),
		widget.NewFormItem("Max words", u.maxWords),
		widget.NewFormItem("Save to", container.NewBorder(nil, nil, nil, choose, u.outDir)),
	)
	return container.NewVBox(form, container.NewHBox(layout.NewSpacer(), u.write))
}

// updateWrite enables the write button when there is something to write.
func (u *ui) updateWrite() {
	_, err := parseMaxWords(u.maxWords.Text)
	if u.result != nil && err == nil && strings.TrimSpace(u.outDir.Text) != "" {
		u.write.Enable()
	} else {
		u.write.Disable()
	}
}

// chooseOutDir lets the user pick the folder in which a folder named after
// the book is created.
func (u *ui) chooseOutDir() {
	d := dialog.NewFolderOpen(func(dir fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, u.win)
			return
		}
		if dir == nil {
			return // cancelled
		}
		u.outDir.SetText(filepath.Join(dir.Path(), prep.DefaultDir(u.path)))
	}, u.win)
	if dir, err := storage.ListerForURI(storage.NewFileURI(filepath.Dir(u.outDir.Text))); err == nil {
		d.SetLocation(dir)
	}
	d.Resize(fyne.NewSize(800, 550))
	d.Show()
}

// writeLessons writes the files, asking first if the folder isn't empty.
func (u *ui) writeLessons() {
	dir := strings.TrimSpace(u.outDir.Text)
	err := u.result.Write(dir, false)
	if errors.Is(err, lesson.ErrNotEmpty) {
		msg := fmt.Sprintf("%s already contains files.\nLesson files with the same names will be overwritten.", dir)
		dialog.ShowConfirm("Folder not empty", msg, func(ok bool) {
			if ok {
				u.wrote(dir, u.result.Write(dir, true))
			}
		}, u.win)
		return
	}
	u.wrote(dir, err)
}

// wrote reports the result of writing the lessons.
func (u *ui) wrote(dir string, err error) {
	if err != nil {
		dialog.ShowError(err, u.win)
		return
	}
	msg := fmt.Sprintf("Wrote %d lessons to\n%s", len(u.result.Lessons), dir)
	if over := u.result.Oversized(); len(over) > 0 {
		msg += fmt.Sprintf("\n\n%d lessons are over the limit because a single sentence is longer than %d words.",
			len(over), u.result.MaxWords)
	}
	dialog.ShowCustomConfirm("Lessons written", "Open folder", "Close", widget.NewLabel(msg), func(open bool) {
		if open {
			u.openFolder(dir)
		}
	}, u.win)
}

// openFolder shows dir in the system's file manager.
func (u *ui) openFolder(dir string) {
	abs, err := filepath.Abs(dir)
	if err == nil {
		var link *url.URL
		if link, err = url.Parse(storage.NewFileURI(abs).String()); err == nil {
			err = fyne.CurrentApp().OpenURL(link)
		}
	}
	if err != nil {
		dialog.ShowError(err, u.win)
	}
}
