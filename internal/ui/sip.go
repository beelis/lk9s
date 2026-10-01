package ui

import (
	"cmp"
	"context"
	"strings"
	"time"

	"github.com/beelis/lk9s/internal/lk"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var sipCols = []column[lk.SIPEntry]{
	{
		header:  "KIND",
		key:     'K',
		display: func(e lk.SIPEntry) string { return string(e.Kind) },
		compare: func(a, b lk.SIPEntry) int { return cmp.Compare(a.Kind, b.Kind) },
	},
	{
		header:  "ID",
		key:     'I',
		display: func(e lk.SIPEntry) string { return e.ID },
		compare: func(a, b lk.SIPEntry) int { return cmp.Compare(a.ID, b.ID) },
	},
	{
		header:  "NAME",
		key:     'N',
		display: func(e lk.SIPEntry) string { return e.Name },
		compare: func(a, b lk.SIPEntry) int { return cmp.Compare(a.Name, b.Name) },
	},
	{
		header:  "NUMBERS",
		key:     'U',
		display: func(e lk.SIPEntry) string { return strings.Join(e.Numbers, ", ") },
		compare: func(a, b lk.SIPEntry) int {
			return cmp.Compare(strings.Join(a.Numbers, ","), strings.Join(b.Numbers, ","))
		},
	},
	{
		header:  "ADDRESS",
		key:     'A',
		display: func(e lk.SIPEntry) string { return e.Address },
		compare: func(a, b lk.SIPEntry) int { return cmp.Compare(a.Address, b.Address) },
	},
}

func sipPage(n nav, initial []lk.SIPEntry) tview.Primitive {
	header := tview.NewTextView().SetText(" ctx: " + n.contextName + " > sip")
	table := newTable(" SIP ")
	status := newStatusBar()
	state := &tableState[lk.SIPEntry]{cols: sipCols, sortAsc: true}
	state.setItems(initial)

	state.render(table)

	ctx, cancel := context.WithCancel(n.ctx)

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			cancel()
			n.pages.SwitchToPage("rooms")

			return nil
		}

		if event.Rune() == 'm' {
			row, _ := table.GetSelection()
			if row > 0 && row <= len(state.sorted) {
				e := state.sorted[row-1]

				n.pages.RemovePage("metadata")
				n.pages.AddPage("metadata", metadataPage(n, e.Name, e.Metadata), true, true)
			}

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
				fetched, err := n.client.ListSIP(ctx)

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

	keys := [][2]string{{"Esc", "back"}, {"m", "metadata"}, {"Shift+letter", "sort"}}

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(table, 0, 1, true).
		AddItem(status, 1, 0, false).
		AddItem(legend(keys), 1, 0, false)
}
