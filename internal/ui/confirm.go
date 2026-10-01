package ui

import (
	"context"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// confirmDeleteRoomPage asks the user to type the room name back before
// deleting it, since the action disconnects every participant and cannot
// be undone. onDone is called on the UI thread after the delete attempt.
// Esc cancels: before the request is sent it just closes the dialog, while
// a delete already in flight is aborted via context cancellation.
func confirmDeleteRoomPage(n nav, roomName string, onDone func(err error)) tview.Primitive {
	const pageName = "confirm-delete"

	ctx, cancel := context.WithCancel(n.ctx)

	warning := tview.NewTextView().SetText(fmt.Sprintf(
		"Delete room %q?\nThis disconnects all participants and cannot be undone.\n\nType the room name to confirm, then press Enter.\nEsc to cancel.",
		roomName,
	))

	status := newStatusBar()
	input := tview.NewInputField().SetLabel("room name: ")

	input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}

		if input.GetText() != roomName {
			status.SetText("[red] name does not match")

			return
		}

		input.SetDisabled(true)
		status.SetText("deleting... (Esc to cancel)")

		go func() {
			defer cancel()

			err := n.client.DeleteRoom(ctx, roomName)

			n.app.QueueUpdateDraw(func() {
				n.pages.RemovePage(pageName)
				onDone(err)
			})
		}()
	})

	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			cancel()
			n.pages.RemovePage(pageName)

			return nil
		}

		return event
	})

	body := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(warning, 0, 1, false).
		AddItem(input, 1, 0, true).
		AddItem(status, 1, 0, false)
	body.SetBorder(true).SetTitle(" confirm delete ")

	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(body, 9, 0, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)
}

// confirmKickParticipantPage asks the user to type the participant's
// identity back before removing them from the room, mirroring
// confirmDeleteRoomPage. onDone is called on the UI thread after the
// removal attempt.
func confirmKickParticipantPage(n nav, room, identity string, onDone func(err error)) tview.Primitive {
	const pageName = "confirm-kick"

	ctx, cancel := context.WithCancel(n.ctx)

	warning := tview.NewTextView().SetText(fmt.Sprintf(
		"Remove participant %q from room %q?\nThis disconnects them immediately.\n\nType the identity to confirm, then press Enter.\nEsc to cancel.",
		identity, room,
	))

	status := newStatusBar()
	input := tview.NewInputField().SetLabel("identity: ")

	input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}

		if input.GetText() != identity {
			status.SetText("[red] identity does not match")

			return
		}

		input.SetDisabled(true)
		status.SetText("removing... (Esc to cancel)")

		go func() {
			defer cancel()

			err := n.client.RemoveParticipant(ctx, room, identity)

			n.app.QueueUpdateDraw(func() {
				n.pages.RemovePage(pageName)
				onDone(err)
			})
		}()
	})

	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			cancel()
			n.pages.RemovePage(pageName)

			return nil
		}

		return event
	})

	body := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(warning, 0, 1, false).
		AddItem(input, 1, 0, true).
		AddItem(status, 1, 0, false)
	body.SetBorder(true).SetTitle(" confirm kick ")

	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(body, 9, 0, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)
}
