package ui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type column[T any] struct {
	header  string
	key     rune
	display func(T) string
	compare func(T, T) int
}

type tableState[T any] struct {
	title      string // resource name shown in the border title, e.g. "rooms"
	scope      string // optional parent shown as "title(scope)"
	cols       []column[T]
	items      []T
	sorted     []T
	sortCol    int
	sortAsc    bool
	resort     bool
	filterText string
}

// headerCell is the styled, non-selectable column header cell.
func headerCell(label string) *tview.TableCell {
	return tview.NewTableCell(label).SetSelectable(false).SetExpansion(1).
		SetTextColor(colorCyan).SetAttributes(tcell.AttrBold)
}

func (s *tableState[T]) setItems(items []T) {
	s.items = items
	s.resort = true
}

// setFilter sets a case-insensitive substring filter matched against every
// column's display text; an empty filter shows all rows.
func (s *tableState[T]) setFilter(f string) {
	f = strings.ToLower(f)
	if f == s.filterText {
		return
	}

	s.filterText = f
	s.resort = true
}

func (s *tableState[T]) matches(item T) bool {
	if s.filterText == "" {
		return true
	}

	for _, c := range s.cols {
		if strings.Contains(strings.ToLower(c.display(item)), s.filterText) {
			return true
		}
	}

	return false
}

func (s *tableState[T]) render(table *tview.Table) {
	table.Clear()

	for col, c := range s.cols {
		label := c.header

		if col == s.sortCol {
			if s.sortAsc {
				label += "↑"
			} else {
				label += "↓"
			}
		}

		table.SetCell(0, col, headerCell(label))
	}

	if s.resort {
		filtered := make([]T, 0, len(s.items))

		for _, item := range s.items {
			if s.matches(item) {
				filtered = append(filtered, item)
			}
		}

		s.sorted = filtered
		slices.SortFunc(s.sorted, func(a, b T) int {
			n := s.cols[s.sortCol].compare(a, b)
			if !s.sortAsc {
				return -n
			}

			return n
		})
		s.resort = false
	}

	for row, item := range s.sorted {
		for col, c := range s.cols {
			table.SetCell(row+1, col, tview.NewTableCell(c.display(item)).SetExpansion(1))
		}
	}

	// tview loops forever when a selectable table has no selectable cell.
	table.SetSelectable(len(s.sorted) > 0, false)

	if s.title != "" {
		table.SetTitle(s.borderTitle())
	}
}

// borderTitle formats the k9s-style table title: name(scope)</filter>[count].
func (s *tableState[T]) borderTitle() string {
	var b strings.Builder

	b.WriteString(" " + s.title)

	if s.scope != "" {
		b.WriteString("(" + s.scope + ")")
	}

	if s.filterText != "" {
		b.WriteString("</" + s.filterText + ">")
	}

	fmt.Fprintf(&b, "[%d] ", len(s.sorted))

	return b.String()
}

func (s *tableState[T]) handleKey(r rune) bool {
	for i, c := range s.cols {
		if c.key != r {
			continue
		}

		if i == s.sortCol {
			s.sortAsc = !s.sortAsc
		} else {
			s.sortCol = i
			s.sortAsc = true
		}

		s.resort = true

		return true
	}

	return false
}
