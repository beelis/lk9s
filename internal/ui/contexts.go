package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mariusgassen/lk9s/internal/config"
	"github.com/rivo/tview"
)

// contextsPage lets the user pick a different project/context while the app
// is running (invoked via the ":projects" command). onSelect is called with
// the chosen context; onCancel is called on Esc.
func contextsPage(n nav, contexts []config.Context, onSelect func(config.Context), onCancel func()) tview.Primitive {
	const pageName = "contexts"

	table := newTable(" Select Project ")
	// tview loops forever when a selectable table has no selectable cell.
	table.SetSelectable(len(contexts) > 0, false)

	headers := []string{"NAME", "URL"}
	for col, h := range headers {
		table.SetCell(0, col, headerCell(h))
	}

	for row, c := range contexts {
		table.SetCell(row+1, 0, tview.NewTableCell(c.Name).SetExpansion(1))
		table.SetCell(row+1, 1, tview.NewTableCell(c.URL).SetExpansion(1))
	}

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			n.pages.RemovePage(pageName)
			onCancel()

			return nil
		}

		return event
	})

	table.SetSelectedFunc(func(row, _ int) {
		if row == 0 || row > len(contexts) {
			return
		}

		n.pages.RemovePage(pageName)
		onSelect(contexts[row-1])
	})

	keys := [][2]string{{"Enter", "select"}, {"Esc", "cancel"}}

	return withHints(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true), keys)
}
