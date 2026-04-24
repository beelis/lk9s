package ui

import (
	"encoding/json"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func metadataPage(n nav, title, content string) tview.Primitive {
	body := tview.NewTextView().SetText(prettyJSON(content))
	body.SetBorder(true).SetTitle(" metadata: " + title + " ")

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			n.pages.RemovePage("metadata")

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

func attributesPage(n nav, title string, attrs map[string]string) tview.Primitive {
	var content string
	if len(attrs) == 0 {
		content = "(no attributes)"
	} else {
		out, err := json.MarshalIndent(attrs, "", "  ")
		if err != nil {
			content = fmt.Sprintf("%v", attrs)
		} else {
			content = string(out)
		}
	}

	body := tview.NewTextView().SetText(content)
	body.SetBorder(true).SetTitle(" attributes: " + title + " ")

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			n.pages.RemovePage("attributes")

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

func prettyJSON(s string) string {
	if s == "" {
		return "(no metadata)"
	}

	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return s
	}

	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return s
	}

	return string(out)
}
