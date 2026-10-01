package ui

import (
	"errors"

	"github.com/beelis/lk9s/internal/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ErrSelectionCancelled is returned by SelectContext when the user backs out
// via Esc instead of picking a context.
var ErrSelectionCancelled = errors.New("selection cancelled")

// SelectContext shows an interactive table and returns the chosen context.
// Esc cancels the selection (returning ErrSelectionCancelled) instead of
// leaving the user stuck with no way out short of Ctrl+C.
func SelectContext(contexts []config.Context) (config.Context, error) {
	app := tview.NewApplication()
	table := tview.NewTable().SetBorders(false).SetSelectable(true, false)
	table.SetTitle(" Select Context ").SetBorder(true)

	cancelled := false

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			cancelled = true

			app.Stop()

			return nil
		}

		return event
	})

	headers := []string{"NAME", "URL"}
	for col, h := range headers {
		table.SetCell(0, col, tview.NewTableCell(h).SetSelectable(false).SetExpansion(1))
	}

	for row, ctx := range contexts {
		table.SetCell(row+1, 0, tview.NewTableCell(ctx.Name).SetExpansion(1))
		table.SetCell(row+1, 1, tview.NewTableCell(ctx.URL).SetExpansion(1))
	}

	var selected config.Context

	table.SetSelectedFunc(func(row, _ int) {
		if row > 0 {
			selected = contexts[row-1]

			app.Stop()
		}
	})

	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true).
		AddItem(legend([][2]string{{"Enter", "select"}, {"Esc", "cancel"}}), 1, 0, false)

	if err := app.SetRoot(root, true).Run(); err != nil {
		return config.Context{}, err
	}

	if cancelled {
		return config.Context{}, ErrSelectionCancelled
	}

	if selected.Name == "" {
		return config.Context{}, errors.New("no context selected")
	}

	return selected, nil
}
