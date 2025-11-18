package pairs

type Pair[T1, T2 any] struct {
	First  T1 `json:"first"`
	Second T2 `json:"second"`
}

func New[T1, T2 any](first T1, second T2) *Pair[T1, T2] {
	return &Pair[T1, T2]{First: first, Second: second}
}

func Compare[T1, T2 any](f1 func(a, b T1) int, f2 func(a, b T2) int) func(a, b *Pair[T1, T2]) int {
	return func(a, b *Pair[T1, T2]) int {
		c := f1(a.First, b.First)
		if c != 0 {
			return c
		}
		return f2(a.Second, b.Second)
	}
}

func CompareFirst[T1, T2 any](f func(a, b T1) int) func(a, b *Pair[T1, T2]) int {
	return func(a, b *Pair[T1, T2]) int {
		return f(a.First, b.First)
	}
}

func CompareSecond[T1, T2 any](f func(a, b T2) int) func(a, b *Pair[T1, T2]) int {
	return func(a, b *Pair[T1, T2]) int {
		return f(a.Second, b.Second)
	}
}

func (p *Pair[T1, T2]) Clone() *Pair[T1, T2] {
	return &Pair[T1, T2]{First: p.First, Second: p.Second}
}
