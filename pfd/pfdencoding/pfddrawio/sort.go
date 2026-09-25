package pfddrawio

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/Kuniwak/pfd-tools/geom"
	"github.com/Kuniwak/pfd-tools/graph"
	"github.com/Kuniwak/pfd-tools/pairs"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/xmldom"
)

type PosRestriction int

const (
	PosLock PosRestriction = iota

	PosFreeV

	PosFreeH

	PosFree
)

func ParsePosRestriction(s string) (PosRestriction, error) {
	switch s {
	case "lock":
		return PosLock, nil
	case "free-v":
		return PosFreeV, nil
	case "free-h":
		return PosFreeH, nil
	case "free":
		return PosFree, nil
	default:
		return 0, fmt.Errorf("pfddrawio.ParsePosRestriction: 不正な位置制約です（lock|free-v|free-h|free）: %q", s)
	}
}

type SortOptions struct {
	DupRankSpan int

	DupRowSpan int

	HGap float64

	VGap float64

	OnlyPages []string

	OnlyNodes []pfd.NodeID

	PosRestriction PosRestriction

	BreakCycles bool
}

func ShouldSortPage(name string, onlyPages []string) bool {
	if len(onlyPages) == 0 {
		return true
	}
	for _, want := range onlyPages {
		if PageNameElementID(want) == PageNameElementID(name) {
			return true
		}
		if IsContextDiagramName(want) && IsContextDiagramName(name) {
			return true
		}
	}
	return false
}

func UnmatchedOnlyPages(onlyPages, pageNames []string) []string {
	var unmatched []string
	for _, want := range onlyPages {
		matched := false
		for _, name := range pageNames {
			if ShouldSortPage(name, []string{want}) {
				matched = true
				break
			}
		}
		if !matched {
			unmatched = append(unmatched, want)
		}
	}
	return unmatched
}

func UnmatchedOnlyNodes(onlyNodes, nodeIDs []pfd.NodeID) []pfd.NodeID {
	present := make(map[pfd.NodeID]struct{}, len(nodeIDs))
	for _, id := range nodeIDs {
		present[id] = struct{}{}
	}
	var unmatched []pfd.NodeID
	for _, want := range onlyNodes {
		if _, ok := present[want]; !ok {
			unmatched = append(unmatched, want)
		}
	}
	return unmatched
}

func SelectCells(vertices []LayoutVertex, onlyNodes []pfd.NodeID) map[CellID]struct{} {
	want := make(map[pfd.NodeID]struct{}, len(onlyNodes))
	for _, id := range onlyNodes {
		want[id] = struct{}{}
	}
	selected := make(map[CellID]struct{})
	for _, v := range vertices {
		if _, ok := want[v.NodeID]; ok {
			selected[v.ID] = struct{}{}
		}
	}
	return selected
}

func (o SortOptions) IsOnlyNodeMode() bool {
	return len(o.OnlyNodes) > 0
}

func DefaultSortOptions() SortOptions {
	return SortOptions{DupRankSpan: 1, DupRowSpan: 5, HGap: 80, VGap: 40}
}

type VertexKind int

const (
	VertexDeliverable VertexKind = iota

	VertexProcess

	VertexConnector
)

func (k VertexKind) IsProcessLike() bool {
	return k == VertexProcess || k == VertexConnector
}

type LayoutVertex struct {
	ID     CellID
	NodeID pfd.NodeID
	Kind   VertexKind
	Rect   geom.Rect
}

type LayoutEdge struct {
	ID         CellID
	Source     CellID
	Target     CellID
	IsFeedback bool
	IsCut      bool
}

func (v LayoutVertex) Label() string {
	if v.NodeID != "" {
		return string(v.NodeID)
	}
	return string(v.ID)
}

func (e LayoutEdge) IsForward() bool {
	return !e.IsFeedback && !e.IsCut
}

func classifyVertexStyle(style StyleMap, value string) (VertexKind, bool) {
	if value == "" && style.IsEllipse() && style.Get("aspect") == "fixed" {
		return VertexConnector, true
	}
	if style.IsRectangle() {
		return VertexDeliverable, true
	}
	if style.IsEllipse() {
		return VertexProcess, true
	}
	return 0, false
}

func CollectPageLayout(dom *DiagramDOM, layers LayerMap, parents ParentMap) ([]LayoutVertex, []LayoutEdge) {
	vertices := make([]LayoutVertex, 0, len(dom.CellsByID))
	for cellID, cell := range dom.CellsByID {
		if layers.IsCommentDescendant(parents, cellID) {
			continue
		}
		if v, _ := cell.GetAttr("vertex", ""); v != "1" {
			continue
		}
		styleStr, _ := cell.GetAttr("style", "")
		style, err := ParseStyle(styleStr)
		if err != nil {
			continue
		}
		value, _ := cell.GetAttr("value", "")
		kind, ok := classifyVertexStyle(style, value)
		if !ok {
			continue
		}
		rect, ok := GeometryRect(cell)
		if !ok {
			continue
		}

		nodeID, _ := VertexNodeID(cell)
		vertices = append(vertices, LayoutVertex{ID: cellID, NodeID: nodeID, Kind: kind, Rect: rect})
	}
	sort.Slice(vertices, func(i, j int) bool {
		return vertices[i].ID.Compare(vertices[j].ID) < 0
	})

	vertexSet := make(map[CellID]struct{}, len(vertices))
	for _, v := range vertices {
		vertexSet[v.ID] = struct{}{}
	}

	edges := make([]LayoutEdge, 0, len(dom.CellsByID))
	for cellID, cell := range dom.CellsByID {
		if layers.IsCommentDescendant(parents, cellID) {
			continue
		}
		if e, _ := cell.GetAttr("edge", ""); e != "1" {
			continue
		}
		src, _ := cell.GetAttr("source", "")
		tgt, _ := cell.GetAttr("target", "")
		if src == "" || tgt == "" {
			continue
		}
		if _, ok := vertexSet[CellID(src)]; !ok {
			continue
		}
		if _, ok := vertexSet[CellID(tgt)]; !ok {
			continue
		}
		styleStr, _ := cell.GetAttr("style", "")
		style, err := ParseStyle(styleStr)
		if err != nil {

			continue
		}
		edges = append(edges, LayoutEdge{
			ID:         cellID,
			Source:     CellID(src),
			Target:     CellID(tgt),
			IsFeedback: style.IsDashed(),
		})
	}
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].ID.Compare(edges[j].ID) < 0
	})
	return vertices, edges
}

func ForwardGraph(vertices []LayoutVertex, edges []LayoutEdge) *graph.Graph {
	edgeCmp := pairs.Compare(graph.Node.Compare, graph.Node.Compare)
	g := &graph.Graph{
		Nodes: sets.New(graph.Node.Compare),
		Edges: sets.New(edgeCmp),
	}
	for _, v := range vertices {
		g.Nodes.Add(graph.Node.Compare, graph.Node(v.ID))
	}
	for _, e := range edges {
		if !e.IsForward() {
			continue
		}
		g.Edges.Add(edgeCmp, pairs.New(graph.Node(e.Source), graph.Node(e.Target)))
	}
	return g
}

func MarkCycleCuts(vertices []LayoutVertex, edges []LayoutEdge) []LayoutEdge {
	out := make([]LayoutEdge, len(edges))
	copy(out, edges)

	kind := make(map[CellID]VertexKind, len(vertices))
	centerX := make(map[CellID]float64, len(vertices))
	for _, v := range vertices {
		kind[v.ID] = v.Kind
		centerX[v.ID] = v.Rect.X + v.Rect.Width/2
	}

	candidates := make([]int, 0, len(out))
	for i, e := range out {
		if !e.IsForward() {
			continue
		}

		src, okSrc := kind[e.Source]
		tgt, okTgt := kind[e.Target]
		if !okSrc || !okTgt || src != VertexDeliverable || !tgt.IsProcessLike() {
			continue
		}
		candidates = append(candidates, i)
	}
	sort.SliceStable(candidates, func(a, b int) bool {
		ea, eb := out[candidates[a]], out[candidates[b]]
		backA := centerX[ea.Source] - centerX[ea.Target]
		backB := centerX[eb.Source] - centerX[eb.Target]
		if backA != backB {
			return backA > backB
		}
		return ea.ID.Compare(eb.ID) < 0
	})

	forward := ForwardGraph(vertices, out)
	onCycle := func(i int) bool {

		return forward.IsReachable(graph.Node(out[i].Target), graph.Node(out[i].Source))
	}

	for _, i := range candidates {
		if !onCycle(i) {
			continue
		}
		out[i].IsCut = true
		forward = ForwardGraph(vertices, out)
	}

	for j := len(candidates) - 1; j >= 0; j-- {
		i := candidates[j]
		if !out[i].IsCut {
			continue
		}
		out[i].IsCut = false
		forward = ForwardGraph(vertices, out)
		if onCycle(i) {
			out[i].IsCut = true
			forward = ForwardGraph(vertices, out)
		}
	}
	return out
}

func CycleHint(vertices []LayoutVertex, edges []LayoutEdge, broke bool) string {
	label := make(map[CellID]string, len(vertices))
	for _, v := range vertices {
		label[v.ID] = v.Label()
	}

	rep := ContractDuplicates(vertices, edges)
	contracted, contractedEdges := ContractLayout(vertices, edges, rep)
	var cuts []string
	for _, e := range MarkCycleCuts(contracted, contractedEdges) {
		if e.IsCut {
			cuts = append(cuts, label[e.Source]+"→"+label[e.Target])
		}
	}

	hint := "循環を作っている辺を特定できませんでした（成果物とプロセスを交互に結ばない辺が混じっていないか pfdlint で確認してください）"
	switch {
	case len(cuts) > 0 && broke:
		hint = "レイアウトのために " + strings.Join(cuts, ", ") +
			" を切りましたが、まだ循環が残っています（成果物→プロセスの実線を含まない循環は切れません。成果物とプロセスを交互に結ばない辺が混じっていないか pfdlint で確認してください）"
	case len(cuts) > 0:
		hint = "循環上の辺: " + strings.Join(cuts, ", ") +
			"（-break-cycles を付けると、レイアウトのためにこれらを切って成果物を 2 箱に分けて整列します。図の意味は変わりません）"
	}

	for _, v := range vertices {
		if rep.Of(v.ID) != v.ID {
			hint += "。同一 ID の成果物セルが複数あるので、複製表示のセルが論理レベルで循環を作っていないかも pfdlint で確認してください"
			break
		}
	}
	return hint
}

func AssignRanks(vertices []LayoutVertex, edges []LayoutEdge) (map[CellID]int, error) {
	order := ForwardGraph(vertices, edges).TopologicalSort()
	if order == nil {
		return nil, fmt.Errorf("pfddrawio.AssignRanks: %w", ErrCycleRemains)
	}

	succ := make(map[CellID][]CellID)
	for _, e := range edges {
		if !e.IsForward() {
			continue
		}
		succ[e.Source] = append(succ[e.Source], e.Target)
	}

	ranks := make(map[CellID]int, len(vertices))
	for _, v := range vertices {
		ranks[v.ID] = 0
	}
	for _, u := range order {
		ru := ranks[CellID(u)]
		for _, w := range succ[CellID(u)] {
			if ru+1 > ranks[w] {
				ranks[w] = ru + 1
			}
		}
	}
	return ranks, nil
}

func (v LayoutVertex) DuplicateGroup() (pfd.NodeID, bool) {
	if v.Kind != VertexDeliverable || v.NodeID == "" {
		return "", false
	}
	return v.NodeID, true
}

type Representatives map[CellID]CellID

func (r Representatives) Of(id CellID) CellID {
	if rep, ok := r[id]; ok {
		return rep
	}
	return id
}

func (r Representatives) SameGroup(a, b CellID) bool {
	return r.Of(a) == r.Of(b)
}

func DuplicateCellID(source CellID, n int) CellID {
	return source + CellID("-dup"+strconv.Itoa(n+1))
}

func NextDuplicateCellID(source CellID, taken map[CellID]struct{}) CellID {
	next := 0

	for n := 0; n <= len(taken); n++ {
		if _, used := taken[DuplicateCellID(source, n)]; used {
			next = n + 1
		}
	}
	return DuplicateCellID(source, next)
}

func ContractDuplicates(vertices []LayoutVertex, edges []LayoutEdge) Representatives {
	produced := make(map[CellID]bool, len(vertices))
	kind := make(map[CellID]VertexKind, len(vertices))
	for _, v := range vertices {
		kind[v.ID] = v.Kind
	}
	for _, e := range edges {
		if !e.IsFeedback && kind[e.Target] == VertexDeliverable {
			produced[e.Target] = true
		}
	}

	best := make(map[pfd.NodeID]LayoutVertex)
	for _, v := range vertices {
		group, ok := v.DuplicateGroup()
		if !ok {
			continue
		}
		cur, seen := best[group]
		if !seen || PrefersAsRepresentative(v.ID, cur.ID, produced) {
			best[group] = v
		}
	}

	rep := make(Representatives, len(vertices))
	for _, v := range vertices {
		rep[v.ID] = v.ID
		if group, ok := v.DuplicateGroup(); ok {
			rep[v.ID] = best[group].ID
		}
	}
	return rep
}

func PrefersAsRepresentative(candidate, current CellID, produced map[CellID]bool) bool {
	if produced[candidate] != produced[current] {
		return produced[candidate]
	}
	return candidate.Compare(current) < 0
}

func SeveredDuplicateGroups(vertices []LayoutVertex, edges []LayoutEdge) []pfd.NodeID {
	return SeveredDuplicateGroupsBy(vertices, edges, func(LayoutVertex) float64 { return 0 })
}

func SameColumnSeveredDuplicateGroups(vertices []LayoutVertex, edges []LayoutEdge) []pfd.NodeID {
	return SeveredDuplicateGroupsBy(vertices, edges, func(v LayoutVertex) float64 { return v.Rect.X })
}

func SeveredDuplicateGroupsBy(vertices []LayoutVertex, edges []LayoutEdge, column func(LayoutVertex) float64) []pfd.NodeID {
	solidIn := make(map[CellID]int, len(vertices))
	solidOut := make(map[CellID]int, len(vertices))
	for _, e := range edges {
		if e.IsFeedback {
			continue
		}
		solidOut[e.Source]++
		solidIn[e.Target]++
	}

	type groupColumn struct {
		Group  pfd.NodeID
		Column float64
	}
	leaf := make(map[groupColumn]bool)
	orphan := make(map[groupColumn]bool)
	for _, v := range vertices {
		group, ok := v.DuplicateGroup()
		if !ok {
			continue
		}
		key := groupColumn{Group: group, Column: column(v)}
		if solidIn[v.ID] > 0 && solidOut[v.ID] == 0 {
			leaf[key] = true
		}
		if solidIn[v.ID] == 0 && solidOut[v.ID] > 0 {
			orphan[key] = true
		}
	}

	var severed []pfd.NodeID
	for key := range leaf {
		if orphan[key] {
			severed = append(severed, key.Group)
		}
	}
	slices.Sort(severed)
	return slices.Compact(severed)
}

func ExtraDuplicateCells(vertices []LayoutVertex, rep Representatives) map[CellID][]CellID {
	extras := make(map[CellID][]CellID)
	for _, v := range vertices {
		if _, ok := v.DuplicateGroup(); !ok {
			continue
		}
		if rep.Of(v.ID) == v.ID {
			continue
		}
		extras[rep.Of(v.ID)] = append(extras[rep.Of(v.ID)], v.ID)
	}
	for _, ids := range extras {
		slices.SortFunc(ids, CellID.Compare)
	}
	return extras
}

func ContractLayout(vertices []LayoutVertex, edges []LayoutEdge, rep Representatives) ([]LayoutVertex, []LayoutEdge) {
	contracted := make([]LayoutVertex, 0, len(vertices))
	for _, v := range vertices {
		if rep.Of(v.ID) == v.ID {
			contracted = append(contracted, v)
		}
	}

	var contractedEdges []LayoutEdge
	for _, e := range edges {
		source := rep.Of(e.Source)
		target := rep.Of(e.Target)
		if source == target {
			continue
		}
		contractedEdges = append(contractedEdges, LayoutEdge{
			ID: e.ID, Source: source, Target: target, IsFeedback: e.IsFeedback, IsCut: e.IsCut,
		})
	}
	return contracted, contractedEdges
}

func ExpandRanks(ranks map[CellID]int, rep Representatives) map[CellID]int {
	expanded := make(map[CellID]int, len(rep))
	for id := range rep {
		expanded[id] = ranks[rep.Of(id)]
	}
	return expanded
}

func SpreadSourceRanks(ranks map[CellID]int, vertices []LayoutVertex, edges []LayoutEdge, rep Representatives) map[CellID]int {
	hasProducer := make(map[CellID]struct{}, len(vertices))
	consumerRank := make(map[CellID]int, len(vertices))
	feedbackRank := make(map[CellID]int, len(vertices))
	for _, e := range edges {
		if rep.SameGroup(e.Source, e.Target) {

			continue
		}
		if e.IsCut {

			continue
		}
		if e.IsFeedback {

			if r, ok := feedbackRank[e.Source]; !ok || ranks[e.Target] > r {
				feedbackRank[e.Source] = ranks[e.Target]
			}
			continue
		}
		hasProducer[e.Target] = struct{}{}
		if r, ok := consumerRank[e.Source]; !ok || ranks[e.Target] < r {
			consumerRank[e.Source] = ranks[e.Target]
		}
	}

	spread := make(map[CellID]int, len(ranks))
	for id, r := range ranks {
		spread[id] = r
	}
	for _, v := range vertices {
		if v.Kind != VertexDeliverable {
			continue
		}
		if _, ok := hasProducer[v.ID]; ok {
			continue
		}
		if nearest, ok := consumerRank[v.ID]; ok {
			spread[v.ID] = max(ranks[v.ID], nearest-1)
			continue
		}

		if farthest, ok := feedbackRank[v.ID]; ok {
			spread[v.ID] = farthest + 1
		}
	}
	return spread
}

func ValidateEdgeInvariants(vertices []LayoutVertex, edges []LayoutEdge) error {
	rep := ContractDuplicates(vertices, edges)
	kind := make(map[CellID]VertexKind, len(vertices))
	for _, v := range vertices {
		kind[v.ID] = v.Kind
	}
	for _, e := range edges {
		if e.IsFeedback {

			if kind[e.Source] != VertexDeliverable || kind[e.Target] == VertexDeliverable {
				return fmt.Errorf("pfddrawio.ValidateEdgeInvariants: フィードバック辺 %s の始端 %s が成果物・終端 %s がプロセスになっていません（不変条件 no-p2d-fb 違反。pfdlint で確認してください）", e.ID, e.Source, e.Target)
			}
			continue
		}
		if !rep.SameGroup(e.Source, e.Target) {
			continue
		}

		return fmt.Errorf("pfddrawio.ValidateEdgeInvariants: 同一 ID の成果物セル %s と %s が実線で結ばれています（不変条件 no-d2d 違反。pfdlint で確認してください）", e.Source, e.Target)
	}
	return nil
}

func AssignDisplayRanks(vertices []LayoutVertex, edges []LayoutEdge) (map[CellID]int, error) {
	if err := ValidateEdgeInvariants(vertices, edges); err != nil {
		return nil, err
	}

	rep := ContractDuplicates(vertices, edges)
	contracted, contractedEdges := ContractLayout(vertices, edges, rep)
	ranks, err := AssignRanks(contracted, contractedEdges)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.AssignDisplayRanks: 同一 ID の成果物を同一視したグラフに循環が残っています: %w", err)
	}
	return SpreadSourceRanks(ExpandRanks(ranks, rep), vertices, edges, rep), nil
}

const orderSweeps = 4

func OrderWithinRanks(vertices []LayoutVertex, edges []LayoutEdge, ranks map[CellID]int) map[CellID]int {
	rectY := make(map[CellID]float64, len(vertices))
	byRank := make(map[int][]CellID)
	maxRank := 0
	for _, v := range vertices {
		rectY[v.ID] = v.Rect.Y
		r := ranks[v.ID]
		byRank[r] = append(byRank[r], v.ID)
		if r > maxRank {
			maxRank = r
		}
	}

	tieLess := func(a, b CellID) bool {
		if rectY[a] != rectY[b] {
			return rectY[a] < rectY[b]
		}
		return a.Compare(b) < 0
	}
	for r := range byRank {
		ids := byRank[r]
		sort.SliceStable(ids, func(i, j int) bool { return tieLess(ids[i], ids[j]) })
	}

	preds := make(map[CellID][]CellID)
	succs := make(map[CellID][]CellID)
	feedbackNbr := make(map[CellID][]CellID)
	for _, e := range edges {
		if e.IsFeedback {

			feedbackNbr[e.Source] = append(feedbackNbr[e.Source], e.Target)
			feedbackNbr[e.Target] = append(feedbackNbr[e.Target], e.Source)
			continue
		}
		if e.IsCut {

			continue
		}
		preds[e.Target] = append(preds[e.Target], e.Source)
		succs[e.Source] = append(succs[e.Source], e.Target)
	}

	solidIsolated := make(map[CellID]struct{}, len(vertices))
	for _, v := range vertices {
		if len(preds[v.ID]) == 0 && len(succs[v.ID]) == 0 {
			solidIsolated[v.ID] = struct{}{}
		}
	}

	pos := make(map[CellID]int)
	refreshPos := func() {
		for _, ids := range byRank {
			for i, id := range ids {
				pos[id] = i
			}
		}
	}

	sortByBarycenter := func(ids []CellID, adj map[CellID][]CellID) {
		key := make(map[CellID]float64, len(ids))
		for i, id := range ids {
			ns := adj[id]
			if _, isolated := solidIsolated[id]; isolated {

				ns = feedbackNbr[id]
			}
			if len(ns) == 0 {
				key[id] = float64(i)
				continue
			}
			sum := 0.0
			for _, n := range ns {
				sum += float64(pos[n])
			}
			key[id] = sum / float64(len(ns))
		}
		sort.SliceStable(ids, func(i, j int) bool {
			if key[ids[i]] != key[ids[j]] {
				return key[ids[i]] < key[ids[j]]
			}
			return tieLess(ids[i], ids[j])
		})
	}

	for s := 0; s < orderSweeps; s++ {
		refreshPos()
		for r := 1; r <= maxRank; r++ {
			sortByBarycenter(byRank[r], preds)
			for i, id := range byRank[r] {
				pos[id] = i
			}
		}
		refreshPos()
		for r := maxRank - 1; r >= 0; r-- {
			sortByBarycenter(byRank[r], succs)
			for i, id := range byRank[r] {
				pos[id] = i
			}
		}
	}

	orders := make(map[CellID]int, len(vertices))
	for _, ids := range byRank {
		for i, id := range ids {
			orders[id] = i
		}
	}
	return orders
}

const alignSweeps = 8

func AlignYCoordinates(orders map[CellID]int, ranks map[CellID]int, edges []LayoutEdge, topCenterY, rowPitch float64) map[CellID]float64 {
	byRank := GroupByRank(ranks)
	maxRank := 0
	for r := range byRank {
		if r > maxRank {
			maxRank = r
		}
		SortByOrder(byRank[r], orders)
	}

	nbr := make(map[CellID][]CellID)
	for _, e := range edges {
		if !e.IsForward() {
			continue
		}
		nbr[e.Source] = append(nbr[e.Source], e.Target)
		nbr[e.Target] = append(nbr[e.Target], e.Source)
	}

	cy := make(map[CellID]float64, len(ranks))
	for _, ids := range byRank {
		for i, id := range ids {
			cy[id] = topCenterY + float64(i)*rowPitch
		}
	}

	placeRank := func(ids []CellID) {
		n := len(ids)
		if n == 0 {
			return
		}

		z := make([]float64, n)
		for i, id := range ids {
			ns := nbr[id]
			target := cy[id]
			if len(ns) > 0 {
				sum := 0.0
				for _, m := range ns {
					sum += cy[m]
				}
				target = sum / float64(len(ns))
			}
			z[i] = target - float64(i)*rowPitch
		}
		zz := IsotonicRegressionL2(z)
		for i, id := range ids {
			cy[id] = zz[i] + float64(i)*rowPitch
		}
	}

	for s := 0; s < alignSweeps; s++ {
		for r := 0; r <= maxRank; r++ {
			placeRank(byRank[r])
		}
		for r := maxRank; r >= 0; r-- {
			placeRank(byRank[r])
		}
	}
	return cy
}

func GroupByRank(ranks map[CellID]int) map[int][]CellID {
	byRank := make(map[int][]CellID)
	for id, r := range ranks {
		byRank[r] = append(byRank[r], id)
	}
	return byRank
}

func SortByOrder(ids []CellID, orders map[CellID]int) {
	sort.SliceStable(ids, func(i, j int) bool {
		if orders[ids[i]] != orders[ids[j]] {
			return orders[ids[i]] < orders[ids[j]]
		}
		return ids[i].Compare(ids[j]) < 0
	})
}

func QuantizeRow(centerY, topCenterY, rowPitch float64) int {
	return int(math.Floor((centerY-topCenterY)/rowPitch + 0.5))
}

func AssignRows(orders, ranks map[CellID]int, centerY map[CellID]float64, topCenterY, rowPitch float64) map[CellID]int {
	rows := make(map[CellID]int, len(ranks))
	for id := range ranks {
		rows[id] = QuantizeRow(centerY[id], topCenterY, rowPitch)
	}

	for _, ids := range GroupByRank(ranks) {
		SortByOrder(ids, orders)
		for i := 1; i < len(ids); i++ {
			if prev := rows[ids[i-1]]; rows[ids[i]] <= prev {
				rows[ids[i]] = prev + 1
			}
		}
	}

	minRow := 0
	first := true
	for _, r := range rows {
		if first || r < minRow {
			minRow, first = r, false
		}
	}
	if minRow != 0 {
		for id := range rows {
			rows[id] -= minRow
		}
	}
	return rows
}

func IsotonicRegressionL2(y []float64) []float64 {
	n := len(y)
	if n == 0 {
		return nil
	}
	vals := make([]float64, 0, n)
	cnts := make([]int, 0, n)
	for _, v := range y {
		vals = append(vals, v)
		cnts = append(cnts, 1)
		for len(vals) >= 2 && vals[len(vals)-2] > vals[len(vals)-1] {
			v2, c2 := vals[len(vals)-1], cnts[len(cnts)-1]
			v1, c1 := vals[len(vals)-2], cnts[len(cnts)-2]
			merged := (v1*float64(c1) + v2*float64(c2)) / float64(c1+c2)
			vals = vals[:len(vals)-2]
			cnts = cnts[:len(cnts)-2]
			vals = append(vals, merged)
			cnts = append(cnts, c1+c2)
		}
	}
	out := make([]float64, 0, n)
	for i := range vals {
		for j := 0; j < cnts[i]; j++ {
			out = append(out, vals[i])
		}
	}
	return out
}

func originXY(vertices []LayoutVertex) (float64, float64) {
	minX, minY := math.Inf(1), math.Inf(1)
	for _, v := range vertices {
		if v.Rect.X < minX {
			minX = v.Rect.X
		}
		if v.Rect.Y < minY {
			minY = v.Rect.Y
		}
	}
	if math.IsInf(minX, 1) {
		minX = 0
	}
	if math.IsInf(minY, 1) {
		minY = 0
	}
	return minX, minY
}

func RowPitch(vertices []LayoutVertex, opts SortOptions) float64 {
	return MaxVertexHeight(vertices) + opts.VGap
}

func TopCenterY(vertices []LayoutVertex, originY float64) float64 {
	return originY + MaxVertexHeight(vertices)/2
}

func MaxVertexHeight(vertices []LayoutVertex) float64 {
	var maxH float64
	for _, v := range vertices {
		if v.Rect.Height > maxH {
			maxH = v.Rect.Height
		}
	}
	return maxH
}

func ComputePositions(vertices []LayoutVertex, edges []LayoutEdge, ranks map[CellID]int, originX, originY float64, opts SortOptions) (map[CellID]geom.Rect, map[CellID]int) {
	orders := OrderWithinRanks(vertices, edges, ranks)

	colWidth := make(map[int]float64)
	maxRank := 0
	for _, v := range vertices {
		r := ranks[v.ID]
		if v.Rect.Width > colWidth[r] {
			colWidth[r] = v.Rect.Width
		}
		if r > maxRank {
			maxRank = r
		}
	}

	colX := make(map[int]float64, maxRank+1)
	x := originX
	for r := 0; r <= maxRank; r++ {
		colX[r] = x
		x += colWidth[r] + opts.HGap
	}
	rowPitch := RowPitch(vertices, opts)
	topCenterY := TopCenterY(vertices, originY)

	centerY := AlignYCoordinates(orders, ranks, edges, topCenterY, rowPitch)
	rows := AssignRows(orders, ranks, centerY, topCenterY, rowPitch)

	positions := make(map[CellID]geom.Rect, len(vertices))
	minY := math.Inf(1)
	for _, v := range vertices {
		r := ranks[v.ID]
		rect := geom.Rect{
			X:      colX[r] + (colWidth[r]-v.Rect.Width)/2,
			Y:      topCenterY + float64(rows[v.ID])*rowPitch - v.Rect.Height/2,
			Width:  v.Rect.Width,
			Height: v.Rect.Height,
		}
		if rect.Y < minY {
			minY = rect.Y
		}
		positions[v.ID] = rect
	}

	if !math.IsInf(minY, 1) && minY != originY {
		shift := originY - minY
		for id, rect := range positions {
			rect.Y += shift
			positions[id] = rect
		}
	}
	return positions, rows
}

func EstimateLayout(vertices []LayoutVertex, edges []LayoutEdge, ranks map[CellID]int, opts SortOptions) (map[CellID]geom.Rect, map[CellID]int) {
	ox, oy := originXY(vertices)
	return ComputePositions(vertices, edges, ranks, ox, oy, opts)
}

func EstimateRows(vertices []LayoutVertex, edges []LayoutEdge, ranks map[CellID]int, opts SortOptions) map[CellID]int {
	_, rows := EstimateLayout(vertices, edges, ranks, opts)
	return rows
}

func PlanPositions(vertices []LayoutVertex, edges []LayoutEdge, opts SortOptions) (map[CellID]geom.Rect, error) {
	ranks, err := AssignDisplayRanks(vertices, edges)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.PlanPositions: %w", err)
	}
	ox, oy := originXY(vertices)
	positions, _ := ComputePositions(vertices, edges, ranks, ox, oy, opts)
	return positions, nil
}

type EdgeEndpoint int

const (
	EndpointSource EdgeEndpoint = iota

	EndpointTarget
)

type DuplicatePlan struct {
	TempID   CellID
	SourceID CellID
	Rank     int
	ProcID   CellID
	ReuseID  CellID
	Rect     geom.Rect
}

type RewirePlan struct {
	EdgeID   CellID
	Endpoint EdgeEndpoint
	NewCell  CellID
}

type LayoutPlan struct {
	Positions  map[CellID]geom.Rect
	Duplicates []DuplicatePlan
	Rewires    []RewirePlan
	Removals   map[CellID]CellID
}

func AssignDuplicateCells(dups []DuplicatePlan, extras map[CellID][]CellID) ([]DuplicatePlan, map[CellID]CellID) {
	assigned := make([]DuplicatePlan, len(dups))
	copy(assigned, dups)

	used := make(map[CellID]int, len(extras))
	for i, d := range assigned {
		stock := extras[d.SourceID]
		if used[d.SourceID] < len(stock) {
			assigned[i].ReuseID = stock[used[d.SourceID]]
			used[d.SourceID]++
		}
	}

	removals := make(map[CellID]CellID)
	for source, stock := range extras {
		for _, id := range stock[used[source]:] {
			removals[id] = source
		}
	}
	return assigned, removals
}

func CanonicalizeRewires(edges []LayoutEdge, rep Representatives) []RewirePlan {
	var rewires []RewirePlan
	for _, e := range edges {
		if source := rep.Of(e.Source); source != e.Source {
			rewires = append(rewires, RewirePlan{EdgeID: e.ID, Endpoint: EndpointSource, NewCell: source})
		}
		if target := rep.Of(e.Target); target != e.Target {
			rewires = append(rewires, RewirePlan{EdgeID: e.ID, Endpoint: EndpointTarget, NewCell: target})
		}
	}
	return rewires
}

type DupCriteria struct {
	Ranks map[CellID]int

	Rows map[CellID]int

	RankSpan int

	RowSpan int
}

func (c DupCriteria) Spans(e LayoutEdge) (rankSpan, rowSpan int, far bool) {
	rs, ok1 := c.Ranks[e.Source]
	rt, ok2 := c.Ranks[e.Target]
	if !ok1 || !ok2 {
		return 0, 0, false
	}
	rankSpan = max(rt-rs, rs-rt)

	ys, ok1 := c.Rows[e.Source]
	yt, ok2 := c.Rows[e.Target]
	if !ok1 || !ok2 {
		return rankSpan, 0, e.IsCut || rankSpan > c.RankSpan
	}
	rowSpan = max(yt-ys, ys-yt)
	return rankSpan, rowSpan, e.IsCut || rankSpan > c.RankSpan || rowSpan > c.RowSpan
}

type duplicationCandidate struct {
	EdgeID CellID

	IsCut bool

	IsFeedback bool
	DelivID    CellID
	ProcID     CellID
	Endpoint   EdgeEndpoint
	DupRank    int

	SameColumn bool
	ProcRow    int
	RankSpan   int
	RowSpan    int
	EdgeOrder  int
}

func (c duplicationCandidate) nearerThan(o duplicationCandidate) bool {
	if c.IsCut != o.IsCut {
		return !c.IsCut
	}
	if c.IsFeedback != o.IsFeedback {
		return !c.IsFeedback
	}
	if c.RankSpan != o.RankSpan {
		return c.RankSpan < o.RankSpan
	}
	if c.RowSpan != o.RowSpan {
		return c.RowSpan < o.RowSpan
	}
	return c.EdgeOrder < o.EdgeOrder
}

func nearestCandidate(cs []duplicationCandidate) duplicationCandidate {
	best := cs[0]
	for _, c := range cs[1:] {
		if c.nearerThan(best) {
			best = c
		}
	}
	return best
}

type DuplicationGroup struct {
	DelivID    CellID
	Rank       int
	IsFeedback bool
}

func (c duplicationCandidate) Group() DuplicationGroup {
	return DuplicationGroup{DelivID: c.DelivID, Rank: c.DupRank, IsFeedback: c.IsFeedback}
}

func assignRowBands(cands []duplicationCandidate, rowSpan int) map[CellID]int {
	byGroup := make(map[DuplicationGroup][]duplicationCandidate)
	for _, c := range cands {
		byGroup[c.Group()] = append(byGroup[c.Group()], c)
	}

	bands := make(map[CellID]int, len(cands))
	for _, cs := range byGroup {
		sort.SliceStable(cs, func(i, j int) bool {
			if cs[i].ProcRow != cs[j].ProcRow {
				return cs[i].ProcRow < cs[j].ProcRow
			}
			return cs[i].EdgeOrder < cs[j].EdgeOrder
		})
		rows := make([]int, len(cs))
		for i, c := range cs {
			rows[i] = c.ProcRow
		}
		for i, band := range GreedyBands(rows, rowSpan) {
			bands[cs[i].EdgeID] = band
		}
	}
	return bands
}

func GreedyBands(points []int, span int) []int {
	if len(points) == 0 {
		return nil
	}
	bands := make([]int, len(points))
	band, bandTop := 0, points[0]
	for i, p := range points {
		if p-bandTop > span {
			band++
			bandTop = p
		}
		bands[i] = band
	}
	return bands
}

func PlanDuplications(vertices []LayoutVertex, edges []LayoutEdge, crit DupCriteria) ([]DuplicatePlan, []RewirePlan) {
	kind := make(map[CellID]VertexKind, len(vertices))
	for _, v := range vertices {
		kind[v.ID] = v.Kind
	}

	incidentCount := make(map[CellID]int, len(vertices))
	solidOut := make(map[CellID]int, len(vertices))
	solidIn := make(map[CellID]int, len(vertices))
	for _, e := range edges {
		if kind[e.Source] == VertexDeliverable {
			incidentCount[e.Source]++
			if !e.IsFeedback {
				solidOut[e.Source]++
			}
		}
		if kind[e.Target] == VertexDeliverable {
			incidentCount[e.Target]++
			if !e.IsFeedback {
				solidIn[e.Target]++
			}
		}
	}

	var ordered []duplicationCandidate
	byDeliv := make(map[CellID][]duplicationCandidate)
	for i, e := range edges {
		rankSpan, rowSpan, far := crit.Spans(e)
		if !far {
			continue
		}

		var delivID, procID CellID
		var endpoint EdgeEndpoint
		switch {
		case kind[e.Source] == VertexDeliverable && kind[e.Target].IsProcessLike():
			delivID, procID, endpoint = e.Source, e.Target, EndpointSource
		case kind[e.Target] == VertexDeliverable && kind[e.Source].IsProcessLike():
			delivID, procID, endpoint = e.Target, e.Source, EndpointTarget
		default:
			continue
		}

		procRank := crit.Ranks[procID]
		dupRank := procRank - 1
		if e.IsFeedback || endpoint == EndpointTarget {
			dupRank = procRank + 1
		}

		cand := duplicationCandidate{
			EdgeID: e.ID, IsCut: e.IsCut, IsFeedback: e.IsFeedback, DelivID: delivID, ProcID: procID, Endpoint: endpoint,

			DupRank: dupRank, SameColumn: dupRank == crit.Ranks[delivID], ProcRow: crit.Rows[procID],
			RankSpan: rankSpan, RowSpan: rowSpan, EdgeOrder: i,
		}
		ordered = append(ordered, cand)
		byDeliv[delivID] = append(byDeliv[delivID], cand)
	}

	keep := make(map[CellID]struct{})
	for delivID, cs := range byDeliv {
		if len(cs) != incidentCount[delivID] {
			continue
		}
		keep[nearestCandidate(cs).EdgeID] = struct{}{}
	}

	for delivID, cs := range byDeliv {
		if solidIn[delivID] == 0 || solidOut[delivID] == 0 {
			continue
		}
		var outCands, inCands []duplicationCandidate
		for _, c := range cs {
			if c.IsFeedback {
				continue
			}
			if c.Endpoint == EndpointSource {
				outCands = append(outCands, c)
			} else {
				inCands = append(inCands, c)
			}
		}
		if len(outCands) == solidOut[delivID] {
			if c := nearestCandidate(outCands); c.SameColumn {
				keep[c.EdgeID] = struct{}{}
			}
		}
		if len(inCands) == solidIn[delivID] {
			if c := nearestCandidate(inCands); c.SameColumn {
				keep[c.EdgeID] = struct{}{}
			}
		}
	}

	for _, c := range ordered {
		if c.IsCut {
			delete(keep, c.EdgeID)
		}
	}

	emitting := make([]duplicationCandidate, 0, len(ordered))
	for _, c := range ordered {
		if _, ok := keep[c.EdgeID]; !ok {
			emitting = append(emitting, c)
		}
	}
	bands := assignRowBands(emitting, crit.RowSpan)

	var dups []DuplicatePlan
	var rewires []RewirePlan

	type bandedGroup struct {
		Group DuplicationGroup
		Band  int
	}
	index := make(map[bandedGroup]CellID)
	n := 0
	for _, c := range emitting {
		key := bandedGroup{Group: c.Group(), Band: bands[c.EdgeID]}
		tempID, exists := index[key]
		if !exists {
			tempID = CellID("dup#" + strconv.Itoa(n))
			n++
			index[key] = tempID

			dups = append(dups, DuplicatePlan{TempID: tempID, SourceID: c.DelivID, Rank: c.DupRank, ProcID: c.ProcID})
		}
		rewires = append(rewires, RewirePlan{EdgeID: c.EdgeID, Endpoint: c.Endpoint, NewCell: tempID})
	}
	return dups, rewires
}

func ApplyRewires(edges []LayoutEdge, rewires []RewirePlan) []LayoutEdge {
	out := make([]LayoutEdge, len(edges))
	copy(out, edges)
	index := make(map[CellID]int, len(out))
	for i, e := range out {
		index[e.ID] = i
	}
	for _, r := range rewires {
		i, ok := index[r.EdgeID]
		if !ok {
			continue
		}
		if r.Endpoint == EndpointSource {
			out[i].Source = r.NewCell
		} else {
			out[i].Target = r.NewCell
		}
		out[i].IsCut = false
	}
	return out
}

func PlanLayout(vertices []LayoutVertex, edges []LayoutEdge, opts SortOptions) (*LayoutPlan, error) {

	if err := ValidateEdgeInvariants(vertices, edges); err != nil {
		return nil, fmt.Errorf("pfddrawio.PlanLayout: %w", err)
	}

	rep := ContractDuplicates(vertices, edges)
	extras := ExtraDuplicateCells(vertices, rep)
	canonVertices, canonEdges := ContractLayout(vertices, edges, rep)
	canonRewires := CanonicalizeRewires(edges, rep)

	if opts.BreakCycles {
		canonEdges = MarkCycleCuts(canonVertices, canonEdges)
	}

	ranks, err := AssignDisplayRanks(canonVertices, canonEdges)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.PlanLayout: %w", err)
	}

	estPositions, estRows := EstimateLayout(canonVertices, canonEdges, ranks, opts)

	crit := DupCriteria{
		Ranks:    ranks,
		Rows:     estRows,
		RankSpan: opts.DupRankSpan,
		RowSpan:  opts.DupRowSpan,
	}
	dups, dupRewires := PlanDuplications(canonVertices, canonEdges, crit)
	dups, removals := AssignDuplicateCells(dups, extras)

	rewires := make([]RewirePlan, 0, len(canonRewires)+len(dupRewires))
	rewires = append(rewires, canonRewires...)
	rewires = append(rewires, dupRewires...)

	sizeOf := make(map[CellID]geom.Rect, len(canonVertices))
	for _, v := range canonVertices {
		sizeOf[v.ID] = v.Rect
	}

	augVertices := make([]LayoutVertex, 0, len(canonVertices)+len(dups))
	for _, v := range canonVertices {
		v.Rect.Y = estPositions[v.ID].Y
		augVertices = append(augVertices, v)
	}
	augRanks := make(map[CellID]int, len(ranks)+len(dups))
	for k, v := range ranks {
		augRanks[k] = v
	}
	for _, d := range dups {

		src := sizeOf[d.SourceID]
		augVertices = append(augVertices, LayoutVertex{
			ID:   d.TempID,
			Kind: VertexDeliverable,
			Rect: geom.Rect{Y: estPositions[d.ProcID].Y, Width: src.Width, Height: src.Height},
		})
		augRanks[d.TempID] = d.Rank
	}
	augEdges := ApplyRewires(canonEdges, dupRewires)

	ox, oy := originXY(vertices)
	positions, _ := ComputePositions(augVertices, augEdges, augRanks, ox, oy, opts)

	dupIdx := make(map[CellID]int, len(dups))
	for i, d := range dups {
		dupIdx[d.TempID] = i
	}
	existing := make(map[CellID]geom.Rect, len(canonVertices))
	for id, rect := range positions {
		if i, ok := dupIdx[id]; ok {
			dups[i].Rect = rect
			continue
		}
		existing[id] = rect
	}

	return &LayoutPlan{Positions: existing, Duplicates: dups, Rewires: rewires, Removals: removals}, nil
}

func AverageCenterY(ids []CellID, rects map[CellID]geom.Rect) (float64, bool) {
	sum := 0.0
	n := 0
	for _, id := range ids {
		if r, ok := rects[id]; ok {
			sum += r.Y + r.Height/2
			n++
		}
	}
	if n == 0 {
		return 0, false
	}
	return sum / float64(n), true
}

func SelectedTopologicalOrder(vertices []LayoutVertex, edges []LayoutEdge, selected map[CellID]struct{}) ([]CellID, error) {
	order := ForwardGraph(vertices, edges).TopologicalSort()
	if order == nil {
		return nil, fmt.Errorf("pfddrawio.SelectedTopologicalOrder: %w", ErrCycleRemains)
	}
	ids := make([]CellID, 0, len(selected))
	for _, n := range order {
		id := CellID(n)
		if _, ok := selected[id]; ok {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func OnlyNodeTarget(cellID CellID, current geom.Rect, edges []LayoutEdge, known map[CellID]geom.Rect, hgap float64) geom.Rect {
	var preds, succs, feedback []CellID
	for _, e := range edges {
		if e.IsFeedback {
			if e.Source == cellID {
				feedback = append(feedback, e.Target)
			}
			if e.Target == cellID {
				feedback = append(feedback, e.Source)
			}
			continue
		}
		if e.IsCut {

			continue
		}
		if e.Target == cellID {
			preds = append(preds, e.Source)
		}
		if e.Source == cellID {
			succs = append(succs, e.Target)
		}
	}

	x := current.X
	hasLeft, hasRight := false, false
	var xLeft, xRight float64
	for _, p := range preds {
		if r, ok := known[p]; ok {
			if v := r.X + r.Width + hgap; !hasLeft || v > xLeft {
				xLeft, hasLeft = v, true
			}
		}
	}
	for _, s := range succs {
		if r, ok := known[s]; ok {
			if v := r.X - hgap - current.Width; !hasRight || v < xRight {
				xRight, hasRight = v, true
			}
		}
	}
	switch {
	case hasLeft && hasRight:
		if xLeft <= xRight {
			x = xLeft
		} else {
			x = (xLeft + xRight) / 2
		}
	case hasLeft:
		x = xLeft
	case hasRight:
		x = xRight
	}

	y := current.Y
	solid := append(append([]CellID{}, preds...), succs...)
	if cy, ok := AverageCenterY(solid, known); ok {
		y = cy - current.Height/2
	} else if cy, ok := AverageCenterY(feedback, known); ok {
		y = cy - current.Height/2
	}

	return geom.Rect{X: x, Y: y, Width: current.Width, Height: current.Height}
}

func ClampToPredecessorFloor(target geom.Rect, id CellID, edges []LayoutEdge, current map[CellID]geom.Rect, selected map[CellID]struct{}, hgap float64) geom.Rect {
	floor := target.X
	hasFloor := false
	for _, e := range edges {
		if !e.IsForward() || e.Target != id {
			continue
		}
		if _, ok := selected[e.Source]; !ok {
			continue
		}
		if r, ok := current[e.Source]; ok {
			if v := r.X + r.Width + hgap; !hasFloor || v > floor {
				floor, hasFloor = v, true
			}
		}
	}
	if hasFloor && floor > target.X {
		target.X = floor
	}
	return target
}

func ClampToSuccessorCeiling(target geom.Rect, id CellID, edges []LayoutEdge, current map[CellID]geom.Rect, selected map[CellID]struct{}, hgap float64) geom.Rect {
	ceiling := target.X
	hasCeiling := false
	for _, e := range edges {
		if !e.IsForward() || e.Source != id {
			continue
		}
		if _, ok := selected[e.Target]; !ok {
			continue
		}
		if r, ok := current[e.Target]; ok {
			if v := r.X - target.Width - hgap; !hasCeiling || v < ceiling {
				ceiling, hasCeiling = v, true
			}
		}
	}
	if hasCeiling && ceiling < target.X {
		target.X = ceiling
	}
	return target
}

func RelaxOnlyNodeTargets(ids []CellID, frozen map[CellID]geom.Rect, edges []LayoutEdge, hgap float64) map[CellID]geom.Rect {
	selected := make(map[CellID]struct{}, len(ids))
	for _, id := range ids {
		selected[id] = struct{}{}
	}
	current := make(map[CellID]geom.Rect, len(frozen))
	for id, r := range frozen {
		current[id] = r
	}

	for _, id := range ids {
		target := OnlyNodeTarget(id, current[id], edges, current, hgap)
		current[id] = ClampToPredecessorFloor(target, id, edges, current, selected, hgap)
	}
	for i := len(ids) - 1; i >= 0; i-- {
		id := ids[i]
		target := OnlyNodeTarget(id, current[id], edges, current, hgap)
		target = ClampToPredecessorFloor(target, id, edges, current, selected, hgap)
		current[id] = ClampToSuccessorCeiling(target, id, edges, current, selected, hgap)
	}

	result := make(map[CellID]geom.Rect, len(ids))
	for _, id := range ids {
		result[id] = current[id]
	}
	return result
}

func OverlapsAny(r geom.Rect, others []geom.Rect) bool {
	for _, o := range others {
		if r.Overlaps(o) {
			return true
		}
	}
	return false
}

func ShiftDownToClear(r geom.Rect, blockers []geom.Rect, step float64) geom.Rect {
	maxBottom := r.Y
	for _, b := range blockers {
		if bottom := b.Y + b.Height; bottom > maxBottom {
			maxBottom = bottom
		}
	}

	if step <= 0 {
		return r
	}
	for r.Y <= maxBottom && OverlapsAny(r, blockers) {
		r.Y += step
	}
	return r
}

func PushClear(cell geom.Rect, blockers []geom.Rect, mode PosRestriction) geom.Rect {
	var over []geom.Rect
	for _, b := range blockers {
		if cell.Overlaps(b) {
			over = append(over, b)
		}
	}
	if len(over) == 0 {
		return cell
	}
	maxBottom, minTop := over[0].Y+over[0].Height, over[0].Y
	maxRight, minLeft := over[0].X+over[0].Width, over[0].X
	for _, b := range over {
		if bottom := b.Y + b.Height; bottom > maxBottom {
			maxBottom = bottom
		}
		if b.Y < minTop {
			minTop = b.Y
		}
		if right := b.X + b.Width; right > maxRight {
			maxRight = right
		}
		if b.X < minLeft {
			minLeft = b.X
		}
	}
	down := maxBottom - cell.Y
	up := (cell.Y + cell.Height) - minTop
	right := maxRight - cell.X
	left := (cell.X + cell.Width) - minLeft

	type move struct{ dist, dx, dy float64 }
	var cands []move
	switch mode {
	case PosFreeV:
		cands = []move{{down, 0, down}, {up, 0, -up}}
	case PosFreeH:
		cands = []move{{right, right, 0}, {left, -left, 0}}
	default:
		cands = []move{{down, 0, down}, {up, 0, -up}, {right, right, 0}, {left, -left, 0}}
	}
	best := cands[0]
	for _, c := range cands[1:] {
		if c.dist < best.dist {
			best = c
		}
	}
	cell.X += best.dx
	cell.Y += best.dy
	return cell
}

func PlanOnlyNodes(vertices []LayoutVertex, edges []LayoutEdge, opts SortOptions) (*LayoutPlan, error) {
	selected := SelectCells(vertices, opts.OnlyNodes)

	if opts.BreakCycles {
		edges = MarkCycleCuts(vertices, edges)
	}

	frozen := make(map[CellID]geom.Rect, len(vertices))
	for _, v := range vertices {
		frozen[v.ID] = v.Rect
	}

	ids, err := SelectedTopologicalOrder(vertices, edges, selected)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.PlanOnlyNodes: %w", err)
	}
	targets := RelaxOnlyNodeTargets(ids, frozen, edges, opts.HGap)

	rowPitch := RowPitch(vertices, opts)
	positions := make(map[CellID]geom.Rect, len(ids))

	placedSelected := func() []geom.Rect {
		var rs []geom.Rect
		for _, prev := range ids {
			if r, ok := positions[prev]; ok {
				rs = append(rs, r)
			}
		}
		return rs
	}

	if opts.PosRestriction == PosLock {

		var immovable []geom.Rect
		for _, v := range vertices {
			if _, sel := selected[v.ID]; !sel {
				immovable = append(immovable, v.Rect)
			}
		}
		for _, id := range ids {

			blockers := append(append([]geom.Rect{}, immovable...), placedSelected()...)
			positions[id] = ShiftDownToClear(targets[id], blockers, rowPitch)
		}
		return &LayoutPlan{Positions: positions}, nil
	}

	for _, id := range ids {
		positions[id] = ShiftDownToClear(targets[id], placedSelected(), rowPitch)
	}

	placed := placedSelected()
	for _, v := range vertices {
		if _, sel := selected[v.ID]; sel {
			continue
		}
		if moved := PushClear(v.Rect, placed, opts.PosRestriction); moved != v.Rect {
			positions[v.ID] = moved
		}
	}
	return &LayoutPlan{Positions: positions}, nil
}

type EstimateBoxRef struct {
	ID   CellID
	Rect geom.Rect
}

func FollowEstimateBoxes(processes []LayoutVertex, boxes []EstimateBoxRef, positions map[CellID]geom.Rect) map[CellID]geom.Rect {
	result := make(map[CellID]geom.Rect, len(boxes))
	for _, b := range boxes {
		for _, p := range processes {
			if p.Kind != VertexProcess {
				continue
			}
			if !IsBelow(p.Rect, b.Rect) {
				continue
			}
			if np, ok := positions[p.ID]; ok {
				result[b.ID] = EstimateBoxRect(np)
			}
			break
		}
	}
	return result
}

func PageCellIDs(root *xmldom.Node) map[CellID]struct{} {
	ids := make(map[CellID]struct{})
	root.Traverse(func(n *xmldom.Node) {
		if n.Kind != xmldom.ElementNode {
			return
		}
		if id, ok := n.GetAttr("id", ""); ok && id != "" {
			ids[CellID(id)] = struct{}{}
		}
	}, nil)
	return ids
}

func DanglingRefs(root *xmldom.Node) []string {
	known := PageCellIDs(root)
	var dangling []string
	root.Traverse(func(n *xmldom.Node) {
		if n.Kind != xmldom.ElementNode {
			return
		}
		for _, attr := range []string{"parent", "source", "target"} {
			ref, ok := n.GetAttr(attr, "")
			if !ok || ref == "" {
				continue
			}
			if _, found := known[CellID(ref)]; !found {
				id, _ := n.GetAttr("id", "")
				dangling = append(dangling, id+"."+attr+"="+ref)
			}
		}
	}, nil)
	slices.Sort(dangling)
	return dangling
}

func RetargetPageRefs(root *xmldom.Node, refs map[CellID]CellID) {
	if len(refs) == 0 {
		return
	}
	root.Traverse(func(n *xmldom.Node) {
		if n.Kind != xmldom.ElementNode {
			return
		}
		for _, attr := range []string{"parent", "source", "target"} {
			id, ok := n.GetAttr(attr, "")
			if !ok {
				continue
			}
			if to, found := refs[CellID(id)]; found {
				n.SetAttr(attr, string(to))
			}
		}
	}, nil)
}

func SetGeometry(cell *xmldom.Node, rect geom.Rect) {
	geo := cell.FirstChildElement("mxGeometry")
	if geo == nil {
		geo = &xmldom.Node{
			Kind: xmldom.ElementNode,
			Start: xml.StartElement{
				Name: xml.Name{Local: "mxGeometry"},
				Attr: []xml.Attr{{Name: xml.Name{Local: "as"}, Value: "geometry"}},
			},
			End: xml.EndElement{Name: xml.Name{Local: "mxGeometry"}},
		}
		cell.AppendChildIndented(cell.ChildElementIndent(), geo)
	}
	geo.SetAttr("x", formatFloat(rect.X))
	geo.SetAttr("y", formatFloat(rect.Y))
	geo.SetAttr("width", formatFloat(rect.Width))
	geo.SetAttr("height", formatFloat(rect.Height))
}

func Sort(r io.Reader, opts SortOptions, logger *slog.Logger) ([]*xmldom.Node, error) {
	nodes, err := xmldom.ParseXML(r)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.Sort: %w", err)
	}

	layers := NewLayerMapFromNodes(nodes)
	parents := NewParentMapFromNodes(nodes)

	boxesByDiagram := make(map[DiagramID][]EstimateBoxRef)
	for _, b := range EstimateBoxes(nodes) {
		boxesByDiagram[b.DiagramID] = append(boxesByDiagram[b.DiagramID], EstimateBoxRef{ID: b.CellID, Rect: b.Rect})
	}

	doms := CollectDiagramDOMs(nodes)

	diagramIDs := make([]DiagramID, 0, len(doms))
	for id := range doms {
		diagramIDs = append(diagramIDs, id)
	}
	sort.Slice(diagramIDs, func(i, j int) bool {
		return CellID(diagramIDs[i]).Compare(CellID(diagramIDs[j])) < 0
	})

	pageNames := PageNames(doms)

	var seenNodeIDs []pfd.NodeID
	seenSet := make(map[pfd.NodeID]struct{})

	for _, diagramID := range diagramIDs {
		dom := doms[diagramID]
		if !ShouldSortPage(dom.Name, opts.OnlyPages) {
			continue
		}
		vertices, edges := CollectPageLayout(dom, layers, parents)
		if len(vertices) == 0 {
			continue
		}
		for _, v := range vertices {
			if v.NodeID == "" {
				continue
			}
			if _, ok := seenSet[v.NodeID]; !ok {
				seenSet[v.NodeID] = struct{}{}
				seenNodeIDs = append(seenNodeIDs, v.NodeID)
			}
		}

		plan := &LayoutPlan{}
		if opts.IsOnlyNodeMode() {
			plan, err = PlanOnlyNodes(vertices, edges, opts)
		} else {
			plan, err = PlanLayout(vertices, edges, opts)
		}
		if err != nil {

			if errors.Is(err, ErrCycleRemains) {
				return nil, fmt.Errorf("pfddrawio.Sort: diagram %q: %w。%s", diagramID, err, CycleHint(vertices, edges, opts.BreakCycles))
			}
			return nil, fmt.Errorf("pfddrawio.Sort: diagram %q: %w", diagramID, err)
		}

		tempToReal := make(map[CellID]CellID, len(plan.Duplicates))
		tempIDs := make(map[CellID]struct{}, len(plan.Duplicates))
		takenIDs := PageCellIDs(dom.Root)
		reused := 0
		for _, d := range plan.Duplicates {
			tempIDs[d.TempID] = struct{}{}
			if d.ReuseID != "" {

				cell, ok := dom.CellsByID[d.ReuseID]
				if !ok {
					return nil, fmt.Errorf("pfddrawio.Sort: diagram %q: 使い回す複製セル %s が見つかりません", diagramID, d.ReuseID)
				}
				tempToReal[d.TempID] = d.ReuseID

				SetGeometry(cell, d.Rect)
				if src, ok := dom.CellsByID[d.SourceID]; ok {
					value, _ := src.GetAttr("value", "")
					cell.SetAttr("value", value)
					style, _ := src.GetAttr("style", "")
					cell.SetAttr("style", style)
				}
				reused++
				logger.Debug("pfddrawio.Sort: reused duplicate deliverable", "diagram", diagramID, "source", d.SourceID, "cell", d.ReuseID)
				continue
			}

			src, ok := dom.CellsByID[d.SourceID]
			if !ok {
				continue
			}
			realID := NextDuplicateCellID(d.SourceID, takenIDs)
			takenIDs[realID] = struct{}{}
			tempToReal[d.TempID] = realID
			clone := src.Clone()
			clone.SetAttr("id", string(realID))
			SetGeometry(clone, d.Rect)
			dom.Root.AppendChildIndented(dom.Root.ChildElementIndent(), clone)
			dom.CellsByID[realID] = clone
			logger.Debug("pfddrawio.Sort: duplicated deliverable", "diagram", diagramID, "source", d.SourceID, "cell", realID)
		}

		for id, rect := range plan.Positions {
			if cell, ok := dom.CellsByID[id]; ok {
				SetGeometry(cell, rect)
			}
		}

		for _, rw := range plan.Rewires {

			newID := rw.NewCell
			if _, isTemp := tempIDs[newID]; isTemp {
				realID, ok := tempToReal[newID]
				if !ok {
					continue
				}
				newID = realID
			}
			edgeCell, ok := dom.CellsByID[rw.EdgeID]
			if !ok {
				continue
			}
			if rw.Endpoint == EndpointSource {
				edgeCell.SetAttr("source", string(newID))
			} else {
				edgeCell.SetAttr("target", string(newID))
			}
		}

		if len(plan.Removals) > 0 {
			RetargetPageRefs(dom.Root, plan.Removals)
			dom.Root.RemoveChildElements(func(n *xmldom.Node) bool {
				id, _ := n.GetAttr("id", "")
				_, ok := plan.Removals[CellID(id)]
				return ok
			})
			for id := range plan.Removals {
				delete(dom.CellsByID, id)
			}
		}

		for _, e := range edges {
			if opts.IsOnlyNodeMode() {
				_, srcMoved := plan.Positions[e.Source]
				_, tgtMoved := plan.Positions[e.Target]
				if !srcMoved && !tgtMoved {
					continue
				}
			}

			edgeCell, ok := dom.CellsByID[e.ID]
			if !ok {
				continue
			}
			styleStr, _ := edgeCell.GetAttr("style", "")
			edgeCell.SetAttr("style", StraightEdgeStyle(styleStr))
			RemoveEdgeWaypoints(edgeCell)
		}

		for boxID, rect := range FollowEstimateBoxes(vertices, boxesByDiagram[diagramID], plan.Positions) {
			if cell, ok := dom.CellsByID[boxID]; ok {
				SetGeometry(cell, rect)
			}
		}

		logger.Debug("pfddrawio.Sort: laid out page", "diagram", diagramID, "vertices", len(vertices),
			"duplicates", len(plan.Duplicates), "reused", reused, "removed", len(plan.Removals))
	}

	if !opts.IsOnlyNodeMode() {
		ApplyDuplicateMarksToNodes(nodes, opts.OnlyPages, logger)
	}

	for _, want := range UnmatchedOnlyPages(opts.OnlyPages, pageNames) {
		logger.Warn("pfddrawio.Sort: -only-page で指定されたページが見つかりません", "page", want)
	}

	for _, want := range UnmatchedOnlyNodes(opts.OnlyNodes, seenNodeIDs) {
		logger.Warn("pfddrawio.Sort: -only-node で指定されたノードが見つかりません", "node", want)
	}
	return nodes, nil
}
