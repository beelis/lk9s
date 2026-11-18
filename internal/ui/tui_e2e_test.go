package ui //nolint:testpackage // drives unexported runApp

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/mariusgassen/lk9s/internal/config"
	"github.com/mariusgassen/lk9s/internal/lk"
)

type fakeClient struct {
	RoomLister

	mu    sync.Mutex
	rooms []lk.Room
}

func (f *fakeClient) ListRooms(context.Context) ([]lk.Room, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]lk.Room(nil), f.rooms...), nil
}

func (f *fakeClient) ListParticipants(context.Context, string) ([]lk.Participant, error) {
	return []lk.Participant{{Identity: "alice", Kind: "STANDARD", State: "ACTIVE"}}, nil
}

func (f *fakeClient) DeleteRoom(_ context.Context, name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	for i, r := range f.rooms {
		if r.Name == name {
			f.rooms = append(f.rooms[:i], f.rooms[i+1:]...)
		}
	}

	return nil
}

type harness struct {
	t      *testing.T
	app    *tview.Application
	screen tcell.SimulationScreen
	done   chan error
}

func startTUI(t *testing.T, client RoomLister, write bool) *harness {
	t.Helper()

	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("init screen: %v", err)
	}

	screen.SetSize(160, 40)

	app := tview.NewApplication().SetScreen(screen)
	cc := config.Context{Name: "test", URL: "http://x", Write: write}
	done := make(chan error, 1)

	go func() {
		done <- runApp(t.Context(), app, func(config.Context) RoomLister { return client }, []config.Context{cc}, cc, "test")
	}()

	h := &harness{t: t, app: app, screen: screen, done: done}

	t.Cleanup(func() {
		app.Stop()
	})

	return h
}

// text snapshots the screen on the UI goroutine, since the simulation
// screen's cell buffer is not safe to read while a draw is in flight.
func (h *harness) text() string {
	out := make(chan string, 1)

	h.app.QueueUpdate(func() { out <- h.snapshot() })

	select {
	case s := <-out:
		return s
	case <-time.After(time.Second):
		return ""
	}
}

func (h *harness) snapshot() string {
	cells, w, _ := h.screen.GetContents()

	var sb strings.Builder

	for i, c := range cells {
		if len(c.Runes) > 0 {
			sb.WriteRune(c.Runes[0])
		} else {
			sb.WriteByte(' ')
		}

		if (i+1)%w == 0 {
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}

func (h *harness) typeText(s string) {
	for _, r := range s {
		h.screen.InjectKey(tcell.KeyRune, r, tcell.ModNone)
	}
}

func (h *harness) key(k tcell.Key) { h.screen.InjectKey(k, 0, tcell.ModNone) }

func (h *harness) waitFor(want string) {
	h.t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(h.text(), want) {
			return
		}

		time.Sleep(25 * time.Millisecond)
	}

	h.t.Fatalf("timed out waiting for %q; screen:\n%s", want, h.text())
}

func (h *harness) waitGone(unwanted string) {
	h.t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !strings.Contains(h.text(), unwanted) {
			return
		}

		time.Sleep(25 * time.Millisecond)
	}

	h.t.Fatalf("timed out waiting for %q to disappear; screen:\n%s", unwanted, h.text())
}

func newFake() *fakeClient {
	return &fakeClient{rooms: []lk.Room{{Name: "alpha", SID: "RM_a"}, {Name: "beta", SID: "RM_b"}}}
}

//nolint:paralleltest // applyTheme mutates tview's global styles
func TestTUIBrowseAndQuit(t *testing.T) {
	h := startTUI(t, newFake(), false)

	h.waitFor("alpha")
	h.waitFor("beta")

	h.key(tcell.KeyEnter)
	h.waitFor("alice")

	h.key(tcell.KeyEscape)
	h.waitFor("beta")

	h.typeText(":quit")
	h.key(tcell.KeyEnter)

	select {
	case err := <-h.done:
		if err != nil {
			t.Fatalf("run: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("app did not quit")
	}
}

//nolint:paralleltest // applyTheme mutates tview's global styles
func TestTUIFilter(t *testing.T) {
	h := startTUI(t, newFake(), false)
	h.waitFor("beta")

	h.typeText("/")
	h.typeText("alp")
	h.key(tcell.KeyEnter)

	h.waitGone("beta")
	h.waitFor("alpha")
}

//nolint:paralleltest // applyTheme mutates tview's global styles
func TestTUIDeleteRefusedOnReadOnly(t *testing.T) {
	fake := newFake()
	h := startTUI(t, fake, false)
	h.waitFor("alpha")

	h.key(tcell.KeyCtrlD)
	h.waitFor("write actions disabled")

	if got, _ := fake.ListRooms(t.Context()); len(got) != 2 {
		t.Fatalf("rooms = %d, want 2", len(got))
	}
}

//nolint:paralleltest // applyTheme mutates tview's global styles
func TestTUIDeleteWithConfirm(t *testing.T) {
	fake := newFake()
	h := startTUI(t, fake, true)
	h.waitFor("alpha")

	h.key(tcell.KeyCtrlD)
	h.typeText("wrong")
	h.key(tcell.KeyEnter)

	if got, _ := fake.ListRooms(t.Context()); len(got) != 2 {
		t.Fatalf("mismatched confirmation deleted a room: %d left", len(got))
	}

	for range len("wrong") {
		h.key(tcell.KeyBackspace2)
	}

	h.typeText("alpha")
	h.key(tcell.KeyEnter)

	h.waitGone("RM_a")
	h.waitFor("RM_b")
}
