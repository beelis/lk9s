package ui //nolint:testpackage // tests unexported helpers

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mariusgassen/lk9s/internal/lk"
	"github.com/rivo/tview"
)

func TestEmptyTableDraws(t *testing.T) {
	t.Parallel()

	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}

	screen.SetSize(120, 30)

	table := newTable(" Rooms ")
	state := &tableState[lk.Room]{title: "rooms", cols: roomCols, sortAsc: true}
	state.setItems(nil)
	state.render(table)

	done := make(chan struct{})

	go func() {
		table.SetRect(0, 0, 120, 30)
		table.Draw(screen)
		table.InputHandler()(tcell.NewEventKey(tcell.KeyDown, 0, 0), func(p tview.Primitive) {})
		table.Draw(screen)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("hang drawing empty table")
	}
}
