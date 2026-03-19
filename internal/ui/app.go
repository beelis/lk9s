package ui

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/beelis/lk9s/internal/lk"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const refreshInterval = 5 * time.Second

type column[T any] struct {
	header  string
	key     rune
	display func(T) string
	compare func(T, T) int
}

type tableState[T any] struct {
	cols    []column[T]
	items   []T
	sorted  []T
	sortCol int
	sortAsc bool
}

func (s *tableState[T]) render(table *tview.Table) {
	table.Clear()

	for col, c := range s.cols {
		label := c.header

		if col == s.sortCol {
			if s.sortAsc {
				label += " ▲"
			} else {
				label += " ▼"
			}
		}

		table.SetCell(0, col, tview.NewTableCell(label).SetSelectable(false).SetExpansion(1))
	}

	s.sorted = slices.Clone(s.items)
	slices.SortFunc(s.sorted, func(a, b T) int {
		n := s.cols[s.sortCol].compare(a, b)
		if !s.sortAsc {
			return -n
		}

		return n
	})

	for row, item := range s.sorted {
		for col, c := range s.cols {
			table.SetCell(row+1, col, tview.NewTableCell(c.display(item)).SetExpansion(1))
		}
	}
}

func (s *tableState[T]) handleKey(r rune) bool {
	for i, c := range s.cols {
		if c.key != r {
			continue
		}

		if i == s.sortCol {
			s.sortAsc = !s.sortAsc
		} else {
			s.sortCol = i
			s.sortAsc = true
		}

		return true
	}

	return false
}

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
		header:  "STATE",
		key:     'S',
		display: func(p lk.Participant) string { return p.State },
		compare: func(a, b lk.Participant) int { return cmp.Compare(a.State, b.State) },
	},
	{
		header:  "TRACKS",
		key:     'T',
		display: func(p lk.Participant) string { return fmt.Sprintf("%d", p.Tracks) },
		compare: func(a, b lk.Participant) int { return cmp.Compare(a.Tracks, b.Tracks) },
	},
	{
		header:  "JOINED",
		key:     'J',
		display: func(p lk.Participant) string { return time.Unix(p.JoinedAt, 0).Format(time.DateTime) },
		compare: func(a, b lk.Participant) int { return cmp.Compare(a.JoinedAt, b.JoinedAt) },
	},
}

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

type roomLister interface {
	ListRooms(ctx context.Context) ([]lk.Room, error)
	ListParticipants(ctx context.Context, room string) ([]lk.Participant, error)
	ListEgresses(ctx context.Context, room string) ([]lk.Egress, error)
}

type nav struct {
	app         *tview.Application
	pages       *tview.Pages
	client      roomLister
	contextName string
}

func Run(client roomLister, contextName string) error {
	rooms, err := client.ListRooms(context.Background())
	if err != nil {
		return fmt.Errorf("fetch rooms: %w", err)
	}

	app := tview.NewApplication()
	pages := tview.NewPages()
	n := nav{app: app, pages: pages, client: client, contextName: contextName}

	pages.AddPage("rooms", roomsPage(n, rooms), true, true)

	return app.SetRoot(pages, true).Run()
}

func roomsInputCapture(n nav, table *tview.Table, state *tableState[lk.Room]) func(*tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := table.GetSelection()

		if event.Rune() == 'e' {
			if row > 0 && row <= len(state.sorted) {
				roomName := state.sorted[row-1].Name

				go func() {
					eg, err := n.client.ListEgresses(context.Background(), roomName)
					if err != nil {
						return
					}

					n.app.QueueUpdateDraw(func() {
						n.pages.RemovePage("egresses")
						n.pages.AddPage("egresses", egressesPage(n, roomName, eg), true, true)
					})
				}()
			}

			return nil
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

func roomsPage(n nav, rooms []lk.Room) tview.Primitive {
	header := tview.NewTextView().SetText(" ctx: " + n.contextName)
	table := newTable(" Rooms ")
	state := &tableState[lk.Room]{cols: roomCols, items: rooms, sortAsc: true}

	state.render(table)
	table.SetInputCapture(roomsInputCapture(n, table, state))

	table.SetSelectedFunc(func(row, _ int) {
		if row == 0 || row > len(state.sorted) {
			return
		}

		roomName := state.sorted[row-1].Name

		go func() {
			pp, err := n.client.ListParticipants(context.Background(), roomName)
			if err != nil {
				return
			}

			n.app.QueueUpdateDraw(func() {
				n.pages.RemovePage("participants")
				n.pages.AddPage("participants", participantsPage(n, roomName, pp), true, true)
			})
		}()
	})

	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for range ticker.C {
			fetched, err := n.client.ListRooms(context.Background())
			if err != nil {
				return
			}

			n.app.QueueUpdateDraw(func() {
				state.items = fetched
				state.render(table)
			})
		}
	}()

	keys := [][2]string{{"Enter", "participants"}, {"e", "egresses"}, {"m", "metadata"}, {"Shift+letter", "sort"}}

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(table, 0, 1, true).
		AddItem(legend(keys), 1, 0, false)
}

func participantsPage(n nav, roomName string, initial []lk.Participant) tview.Primitive {
	header := tview.NewTextView().SetText(" ctx: " + n.contextName + " > " + roomName)
	table := newTable(" Participants ")
	state := &tableState[lk.Participant]{cols: participantCols, items: initial, sortAsc: true}

	state.render(table)

	ctx, cancel := context.WithCancel(context.Background())

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
				if err != nil {
					return
				}

				n.app.QueueUpdateDraw(func() {
					state.items = fetched
					state.render(table)
				})
			}
		}
	}()

	keys := [][2]string{{"Esc", "back"}, {"m", "metadata"}, {"Shift+letter", "sort"}}

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(table, 0, 1, true).
		AddItem(legend(keys), 1, 0, false)
}

func egressesPage(n nav, roomName string, initial []lk.Egress) tview.Primitive {
	header := tview.NewTextView().SetText(" ctx: " + n.contextName + " > " + roomName + " > egresses")
	table := newTable(" Egresses ")
	state := &tableState[lk.Egress]{cols: egressCols, items: initial, sortAsc: true}

	state.render(table)

	ctx, cancel := context.WithCancel(context.Background())

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
				if err != nil {
					return
				}

				n.app.QueueUpdateDraw(func() {
					state.items = fetched
					state.render(table)
				})
			}
		}
	}()

	keys := [][2]string{{"Esc", "back"}, {"Shift+letter", "sort"}}

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(table, 0, 1, true).
		AddItem(legend(keys), 1, 0, false)
}

func metadataPage(n nav, title, content string) tview.Primitive {
	body := tview.NewTextView().SetText(prettyJSON(content))
	body.SetBorder(true).SetTitle(" metadata: " + title + " ")

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			n.pages.RemovePage("metadata")

			return nil
		}

		return event
	})

	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(body, 0, 3, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)
}

func prettyJSON(s string) string {
	if s == "" {
		return "(no metadata)"
	}

	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return s
	}

	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return s
	}

	return string(out)
}

func newTable(title string) *tview.Table {
	table := tview.NewTable().SetBorders(false).SetSelectable(true, false)
	table.SetTitle(title).SetBorder(true)

	return table
}

func legend(entries [][2]string) *tview.TextView {
	text := ""

	for _, e := range entries {
		if text != "" {
			text += "  "
		}

		text += "<" + e[0] + "> " + e[1]
	}

	return tview.NewTextView().SetText(" " + text)
}
