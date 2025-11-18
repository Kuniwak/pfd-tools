package pfddrawio

import (
	"errors"
	"strings"

	"github.com/Kuniwak/pfd-tools/pfd"
)

type CellError struct {
	Locations []DrawIOLocation
	Wrapped   error
}

func NewCellErrorByError(wrapped error, locations ...DrawIOLocation) CellError {
	return CellError{Locations: locations, Wrapped: wrapped}
}

func NewCellErrorByMessage(message string, locations ...DrawIOLocation) CellError {
	return NewCellErrorByError(errors.New(message), locations...)
}

func NewCellErrorByPFDError(err pfd.Error, srcMap *SourceMap) CellError {
	locs := make([]DrawIOLocation, 0, len(err.Locations))
	for _, loc := range err.Locations {
		if loc.IsNode {
			locs = append(locs, srcMap.NodeIDMap[loc.NodeID].Slice()...)
		} else {
			locs = append(locs, srcMap.EdgeIDMap[loc.EdgeSourceID][loc.EdgeTargetID].Slice()...)
		}
	}
	return NewCellErrorByError(err, locs...)
}

func (e CellError) Write(sb *strings.Builder) {
	sb.WriteString(e.Wrapped.Error())
	sb.WriteString(": [")
	for i, id := range e.Locations {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("{diagramID:")
		sb.WriteString(string(id.DiagramID))
		sb.WriteString(", cellID:")
		sb.WriteString(string(id.CellID))
		sb.WriteString("}")
	}
	sb.WriteString("]")
}

func (e CellError) Error() string {
	sb := &strings.Builder{}
	e.Write(sb)
	return sb.String()
}

func (e CellError) Unwrap() error {
	return e.Wrapped
}

type CellErrors []CellError

func NewCellErrorsByPFDErrors(errs pfd.Errors, srcMap *SourceMap) CellErrors {
	cellErrors := make(CellErrors, 0, len(errs))
	for _, err := range errs {
		cellErrors = append(cellErrors, NewCellErrorByPFDError(err, srcMap))
	}
	return cellErrors
}

func (e CellErrors) Error() string {
	if len(e) == 0 {
		panic("pfddrawio.CellErrors: empty")
	}

	sb := &strings.Builder{}
	for i, err := range []CellError(e) {
		if i > 0 {
			sb.WriteString("\n")
		}
		err.Write(sb)
	}
	return sb.String()
}
