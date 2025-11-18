package parser

import (
	"cmp"
	"slices"

	"github.com/Kuniwak/pfd-tools/sets"
)

func Or(predicates ...func(rune) bool) func(rune) bool {
	return func(r rune) bool {
		for _, predicate := range predicates {
			if predicate(r) {
				return true
			}
		}
		return false
	}
}

func IsDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func IsAlpha(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func Contains(s *sets.Set[rune]) func(rune) bool {
	return func(r rune) bool {
		return s.Contains(cmp.Compare, r)
	}
}

func AdvanceUntil(predicate func(rune) bool, text []rune, index int) ([]rune, int) {
	if index >= len(text) {
		return nil, index
	}
	runes := make([]rune, 0, len(text)-index)
	newIndex := index
	for newIndex < len(text) {
		r := text[newIndex]
		if !predicate(r) {
			return runes, newIndex
		}
		runes = append(runes, r)
		newIndex++
	}
	return runes, newIndex
}

func SkipRune(skippingRunes *sets.Set[rune], text []rune, index int) int {
	newIndex := index
	for newIndex < len(text) {
		r := text[newIndex]
		if skippingRunes.Contains(cmp.Compare, r) {
			newIndex++
		} else {
			break
		}
	}
	return newIndex
}

func ExpectKeyword(expected []rune, text []rune, index int) (bool, int) {
	expectedLen := len(expected)
	if index >= len(text) || index+expectedLen > len(text) {
		return false, index
	}
	if slices.Equal(text[index:index+expectedLen], expected) {
		return true, index + expectedLen
	}
	return false, index
}
