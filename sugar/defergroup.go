package sugar

import "fmt"

// DeferGroup provides an idiom to execute defer cleanup in each loop iteration.
// Since defer is executed when leaving a function, using defer in a loop causes
// defers to accumulate until leaving the function containing the loop.
type DeferGroup struct{}

// Run executes the given function immediately.
func (g DeferGroup) Run(f func() error) error {
	if err := f(); err != nil {
		return fmt.Errorf("sugar.DeferGroup.Run: %w", err)
	}
	return nil
}
