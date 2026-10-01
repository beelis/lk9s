package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// commands are the recognized ":" command names, offered as autocomplete
// entries in the order shown.
var commands = []string{"create-room", "projects", "ctx", "rooms", "sip", "token", "quit"}

// newCommandBar builds the k9s-style ":" command line that stays docked at
// the top of the screen for the lifetime of the app (see Run). Pressing ":"
// anywhere focuses it; blur is called to return focus to the underlying
// view, on both cancel and submit. onExec is called with the trimmed,
// lower-cased command once the user submits non-empty text.
func newCommandBar(blur func(), onExec func(cmd string)) *tview.InputField {
	input := tview.NewInputField().SetLabel(" : ")

	input.SetAutocompleteFunc(func(currentText string) []string {
		currentText = strings.ToLower(strings.TrimSpace(currentText))
		if currentText == "" {
			return nil
		}

		var matches []string

		for _, c := range commands {
			if strings.HasPrefix(c, currentText) {
				matches = append(matches, c)
			}
		}

		return matches
	})

	submit := func(text string) {
		input.SetText("")
		blur()

		text = strings.ToLower(strings.TrimSpace(text))
		if text == "" {
			return
		}

		onExec(text)
	}

	// Enter on a highlighted autocomplete entry both fills and submits;
	// arrow-key navigation only fills, keeping the list open.
	input.SetAutocompletedFunc(func(text string, index int, source int) bool {
		input.SetText(text)

		if source == tview.AutocompletedNavigate {
			return false
		}

		if source == tview.AutocompletedEnter {
			submit(text)
		}

		return true
	})

	input.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEscape:
			input.SetText("")
			blur()
		case tcell.KeyEnter:
			submit(input.GetText())
		}
	})

	return input
}
