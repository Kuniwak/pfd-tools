package sugar

import "fmt"

type DeferGroup struct{}

func (g DeferGroup) Run(f func() error) error {
	if err := f(); err != nil {
		return fmt.Errorf("sugar.DeferGroup.Run: %w", err)
	}
	return nil
}
