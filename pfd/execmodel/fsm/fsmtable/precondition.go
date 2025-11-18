package fsmtable

import (
	"cmp"
	"fmt"
	"strings"

	"github.com/Kuniwak/pfd-tools/parser"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/sets"
)

const (
	PreconditionColumnHeaderJa = "開始条件"
	PreconditionColumnHeaderEn = "Start Condition"
)

var DefaultPreconditionColumnMatchFunc = pfd.ColumnMatchFunc(sets.New(
	strings.Compare,
	PreconditionColumnHeaderJa,
	PreconditionColumnHeaderEn,
))

func RawPreconditionMap(t *pfd.AtomicProcessTable, matchFunc pfd.ColumnSelectFunc) (map[pfd.AtomicProcessID]string, error) {
	m := make(map[pfd.AtomicProcessID]string, len(t.Rows))

	idx := matchFunc(t.ExtraHeaders)
	if idx < 0 {
		return nil, fmt.Errorf("fsmtable.RawPreconditionMap: missing precondition column")
	}
	for _, row := range t.Rows {
		m[row.ID] = row.ExtraCells[idx]
	}
	return m, nil
}

func ValidatePreconditionMap(m map[pfd.AtomicProcessID]string) (map[pfd.AtomicProcessID]*fsm.Precondition, error) {
	m2 := make(map[pfd.AtomicProcessID]*fsm.Precondition, len(m))
	for ap, preconditionText := range m {
		precondition, err := ValidatePrecondition(preconditionText, ap)
		if err != nil {
			return nil, fmt.Errorf("fsmtable.ValidatePreconditionMap: %w", err)
		}
		m2[ap] = precondition
	}
	return m2, nil
}

func ValidatePrecondition(s string, ap pfd.AtomicProcessID) (*fsm.Precondition, error) {
	precondition, err := ParsePrecondition(s, ap)
	if err != nil {
		return nil, fmt.Errorf("fsmtable.ValidatePrecondition: %w", err)
	}
	return precondition, nil
}

func PreconditionFuncByTableFunc(table *pfd.AtomicProcessTable, matchFunc pfd.ColumnSelectFunc) (map[pfd.AtomicProcessID]*fsm.Precondition, error) {
	rawPreconditionMap, err := RawPreconditionMap(table, matchFunc)
	if err != nil {
		return nil, fmt.Errorf("fsmtable.PreconditionFuncByTableFunc: %w", err)
	}
	preconditionMap, err := ValidatePreconditionMap(rawPreconditionMap)
	if err != nil {
		return nil, fmt.Errorf("fsmtable.PreconditionFuncByTableFunc: %w", err)
	}
	return preconditionMap, nil
}

// ParsePrecondition parses the following syntax rules:
//
//		precondition = *SP *1(or_expr)
//
//		or_expr      = and_expr *( "||" *SP and_expr)
//		and_expr     = primary  *( "&&" *SP primary )
//
//		primary      = "(" *SP precondition ")" *SP
//		             / "\complete(" *SP ("*" / node_id) *SP ")" *SP
//		             / "\exec(" *SP node_id *SP ")" *SP
//		             / "!" *SP precondition *SP
//
//		node_id      = *(DIGIT / ALPHA / "_" / "-" / ".") 1*(DIGIT / ALPHA)
//	 	SP           = " "
func ParsePrecondition(s string, ap pfd.AtomicProcessID) (*fsm.Precondition, error) {
	rs := []rune(s)

	newIndex := parser.SkipRune(Whitespaces, rs, 0)
	if newIndex == len(rs) {
		return &fsm.Precondition{
			Type: fsm.PreconditionTypeTrue,
		}, nil
	}

	p, newIndex := parseOrExpression(rs, newIndex, ap)
	if newIndex != len(rs) {
		return nil, fmt.Errorf("fsmtable.ParsePrecondition: trailing garbage: %q", string(rs[newIndex:]))
	}
	if p == nil {
		return nil, fmt.Errorf("fsmtable.ParsePrecondition: syntax error")
	}
	return p, nil
}

var (
	OrKeyword               = []rune{'|', '|'}
	AndKeyword              = []rune{'&', '&'}
	ParenthesesOpenKeyword  = []rune{'('}
	ParenthesesCloseKeyword = []rune{')'}
	ArrowKeyword            = []rune{'-', '>'}
	NotKeyword              = []rune{'!'}
	CompleteBeginKeyword    = []rune(`\complete(`)
	CompleteEndKeyword      = []rune(`)`)
	AsteriskKeyword         = []rune{'*'}
	ExecutableBeginKeyword  = []rune(`\exec(`)
	ExecutableEndKeyword    = []rune(`)`)
	TrueKeyword             = []rune(`\true`)
	Whitespaces             = sets.New(cmp.Compare, ' ')
	IdentifierSymbols       = sets.New(cmp.Compare, '_', '-', '.')
)

func parseOrExpression(s []rune, index int, ap pfd.AtomicProcessID) (*fsm.Precondition, int) {
	p, newIndex := parseAndExpression(s, index, ap)
	if p == nil {
		return nil, index
	}

	ps := make([]*fsm.Precondition, 1)
	ps[0] = p

	var ok bool
	for {
		ok, newIndex = parser.ExpectKeyword(OrKeyword, s, newIndex)
		if !ok {
			break
		}

		newIndex = parser.SkipRune(Whitespaces, s, newIndex)

		p, newIndex = parseAndExpression(s, newIndex, ap)
		if p == nil {
			break
		}

		ps = append(ps, p)
	}

	if len(ps) == 1 {
		return ps[0], newIndex
	}

	return fsm.NewOrPrecondition(ps...), newIndex
}

func parseAndExpression(s []rune, index int, ap pfd.AtomicProcessID) (*fsm.Precondition, int) {
	p, newIndex := parsePrimary(s, index, ap)
	if p == nil {
		return nil, index
	}

	ps := make([]*fsm.Precondition, 1)
	ps[0] = p

	var ok bool
	for {
		ok, newIndex = parser.ExpectKeyword(AndKeyword, s, newIndex)
		if !ok {
			break
		}

		newIndex = parser.SkipRune(Whitespaces, s, newIndex)

		p, newIndex = parsePrimary(s, newIndex, ap)
		if p == nil {
			break
		}

		ps = append(ps, p)
	}

	if len(ps) == 1 {
		return ps[0], newIndex
	}

	return fsm.NewAndPrecondition(ps...), newIndex
}

func parsePrimary(s []rune, index int, ap pfd.AtomicProcessID) (*fsm.Precondition, int) {
	p, newIndex := parseParenthesizedPrecondition(s, index, ap)
	if p != nil {
		return p, newIndex
	}

	p, newIndex = parseComplete(s, newIndex, ap)
	if p != nil {
		return p, newIndex
	}

	p, newIndex = parseExecutable(s, newIndex)
	if p != nil {
		return p, newIndex
	}

	p, newIndex = parseNotExpression(s, newIndex, ap)
	if p != nil {
		return p, newIndex
	}

	p, newIndex = parseTrue(s, newIndex)
	if p != nil {
		return p, newIndex
	}

	return nil, index
}

func parseComplete(s []rune, index int, ap pfd.AtomicProcessID) (*fsm.Precondition, int) {
	ok, newIndex := parser.ExpectKeyword(CompleteBeginKeyword, s, index)
	if !ok {
		return nil, index
	}
	newIndex = parser.SkipRune(Whitespaces, s, newIndex)

	ok, newIndex = parser.ExpectKeyword(AsteriskKeyword, s, newIndex)
	if ok {
		newIndex = parser.SkipRune(Whitespaces, s, newIndex)

		ok, newIndex = parser.ExpectKeyword(CompleteEndKeyword, s, newIndex)
		if ok {
			return fsm.NewAllBackwardReachableFeedbackSourcesCompleted(ap), parser.SkipRune(Whitespaces, s, newIndex)
		}
		return nil, index
	}

	newIndex = parser.SkipRune(Whitespaces, s, newIndex)

	var id pfd.NodeID
	ok, id, newIndex = parseNodeID(s, newIndex)
	if !ok {
		return nil, index
	}

	ok, newIndex = parser.ExpectKeyword(CompleteEndKeyword, s, newIndex)
	if !ok {
		return nil, index
	}

	return fsm.NewFeedbackSourceCompletedPrecondition(pfd.AtomicDeliverableID(id)), parser.SkipRune(Whitespaces, s, newIndex)
}

func parseExecutable(s []rune, index int) (*fsm.Precondition, int) {
	ok, newIndex := parser.ExpectKeyword(ExecutableBeginKeyword, s, index)
	if !ok {
		return nil, index
	}
	newIndex = parser.SkipRune(Whitespaces, s, newIndex)

	var id pfd.NodeID
	ok, id, newIndex = parseNodeID(s, newIndex)
	if !ok {
		return nil, index
	}

	ok, newIndex = parser.ExpectKeyword(ExecutableEndKeyword, s, newIndex)
	if !ok {
		return nil, index
	}

	return fsm.NewExecutablePrecondition(pfd.AtomicProcessID(id)), parser.SkipRune(Whitespaces, s, newIndex)
}

func parseNotExpression(s []rune, index int, ap pfd.AtomicProcessID) (*fsm.Precondition, int) {
	ok, newIndex := parser.ExpectKeyword(NotKeyword, s, index)
	if !ok {
		return nil, index
	}
	newIndex = parser.SkipRune(Whitespaces, s, newIndex)

	var p *fsm.Precondition
	p, newIndex = parseOrExpression(s, newIndex, ap)
	if p == nil {
		return nil, index
	}
	return fsm.NewNotPrecondition(p), parser.SkipRune(Whitespaces, s, newIndex)
}

func parseTrue(s []rune, index int) (*fsm.Precondition, int) {
	ok, newIndex := parser.ExpectKeyword(TrueKeyword, s, index)
	if !ok {
		return nil, index
	}
	return fsm.NewTruePrecondition(), parser.SkipRune(Whitespaces, s, newIndex)
}

var isNodeIDPrefixRune = parser.Or(parser.IsDigit, parser.IsAlpha, parser.Contains(IdentifierSymbols))

func parseNodeID(s []rune, index int) (bool, pfd.NodeID, int) {
	runes, newIndex := parser.AdvanceUntil(isNodeIDPrefixRune, s, index)

	if len(runes) == 0 {
		return false, "", index
	}

	if runes[len(runes)-1] == '-' {
		if newIndex < len(s) && s[newIndex] == '>' {
			return true, pfd.NodeID(runes[:len(runes)-1]), newIndex - 1
		}
	}

	return true, pfd.NodeID(runes), parser.SkipRune(Whitespaces, s, newIndex)
}

func parseParenthesizedPrecondition(s []rune, index int, ap pfd.AtomicProcessID) (*fsm.Precondition, int) {
	ok, newIndex := parser.ExpectKeyword(ParenthesesOpenKeyword, s, index)
	if !ok {
		return nil, index
	}

	newIndex = parser.SkipRune(Whitespaces, s, newIndex)

	p, newIndex := parseOrExpression(s, newIndex, ap)
	if p != nil {
		return p, newIndex
	}

	ok, newIndex = parser.ExpectKeyword(ParenthesesCloseKeyword, s, newIndex)
	if !ok {
		return nil, index
	}

	return p, parser.SkipRune(Whitespaces, s, newIndex)
}
