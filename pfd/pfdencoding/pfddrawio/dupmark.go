package pfddrawio

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/xmldom"
)

const DuplicateMark = "＊"

const duplicateMarks = "＊*"

func IsDuplicateMark(r rune) bool {
	return strings.ContainsRune(duplicateMarks, r)
}

func SplitDuplicateMark(id pfd.NodeID) (pfd.NodeID, string) {
	base := pfd.NodeID(strings.TrimRightFunc(string(id), IsDuplicateMark))
	if base == id {
		return id, ""
	}
	if _, err := pfd.ParseDeliverableID(base); err != nil {
		return id, ""
	}

	mark := string(id[len(base):])
	return base, string([]rune(mark)[:1])
}

func TextOutsideTags(s string) (string, []int) {
	var text strings.Builder
	offsets := make([]int, 0, len(s))
	inTag := false
	for i := range len(s) {
		switch {
		case s[i] == '<':
			inTag = true
		case s[i] == '>':
			inTag = false
		case !inTag:
			text.WriteByte(s[i])
			offsets = append(offsets, i)
		}
	}
	return text.String(), offsets
}

func RewriteDuplicateMark(value ValueHTML, dup bool) (ValueHTML, bool) {
	var sb strings.Builder
	if err := ParseValueHTML(value, &sb); err != nil {
		return value, false
	}
	id, _, err := ParseVertexValue(sb.String())
	if err != nil {
		return value, false
	}
	if _, err := pfd.ParseDeliverableID(id); err != nil {
		return value, false
	}

	raw := string(value)
	text, offsets := TextOutsideTags(raw)
	idx := strings.Index(text, string(id))
	if idx < 0 {

		return value, false
	}
	baseEnd := offsets[idx+len(id)-1] + 1

	mark := ""
	var tags strings.Builder
	i := baseEnd
	for i < len(raw) {
		if raw[i] == '<' {
			closing := strings.IndexByte(raw[i:], '>')
			if closing < 0 {
				break
			}
			tags.WriteString(raw[i : i+closing+1])
			i += closing + 1
			continue
		}
		r, size := utf8.DecodeRuneInString(raw[i:])
		if !IsDuplicateMark(r) {
			break
		}
		if mark == "" {
			mark = string(r)
		}
		i += size
	}

	next := ""
	if dup {
		next = mark
		if next == "" {
			next = DuplicateMark
		}
	}
	return ValueHTML(raw[:baseEnd] + next + tags.String() + raw[i:]), true
}

func DuplicateMarks(vertices []LayoutVertex, edges []LayoutEdge) map[CellID]bool {
	rep := ContractDuplicates(vertices, edges)
	marks := make(map[CellID]bool, len(vertices))
	for _, v := range vertices {
		if _, ok := v.DuplicateGroup(); !ok {
			continue
		}
		marks[v.ID] = rep.Of(v.ID) != v.ID
	}
	return marks
}

var AllPages []string

type DupMarkOptions struct {
	OnlyPages []string
}

func ApplyDuplicateMarksToNodes(nodes []*xmldom.Node, onlyPages []string, logger *slog.Logger) int {
	layers := NewLayerMapFromNodes(nodes)
	parents := NewParentMapFromNodes(nodes)

	rewritten := 0
	for _, dom := range CollectDiagramDOMs(nodes) {
		if !ShouldSortPage(dom.Name, onlyPages) {
			continue
		}
		vertices, edges := CollectPageLayout(dom, layers, parents)
		for cellID, dup := range DuplicateMarks(vertices, edges) {
			cell, ok := dom.CellsByID[cellID]
			if !ok {
				continue
			}
			value, _ := cell.GetAttr("value", "")
			next, ok := RewriteDuplicateMark(ValueHTML(value), dup)
			if !ok {

				logger.Debug("pfddrawio.ApplyDuplicateMarksToNodes: skipped a cell without a markable deliverable ID", "cell", cellID, "value", value)
				continue
			}
			if string(next) == value {
				continue
			}
			cell.SetAttr("value", string(next))
			rewritten++
			logger.Debug("pfddrawio.ApplyDuplicateMarksToNodes: rewrote a label", "cell", cellID, "from", value, "to", string(next))
		}
	}
	return rewritten
}

func MarkDuplicates(r io.Reader, opts DupMarkOptions, logger *slog.Logger) ([]*xmldom.Node, error) {
	nodes, err := xmldom.ParseXML(r)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.MarkDuplicates: %w", err)
	}

	rewritten := ApplyDuplicateMarksToNodes(nodes, opts.OnlyPages, logger)
	logger.Debug("pfddrawio.MarkDuplicates: marked duplicated deliverables", "cells", rewritten)

	for _, want := range UnmatchedOnlyPages(opts.OnlyPages, PageNames(CollectDiagramDOMs(nodes))) {
		logger.Warn("pfddrawio.MarkDuplicates: -only-page で指定されたページが見つかりません", "page", want)
	}
	return nodes, nil
}
