package ui

import (
	"context"
	"fmt"
	"strings"
	"syscall"
	"time"

	"os"
	"os/signal"

	"github.com/beelis/lk9s/internal/config"
	"github.com/beelis/lk9s/internal/lk"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const refreshInterval = 5 * time.Second

// RoomLister is the subset of *lk.Client the UI depends on.
type RoomLister interface {
	ListRooms(ctx context.Context) ([]lk.Room, error)
	ListParticipants(ctx context.Context, room string) ([]lk.Participant, error)
	ListEgresses(ctx context.Context, room string) ([]lk.Egress, error)
	ListAgentDispatches(ctx context.Context, room string) ([]lk.AgentDispatch, error)
	ListSIP(ctx context.Context) ([]lk.SIPEntry, error)
	DeleteRoom(ctx context.Context, room string) error
	CreateRoom(ctx context.Context, name string, emptyTimeout, maxParticipants uint32) (lk.Room, error)
	RemoveParticipant(ctx context.Context, room, identity string) error
	SetTrackMuted(ctx context.Context, room, identity, trackSID string, muted bool) error
	UpdatePermission(ctx context.Context, room, identity string, perm lk.Permission) error
	CreateToken(identity, room string, ttl time.Duration, grant lk.TokenGrant) (string, error)
}

type nav struct {
	app         *tview.Application
	pages       *tview.Pages
	client      RoomLister
	contextName string
	write       bool // whether destructive/mutating actions are allowed for the current context
	version     string
	ctx         context.Context
}

// requireWrite reports whether destructive/mutating actions are allowed for
// the current context. If not, it leaves an explanatory message on status
// and returns false so the caller can bail out before performing the action.
func requireWrite(n nav, status *tview.TextView) bool {
	if n.write {
		return true
	}

	status.SetText("[yellow] write actions disabled for this context (set `write: true` in ~/.lk9s.yaml)")

	return false
}

// writeTag returns a header suffix flagging that write actions are enabled
// for the current context. It's intentionally silent in the default
// (read-only) mode, so the normal, safe state stays visually quiet and the
// dangerous one stands out.
func writeTag(n nav) string {
	if !n.write {
		return ""
	}

	return " [red::b][WRITE][-:-:-]"
}

// Run starts the TUI, connected to dial(current). contexts is the full list
// of configured projects, offered by the ":projects" command to switch mid
// session.
func Run(dial func(config.Context) RoomLister, contexts []config.Context, current config.Context, version string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	app := tview.NewApplication()
	pages := tview.NewPages()
	n := nav{
		app: app, pages: pages, client: dial(current), contextName: current.Name,
		write: current.Write, version: version, ctx: ctx,
	}

	// cancelRooms stops the previous rooms page's background refresh (and,
	// transitively, any subpage opened from it) when a project switch
	// replaces it, instead of leaking a poller per switch.
	var cancelRooms context.CancelFunc

	showRooms := func(nn nav) {
		if cancelRooms != nil {
			cancelRooms()
		}

		roomsCtx, roomsCancel := context.WithCancel(ctx)
		cancelRooms = roomsCancel
		nn.ctx = roomsCtx

		pages.RemovePage("rooms")
		pages.AddPage("rooms", roomsPage(nn), true, true)
		pages.SwitchToPage("rooms")
	}

	// active tracks the nav for whichever context is currently selected, so
	// the global ":" command dispatch below (which outlives any one
	// context) always acts against the right client instead of the one
	// dial(current) built at startup.
	active := n

	switchContext := func(c config.Context) {
		nn := n
		nn.client = dial(c)
		nn.contextName = c.Name
		nn.write = c.Write
		active = nn
		showRooms(nn)
	}

	showRooms(n)

	// cmdBar is a single persistent widget, docked at the top of the screen
	// for the app's entire lifetime (k9s style), rather than a page overlay
	// recreated on every ":" press.
	var cmdBar *tview.InputField

	blur := func() { app.SetFocus(pages) }

	cmdBar = newCommandBar(blur, func(cmd string) {
		switch cmd {
		case "projects", "ctx":
			prev, _ := pages.GetFrontPage()

			pages.AddPage("contexts", contextsPage(n, contexts, switchContext, func() {
				pages.SwitchToPage(prev)
			}), true, true)
			pages.SwitchToPage("contexts")
		case "rooms":
			pages.SwitchToPage("rooms")
		case "create-room":
			pages.AddPage("create-room", roomCreatePage(active), true, true)
		case "token":
			pages.AddPage("token-form", tokenFormPage(active, "", ""), true, true)
		case "sip":
			// Errors aren't surfaced here (there's no status bar at this
			// level); sipPage's own refresh ticker will report a fetch
			// failure via its status bar within refreshInterval.
			go func() {
				entries, _ := active.client.ListSIP(active.ctx)

				app.QueueUpdateDraw(func() {
					pages.RemovePage("sip")
					pages.AddPage("sip", sipPage(active, entries), true, true)
				})
			}()
		case "quit", "q":
			app.Stop()
		}
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() != ':' {
			return event
		}

		if _, ok := app.GetFocus().(*tview.InputField); ok {
			return event
		}

		app.SetFocus(cmdBar)

		return nil
	})

	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(cmdBar, 1, 0, false).
		AddItem(pages, 0, 1, true)

	return app.SetRoot(root, true).Run()
}

func newTable(title string) *tview.Table {
	table := tview.NewTable().SetBorders(false).SetSelectable(true, false).SetFixed(1, 0)
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

// updateListStatus is like updateStatus but, on success, reports the row
// count and the time of the refresh that produced it instead of clearing
// the bar.
func updateListStatus(bar *tview.TextView, err error, count int) {
	if err != nil {
		bar.SetText("[red] error: " + err.Error())

		return
	}

	bar.SetText(fmt.Sprintf(" %d item(s) · refreshed %s", count, time.Now().Format("15:04:05")))
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
