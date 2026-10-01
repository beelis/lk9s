package ui

import (
	"cmp"
	"context"
	"time"

	"github.com/beelis/lk9s/internal/lk"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var agentDispatchCols = []column[lk.AgentDispatch]{
	{
		header:  "ID",
		key:     'I',
		display: func(d lk.AgentDispatch) string { return d.ID },
		compare: func(a, b lk.AgentDispatch) int { return cmp.Compare(a.ID, b.ID) },
	},
	{
		header:  "AGENT",
		key:     'A',
		display: func(d lk.AgentDispatch) string { return d.AgentName },
		compare: func(a, b lk.AgentDispatch) int { return cmp.Compare(a.AgentName, b.AgentName) },
	},
	{
		header:  "JOB STATUS",
		key:     'J',
		display: func(d lk.AgentDispatch) string { return d.JobStatus },
		compare: func(a, b lk.AgentDispatch) int { return cmp.Compare(a.JobStatus, b.JobStatus) },
	},
	{
		header:  "ERROR",
		key:     'E',
		display: func(d lk.AgentDispatch) string { return d.JobError },
		compare: func(a, b lk.AgentDispatch) int { return cmp.Compare(a.JobError, b.JobError) },
	},
	{
		header: "CREATED",
		key:    'C',
		display: func(d lk.AgentDispatch) string {
			if d.CreatedAt == 0 {
				return "-"
			}

			return time.Unix(d.CreatedAt, 0).Format(time.DateTime)
		},
		compare: func(a, b lk.AgentDispatch) int { return cmp.Compare(a.CreatedAt, b.CreatedAt) },
	},
	{
		header: "DELETED",
		key:    'D',
		display: func(d lk.AgentDispatch) string {
			if d.DeletedAt == 0 {
				return "-"
			}

			return time.Unix(d.DeletedAt, 0).Format(time.DateTime)
		},
		compare: func(a, b lk.AgentDispatch) int { return cmp.Compare(a.DeletedAt, b.DeletedAt) },
	},
}

func agentsPage(n nav, roomName string, initial []lk.AgentDispatch) tview.Primitive {
	header := tview.NewTextView().SetText(" ctx: " + n.contextName + " > " + roomName + " > agents")
	table := newTable(" Agent Dispatches ")
	status := newStatusBar()
	state := &tableState[lk.AgentDispatch]{cols: agentDispatchCols, sortAsc: true}
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
				fetched, err := n.client.ListAgentDispatches(ctx, roomName)

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
