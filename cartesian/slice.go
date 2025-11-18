package cartesian

import "github.com/Kuniwak/pfd-tools/pairs"

func Slice[T any, S any](xs []T, ys []S) []pairs.Pair[T, S] {
	res := make([]pairs.Pair[T, S], 0)
	for _, x := range xs {
		for _, y := range ys {
			res = append(res, pairs.Pair[T, S]{First: x, Second: y})
		}
	}
	return res
}
