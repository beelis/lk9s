package ui

import (
	"cmp"
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mariusgassen/lk9s/internal/lk"
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

// validRow reports whether row (1-based, below the header) addresses an item.
func validRow[T any](items []T, row int) bool {
	return row > 0 && row <= len(items)
}

//nolint:gocognit,funlen // key-binding dispatch
func roomsInputCapture(
	n nav,
	table *tview.Table,
	state *tableState[lk.Room],
	status *tview.TextView,
	refresh func(keepStatus error),
) func(*tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := table.GetSelection()

		if event.Rune() == '/' {
			n.pages.AddPage("filter", filterBarPage(n, state.filterText, func(f string) {
				state.setFilter(f)
				state.render(table)
			}), true, true)

			return nil
		}

		if event.Key() == tcell.KeyCtrlE {
			if text := status.GetText(true); text != "" {
				n.pages.RemovePage("status-detail")
				n.pages.AddPage("status-detail", statusDetailPage(n, text), true, true)
			}

			return nil
		}

		if event.Key() == tcell.KeyCtrlD {
			if validRow(state.sorted, row) && requireWrite(n, status) {
				roomName := state.sorted[row-1].Name

				n.pages.RemovePage("confirm-delete")
				n.pages.AddPage("confirm-delete", confirmDeleteRoomPage(n, roomName, func(err error) {
					updateStatus(status, err)

					go refresh(err)
				}), true, true)
			}

			return nil
		}

		if event.Rune() == 'e' {
			if validRow(state.sorted, row) {
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
			if validRow(state.sorted, row) {
				r := state.sorted[row-1]

				n.pages.RemovePage("metadata")
				n.pages.AddPage("metadata", metadataPage(n, r.Name, r.Metadata), true, true)
			}

			return nil
		}

		if event.Rune() == 'u' {
			if validRow(state.sorted, row) && requireWrite(n, status) {
				r := state.sorted[row-1]

				n.pages.RemovePage("update-metadata")
				n.pages.AddPage("update-metadata", formPage(n, "update-metadata", " room metadata: "+r.Name+" ",
					[]formField{{label: "metadata", initial: r.Metadata}},
					func(ctx context.Context, v map[string]string) error {
						return n.client.UpdateRoomMetadata(ctx, r.Name, v["metadata"])
					},
					func() { go refresh(nil) },
				), true, true)
			}

			return nil
		}

		if event.Rune() == 'd' {
			if validRow(state.sorted, row) && requireWrite(n, status) {
				roomName := state.sorted[row-1].Name

				n.pages.RemovePage("send-data")
				n.pages.AddPage("send-data", sendDataPage(n, roomName, status), true, true)
			}

			return nil
		}

		if event.Rune() == 'i' {
			if validRow(state.sorted, row) {
				r := state.sorted[row-1]

				n.pages.RemovePage("roominfo")
				n.pages.AddPage("roominfo", roomInfoPage(n, r), true, true)
			}

			return nil
		}

		if event.Rune() == 'g' {
			if validRow(state.sorted, row) {
				roomName := state.sorted[row-1].Name

				go func() {
					dd, err := n.client.ListAgentDispatches(n.ctx, roomName)

					n.app.QueueUpdateDraw(func() {
						updateStatus(status, err)

						if err != nil {
							return
						}

						n.pages.RemovePage("agents")
						n.pages.AddPage("agents", agentsPage(n, roomName, dd), true, true)
					})
				}()

				return nil
			}

			return event
		}

		if event.Rune() == 'T' {
			if validRow(state.sorted, row) {
				roomName := state.sorted[row-1].Name

				n.pages.RemovePage("token-form")
				n.pages.AddPage("token-form", tokenFormPage(n, "", roomName), true, true)
			}

			return nil
		}

		if event.Rune() == 's' {
			go func() {
				entries, err := n.client.ListSIP(n.ctx)

				n.app.QueueUpdateDraw(func() {
					updateStatus(status, err)

					if err != nil {
						return
					}

					n.pages.RemovePage("sip")
					n.pages.AddPage("sip", sipPage(n, entries), true, true)
				})
			}()

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
	table := newTable(" Rooms ")
	status := newStatusBar()
	state := &tableState[lk.Room]{title: "rooms", cols: roomCols, sortAsc: true}

	status.SetText(" loading...")
	state.render(table)

	// keepStatus, when non-nil, leaves the status bar as the caller set it
	// (e.g. an error from an action that just ran) instead of overwriting
	// it with this refresh's own result; the room list is always updated.
	fetchAndRender := func(keepStatus error) {
		fetched, err := n.client.ListRooms(n.ctx)

		n.app.QueueUpdateDraw(func() {
			if err == nil {
				state.setItems(fetched)
				state.render(table)
			}

			if keepStatus != nil {
				return
			}

			updateListStatus(status, err, len(fetched))
		})
	}

	table.SetInputCapture(roomsInputCapture(n, table, state, status, fetchAndRender))

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

	go func() {
		fetchAndRender(nil)

		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-n.ctx.Done():
				return
			case <-ticker.C:
				fetchAndRender(nil)
			}
		}
	}()

	keys := [][2]string{
		{"Enter", "participants"}, {"e", "egresses"}, {"m", "metadata"}, {"u", "update metadata"},
		{"d", "send data"}, {"i", "room info"},
		{"g", "agents"}, {"s", "sip"}, {"T", "token"}, {"Ctrl+D", "delete"}, {"Ctrl+E", "error detail"},
		{"Shift+letter", "sort"}, {"/", "filter"}, {":", "command"},
	}

	return withHints(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true).
		AddItem(status, 1, 0, false), keys)
}
