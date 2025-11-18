package sets

import (
	"iter"
	"slices"
)

// Set is a set.
type Set[T any] []T

// New creates a set from the given elements.
func New[T any](f func(a, b T) int, xs ...T) *Set[T] {
	ys := Set[T](make([]T, 0, len(xs)))
	for _, x := range xs {
		ys.Add(f, x)
	}
	return &ys
}

// NewWithCapacity creates a set from the given element count.
func NewWithCapacity[T any](capacity int) *Set[T] {
	xs := make([]T, 0, capacity)
	ys := Set[T](xs)
	return &ys
}

// Compare compares sets. To create a total order, the ordering differs from set inclusion relationships.
func Compare[T any](f func(a, b T) int) func(a, b *Set[T]) int {
	return func(a, b *Set[T]) int {
		return slices.CompareFunc(*a, *b, f)
	}
}

// CompareFamily compares families of sets. To create a total order, the ordering differs from set inclusion relationships.
func CompareFamily[T any](f func(a, b T) int) func(a, b *Set[*Set[T]]) int {
	f2 := Compare(f)
	return func(a, b *Set[*Set[T]]) int {
		return slices.CompareFunc(*a, *b, f2)
	}
}

func SliceOfFamily[T any](a *Set[*Set[T]]) [][]T {
	xs := make([][]T, 0, a.Len())
	for _, s := range a.Iter() {
		xs = append(xs, s.Slice())
	}
	return xs
}

func SliceOfSetOfFamily[T any](a *Set[*Set[*Set[T]]]) [][][]T {
	xs := make([][][]T, 0, a.Len())
	for _, s := range a.Iter() {
		xs = append(xs, SliceOfFamily(s))
	}
	return xs
}

// Len returns the number of elements.
func (s *Set[T]) Len() int {
	if s == nil {
		return 0
	}
	return len(*s)
}

// Iter returns the elements of the set with an iterator.
func (s *Set[T]) Iter() iter.Seq2[int, T] {
	if s == nil {
		return slices.All[[]T](nil)
	}
	return slices.All(*s)
}

// At returns the element at the given index and true if the given index is within the range of the sequence
// when the set is arranged in total order. If the index is not within the range of the sequence,
// it returns an undefined value and false.
func (s *Set[T]) At(idx int) (T, bool) {
	if idx >= len(*s) {
		var zero T
		return zero, false
	}
	return (*s)[idx], true
}

// Slice returns a sequence of elements arranged in total order.
func (s *Set[T]) Slice() []T {
	xs := slices.Clone(*s)
	return xs
}

// Clone returns a shallow copy of the set.
func (s *Set[T]) Clone() *Set[T] {
	xs := slices.Clone(*s)
	return &xs
}

// Clear empties the set.
func (s *Set[T]) Clear() { *s = (*s)[:0] }

// Contains returns whether the element belongs to the set.
func (s *Set[T]) Contains(f func(a, b T) int, v T) bool {
	_, ok := slices.BinarySearchFunc(*s, v, f)
	return ok
}

// IndexOf returns the index when the set's elements are arranged in total order. Returns -1 if the element does not exist.
func (s *Set[T]) IndexOf(f func(a, b T) int, v T) int {
	i, ok := slices.BinarySearchFunc(*s, v, f)
	if !ok {
		return -1
	}
	return i
}

// Add adds an element.
func (s *Set[T]) Add(f func(a, b T) int, v T) {
	i, ok := slices.BinarySearchFunc(*s, v, f)
	if ok {
		return
	}
	*s = append(*s, v)         // reserve one space at the end
	copy((*s)[i+1:], (*s)[i:]) // shift right
	(*s)[i] = v
}

// Remove removes an element.
func (s *Set[T]) Remove(f func(a, b T) int, v T) {
	i, ok := slices.BinarySearchFunc(*s, v, f)
	if !ok {
		return
	}
	*s = append((*s)[:i], (*s)[i+1:]...)
}

// Union updates itself with the union.
func (s *Set[T]) Union(f func(a, b T) int, t *Set[T]) {
	*s = mergeUnion(*s, *t, f)
}

// Intersection updates itself with the intersection.
func (s *Set[T]) Intersection(f func(a, b T) int, t *Set[T]) {
	*s = mergeIntersection(*s, *t, f)
}

// Difference updates itself with the set difference (s \ t).
func (s *Set[T]) Difference(f func(a, b T) int, t *Set[T]) {
	*s = mergeDifference(*s, *t, f)
}

// IsSubsetOf determines if s ⊆ t.
func (s *Set[T]) IsSubsetOf(f func(a, b T) int, t *Set[T]) bool {
	i, j := 0, 0
	for i < len(*s) && j < len(*t) {
		if f((*s)[i], (*t)[j]) == 0 {
			i, j = i+1, j+1
		} else if f((*s)[i], (*t)[j]) < 0 {
			return false
		} else {
			j++
		}
	}
	return i == len(*s)
}

// IsProperSubsetOf determines if s ⊊ t.
func (s *Set[T]) IsProperSubsetOf(f func(a, b T) int, t *Set[T]) bool {
	return s.IsSubsetOf(f, t) && !Equal(f)(s, t)
}

// IsDisjointWith determines if s and t are disjoint (no common elements).
func (s *Set[T]) IsDisjointWith(f func(a, b T) int, t *Set[T]) bool {
	i, j := 0, 0
	for i < len(*s) && j < len(*t) {
		if f((*s)[i], (*t)[j]) == 0 {
			return false
		}
		if f((*s)[i], (*t)[j]) < 0 {
			i++
		} else {
			j++
		}
	}
	return true
}

func Equal[T any](f func(a, b T) int) func(a, b *Set[T]) bool {
	return func(a, b *Set[T]) bool {
		return IsEqual(f, a, b)
	}
}

func IsEqual[T any](f func(a, b T) int, a, b *Set[T]) bool {
	return slices.EqualFunc(*a, *b, func(a, b T) bool { return f(a, b) == 0 })
}

func mergeUnion[T any](a, b []T, f func(a, b T) int) []T {
	i, j := 0, 0
	res := make([]T, 0, len(a)+len(b))
	for i < len(a) && j < len(b) {
		if f(a[i], b[j]) == 0 {
			res = append(res, a[i])
			i, j = i+1, j+1
		} else if f(a[i], b[j]) < 0 {
			res = append(res, a[i])
			i++
		} else {
			res = append(res, b[j])
			j++
		}
	}
	res = append(res, a[i:]...)
	res = append(res, b[j:]...)
	return res
}

func mergeIntersection[T any](a, b []T, f func(a, b T) int) []T {
	i, j := 0, 0
	res := make([]T, 0, min(len(a), len(b)))
	for i < len(a) && j < len(b) {
		if f(a[i], b[j]) == 0 {
			res = append(res, a[i])
			i, j = i+1, j+1
		} else if f(a[i], b[j]) < 0 {
			i++
		} else {
			j++
		}
	}
	return res
}

func mergeDifference[T any](a, b []T, f func(a, b T) int) []T {
	i, j := 0, 0
	res := make([]T, 0, len(a))
	for i < len(a) && j < len(b) {
		if f(a[i], b[j]) == 0 {
			i, j = i+1, j+1
		} else if f(a[i], b[j]) < 0 {
			res = append(res, a[i])
			i++
		} else {
			j++
		}
	}
	res = append(res, a[i:]...)
	return res
}

// PowerSet returns all subsets (including the empty set).
func PowerSet[T any](s *Set[T], f func(a, b T) int) *Set[*Set[T]] {
	n := s.Len()
	subsets := make([]*Set[T], 0, 1<<uint(min(n, 30))) // reserve conservatively to avoid under/over allocation

	var cur []T
	var dfs func(i int)
	dfs = func(i int) {
		if i == n {
			ys := Set[T](slices.Clone(cur))
			subsets = append(subsets, &ys)
			return
		}
		// don't take
		dfs(i + 1)
		// take (input order = already sorted, so order is preserved)
		cur = append(cur, (*s)[i])
		dfs(i + 1)
		cur = cur[:len(cur)-1]
	}
	dfs(0)

	// each subset preserves element order, but arrange the entire collection in total order
	slices.SortFunc(subsets, Compare(f))

	res := Set[*Set[T]](subsets)
	return &res
}
