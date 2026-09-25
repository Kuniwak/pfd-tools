package sugar_test

import (
	"errors"
	"testing"

	"github.com/Kuniwak/pfd-tools/sugar"
)

func TestMust(t *testing.T) {
	t.Run("returns the value when err is nil", func(t *testing.T) {
		got := sugar.Must(valueOrError(42, nil))
		if got != 42 {
			t.Errorf("Must = %d, want 42", got)
		}
	})

	t.Run("panics when err is non-nil", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("Must did not panic on error")
			}
		}()
		_ = sugar.Must(valueOrError(0, errors.New("boom")))
	})
}

func valueOrError[T any](v T, err error) (T, error) {
	return v, err
}
