package ui

import (
	"slices"
	"strings"

	"github.com/rivo/tview"
)

type column[T any] struct {
	header  string
	key     rune
	display func(T) string
	compare func(T, T) int
}

type tableState[T any] struct {
	cols       []column[T]
	items      []T
	sorted     []T
	sortCol    int
	sortAsc    bool
	resort     bool
	filterText string
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
				label += " ▲"
			} else {
				label += " ▼"
			}
		}

		table.SetCell(0, col, tview.NewTableCell(label).SetSelectable(false).SetExpansion(1))
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
