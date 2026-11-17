package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Colors from the LiveKit brand assets (LK_icon_darkbg.svg).
var colorCyan = tcell.NewHexColor(0x1FD5F9)

const (
	tagCyan  = "[#1fd5f9]"
	tagWhite = "[white]"
)

func applyTheme() {
	tview.Styles.BorderColor = colorCyan
	tview.Styles.TitleColor = colorCyan
	tview.Styles.GraphicsColor = colorCyan
}

// logoCells is the LiveKit icon followed by "9s", on a 15x5 grid: W is white,
// C the cyan chevron.
var logoCells = [5]string{
	"WW....C.WWW....",
	"WW...C..W.W..WW",
	"WW..C...WWW.W..",
	"WW...C....W...W",
	"WWWW..C.WWW.WW.",
}

func logo() string {
	var b strings.Builder

	for i, row := range logoCells {
		if i > 0 {
			b.WriteByte('\n')
		}

		for _, c := range row {
			switch c {
			case 'W':
				b.WriteString(tagWhite + "██")
			case 'C':
				b.WriteString(tagCyan + "██")
			default:
				b.WriteString("  ")
			}
		}
	}

	return b.String()
}
