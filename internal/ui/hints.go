package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

// hinted is a page root that carries the key hints shown in the header while
// it is the front page.
type hinted struct {
	*tview.Flex
	keys [][2]string
}

func withHints(f *tview.Flex, keys [][2]string) tview.Primitive {
	return &hinted{Flex: f, keys: keys}
}

// formatHints lays out "<key> action" pairs in columns of rows lines each,
// like the k9s header.
func formatHints(keys [][2]string, rows int) string {
	lines := make([]strings.Builder, rows)

	for start := 0; start < len(keys); start += rows {
		chunk := keys[start:min(start+rows, len(keys))]
		kw, aw := 0, 0

		for _, k := range chunk {
			kw = max(kw, len(k[0])+2)
			aw = max(aw, len([]rune(k[1])))
		}

		for r := range rows {
			if r >= len(chunk) {
				fmt.Fprint(&lines[r], strings.Repeat(" ", kw+aw+4))

				continue
			}

			fmt.Fprintf(&lines[r], "%s%-*s[-] %-*s   ",
				tagCyan, kw, tview.Escape("<"+chunk[r][0]+">"), aw, tview.Escape(chunk[r][1]))
		}
	}

	out := make([]string, rows)
	for i := range lines {
		out[i] = lines[i].String()
	}

	return strings.Join(out, "\n")
}
