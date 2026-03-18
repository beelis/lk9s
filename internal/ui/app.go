package ui

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/beelis/lk9s/internal/lk"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const refreshInterval = 5 * time.Second

type column struct {
	header  string
	key     rune
	display func(r lk.Room) string
	compare func(a, b lk.Room) int
}

var columns = []column{
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

type roomLister interface {
	ListRooms(ctx context.Context) ([]lk.Room, error)
}

func Run(client roomLister, contextName string) error {
	app := tview.NewApplication()
	table := newTable()
	header := tview.NewTextView().SetText(" ctx: " + contextName)

	sortCol := 0
	sortAsc := true

	rooms, err := client.ListRooms(context.Background())
	if err != nil {
		return fmt.Errorf("fetch rooms: %w", err)
	}

	populateTable(table, rooms, sortCol, sortAsc)

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		col := sortKeyCol(event.Rune())
		if col < 0 {
			return event
		}

		if col == sortCol {
			sortAsc = !sortAsc
		} else {
			sortCol = col
			sortAsc = true
		}

		populateTable(table, rooms, sortCol, sortAsc)

		return nil
	})

	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for range ticker.C {
			fetched, err := client.ListRooms(context.Background())
			if err != nil {
				return
			}

			app.QueueUpdateDraw(func() {
				rooms = fetched
				populateTable(table, rooms, sortCol, sortAsc)
			})
		}
	}()

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(table, 0, 1, true)

	return app.SetRoot(layout, true).Run()
}

func sortKeyCol(r rune) int {
	for i, c := range columns {
		if c.key == r {
			return i
		}
	}

	return -1
}

func newTable() *tview.Table {
	table := tview.NewTable().SetBorders(false).SetSelectable(true, false)
	table.SetTitle(" Rooms ").SetBorder(true)

	return table
}

func populateTable(table *tview.Table, rooms []lk.Room, sortCol int, sortAsc bool) {
	table.Clear()

	for col, c := range columns {
		label := c.header

		if col == sortCol {
			if sortAsc {
				label += " ▲"
			} else {
				label += " ▼"
			}
		}

		table.SetCell(0, col, tview.NewTableCell(label).SetSelectable(false).SetExpansion(1))
	}

	for row, r := range sortedRooms(rooms, sortCol, sortAsc) {
		for col, c := range columns {
			table.SetCell(row+1, col, tview.NewTableCell(c.display(r)).SetExpansion(1))
		}
	}
}

func sortedRooms(rooms []lk.Room, col int, asc bool) []lk.Room {
	sorted := slices.Clone(rooms)
	slices.SortFunc(sorted, func(a, b lk.Room) int {
		n := columns[col].compare(a, b)
		if !asc {
			return -n
		}

		return n
	})

	return sorted
}
