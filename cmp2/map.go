package cmp2

import (
	"maps"
	"slices"
)

func CompareSlice[S ~[]E, E any](cmp func(E, E) int) func(S, S) int {
	return func(a, b S) int {
		return slices.CompareFunc(a, b, cmp)
	}
}

func CompareMap[K comparable, V any](a, b map[K]V, compareKeyFunc func(a, b K) int, compareValueFunc func(a, b V) int) int {
	k1 := slices.Collect(maps.Keys(a))
	k2 := slices.Collect(maps.Keys(b))
	slices.SortFunc(k1, compareKeyFunc)
	slices.SortFunc(k2, compareKeyFunc)
	c := slices.CompareFunc(k1, k2, compareKeyFunc)
	if c != 0 {
		return c
	}

	v1 := make([]V, 0, len(a))
	v2 := make([]V, 0, len(b))
	for _, k := range k1 {
		v1 = append(v1, a[k])
	}
	for _, k := range k2 {
		v2 = append(v2, b[k])
	}
	return slices.CompareFunc(v1, v2, compareValueFunc)
}
