package ui

import (
	"context"
	"errors"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// typeToConfirm describes a destructive action that must be confirmed by
// typing a value back.
type typeToConfirm struct {
	pageName string
	title    string
	label    string // input label, e.g. "room name: "
	mismatch string // status shown when the typed text differs
	progress string // status shown while the action is in flight
	warning  string
	expected string
	action   func(ctx context.Context) error
}

// confirmPage asks the user to type expected back before running action,
// which is aborted via context cancellation if the user presses Esc while
// it is in flight. onDone is called on the UI thread after the attempt; an
// attempt aborted by the user reports a nil error, since the outcome is
// unknown and the caller's refresh shows the real state.
func confirmPage(n nav, c typeToConfirm, onDone func(err error)) tview.Primitive {
	ctx, cancel := context.WithCancel(n.ctx)

	aborted := false

	warning := tview.NewTextView().SetText(c.warning)
	status := newStatusBar()
	input := tview.NewInputField().SetLabel(c.label)

	input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}

		if input.GetText() != c.expected {
			status.SetText(c.mismatch)

			return
		}

		input.SetDisabled(true)
		status.SetText(c.progress + " (Esc to cancel)")

		go func() {
			defer cancel()

			err := c.action(ctx)

			n.app.QueueUpdateDraw(func() {
				n.pages.RemovePage(c.pageName)

				if aborted && errors.Is(err, context.Canceled) {
					err = nil
				}

				onDone(err)
			})
		}()
	})

	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			aborted = true

			cancel()
			n.pages.RemovePage(c.pageName)

			return nil
		}

		return event
	})

	body := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(warning, 0, 1, false).
		AddItem(input, 1, 0, true).
		AddItem(status, 1, 0, false)
	body.SetBorder(true).SetTitle(c.title)

	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(body, 9, 0, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)
}

// confirmDeleteRoomPage asks the user to type the room name back before
// deleting it, since the action disconnects every participant and cannot
// be undone. onDone is called on the UI thread after the delete attempt.
func confirmDeleteRoomPage(n nav, roomName string, onDone func(err error)) tview.Primitive {
	return confirmPage(n, typeToConfirm{
		pageName: "confirm-delete",
		title:    " confirm delete ",
		label:    "room name: ",
		mismatch: "[red] name does not match",
		progress: "deleting...",
		warning: fmt.Sprintf(
			"Delete room %q?\nThis disconnects all participants and cannot be undone.\n\n"+
				"Type the room name to confirm, then press Enter.\nEsc to cancel.",
			roomName,
		),
		expected: roomName,
		action:   func(ctx context.Context) error { return n.client.DeleteRoom(ctx, roomName) },
	}, onDone)
}

// confirmKickParticipantPage asks the user to type the participant's
// identity back before removing them from the room. onDone is called on the
// UI thread after the removal attempt.
func confirmKickParticipantPage(n nav, room, identity string, onDone func(err error)) tview.Primitive {
	return confirmPage(n, typeToConfirm{
		pageName: "confirm-kick",
		title:    " confirm kick ",
		label:    "identity: ",
		mismatch: "[red] identity does not match",
		progress: "removing...",
		warning: fmt.Sprintf(
			"Remove participant %q from room %q?\nThis disconnects them immediately.\n\n"+
				"Type the identity to confirm, then press Enter.\nEsc to cancel.",
			identity, room,
		),
		expected: identity,
		action:   func(ctx context.Context) error { return n.client.RemoveParticipant(ctx, room, identity) },
	}, onDone)
}

// confirmStopEgressPage asks the user to type the egress ID back before
// stopping it, which ends the recording or stream. onDone is called on the UI
// thread after the attempt.
func confirmStopEgressPage(n nav, id string, onDone func(err error)) tview.Primitive {
	return confirmPage(n, typeToConfirm{
		pageName: "confirm-stop-egress",
		title:    " confirm stop ",
		label:    "egress id: ",
		mismatch: "[red] id does not match",
		progress: "stopping...",
		warning: fmt.Sprintf(
			"Stop egress %q?\nThe recording or stream ends.\n\n"+
				"Type the egress ID to confirm, then press Enter.\nEsc to cancel.",
			id,
		),
		expected: id,
		action:   func(ctx context.Context) error { return n.client.StopEgress(ctx, id) },
	}, onDone)
}

// confirmDeleteDispatchPage asks the user to type the dispatch ID back before
// deleting the agent dispatch. onDone is called on the UI thread after the
// attempt.
func confirmDeleteDispatchPage(n nav, room, id string, onDone func(err error)) tview.Primitive {
	return confirmPage(n, typeToConfirm{
		pageName: "confirm-delete-dispatch",
		title:    " confirm delete ",
		label:    "dispatch id: ",
		mismatch: "[red] id does not match",
		progress: "deleting...",
		warning: fmt.Sprintf(
			"Delete agent dispatch %q in room %q?\n\n"+
				"Type the dispatch ID to confirm, then press Enter.\nEsc to cancel.",
			id, room,
		),
		expected: id,
		action:   func(ctx context.Context) error { return n.client.DeleteDispatch(ctx, room, id) },
	}, onDone)
}
