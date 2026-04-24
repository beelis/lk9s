package ui

import (
	"fmt"
	"strings"

	"github.com/beelis/lk9s/internal/lk"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func permissionsPage(n nav, identity string, perm lk.Permission) tview.Primitive {
	var b strings.Builder
	flag := func(label string, v bool) {
		fmt.Fprintf(&b, "%-22s %t\n", label, v)
	}

	flag("CanPublish", perm.CanPublish)
	flag("CanSubscribe", perm.CanSubscribe)
	flag("CanPublishData", perm.CanPublishData)
	flag("CanUpdateMetadata", perm.CanUpdateMetadata)
	flag("Hidden", perm.Hidden)
	flag("Recorder", perm.Recorder)

	if len(perm.CanPublishSources) == 0 {
		fmt.Fprintf(&b, "%-22s (all)\n", "CanPublishSources")
	} else {
		fmt.Fprintf(&b, "%-22s %s\n", "CanPublishSources", strings.Join(perm.CanPublishSources, ", "))
	}

	body := tview.NewTextView().SetText(b.String())
	body.SetBorder(true).SetTitle(" permissions: " + identity + " ")

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			n.pages.RemovePage("permissions")

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
