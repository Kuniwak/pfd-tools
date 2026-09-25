package cmd

import (
	"strings"

	"github.com/Kuniwak/pfd-tools/spreadsheet"
)

func NeededResourceSets(headCount int) string {
	parts := make([]string, 0, headCount)
	for i := 1; i <= headCount; i++ {
		parts = append(parts, spreadsheet.ColumnLabel(i)+":1")
	}
	return strings.Join(parts, ";")
}
