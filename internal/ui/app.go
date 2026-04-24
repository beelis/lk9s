package ui

import (
	"context"
	"strings"
	"time"

	"github.com/beelis/lk9s/internal/lk"
	"github.com/rivo/tview"
)

const refreshInterval = 5 * time.Second

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
	ctx         context.Context
}

func Run(client roomLister, contextName string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := tview.NewApplication()
	pages := tview.NewPages()
	n := nav{app: app, pages: pages, client: client, contextName: contextName, ctx: ctx}

	pages.AddPage("rooms", roomsPage(n), true, true)

	return app.SetRoot(pages, true).Run()
}

func newTable(title string) *tview.Table {
	table := tview.NewTable().SetBorders(false).SetSelectable(true, false)
	table.SetTitle(title).SetBorder(true)

	return table
}

func newStatusBar() *tview.TextView {
	return tview.NewTextView().SetDynamicColors(true)
}

func updateStatus(bar *tview.TextView, err error) {
	if err != nil {
		bar.SetText("[red] error: " + err.Error())

		return
	}

	bar.SetText("")
}

func legend(entries [][2]string) *tview.TextView {
	var b strings.Builder

	for _, e := range entries {
		if b.Len() > 0 {
			b.WriteString("  ")
		}

		b.WriteString("<")
		b.WriteString(e[0])
		b.WriteString("> ")
		b.WriteString(e[1])
	}

	return tview.NewTextView().SetText(" " + b.String())
}
