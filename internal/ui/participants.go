package ui

import (
	"cmp"
	"context"
	"time"

	"github.com/beelis/lk9s/internal/lk"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var participantCols = []column[lk.Participant]{
	{
		header:  "IDENTITY",
		key:     'I',
		display: func(p lk.Participant) string { return p.Identity },
		compare: func(a, b lk.Participant) int { return cmp.Compare(a.Identity, b.Identity) },
	},
	{
		header:  "NAME",
		key:     'N',
		display: func(p lk.Participant) string { return p.Name },
		compare: func(a, b lk.Participant) int { return cmp.Compare(a.Name, b.Name) },
	},
	{
		header:  "KIND",
		key:     'K',
		display: func(p lk.Participant) string { return p.Kind },
		compare: func(a, b lk.Participant) int { return cmp.Compare(a.Kind, b.Kind) },
	},
	{
		header:  "STATE",
		key:     'S',
		display: func(p lk.Participant) string { return p.State },
		compare: func(a, b lk.Participant) int { return cmp.Compare(a.State, b.State) },
	},
	{
		header:  "MIC",
		key:     'M',
		display: func(p lk.Participant) string { return p.Mic.String() },
		compare: func(a, b lk.Participant) int { return cmp.Compare(a.Mic, b.Mic) },
	},
	{
		header:  "CAM",
		key:     'C',
		display: func(p lk.Participant) string { return p.Camera.String() },
		compare: func(a, b lk.Participant) int { return cmp.Compare(a.Camera, b.Camera) },
	},
	{
		header:  "SCREEN",
		key:     'R',
		display: func(p lk.Participant) string { return p.Screen.String() },
		compare: func(a, b lk.Participant) int { return cmp.Compare(a.Screen, b.Screen) },
	},
	{
		header:  "SCR.AUD",
		key:     'A',
		display: func(p lk.Participant) string { return p.ScreenAudio.String() },
		compare: func(a, b lk.Participant) int { return cmp.Compare(a.ScreenAudio, b.ScreenAudio) },
	},
	{
		header:  "JOINED",
		key:     'J',
		display: func(p lk.Participant) string { return time.Unix(p.JoinedAt, 0).Format(time.DateTime) },
		compare: func(a, b lk.Participant) int { return cmp.Compare(a.JoinedAt, b.JoinedAt) },
	},
}

//nolint:gocognit,funlen
func participantsPage(n nav, roomName string, initial []lk.Participant) tview.Primitive {
	header := tview.NewTextView().SetText(" ctx: " + n.contextName + " > " + roomName)
	table := newTable(" Participants ")
	status := newStatusBar()
	state := &tableState[lk.Participant]{cols: participantCols, sortAsc: true}
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
				p := state.sorted[row-1]

				n.pages.RemovePage("metadata")
				n.pages.AddPage("metadata", metadataPage(n, p.Identity, p.Metadata), true, true)
			}

			return nil
		}

		if event.Rune() == 'a' {
			row, _ := table.GetSelection()
			if row > 0 && row <= len(state.sorted) {
				p := state.sorted[row-1]

				n.pages.RemovePage("attributes")
				n.pages.AddPage("attributes", attributesPage(n, p.Identity, p.Attributes), true, true)
			}

			return nil
		}

		if event.Rune() == 'p' {
			row, _ := table.GetSelection()
			if row > 0 && row <= len(state.sorted) {
				p := state.sorted[row-1]

				n.pages.RemovePage("permissions")
				n.pages.AddPage("permissions", permissionsPage(n, p.Identity, p.Permission), true, true)
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
				fetched, err := n.client.ListParticipants(ctx, roomName)

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

	keys := [][2]string{
		{"Esc", "back"}, {"m", "metadata"},
		{"a", "attributes"},
		{"p", "permissions"},
		{"Shift+letter", "sort"},
	}

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(table, 0, 1, true).
		AddItem(status, 1, 0, false).
		AddItem(legend(keys), 1, 0, false)
}
