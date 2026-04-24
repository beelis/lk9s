package ui

import (
	"cmp"
	"context"
	"time"

	"github.com/beelis/lk9s/internal/lk"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var egressCols = []column[lk.Egress]{
	{
		header:  "ID",
		key:     'I',
		display: func(e lk.Egress) string { return e.ID },
		compare: func(a, b lk.Egress) int { return cmp.Compare(a.ID, b.ID) },
	},
	{
		header:  "STATUS",
		key:     'S',
		display: func(e lk.Egress) string { return e.Status },
		compare: func(a, b lk.Egress) int { return cmp.Compare(a.Status, b.Status) },
	},
	{
		header:  "TYPE",
		key:     'T',
		display: func(e lk.Egress) string { return e.Type },
		compare: func(a, b lk.Egress) int { return cmp.Compare(a.Type, b.Type) },
	},
	{
		header: "STARTED",
		key:    'A',
		display: func(e lk.Egress) string {
			if e.StartedAt == 0 {
				return "-"
			}

			return time.Unix(0, e.StartedAt).Format(time.DateTime)
		},
		compare: func(a, b lk.Egress) int { return cmp.Compare(a.StartedAt, b.StartedAt) },
	},
	{
		header:  "ERROR",
		key:     'E',
		display: func(e lk.Egress) string { return e.Error },
		compare: func(a, b lk.Egress) int { return cmp.Compare(a.Error, b.Error) },
	},
}

func egressesPage(n nav, roomName string, initial []lk.Egress) tview.Primitive {
	header := tview.NewTextView().SetText(" ctx: " + n.contextName + " > " + roomName + " > egresses")
	table := newTable(" Egresses ")
	status := newStatusBar()
	state := &tableState[lk.Egress]{cols: egressCols, sortAsc: true}
	state.setItems(initial)

	state.render(table)

	ctx, cancel := context.WithCancel(n.ctx)

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			cancel()
			n.pages.SwitchToPage("rooms")

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
				fetched, err := n.client.ListEgresses(ctx, roomName)

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

	keys := [][2]string{{"Esc", "back"}, {"Shift+letter", "sort"}}

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(table, 0, 1, true).
		AddItem(status, 1, 0, false).
		AddItem(legend(keys), 1, 0, false)
}
