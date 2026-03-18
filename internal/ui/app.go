package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/beelis/lk9s/internal/lk"
	"github.com/rivo/tview"
)

const refreshInterval = 5 * time.Second

type roomLister interface {
	ListRooms(ctx context.Context) ([]lk.Room, error)
}

func Run(client roomLister, contextName string) error {
	app := tview.NewApplication()
	table := newTable()

	header := tview.NewTextView().SetText(" ctx: " + contextName)

	rooms, err := client.ListRooms(context.Background())
	if err != nil {
		return fmt.Errorf("fetch rooms: %w", err)
	}

	populateTable(table, rooms)

	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for range ticker.C {
			rooms, err := client.ListRooms(context.Background())
			if err != nil {
				return
			}

			app.QueueUpdateDraw(func() {
				populateTable(table, rooms)
			})
		}
	}()

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(table, 0, 1, true)

	return app.SetRoot(layout, true).Run()
}

func newTable() *tview.Table {
	table := tview.NewTable().SetBorders(false).SetSelectable(true, false)
	table.SetTitle(" Rooms ").SetBorder(true)

	headers := []string{"NAME", "SID", "PARTICIPANTS", "CREATED"}
	for col, h := range headers {
		table.SetCell(0, col, tview.NewTableCell(h).SetSelectable(false).SetExpansion(1))
	}

	return table
}

func populateTable(table *tview.Table, rooms []lk.Room) {
	table.Clear()

	headers := []string{"NAME", "SID", "PARTICIPANTS", "CREATED"}
	for col, h := range headers {
		table.SetCell(0, col, tview.NewTableCell(h).SetSelectable(false).SetExpansion(1))
	}

	for row, r := range rooms {
		created := time.Unix(r.CreationTime, 0).Format(time.DateTime)

		values := []string{r.Name, r.SID, fmt.Sprintf("%d", r.NumParticipants), created}
		for col, v := range values {
			table.SetCell(row+1, col, tview.NewTableCell(v).SetExpansion(1))
		}
	}
}
