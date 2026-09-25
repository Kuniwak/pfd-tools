package pfddrawio

type EffectiveEdge struct {
	Source     CellID
	Target     CellID
	IsFeedback bool
	Origins    []CellID
}

func SpliceConnectors(cells []Cell) []EffectiveEdge {
	isConnector := make(map[CellID]bool)
	for _, cell := range cells {
		if cell.IsConnector() {
			isConnector[cell.ID] = true
		}
	}

	outgoing := make(map[CellID][]Cell)
	for _, cell := range cells {
		if cell.IsEdge && isConnector[cell.Source] {
			outgoing[cell.Source] = append(outgoing[cell.Source], cell)
		}
	}

	effectiveEdges := make([]EffectiveEdge, 0)

	visited := make(map[CellID]bool)
	var expandChain func(source CellID, conn CellID, isFeedback bool, origins []CellID)
	expandChain = func(source CellID, conn CellID, isFeedback bool, origins []CellID) {
		visited[conn] = true
		for _, out := range outgoing[conn] {
			nextFeedback := isFeedback || out.Style.IsDashed()
			nextOrigins := append(append([]CellID{}, origins...), out.ID)
			if isConnector[out.Target] {
				if visited[out.Target] {
					continue
				}
				expandChain(source, out.Target, nextFeedback, nextOrigins)
			} else {
				effectiveEdges = append(effectiveEdges, EffectiveEdge{
					Source:     source,
					Target:     out.Target,
					IsFeedback: nextFeedback,
					Origins:    nextOrigins,
				})
			}
		}
		visited[conn] = false
	}

	for _, cell := range cells {
		if !cell.IsEdge {
			continue
		}
		if isConnector[cell.Source] {

			continue
		}
		if !isConnector[cell.Target] {

			effectiveEdges = append(effectiveEdges, EffectiveEdge{
				Source:     cell.Source,
				Target:     cell.Target,
				IsFeedback: cell.Style.IsDashed(),
				Origins:    []CellID{cell.ID},
			})
			continue
		}

		expandChain(cell.Source, cell.Target, cell.Style.IsDashed(), []CellID{cell.ID})
	}

	return effectiveEdges
}

func DanglingConnectors(cells []Cell) []CellID {
	isConnector := make(map[CellID]bool)
	for _, cell := range cells {
		if cell.IsConnector() {
			isConnector[cell.ID] = true
		}
	}

	hasIn := make(map[CellID]bool)
	hasOut := make(map[CellID]bool)
	for _, cell := range cells {
		if !cell.IsEdge {
			continue
		}
		if isConnector[cell.Target] {
			hasIn[cell.Target] = true
		}
		if isConnector[cell.Source] {
			hasOut[cell.Source] = true
		}
	}

	var dangling []CellID
	for _, cell := range cells {
		if !cell.IsConnector() {
			continue
		}
		if hasIn[cell.ID] != hasOut[cell.ID] {
			dangling = append(dangling, cell.ID)
		}
	}
	return dangling
}
