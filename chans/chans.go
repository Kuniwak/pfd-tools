package chans

func Slice[T any](ch <-chan T) []T {
	res := make([]T, 0)
	for v := range ch {
		res = append(res, v)
	}
	return res
}

func From[T any](values []T) <-chan T {
	ch := make(chan T)
	go func() {
		for _, v := range values {
			ch <- v
		}
		close(ch)
	}()
	return ch
}
