package ui

import (
	"github.com/beelis/lk9s/internal/config"
	"github.com/rivo/tview"
)

// SelectContext shows an interactive table and returns the chosen context.
func SelectContext(contexts []config.Context) (config.Context, error) {
	app := tview.NewApplication()
	table := tview.NewTable().SetBorders(false).SetSelectable(true, false)
	table.SetTitle(" Select Context ").SetBorder(true)

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

	if err := app.SetRoot(table, true).Run(); err != nil {
		return config.Context{}, err
	}

	return selected, nil
}
