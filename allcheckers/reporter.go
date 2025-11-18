package allcheckers

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/locale"
)

type Func func(ch <-chan checkers.Problem) (int, error)

var (
	tab     = []byte{'\t'}
	sep     = []byte(", ")
	newLine = []byte{'\n'}
)

func NewSorted(reporterFunc Func) Func {
	sb := &strings.Builder{}
	return func(ch <-chan checkers.Problem) (int, error) {
		ps := make([]checkers.Problem, 0)
		for problem := range ch {
			ps = append(ps, problem)
		}

		slices.SortFunc(ps, func(a, b checkers.Problem) int {
			return checkers.CompareProblem(a, b, sb)
		})

		ch2 := make(chan checkers.Problem, len(ps))
		for _, problem := range ps {
			ch2 <- problem
		}
		close(ch2)

		count, err := reporterFunc(ch2)
		if err != nil {
			return 0, fmt.Errorf("NewSorted: %w", err)
		}
		return count, nil
	}
}

func WriteTSVRow(w io.Writer, problem checkers.Problem, l locale.Locale) {
	io.WriteString(w, problem.Severity.String())
	w.Write(tab)
	io.WriteString(w, string(problem.ProblemID))
	w.Write(tab)
	io.WriteString(w, Message(problem.ProblemID, l))
	w.Write(tab)
	for i, loc := range problem.Locations {
		if i > 0 {
			w.Write(sep)
		}
		loc.Write(w)
	}
	w.Write(newLine)
}

func WriteTSV(w io.Writer, ps []checkers.Problem, l locale.Locale) {
	for _, problem := range ps {
		WriteTSVRow(w, problem, l)
	}
}

func NewTSV(w io.Writer, l locale.Locale) Func {
	return func(ch <-chan checkers.Problem) (int, error) {
		count := 0
		for problem := range ch {
			WriteTSVRow(w, problem, l)
			count++
		}
		return count, nil
	}
}

func NewJSON(w io.Writer, l locale.Locale) Func {
	return func(ch <-chan checkers.Problem) (int, error) {
		count := 0
		for problem := range ch {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"locations":  problem.Locations,
				"problem_id": problem.ProblemID,
				"severity":   problem.Severity.String(),
				"message":    Message(problem.ProblemID, l),
			})
		}
		return count, nil
	}
}
