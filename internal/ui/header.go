package ui

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"runtime/debug"
	"strings"
	"time"

	"github.com/rivo/tview"
)

const headerHeight = 5

type headerInfo struct {
	context string
	url     string
	write   bool
}

// sdkVersion returns the server-sdk-go version this binary was built with.
func sdkVersion() string {
	const path = "github.com/livekit/server-sdk-go/v2"

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, d := range info.Deps {
			if d.Path == path {
				return d.Version
			}
		}
	}

	return "n/a"
}

// lkCLIVersion returns the version of the local `lk` CLI, or "n/a" when it is
// not installed or does not answer in time.
func lkCLIVersion(ctx context.Context) string {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "lk", "--version").Output()
	if err != nil {
		return "n/a"
	}

	fields := strings.Fields(string(out))
	if len(fields) == 0 {
		return "n/a"
	}

	return fields[len(fields)-1]
}

func infoText(h headerInfo, version, sdk, lkCLI string) string {
	mode := "[green]RO"
	if h.write {
		mode = "[red::b]RW"
	}

	host := h.url
	if u, err := url.Parse(h.url); err == nil && u.Host != "" {
		host = u.Host
	}

	label := func(name string) string { return tagCyan + name + "[-:-:-]" }

	return fmt.Sprintf("%s %s [%s[-:-:-]]\n%s %s\n%s %s\n%s %s\n%s %s",
		label("Context:  "), h.context, mode,
		label("URL:      "), host,
		label("LK9s Rev: "), version,
		label("SDK:      "), sdk,
		label("lk CLI:   "), lkCLI)
}

// newHeader builds the k9s-style header: context info on the left, logo on
// the right. The returned func refreshes the context fields after a project
// switch.
func newHeader(version, lkCLI string) (*tview.Flex, func(headerInfo), func([][2]string)) {
	info := tview.NewTextView().SetDynamicColors(true)

	update := func(h headerInfo) { info.SetText(infoText(h, version, sdkVersion(), lkCLI)) }

	logoView := tview.NewTextView().SetDynamicColors(true).SetText(logo())

	hints := tview.NewTextView().SetDynamicColors(true)

	setHints := func(keys [][2]string) { hints.SetText(formatHints(keys, headerHeight)) }

	header := tview.NewFlex().
		AddItem(info, 44, 0, false).
		AddItem(hints, 0, 1, false).
		AddItem(logoView, 31, 0, false)

	return header, update, setHints
}
