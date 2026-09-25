package sugar

import "fmt"

func Must[T any](v T, err error) T {
	if err != nil {
		panic(fmt.Errorf("sugar.Must: %w", err))
	}
	return v
}
