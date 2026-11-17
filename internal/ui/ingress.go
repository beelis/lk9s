package ui

import (
	"cmp"
	"context"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mariusgassen/lk9s/internal/lk"
	"github.com/rivo/tview"
)

var ingressCols = []column[lk.Ingress]{
	{
		header:  "ID",
		key:     'I',
		display: func(i lk.Ingress) string { return i.ID },
		compare: func(a, b lk.Ingress) int { return cmp.Compare(a.ID, b.ID) },
	},
	{
		header:  "NAME",
		key:     'N',
		display: func(i lk.Ingress) string { return i.Name },
		compare: func(a, b lk.Ingress) int { return cmp.Compare(a.Name, b.Name) },
	},
	{
		header:  "INPUT",
		key:     'T',
		display: func(i lk.Ingress) string { return i.Input },
		compare: func(a, b lk.Ingress) int { return cmp.Compare(a.Input, b.Input) },
	},
	{
		header:  "STATUS",
		key:     'S',
		display: func(i lk.Ingress) string { return i.Status },
		compare: func(a, b lk.Ingress) int { return cmp.Compare(a.Status, b.Status) },
	},
	{
		header:  "ROOM",
		key:     'R',
		display: func(i lk.Ingress) string { return i.Room },
		compare: func(a, b lk.Ingress) int { return cmp.Compare(a.Room, b.Room) },
	},
	{
		header:  "IDENTITY",
		key:     'D',
		display: func(i lk.Ingress) string { return i.Identity },
		compare: func(a, b lk.Ingress) int { return cmp.Compare(a.Identity, b.Identity) },
	},
	{
		header:  "ERROR",
		key:     'E',
		display: func(i lk.Ingress) string { return i.Error },
		compare: func(a, b lk.Ingress) int { return cmp.Compare(a.Error, b.Error) },
	},
}

func ingressPage(n nav, initial []lk.Ingress) tview.Primitive {
	header := tview.NewTextView().SetText(" ctx: " + n.contextName + " > ingress")
	table := newTable(" Ingress ")
	status := newStatusBar()
	state := &tableState[lk.Ingress]{title: "ingress", cols: ingressCols, sortAsc: true}
	state.setItems(initial)

	state.render(table)

	ctx, cancel := context.WithCancel(n.ctx)

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			cancel()
			n.pages.RemovePage("ingress")

			return nil
		}

		if !state.handleKey(event.Rune()) {
			return event
		}

		state.render(table)

		return nil
	})

	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fetched, err := n.client.ListIngress(ctx, "")

				n.app.QueueUpdateDraw(func() {
					updateStatus(status, err)

					if err != nil {
						return
					}

					state.setItems(fetched)
					state.render(table)
				})
			}
		}
	}()

	keys := [][2]string{{"Esc", "back"}, {"Shift+letter", "sort"}, {":", "command"}}

	return withHints(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(table, 0, 1, true).
		AddItem(status, 1, 0, false), keys)
}

// showIngress fetches all ingresses and opens the ingress page, or an error
// page when the request fails.
func showIngress(n nav, pages *tview.Pages, app *tview.Application) {
	go func() {
		items, err := n.client.ListIngress(n.ctx, "")

		app.QueueUpdateDraw(func() {
			if err != nil {
				pages.RemovePage("status-detail")
				pages.AddPage("status-detail", statusDetailPage(n, "error: "+err.Error()), true, true)

				return
			}

			pages.RemovePage("ingress")
			pages.AddPage("ingress", ingressPage(n, items), true, true)
		})
	}()
}
