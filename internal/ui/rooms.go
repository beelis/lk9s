package ui

import (
	"cmp"
	"fmt"
	"time"

	"github.com/beelis/lk9s/internal/lk"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var roomCols = []column[lk.Room]{
	{
		header:  "NAME",
		key:     'N',
		display: func(r lk.Room) string { return r.Name },
		compare: func(a, b lk.Room) int { return cmp.Compare(a.Name, b.Name) },
	},
	{
		header:  "SID",
		key:     'S',
		display: func(r lk.Room) string { return r.SID },
		compare: func(a, b lk.Room) int { return cmp.Compare(a.SID, b.SID) },
	},
	{
		header:  "PARTICIPANTS",
		key:     'P',
		display: func(r lk.Room) string { return fmt.Sprintf("%d", r.NumParticipants) },
		compare: func(a, b lk.Room) int { return cmp.Compare(a.NumParticipants, b.NumParticipants) },
	},
	{
		header:  "CREATED",
		key:     'C',
		display: func(r lk.Room) string { return time.Unix(r.CreationTime, 0).Format(time.DateTime) },
		compare: func(a, b lk.Room) int { return cmp.Compare(a.CreationTime, b.CreationTime) },
	},
}

func roomsInputCapture(
	n nav,
	table *tview.Table,
	state *tableState[lk.Room],
	status *tview.TextView,
) func(*tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := table.GetSelection()

		if event.Rune() == 'e' {
			if row > 0 && row <= len(state.sorted) {
				roomName := state.sorted[row-1].Name

				go func() {
					eg, err := n.client.ListEgresses(n.ctx, roomName)

					n.app.QueueUpdateDraw(func() {
						updateStatus(status, err)

						if err != nil {
							return
						}

						n.pages.RemovePage("egresses")
						n.pages.AddPage("egresses", egressesPage(n, roomName, eg), true, true)
					})
				}()

				return nil
			}

			return event
		}

		if event.Rune() == 'm' {
			if row > 0 && row <= len(state.sorted) {
				r := state.sorted[row-1]

				n.pages.RemovePage("metadata")
				n.pages.AddPage("metadata", metadataPage(n, r.Name, r.Metadata), true, true)
			}

			return nil
		}

		if !state.handleKey(event.Rune()) {
			return event
		}

		state.render(table)

		return nil
	}
}

func roomsPage(n nav) tview.Primitive {
	header := tview.NewTextView().SetText(" ctx: " + n.contextName)
	table := newTable(" Rooms ")
	status := newStatusBar()
	state := &tableState[lk.Room]{cols: roomCols, sortAsc: true}

	status.SetText(" loading...")
	state.render(table)
	table.SetInputCapture(roomsInputCapture(n, table, state, status))

	table.SetSelectedFunc(func(row, _ int) {
		if row == 0 || row > len(state.sorted) {
			return
		}

		roomName := state.sorted[row-1].Name

		go func() {
			pp, err := n.client.ListParticipants(n.ctx, roomName)

			n.app.QueueUpdateDraw(func() {
				updateStatus(status, err)

				if err != nil {
					return
				}

				n.pages.RemovePage("participants")
				n.pages.AddPage("participants", participantsPage(n, roomName, pp), true, true)
			})
		}()
	})

	fetchAndRender := func() {
		fetched, err := n.client.ListRooms(n.ctx)

		n.app.QueueUpdateDraw(func() {
			updateStatus(status, err)

			if err != nil {
				return
			}

			state.setItems(fetched)
			state.render(table)
		})
	}

	go func() {
		fetchAndRender()

		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-n.ctx.Done():
				return
			case <-ticker.C:
				fetchAndRender()
			}
		}
	}()

	keys := [][2]string{{"Enter", "participants"}, {"e", "egresses"}, {"m", "metadata"}, {"Shift+letter", "sort"}}

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(table, 0, 1, true).
		AddItem(status, 1, 0, false).
		AddItem(legend(keys), 1, 0, false)
}
