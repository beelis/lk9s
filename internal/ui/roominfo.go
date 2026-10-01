package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/beelis/lk9s/internal/lk"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func roomInfoPage(n nav, r lk.Room) tview.Primitive {
	var b strings.Builder

	field := func(label string, v any) {
		fmt.Fprintf(&b, "%-20s %v\n", label, v)
	}

	field("SID", r.SID)
	field("NumParticipants", r.NumParticipants)
	field("NumPublishers", r.NumPublishers)
	field("ActiveRecording", r.ActiveRecording)
	field("MaxParticipants", r.MaxParticipants)
	field("EmptyTimeout", fmt.Sprintf("%ds", r.EmptyTimeout))
	field("DepartureTimeout", fmt.Sprintf("%ds", r.DepartureTimeout))
	field("Created", time.Unix(r.CreationTime, 0).Format(time.DateTime))

	b.WriteString("\nEnabled codecs\n")

	if len(r.EnabledCodecs) == 0 {
		b.WriteString("  (none set — server defaults apply)\n")
	} else {
		for _, c := range r.EnabledCodecs {
			if c.FmtpLine != "" {
				fmt.Fprintf(&b, "  %-16s fmtp=%s\n", c.MimeType, c.FmtpLine)
			} else {
				fmt.Fprintf(&b, "  %s\n", c.MimeType)
			}
		}
	}

	body := tview.NewTextView().SetText(b.String())
	body.SetBorder(true).SetTitle(" room info: " + r.Name + " ")

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			n.pages.RemovePage("roominfo")

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
