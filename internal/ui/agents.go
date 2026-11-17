package ui

import (
	"cmp"
	"context"
	"errors"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mariusgassen/lk9s/internal/lk"
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
	state := &tableState[lk.AgentDispatch]{title: "agent-dispatches", cols: agentDispatchCols, sortAsc: true}
	state.setItems(initial)

	state.render(table)

	ctx, cancel := context.WithCancel(n.ctx)

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			cancel()
			n.pages.SwitchToPage("rooms")

			return nil
		}

		if event.Key() == tcell.KeyCtrlD {
			row, _ := table.GetSelection()
			if row > 0 && row <= len(state.sorted) && requireWrite(n, status) {
				confirmDeleteDispatch(n, roomName, state.sorted[row-1].ID, status)
			}

			return nil
		}

		if event.Rune() == 'c' {
			if requireWrite(n, status) {
				createDispatchForm(n, roomName, status)
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

	keys := [][2]string{{"c", "dispatch agent"}, {"Ctrl+D", "delete"}, {"Esc", "back"}, {"Shift+letter", "sort"}}

	return withHints(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(table, 0, 1, true).
		AddItem(status, 1, 0, false), keys)
}

func confirmDeleteDispatch(n nav, room, id string, status *tview.TextView) {
	n.pages.RemovePage("confirm-delete-dispatch")
	n.pages.AddPage("confirm-delete-dispatch", confirmDeleteDispatchPage(n, room, id, func(err error) {
		updateStatus(status, err)
	}), true, true)
}

func createDispatchForm(n nav, room string, status *tview.TextView) {
	const (
		fieldAgent    = "agent name"
		fieldMetadata = "metadata (optional)"
	)

	n.pages.RemovePage("create-dispatch")
	n.pages.AddPage("create-dispatch", formPage(n, "create-dispatch", " dispatch agent ",
		[]formField{{label: fieldAgent}, {label: fieldMetadata}},
		func(ctx context.Context, v map[string]string) error {
			if v[fieldAgent] == "" {
				return errors.New("agent name is required")
			}

			return n.client.CreateDispatch(ctx, room, v[fieldAgent], v[fieldMetadata])
		},
		func() { status.SetText(" dispatch created") },
	), true, true)
}
