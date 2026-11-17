package ui

import (
	"context"
	"strings"

	"github.com/rivo/tview"
)

type formField struct {
	label   string
	initial string
}

// formPage shows a modal form of text inputs. submit runs off the UI thread;
// on success the page closes and onDone is called on the UI thread, on error
// the page stays open with the error shown.
func formPage(
	n nav, pageName, title string, fields []formField,
	submit func(ctx context.Context, values map[string]string) error, onDone func(),
) tview.Primitive {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle(title)

	for _, f := range fields {
		form.AddInputField(f.label, f.initial, 50, nil, nil)
	}

	form.AddTextView("", "", 50, 1, true, false)
	statusIdx := form.GetFormItemCount() - 1

	setStatus := func(msg string) {
		if tv, ok := form.GetFormItem(statusIdx).(*tview.TextView); ok {
			tv.SetText(msg)
		}
	}

	closePage := func() { n.pages.RemovePage(pageName) }

	form.AddButton("ok", func() {
		values := make(map[string]string, len(fields))
		for _, f := range fields {
			values[f.label] = inputText(form, f.label)
		}

		setStatus("working...")

		go func() {
			err := submit(n.ctx, values)

			n.app.QueueUpdateDraw(func() {
				if err != nil {
					setStatus("[red] " + err.Error())

					return
				}

				closePage()
				onDone()
			})
		}()
	})
	form.AddButton("cancel", closePage)
	form.SetCancelFunc(closePage)

	return centered(form, 2*len(fields)+8, 80)
}

const (
	fieldTopic      = "topic (optional)"
	fieldPayload    = "payload"
	fieldRecipients = "identities (comma-separated, empty = all)"
	fieldLossy      = "lossy (y/n)"
)

// sendDataPage prompts for a data packet and sends it to the room via the
// server API.
func sendDataPage(n nav, room string, status *tview.TextView) tview.Primitive {
	return formPage(n, "send-data", " send data: "+room+" ",
		[]formField{
			{label: fieldTopic}, {label: fieldPayload}, {label: fieldRecipients}, {label: fieldLossy, initial: "n"},
		},
		func(ctx context.Context, v map[string]string) error {
			var identities []string

			for id := range strings.SplitSeq(v[fieldRecipients], ",") {
				if id = strings.TrimSpace(id); id != "" {
					identities = append(identities, id)
				}
			}

			lossy := strings.EqualFold(strings.TrimSpace(v[fieldLossy]), "y")

			return n.client.SendData(ctx, room, v[fieldTopic], []byte(v[fieldPayload]), identities, !lossy)
		},
		func() { status.SetText(" data sent") },
	)
}
