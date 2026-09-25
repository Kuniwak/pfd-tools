package pfddrawio

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/Kuniwak/pfd-tools/geom"
	"github.com/Kuniwak/pfd-tools/xmldom"
)

type FixOptions struct {
	DeleteUnhit bool
	ExpandX     float64
	ExpandY     float64
}

func Fix(r io.Reader, opts FixOptions, logger *slog.Logger) ([]*xmldom.Node, error) {
	nodes, err := xmldom.ParseXML(r)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.Fix: %w", err)
	}

	layerMap := NewLayerMapFromNodes(nodes)
	parents := NewParentMapFromNodes(nodes)
	hitboxes, edges, edgeNodes := collectFixTargets(nodes, layerMap, parents, opts.ExpandX, opts.ExpandY)
	plan := PlanFix(hitboxes, edges)

	for _, c := range plan.Connections {
		node := edgeNodes[c.ID]
		node.SetAttr("source", string(c.Source))
		node.SetAttr("target", string(c.Target))
		RemoveEdgeEndpointPoints(node)
		logger.Debug("pfddrawio.Fix: connected edge", "cell", c.ID, "source", c.Source, "target", c.Target)
	}

	for _, id := range plan.Ambiguous {
		logger.Warn("pfddrawio.Fix: ambiguous edge endpoint (overlapping hitboxes); left unchanged", "cell", id)
	}

	deleteSet := make(map[CellID]struct{})
	for _, id := range plan.Unhit {
		if opts.DeleteUnhit {
			logger.Warn("pfddrawio.Fix: deleting unhit edge (endpoint not in any hitbox)", "cell", id)
			deleteSet[id] = struct{}{}
		} else {
			logger.Warn("pfddrawio.Fix: unhit edge (endpoint not in any hitbox); left unchanged", "cell", id)
		}
	}

	for _, id := range plan.SelfLoop {
		if opts.DeleteUnhit {
			logger.Warn("pfddrawio.Fix: deleting self-loop edge (source and target resolve to the same vertex)", "cell", id)
			deleteSet[id] = struct{}{}
		} else {
			logger.Warn("pfddrawio.Fix: self-loop edge (source and target resolve to the same vertex); left unchanged", "cell", id)
		}
	}

	if len(deleteSet) > 0 {
		removeCells(nodes, deleteSet)
	}

	return nodes, nil
}

func collectFixTargets(nodes []*xmldom.Node, layerMap LayerMap, parents ParentMap, expandX, expandY float64) ([]Hitbox, []EdgeGeom, map[CellID]*xmldom.Node) {
	var hitboxes []Hitbox
	var edges []EdgeGeom
	edgeNodes := make(map[CellID]*xmldom.Node)

	for _, root := range nodes {
		root.Traverse(func(n *xmldom.Node) {
			if n.Kind != xmldom.ElementNode || n.Start.Name.Local != "mxCell" {
				return
			}
			id, _ := n.GetAttr("id", "")
			cellID := CellID(id)
			if layerMap.IsCommentDescendant(parents, cellID) {
				return
			}

			if v, _ := n.GetAttr("vertex", ""); v == "1" {
				rect, ok := vertexHitboxRect(n)
				if !ok {
					return
				}
				hitboxes = append(hitboxes, Hitbox{ID: cellID, Rect: rect.Scale(expandX, expandY)})
				return
			}

			if e, _ := n.GetAttr("edge", ""); e == "1" {
				src, _ := n.GetAttr("source", "")
				tgt, _ := n.GetAttr("target", "")
				sp, tp := edgePoints(n)
				edges = append(edges, EdgeGeom{
					ID:          cellID,
					Source:      CellID(src),
					Target:      CellID(tgt),
					SourcePoint: sp,
					TargetPoint: tp,
				})
				edgeNodes[cellID] = n
			}
		}, nil)
	}

	return hitboxes, edges, edgeNodes
}

func vertexHitboxRect(n *xmldom.Node) (geom.Rect, bool) {
	styleStr, _ := n.GetAttr("style", "")
	styleMap, err := ParseStyle(styleStr)
	if err != nil {
		return geom.Rect{}, false
	}
	if !styleMap.IsRectangle() && !styleMap.IsEllipse() {
		return geom.Rect{}, false
	}
	return GeometryRect(n)
}

func edgePoints(n *xmldom.Node) (*geom.Point, *geom.Point) {
	geo := n.FirstChildElement("mxGeometry")
	if geo == nil {
		return nil, nil
	}

	var sp, tp *geom.Point
	for _, ch := range geo.Children {
		if ch.Kind != xmldom.ElementNode || ch.Start.Name.Local != "mxPoint" {
			continue
		}
		as, _ := ch.GetAttr("as", "")
		x, _ := ch.FloatAttr("x")
		y, _ := ch.FloatAttr("y")
		p := geom.Point{X: x, Y: y}
		switch as {
		case "sourcePoint":
			sp = &p
		case "targetPoint":
			tp = &p
		}
	}
	return sp, tp
}

func removeCells(nodes []*xmldom.Node, deleteSet map[CellID]struct{}) {
	var prune func(n *xmldom.Node)
	prune = func(n *xmldom.Node) {
		filtered := n.Children[:0]
		for _, ch := range n.Children {
			if ch.Kind == xmldom.ElementNode && ch.Start.Name.Local == "mxCell" {
				if id, _ := ch.GetAttr("id", ""); id != "" {
					if _, ok := deleteSet[CellID(id)]; ok {
						continue
					}
				}
			}
			filtered = append(filtered, ch)
		}
		n.Children = filtered
		for _, ch := range n.Children {
			prune(ch)
		}
	}
	for _, n := range nodes {
		prune(n)
	}
}

type Hitbox struct {
	ID   CellID
	Rect geom.Rect
}

type EdgeGeom struct {
	ID          CellID
	Source      CellID
	Target      CellID
	SourcePoint *geom.Point
	TargetPoint *geom.Point
}

type Connection struct {
	ID     CellID
	Source CellID
	Target CellID
}

type FixPlan struct {
	Connections []Connection
	Unhit       []CellID
	Ambiguous   []CellID

	SelfLoop []CellID
}

type sideStatus int

const (
	statusResolved sideStatus = iota
	statusUnhit
	statusAmbiguous
)

func PlanFix(hitboxes []Hitbox, edges []EdgeGeom) FixPlan {
	var plan FixPlan
	for _, e := range edges {
		if e.Source != "" && e.Target != "" {
			continue
		}

		srcID, srcStatus := resolveSide(e.Source, e.SourcePoint, hitboxes)
		tgtID, tgtStatus := resolveSide(e.Target, e.TargetPoint, hitboxes)

		switch {
		case srcStatus == statusAmbiguous || tgtStatus == statusAmbiguous:
			plan.Ambiguous = append(plan.Ambiguous, e.ID)
		case srcStatus == statusUnhit || tgtStatus == statusUnhit:
			plan.Unhit = append(plan.Unhit, e.ID)
		case srcID == tgtID:

			plan.SelfLoop = append(plan.SelfLoop, e.ID)
		default:
			plan.Connections = append(plan.Connections, Connection{ID: e.ID, Source: srcID, Target: tgtID})
		}
	}
	return plan
}

func resolveSide(existing CellID, pt *geom.Point, hitboxes []Hitbox) (CellID, sideStatus) {
	if existing != "" {
		return existing, statusResolved
	}
	if pt == nil {
		return "", statusUnhit
	}

	var matched CellID
	count := 0
	for _, h := range hitboxes {
		if h.Rect.Contains(*pt) {
			count++
			matched = h.ID
		}
	}

	switch count {
	case 0:
		return "", statusUnhit
	case 1:
		return matched, statusResolved
	default:
		return "", statusAmbiguous
	}
}
