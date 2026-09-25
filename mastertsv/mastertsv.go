package mastertsv

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/table/tabletsv"
)

const (
	RowColumnHeader = "MasterRow"
	BarColumnHeader = "MasterBar"
)

type Row string

func (r Row) Compare(b Row) int {
	return strings.Compare(string(r), string(b))
}

type Bar string

func (b Bar) Compare(c Bar) int {
	return strings.Compare(string(b), string(c))
}

type Classification struct {
	ID   pfd.AtomicProcessID
	Rows *sets.Set[Row]
	Bar  Bar
}

type Table struct {
	Classifications []*Classification
}

func Parse(r io.Reader) (*Table, error) {
	table, err := tabletsv.ParseTable(r)
	if err != nil {
		return nil, fmt.Errorf("mastertsv.Parse: %w", err)
	}

	idIndex := slices.Index(table.Header, tabletsv.IDColumnHeader)
	if idIndex < 0 {
		return nil, fmt.Errorf("mastertsv.Parse: 分類表には %q 列が必要ですが、ヘッダは %v でした", tabletsv.IDColumnHeader, table.Header)
	}
	rowIndex := slices.Index(table.Header, RowColumnHeader)
	if rowIndex < 0 {
		return nil, fmt.Errorf("mastertsv.Parse: 分類表には %q 列が必要ですが、ヘッダは %v でした", RowColumnHeader, table.Header)
	}
	barIndex := slices.Index(table.Header, BarColumnHeader)
	if barIndex < 0 {
		return nil, fmt.Errorf("mastertsv.Parse: 分類表には %q 列が必要ですが、ヘッダは %v でした", BarColumnHeader, table.Header)
	}

	seen := make(map[pfd.AtomicProcessID]struct{}, len(table.Rows))
	cs := make([]*Classification, 0, len(table.Rows))
	for _, row := range table.Rows {
		id := pfd.AtomicProcessID(strings.TrimSpace(row[idIndex]))
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			return nil, fmt.Errorf("mastertsv.Parse: ID %q が複数行に現れています（1 つの ID は 1 行で分類してください）", id)
		}
		seen[id] = struct{}{}

		rows := ParseRows(row[rowIndex])
		bar := Bar(strings.TrimSpace(row[barIndex]))
		if rows.Len() == 0 || bar == "" {
			continue
		}
		cs = append(cs, &Classification{ID: id, Rows: rows, Bar: bar})
	}
	return &Table{Classifications: cs}, nil
}

const (
	DescriptionColumnHeaderEn = "Description"
	DescriptionColumnHeaderJa = "説明"
)

func ParseDescriptions(r io.Reader) (map[string]string, error) {
	table, err := tabletsv.ParseTable(r)
	if err != nil {
		return nil, fmt.Errorf("mastertsv.ParseDescriptions: %w", err)
	}

	idIndex := slices.Index(table.Header, tabletsv.IDColumnHeader)
	if idIndex < 0 {
		return nil, fmt.Errorf("mastertsv.ParseDescriptions: メタ表には %q 列が必要ですが、ヘッダは %v でした", tabletsv.IDColumnHeader, table.Header)
	}
	descIndex := slices.Index(table.Header, DescriptionColumnHeaderEn)
	if descIndex < 0 {
		descIndex = slices.Index(table.Header, DescriptionColumnHeaderJa)
	}
	if descIndex < 0 {
		return nil, fmt.Errorf("mastertsv.ParseDescriptions: メタ表には %q または %q 列が必要ですが、ヘッダは %v でした", DescriptionColumnHeaderEn, DescriptionColumnHeaderJa, table.Header)
	}

	descs := make(map[string]string, len(table.Rows))
	for _, row := range table.Rows {
		id := strings.TrimSpace(row[idIndex])
		if id == "" {
			continue
		}
		descs[id] = row[descIndex]
	}
	return descs, nil
}

func RequireKnownIDs(t *Table, known *sets.Set[pfd.AtomicProcessID]) error {
	for _, c := range t.Classifications {
		if !known.Contains(pfd.AtomicProcessID.Compare, c.ID) {
			return fmt.Errorf("mastertsv.RequireKnownIDs: 分類表に未知の ID %q があります（タイポか、別の対象の分類表を疑ってください）", c.ID)
		}
	}
	return nil
}

func ParseRows(s string) *sets.Set[Row] {
	rows := sets.New(Row.Compare)
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		rows.Add(Row.Compare, Row(part))
	}
	return rows
}
