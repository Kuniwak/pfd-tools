package sugar

func IgnoreError(f func() error) {
	_ = f()
}
