package pfddrawio

import (
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/xmldom"
)

const EstimateVolumeColumnSuffix = "作業量"

type EstimatePoint struct {
	Label string
	Value string
}

var estimateEntryPattern = regexp.MustCompile(`([^\s:：]+)[:：]\s*([0-9]+(?:\.[0-9]+)?)?d`)

func ParseEstimateBoxValue(value string) ([]EstimatePoint, error) {
	var sb strings.Builder
	if err := ParseValueHTML(ValueHTML(value), &sb); err != nil {
		return nil, fmt.Errorf("pfddrawio.ParseEstimateBoxValue: %w", err)
	}
	value = strings.Map(func(r rune) rune {
		if r == '\u00A0' || r == '\u3000' {
			return ' '
		}
		return r
	}, sb.String())

	matches := estimateEntryPattern.FindAllStringSubmatchIndex(value, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("pfddrawio.ParseEstimateBoxValue: no estimate points: %q", value)
	}

	points := make([]EstimatePoint, 0, len(matches))
	seen := make(map[string]bool, len(matches))
	prevEnd := 0
	for _, m := range matches {
		if rest := strings.TrimSpace(value[prevEnd:m[0]]); rest != "" {
			return nil, fmt.Errorf("pfddrawio.ParseEstimateBoxValue: unexpected text %q in %q", rest, value)
		}
		prevEnd = m[1]

		label := value[m[2]:m[3]]
		if seen[label] {
			return nil, fmt.Errorf("pfddrawio.ParseEstimateBoxValue: duplicated label %q in %q", label, value)
		}
		seen[label] = true

		num := ""
		if m[4] >= 0 {
			num = value[m[4]:m[5]]
		}
		points = append(points, EstimatePoint{Label: label, Value: num})
	}
	if rest := strings.TrimSpace(value[prevEnd:]); rest != "" {
		return nil, fmt.Errorf("pfddrawio.ParseEstimateBoxValue: unexpected text %q in %q", rest, value)
	}
	return points, nil
}

func EstimateBoxesByProcess(nodes []*xmldom.Node, logger *slog.Logger) (map[pfd.NodeID][]EstimateBox, error) {
	diagrams, err := ParseNodes(nodes, logger)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.EstimateBoxesByProcess: %w", err)
	}
	p, srcMap, err := NormalizeDiagrams("", diagrams, logger)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.EstimateBoxesByProcess: %w", err)
	}
	return EstimateBoxesByProcessWith(p, srcMap, nodes, logger), nil
}

func EstimateBoxesByProcessWith(p *pfd.PFD, srcMap *SourceMap, nodes []*xmldom.Node, logger *slog.Logger) map[pfd.NodeID][]EstimateBox {
	doms := CollectDiagramDOMs(nodes)
	boxesByDiagram := make(map[DiagramID][]EstimateBox)
	for _, b := range EstimateBoxes(nodes) {
		boxesByDiagram[b.DiagramID] = append(boxesByDiagram[b.DiagramID], b)
	}

	result := make(map[pfd.NodeID][]EstimateBox)
	for _, id := range p.AtomicProcesses().Slice() {
		locs, ok := srcMap.NodeIDMap[id]
		if !ok {
			logger.Warn("pfddrawio.EstimateBoxesByProcess: atomic process has no source location; skipped", "process", id)
			continue
		}
		for _, loc := range locs.Slice() {
			dom, ok := doms[loc.DiagramID]
			if !ok {
				logger.Warn("pfddrawio.EstimateBoxesByProcess: diagram not found in DOM; skipped", "process", id, "diagramID", loc.DiagramID)
				continue
			}
			procNode, ok := dom.CellsByID[loc.CellID]
			if !ok {
				logger.Warn("pfddrawio.EstimateBoxesByProcess: process cell not found in DOM; skipped", "process", id, "diagramID", loc.DiagramID, "cellID", loc.CellID)
				continue
			}
			procRect, ok := GeometryRect(procNode)
			if !ok {
				logger.Warn("pfddrawio.EstimateBoxesByProcess: process cell has no geometry; skipped", "process", id, "diagramID", loc.DiagramID, "cellID", loc.CellID)
				continue
			}
			for _, b := range boxesByDiagram[loc.DiagramID] {
				if IsBelow(procRect, b.Rect) {
					result[id] = append(result[id], b)
				}
			}
		}
	}
	return result
}

func EstimateTable(nodes []*xmldom.Node, logger *slog.Logger) (*pfd.AtomicProcessTable, error) {
	diagrams, err := ParseNodes(nodes, logger)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.EstimateTable: %w", err)
	}
	p, srcMap, err := NormalizeDiagrams("", diagrams, logger)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.EstimateTable: %w", err)
	}
	boxesByProcess := EstimateBoxesByProcessWith(p, srcMap, nodes, logger)

	type boxKey struct {
		DiagramID DiagramID
		CellID    CellID
	}
	attached := make(map[boxKey]bool)
	for _, boxes := range boxesByProcess {
		for _, b := range boxes {
			attached[boxKey{b.DiagramID, b.CellID}] = true
		}
	}
	var labels []string
	seenLabel := make(map[string]bool)
	for _, b := range EstimateBoxes(nodes) {
		if !attached[boxKey{b.DiagramID, b.CellID}] {
			logger.Warn("pfddrawio.EstimateTable: estimate box is not below any atomic process; ignored", "diagramID", b.DiagramID, "cellID", b.CellID, "value", b.Value)
			continue
		}
		points, err := ParseEstimateBoxValue(b.Value)
		if err != nil {
			continue
		}
		for _, point := range points {
			if !seenLabel[point.Label] {
				seenLabel[point.Label] = true
				labels = append(labels, point.Label)
			}
		}
	}

	nodeMap := pfd.NewNodeMap(p.Nodes, logger)
	table := pfd.NewAtomicProcessTable(p, nodeMap)
	table.ExtraHeaders = make([]string, 0, len(labels))
	for _, label := range labels {
		table.ExtraHeaders = append(table.ExtraHeaders, label+EstimateVolumeColumnSuffix)
	}
	slices.SortFunc(table.Rows, (*pfd.AtomicProcessRow).Compare)

	for _, row := range table.Rows {
		row.ExtraCells = make([]string, len(labels))
		boxes := boxesByProcess[pfd.NodeID(row.ID)]
		if len(boxes) == 0 {
			logger.Warn("pfddrawio.EstimateTable: atomic process has no estimate box", "process", row.ID)
			continue
		}

		values := make(map[string]string, len(labels))
		conflicted := make(map[string]bool, len(labels))
		for _, b := range boxes {
			points, err := ParseEstimateBoxValue(b.Value)
			if err != nil {
				logger.Warn("pfddrawio.EstimateTable: unparsable estimate box; ignored", "process", row.ID, "value", b.Value, "error", err)
				continue
			}
			for _, point := range points {
				if point.Value == "" {
					logger.Warn("pfddrawio.EstimateTable: estimate is not filled", "process", row.ID, "label", point.Label)
					continue
				}
				if prev, ok := values[point.Label]; ok && prev != point.Value {
					conflicted[point.Label] = true
					logger.Warn("pfddrawio.EstimateTable: conflicting estimate values", "process", row.ID, "label", point.Label, "values", []string{prev, point.Value})
					continue
				}
				values[point.Label] = point.Value
			}
		}
		for i, label := range labels {
			if conflicted[label] {
				continue
			}
			row.ExtraCells[i] = values[label]
		}
	}

	return table, nil
}
