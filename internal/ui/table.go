package ui

import (
	"slices"

	"github.com/rivo/tview"
)

type column[T any] struct {
	header  string
	key     rune
	display func(T) string
	compare func(T, T) int
}

type tableState[T any] struct {
	cols    []column[T]
	items   []T
	sorted  []T
	sortCol int
	sortAsc bool
	resort  bool
}

func (s *tableState[T]) setItems(items []T) {
	s.items = items
	s.resort = true
}

func (s *tableState[T]) render(table *tview.Table) {
	table.Clear()

	for col, c := range s.cols {
		label := c.header

		if col == s.sortCol {
			if s.sortAsc {
				label += " ▲"
			} else {
				label += " ▼"
			}
		}

		table.SetCell(0, col, tview.NewTableCell(label).SetSelectable(false).SetExpansion(1))
	}

	if s.resort {
		s.sorted = slices.Clone(s.items)
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
