package ui

import (
	"strconv"
	"strings"

	"github.com/rivo/tview"
)

// roomCreatePage prompts for a room name and optional empty-timeout /
// max-participants limits, then creates it via CreateRoom. On success the
// page closes and the new room shows up on the rooms page's next periodic
// refresh; on error it stays open with the error shown so the input can be
// corrected. Blocked entirely (no inputs shown) when the context is
// read-only.
func roomCreatePage(n nav) tview.Primitive {
	const pageName = "create-room"

	form := tview.NewForm()
	form.SetBorder(true).SetTitle(" create room ")

	cancel := func() { n.pages.RemovePage(pageName) }

	if !n.write {
		form.AddTextView("", "write actions disabled for this context\n(set `write: true` in ~/.lk9s.yaml)", 40, 2, true, false)
		form.AddButton("close", cancel)
		form.SetCancelFunc(cancel)

		return centered(form, 9, 50)
	}

	form.AddInputField("name", "", 40, nil, nil)
	form.AddInputField("empty timeout (s, optional)", "", 10, nil, nil)
	form.AddInputField("max participants (optional)", "", 10, nil, nil)

	form.AddTextView("", "", 40, 1, true, false)
	statusIdx := form.GetFormItemCount() - 1

	setStatus := func(msg string) {
		if tv, ok := form.GetFormItem(statusIdx).(*tview.TextView); ok {
			tv.SetText(msg)
		}
	}

	create := func() {
		name := strings.TrimSpace(form.GetFormItemByLabel("name").(*tview.InputField).GetText())
		if name == "" {
			setStatus("[red] name is required")

			return
		}

		emptyTimeout, err := parseOptionalUint(form.GetFormItemByLabel("empty timeout (s, optional)").(*tview.InputField).GetText())
		if err != nil {
			setStatus("[red] empty timeout: " + err.Error())

			return
		}

		maxParticipants, err := parseOptionalUint(form.GetFormItemByLabel("max participants (optional)").(*tview.InputField).GetText())
		if err != nil {
			setStatus("[red] max participants: " + err.Error())

			return
		}

		form.SetTitle(" create room (creating...) ")

		go func() {
			_, err := n.client.CreateRoom(n.ctx, name, emptyTimeout, maxParticipants)

			n.app.QueueUpdateDraw(func() {
				if err != nil {
					form.SetTitle(" create room ")
					setStatus("[red] " + err.Error())

					return
				}

				n.pages.RemovePage(pageName)
			})
		}()
	}

	form.AddButton("create", create)
	form.AddButton("cancel", cancel)
	form.SetCancelFunc(cancel)

	return centered(form, 13, 60)
}

// parseOptionalUint parses s as a uint32, treating an empty string as 0
// (meaning "use the server default").
func parseOptionalUint(s string) (uint32, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}

	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint32(v), nil
}

// centered wraps p in a fixed-size box centered on screen.
func centered(p tview.Primitive, height, width int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 0, true).
			AddItem(nil, 0, 1, false), width, 0, true).
		AddItem(nil, 0, 1, false)
}
