package ui

import (
	"cmp"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/beelis/lk9s/internal/lk"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var trackCols = []column[lk.Track]{
	{
		header:  "SID",
		key:     'S',
		display: func(t lk.Track) string { return t.SID },
		compare: func(a, b lk.Track) int { return cmp.Compare(a.SID, b.SID) },
	},
	{
		header:  "TYPE",
		key:     'T',
		display: func(t lk.Track) string { return t.Type },
		compare: func(a, b lk.Track) int { return cmp.Compare(a.Type, b.Type) },
	},
	{
		header:  "SOURCE",
		key:     'O',
		display: func(t lk.Track) string { return t.Source },
		compare: func(a, b lk.Track) int { return cmp.Compare(a.Source, b.Source) },
	},
	{
		header:  "MUTED",
		key:     'U',
		display: func(t lk.Track) string { return fmt.Sprintf("%t", t.Muted) },
		compare: func(a, b lk.Track) int { return cmp.Compare(fmt.Sprint(a.Muted), fmt.Sprint(b.Muted)) },
	},
	{
		header:  "NAME",
		key:     'N',
		display: func(t lk.Track) string { return t.Name },
		compare: func(a, b lk.Track) int { return cmp.Compare(a.Name, b.Name) },
	},
}

// tracksPage lists a participant's published tracks. Enter opens full
// per-track detail (codecs, layers, ...); "m" toggles mute, gated by the
// context's write setting.
//
//nolint:funlen
func tracksPage(n nav, room, identity string, initial []lk.Track) tview.Primitive {
	header := tview.NewTextView().SetDynamicColors(true).
		SetText(fmt.Sprintf(" ctx: %s%s > %s > %s", n.contextName, writeTag(n), room, identity))
	table := newTable(fmt.Sprintf(" Tracks: %s ", identity))
	status := newStatusBar()
	state := &tableState[lk.Track]{cols: trackCols, sortAsc: true}
	state.setItems(initial)

	state.render(table)
	updateListStatus(status, nil, len(initial))

	ctx, cancel := context.WithCancel(n.ctx)

	fetchAndRender := func(keepStatus error) {
		pp, err := n.client.ListParticipants(ctx, room)

		n.app.QueueUpdateDraw(func() {
			if err == nil {
				state.setItems(participantTracks(pp, identity))
				state.render(table)
			}

			if keepStatus != nil {
				return
			}

			updateListStatus(status, err, len(state.items))
		})
	}

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			cancel()
			n.pages.RemovePage("tracks")

			return nil
		}

		if event.Rune() == '/' {
			n.pages.AddPage("filter", filterBarPage(n, state.filterText, func(f string) {
				state.setFilter(f)
				state.render(table)
			}), true, true)

			return nil
		}

		if event.Rune() == 'm' {
			row, _ := table.GetSelection()
			if row > 0 && row <= len(state.sorted) && requireWrite(n, status) {
				t := state.sorted[row-1]

				status.SetText("updating...")

				go func() {
					err := n.client.SetTrackMuted(ctx, room, identity, t.SID, !t.Muted)

					n.app.QueueUpdateDraw(func() {
						updateStatus(status, err)
						go fetchAndRender(err)
					})
				}()
			}

			return nil
		}

		if !state.handleKey(event.Rune()) {
			return event
		}

		state.render(table)

		return nil
	})

	table.SetSelectedFunc(func(row, _ int) {
		if row == 0 || row > len(state.sorted) {
			return
		}

		n.pages.RemovePage("track-detail")
		n.pages.AddPage("track-detail", trackDetailPage(n, state.sorted[row-1]), true, true)
	})

	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fetchAndRender(nil)
			}
		}
	}()

	keys := [][2]string{
		{"Esc", "back"}, {"Enter", "detail"}, {"m", "mute/unmute"}, {"/", "filter"},
		{"Shift+letter", "sort"},
	}

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(table, 0, 1, true).
		AddItem(status, 1, 0, false).
		AddItem(legend(keys), 1, 0, false)
}

func participantTracks(pp []lk.Participant, identity string) []lk.Track {
	for _, p := range pp {
		if p.Identity == identity {
			return p.Tracks
		}
	}

	return nil
}

func trackDetailPage(n nav, t lk.Track) tview.Primitive {
	var b strings.Builder
	writeTrack(&b, t)

	body := tview.NewTextView().SetText(b.String()).SetScrollable(true)
	body.SetBorder(true).SetTitle(" track: " + t.SID + " ")

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			n.pages.RemovePage("track-detail")

			return nil
		}

		return event
	})

	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(body, 0, 6, true).
			AddItem(nil, 0, 1, false), 0, 4, true).
		AddItem(nil, 0, 1, false)
}

func writeTrack(b *strings.Builder, t lk.Track) {
	field := func(label string, v any) {
		fmt.Fprintf(b, "  %-18s %v\n", label, v)
	}

	fmt.Fprintf(b, "%s  %s / %s\n", t.SID, t.Type, t.Source)
	field("Name", t.Name)
	field("MimeType", t.MimeType)
	field("Muted", t.Muted)
	field("MID", t.MID)
	field("Stream", t.Stream)
	field("Encryption", t.Encryption)
	field("BackupCodecPolicy", t.BackupCodecPolicy)

	if t.Version != 0 {
		field("Version", time.UnixMicro(t.Version).Format(time.DateTime))
	}

	if t.Type == "VIDEO" {
		field("Resolution", fmt.Sprintf("%dx%d", t.Width, t.Height))
		field("Simulcast", t.Simulcast)
	}

	if t.Type == "AUDIO" {
		field("Stereo", t.Stereo)
		field("DisableDTX", t.DisableDTX)
		field("DisableRED", t.DisableRED)
		field("AudioFeatures", strings.Join(t.AudioFeatures, ", "))
	}

	writeLayers(b, "  ", t.Layers)

	for _, c := range t.Codecs {
		fmt.Fprintf(b, "  Codec %s\n", c.MimeType)
		fmt.Fprintf(b, "    %-16s %s\n", "MID", c.MID)
		fmt.Fprintf(b, "    %-16s %s\n", "CID", c.CID)

		if c.SDPCID != "" {
			fmt.Fprintf(b, "    %-16s %s\n", "SDP CID", c.SDPCID)
		}

		if t.Type == "VIDEO" {
			fmt.Fprintf(b, "    %-16s %s\n", "LayerMode", c.LayerMode)
		}

		writeLayers(b, "    ", c.Layers)
	}
}

func writeLayers(b *strings.Builder, indent string, layers []lk.VideoLayer) {
	for _, l := range layers {
		fmt.Fprintf(b, "%sLayer %-6s %4dx%-4d %7.0f kbps  ssrc=%d rtx=%d spatial=%d rid=%q\n",
			indent, l.Quality, l.Width, l.Height, float64(l.Bitrate)/1000,
			l.SSRC, l.RepairSSRC, l.SpatialLayer, l.RID)
	}
}
