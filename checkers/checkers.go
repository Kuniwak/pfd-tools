package checkers

import (
	"fmt"

	"golang.org/x/sync/errgroup"
)

type CheckerID string

type Checker[T any] interface {
	Check(T, chan<- Problem) error
}

type AtomicChecker[T any] struct {
	ID              CheckerID
	AvailableIfFunc func(T) bool
	CheckFunc       func(T, chan<- Problem) error
}

func (c AtomicChecker[T]) Check(t T, ch chan<- Problem) error {
	if !c.AvailableIfFunc(t) {
		return nil
	}
	return c.CheckFunc(t, ch)
}

type ParallelChecker[T any] struct {
	cs []Checker[T]
}

func NewParallelChecker[T any](cs ...Checker[T]) ParallelChecker[T] {
	return ParallelChecker[T]{cs: cs}
}

func (c ParallelChecker[T]) Check(t T, ch chan<- Problem) error {
	var eg errgroup.Group
	for _, c := range c.cs {
		eg.Go(func() error {
			if err := c.Check(t, ch); err != nil {
				return fmt.Errorf("checkers.ParallelChecker.Check: %w", err)
			}
			return nil
		})
	}
	return eg.Wait()
}
