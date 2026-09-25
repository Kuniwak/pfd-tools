package emphasis

import (
	"fmt"
	"io"
	"log/slog"
	"maps"
	"slices"
	"strings"

	"github.com/Kuniwak/pfd-tools/table/tabletsv"
)

const (
	MermaidTag    = "crit"
	PlantUMLColor = "Salmon"
)

const (
	ColumnHeader = "Emphasis"
	TrueValue    = "TRUE"
)

type Set struct {
	ids map[string]struct{}
}

func (s *Set) Enabled() bool {
	return s != nil
}

func MustParse(tsv string) *Set {
	s, err := Parse(strings.NewReader(tsv))
	if err != nil {
		panic(fmt.Sprintf("emphasis.MustParse: %s", err))
	}
	return s
}

func Parse(r io.Reader) (*Set, error) {
	table, err := tabletsv.ParseTable(r)
	if err != nil {
		return nil, fmt.Errorf("emphasis.Parse: %w", err)
	}

	idIndex := slices.Index(table.Header, tabletsv.IDColumnHeader)
	if idIndex < 0 {
		return nil, fmt.Errorf("emphasis.Parse: 強調 ID 表には %q 列が必要ですが、ヘッダは %v でした", tabletsv.IDColumnHeader, table.Header)
	}

	ids := make(map[string]struct{}, len(table.Rows))
	for _, row := range table.Rows {
		if row[idIndex] == "" {
			continue
		}
		ids[row[idIndex]] = struct{}{}
	}
	return &Set{ids: ids}, nil
}

func (s *Set) Contains(id string) bool {
	if s == nil {
		return false
	}
	_, ok := s.ids[id]
	return ok
}

func (s *Set) IDs() []string {
	if s == nil {
		return nil
	}
	return slices.Sorted(maps.Keys(s.ids))
}

func (s *Set) WarnUnknownIDs(logger *slog.Logger, known []string) {
	if s == nil {
		return
	}

	knownIDs := make(map[string]struct{}, len(known))
	for _, id := range known {
		knownIDs[id] = struct{}{}
	}

	var unknown []string
	for _, id := range s.IDs() {
		if _, ok := knownIDs[id]; !ok {
			unknown = append(unknown, id)
		}
	}
	if len(unknown) == 0 {
		return
	}

	logger.Warn("強調 ID 表に、強調対象に存在しない ID が含まれています（無視します）", "ids", strings.Join(unknown, ", "))
}
