package ui

import (
	"cmp"
	"context"
	"fmt"
	"sync"
	"time"
	"unicode"

	"github.com/beelis/lk9s/internal/lk"
	"github.com/gdamore/tcell/v2"
	"github.com/livekit/protocol/livekit"
	"github.com/rivo/tview"
)

type roomRow struct {
	lk.Room
	NumSIPUsers     uint32
	NumAgents       uint32
	NumParticipants uint32
	NumEgress       uint32
}

var roomColsBasic = []column[roomRow]{
	{
		header:  "NAME",
		key:     'N',
		display: func(r roomRow) string { return r.Name },
		compare: func(a, b roomRow) int { return cmp.Compare(a.Name, b.Name) },
	},
	{
		header:  "SID",
		key:     'S',
		display: func(r roomRow) string { return r.SID },
		compare: func(a, b roomRow) int { return cmp.Compare(a.SID, b.SID) },
	},
	{
		header:  "PARTICIPANTS",
		key:     'P',
		display: func(r roomRow) string { return fmt.Sprintf("%d", r.NumParticipants) },
		compare: func(a, b roomRow) int { return cmp.Compare(a.NumParticipants, b.NumParticipants) },
	},
	{
		header:  "CREATED",
		key:     'C',
		display: func(r roomRow) string { return time.Unix(r.CreationTime, 0).Format(time.DateTime) },
		compare: func(a, b roomRow) int { return cmp.Compare(a.CreationTime, b.CreationTime) },
	},
}

var roomColsExtended = []column[roomRow]{
	{
		header:  "NAME",
		key:     'N',
		display: func(r roomRow) string { return r.Name },
		compare: func(a, b roomRow) int { return cmp.Compare(a.Name, b.Name) },
	},
	{
		header:  "SID",
		key:     'S',
		display: func(r roomRow) string { return r.SID },
		compare: func(a, b roomRow) int { return cmp.Compare(a.SID, b.SID) },
	},
	{
		header:  "SIP",
		key:     'I',
		display: func(r roomRow) string { return fmt.Sprintf("%d", r.NumSIPUsers) },
		compare: func(a, b roomRow) int { return cmp.Compare(a.NumSIPUsers, b.NumSIPUsers) },
	},
	{
		header:  "AGENT",
		key:     'A',
		display: func(r roomRow) string { return fmt.Sprintf("%d", r.NumAgents) },
		compare: func(a, b roomRow) int { return cmp.Compare(a.NumAgents, b.NumAgents) },
	},
	{
		header:  "PARTICIPANT",
		key:     'P',
		display: func(r roomRow) string { return fmt.Sprintf("%d", r.NumParticipants) },
		compare: func(a, b roomRow) int { return cmp.Compare(a.NumParticipants, b.NumParticipants) },
	},
	{
		header:  "EGRESS",
		key:     'G',
		display: func(r roomRow) string { return fmt.Sprintf("%d", r.NumEgress) },
		compare: func(a, b roomRow) int { return cmp.Compare(a.NumEgress, b.NumEgress) },
	},
}

func roomRows(rooms []lk.Room) []roomRow {
	rows := make([]roomRow, len(rooms))
	for i, r := range rooms {
		rows[i] = roomRow{Room: r}
	}

	return rows
}

func countParticipantTypes(participants []lk.Participant) (sip, agent, participant, egress uint32) {
	for _, p := range participants {
		switch p.Kind {
		case livekit.ParticipantInfo_SIP:
			sip++
		case livekit.ParticipantInfo_AGENT:
			agent++
		case livekit.ParticipantInfo_STANDARD:
			participant++
		case livekit.ParticipantInfo_EGRESS:
			egress++
		}
	}

	return sip, agent, participant, egress
}

func populateRoomCounts(ctx context.Context, client roomLister, rows []roomRow) ([]roomRow, error) {
	var wg sync.WaitGroup
	errCh := make(chan error, len(rows))
	sem := make(chan struct{}, 4)

	for i := range rows {
		wg.Add(1)
		sem <- struct{}{}

		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()

			participants, err := client.ListParticipants(ctx, rows[i].Name)
			if err != nil {
				errCh <- err
				return
			}

			rows[i].NumSIPUsers, rows[i].NumAgents, rows[i].NumParticipants, rows[i].NumEgress = countParticipantTypes(participants)
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		return rows, err
	}

	return rows, nil
}

func roomsInputCapture(
	n nav,
	table *tview.Table,
	state *tableState[roomRow],
	status *tview.TextView,
	header *tview.TextView,
	isExtended *bool,
	refresh func(),
) func(*tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := table.GetSelection()
		r := unicode.ToLower(event.Rune())

		if r == 'e' {
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

		if r == 'm' {
			if row > 0 && row <= len(state.sorted) {
				r := state.sorted[row-1]

				n.pages.RemovePage("metadata")
				n.pages.AddPage("metadata", metadataPage(n, r.Name, r.Metadata), true, true)
			}

			return nil
		}

		if r == 'v' {
			*isExtended = !*isExtended
			if *isExtended {
				state.cols = roomColsExtended
				header.SetText(fmt.Sprintf(" %-10s %s | mode: extended", "ctx:", n.contextName))
			} else {
				state.cols = roomColsBasic
				header.SetText(fmt.Sprintf(" %-10s %s | mode: basic", "ctx:", n.contextName))
			}
			state.sortCol = 0
			state.sortAsc = true
			state.resort = true
			refresh()

			return nil
		}

		if !state.handleKey(r) {
			return event
		}

		state.render(table)

		return nil
	}
}

func roomsPage(n nav) tview.Primitive {
	version := tview.NewTextView().SetText(fmt.Sprintf(" %-10s %s", "LK9s Rev:", n.version))
	extended := false
	header := tview.NewTextView().SetText(fmt.Sprintf(" %-10s %s | mode: basic", "ctx:", n.contextName))
	table := newTable(" Rooms ")
	status := newStatusBar()
	state := &tableState[roomRow]{cols: roomColsBasic, sortAsc: true}

	status.SetText(" loading...")
	state.render(table)

	var fetchAndRender func()
	table.SetInputCapture(roomsInputCapture(n, table, state, status, header, &extended, func() {
		go fetchAndRender()
	}))

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

	fetchAndRender = func() {
		fetched, err := n.client.ListRooms(n.ctx)

		n.app.QueueUpdateDraw(func() {
			updateStatus(status, err)

			if err != nil {
				return
			}

			rows := roomRows(fetched)
			if extended {
				rows, err = populateRoomCounts(n.ctx, n.client, rows)
				if err != nil {
					updateStatus(status, err)
				}
			}

			state.setItems(rows)
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

	keys := [][2]string{{"Enter", "participants"}, {"e", "egresses"}, {"m", "metadata"}, {"v", "toggle view"}, {"Shift+letter", "sort"}}

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(version, 1, 0, false).
		AddItem(header, 1, 0, false).
		AddItem(table, 0, 1, true).
		AddItem(status, 1, 0, false).
		AddItem(legend(keys), 1, 0, false)
}
