package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// statusDetailPage shows the full text of a status bar line that may be
// cut off by the terminal width, most usefully a red error message.
func statusDetailPage(n nav, content string) tview.Primitive {
	body := tview.NewTextView().SetText(content)
	body.SetBorder(true).SetTitle(" status ")

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			n.pages.RemovePage("status-detail")

			return nil
		}

		return event
	})

	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(body, 0, 3, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)
}
