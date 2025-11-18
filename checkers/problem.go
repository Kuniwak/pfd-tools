package checkers

import (
	"slices"
	"strings"
)

type ProblemID string

type Problem struct {
	Locations []Location
	ProblemID ProblemID
	Severity  Severity
}

func NewProblem(problemID ProblemID, severity Severity, locations ...Location) Problem {
	return Problem{Locations: locations, ProblemID: problemID, Severity: severity}
}

func CompareProblem(a, b Problem, sb *strings.Builder) int {
	var i int

	i = CompareSeverity(a.Severity, b.Severity)
	if i != 0 {
		return i
	}

	i = slices.CompareFunc(a.Locations, b.Locations, func(a, b Location) int {
		return CompareLocation(a, b, sb)
	})
	if i != 0 {
		return i
	}

	return strings.Compare(string(a.ProblemID), string(b.ProblemID))
}
