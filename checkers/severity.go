package checkers

import (
	"fmt"
)

type Severity int

func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "ERROR"
	case SeverityStyleProblem:
		return "STYLE_PROBLEM"
	case SeverityWarning:
		return "WARNING"
	default:
		panic(fmt.Sprintf("unknown severity: %d", s))
	}
}

func CompareSeverity(a, b Severity) int {
	return int(a) - int(b)
}

const (
	SeverityError Severity = iota
	SeverityStyleProblem
	SeverityWarning
)
