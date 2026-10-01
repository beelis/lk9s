package ui

import (
	"fmt"
	"time"

	"github.com/beelis/lk9s/internal/lk"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// activityEvent is a single synthesized activity-log line. LiveKit has no
// lightweight way to subscribe to room events over the plain server API:
// webhooks need a public endpoint registered on the LiveKit project ahead
// of time, and joining as a real WebRTC participant to receive live
// callbacks would pull in a much heavier client stack than the rest of
// this Twirp-only tool needs. Instead, this diffs consecutive
// participant-list snapshots taken on the participants page's existing
// refresh ticker, so events land on that cadence (refreshInterval) rather
// than instantly.
type activityEvent struct {
	At   time.Time
	Text string
}

const maxActivityEvents = 200

func snapshotParticipants(pp []lk.Participant) map[string]lk.Participant {
	m := make(map[string]lk.Participant, len(pp))
	for _, p := range pp {
		m[p.Identity] = p
	}

	return m
}

// diffParticipants compares two participant snapshots of the same room and
// returns the join/leave/track events that explain the difference.
func diffParticipants(prev map[string]lk.Participant, current []lk.Participant) []activityEvent {
	now := time.Now()

	var events []activityEvent

	seen := make(map[string]bool, len(current))

	for _, p := range current {
		seen[p.Identity] = true

		old, existed := prev[p.Identity]
		if !existed {
			events = append(events, activityEvent{now, p.Identity + " joined"})

			continue
		}

		events = appendTrackDiff(events, now, p.Identity, "mic", old.Mic, p.Mic)
		events = appendTrackDiff(events, now, p.Identity, "camera", old.Camera, p.Camera)
		events = appendTrackDiff(events, now, p.Identity, "screen", old.Screen, p.Screen)
		events = appendTrackDiff(events, now, p.Identity, "screen audio", old.ScreenAudio, p.ScreenAudio)
	}

	for identity := range prev {
		if !seen[identity] {
			events = append(events, activityEvent{now, identity + " left"})
		}
	}

	return events
}

func appendTrackDiff(events []activityEvent, now time.Time, identity, label string, old, cur lk.TrackState) []activityEvent {
	if old == cur {
		return events
	}

	var text string

	switch {
	case cur == lk.TrackActive && old == lk.TrackAbsent:
		text = fmt.Sprintf("%s published %s", identity, label)
	case cur == lk.TrackActive:
		text = fmt.Sprintf("%s unmuted %s", identity, label)
	case cur == lk.TrackMuted:
		text = fmt.Sprintf("%s muted %s", identity, label)
	case cur == lk.TrackAbsent:
		text = fmt.Sprintf("%s unpublished %s", identity, label)
	default:
		return events
	}

	return append(events, activityEvent{now, text})
}

func formatActivityEvent(e activityEvent) string {
	return e.At.Format("15:04:05") + "  " + e.Text
}

// activityPage shows a room's synthesized activity log. onOpen is called
// with the page's TextView so the caller can append live lines to it while
// this page is on screen; onClose is called when the page is dismissed so
// the caller stops doing so.
func activityPage(n nav, room string, initial []activityEvent, onOpen func(*tview.TextView), onClose func()) tview.Primitive {
	const pageName = "activity"

	body := tview.NewTextView().SetScrollable(true).SetMaxLines(maxActivityEvents)
	body.SetBorder(true).SetTitle(" activity: " + room + " ")

	if len(initial) == 0 {
		fmt.Fprint(body, "(no activity observed yet; new events appear as they're detected)")
	} else {
		for _, e := range initial {
			fmt.Fprintln(body, formatActivityEvent(e))
		}
	}

	body.ScrollToEnd()

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			onClose()
			n.pages.RemovePage(pageName)

			return nil
		}

		return event
	})

	onOpen(body)

	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(body, 0, 3, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)
}

func appendActivityLines(tv *tview.TextView, events []activityEvent) {
	for _, e := range events {
		fmt.Fprintln(tv, formatActivityEvent(e))
	}

	tv.ScrollToEnd()
}
