package ui

import "github.com/rivo/tview"

func inputText(form *tview.Form, label string) string {
	if in, ok := form.GetFormItemByLabel(label).(*tview.InputField); ok {
		return in.GetText()
	}

	return ""
}

func isChecked(form *tview.Form, label string) bool {
	if cb, ok := form.GetFormItemByLabel(label).(*tview.Checkbox); ok {
		return cb.IsChecked()
	}

	return false
}
