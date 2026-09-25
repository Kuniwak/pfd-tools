package pfddrawio

import "github.com/Kuniwak/pfd-tools/xmldom"

func RemoveEdgeWaypoints(cell *xmldom.Node) {
	geo := cell.FirstChildElement("mxGeometry")
	if geo == nil {
		return
	}
	geo.RemoveChildElements(func(ch *xmldom.Node) bool {
		if ch.Start.Name.Local != "Array" {
			return false
		}
		as, _ := ch.GetAttr("as", "")
		return as == "points"
	})
}

func RemoveEdgeEndpointPoints(cell *xmldom.Node) {
	geo := cell.FirstChildElement("mxGeometry")
	if geo == nil {
		return
	}

	geo.RemoveChildElements(func(ch *xmldom.Node) bool {
		if ch.Start.Name.Local != "mxPoint" {
			return false
		}
		as, _ := ch.GetAttr("as", "")
		return as == "sourcePoint" || as == "targetPoint"
	})
}
