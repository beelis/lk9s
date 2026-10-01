package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// filterBarPage shows a k9s-style "/" filter line, mirroring commandBarPage.
// onChange fires on every keystroke so the table behind it updates live.
// Enter keeps the current text and closes the bar; Esc clears the filter
// (calling onChange("")) and closes it.
func filterBarPage(n nav, initial string, onChange func(text string)) tview.Primitive {
	const pageName = "filter"

	input := tview.NewInputField().SetLabel(" / ").SetText(initial)

	input.SetChangedFunc(onChange)

	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			onChange("")
		}

		n.pages.RemovePage(pageName)
	})

	body := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(input, 1, 0, true)
	body.SetBorder(true)

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(body, 3, 0, true).
		AddItem(nil, 0, 1, false)
}
