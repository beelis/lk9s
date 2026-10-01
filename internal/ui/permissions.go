package ui

import (
	"strings"

	"github.com/beelis/lk9s/internal/lk"
	"github.com/rivo/tview"
)

// permissionsPage shows a participant's permissions as an editable form.
// Saving is gated by the context's write setting. CanPublishSources isn't
// editable here (it's a source list, not a flag) but is preserved
// unchanged on save. onSaved is called on the UI thread after a successful
// save, so the caller can refresh its own view.
func permissionsPage(n nav, room, identity string, perm lk.Permission, onSaved func()) tview.Primitive {
	const pageName = "permissions"

	form := tview.NewForm()
	form.SetBorder(true).SetTitle(" permissions: " + identity + " ")

	cancel := func() { n.pages.RemovePage(pageName) }

	form.AddCheckbox("CanPublish", perm.CanPublish, nil)
	form.AddCheckbox("CanSubscribe", perm.CanSubscribe, nil)
	form.AddCheckbox("CanPublishData", perm.CanPublishData, nil)
	form.AddCheckbox("CanUpdateMetadata", perm.CanUpdateMetadata, nil)
	form.AddCheckbox("CanSubscribeMetrics", perm.CanSubscribeMetrics, nil)
	form.AddCheckbox("CanManageAgentSession", perm.CanManageAgentSession, nil)
	form.AddCheckbox("Hidden", perm.Hidden, nil)
	form.AddCheckbox("Recorder", perm.Recorder, nil)

	sources := "(all)"
	if len(perm.CanPublishSources) > 0 {
		sources = strings.Join(perm.CanPublishSources, ", ")
	}

	form.AddTextView("CanPublishSources", sources, 40, 1, true, false)

	form.AddTextView("", "", 40, 1, true, false)
	statusIdx := form.GetFormItemCount() - 1

	setStatus := func(msg string) {
		if tv, ok := form.GetFormItem(statusIdx).(*tview.TextView); ok {
			tv.SetText(msg)
		}
	}

	if !n.write {
		setStatus("[yellow] write actions disabled for this context (set `write: true` in ~/.lk9s.yaml)")
	}

	checkbox := func(label string) bool {
		return form.GetFormItemByLabel(label).(*tview.Checkbox).IsChecked()
	}

	save := func() {
		if !n.write {
			setStatus("[yellow] write actions disabled for this context (set `write: true` in ~/.lk9s.yaml)")

			return
		}

		updated := lk.Permission{
			CanPublish:            checkbox("CanPublish"),
			CanSubscribe:          checkbox("CanSubscribe"),
			CanPublishData:        checkbox("CanPublishData"),
			CanUpdateMetadata:     checkbox("CanUpdateMetadata"),
			CanSubscribeMetrics:   checkbox("CanSubscribeMetrics"),
			CanManageAgentSession: checkbox("CanManageAgentSession"),
			Hidden:                checkbox("Hidden"),
			Recorder:              checkbox("Recorder"),
			CanPublishSources:     perm.CanPublishSources,
		}

		setStatus("saving...")

		go func() {
			err := n.client.UpdatePermission(n.ctx, room, identity, updated)

			n.app.QueueUpdateDraw(func() {
				if err != nil {
					setStatus("[red] " + err.Error())

					return
				}

				n.pages.RemovePage(pageName)

				if onSaved != nil {
					onSaved()
				}
			})
		}()
	}

	form.AddButton("save", save)
	form.AddButton("cancel", cancel)
	form.SetCancelFunc(cancel)

	return centered(form, 16, 60)
}
