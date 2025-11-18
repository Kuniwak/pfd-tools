package fsm

import (
	"testing"
)

func TestExponentialReworkVolumeFunc(t *testing.T) {
	init := ConstInitialVolumeFunc(100)
	f := ExponentialReworkVolumeFunc(0.1, init)

	t.Run("1/5", func(t *testing.T) {
		vol := f("P1", 1)
		if !vol.ApproximateEqual(10) {
			t.Errorf("got %v, expected %v", vol, 10)
		}
	})
	t.Run("2/5", func(t *testing.T) {
		vol := f("P1", 2)
		if !vol.ApproximateEqual(1) {
			t.Errorf("got %v, expected %v", vol, 1)
		}
	})
	t.Run("3/5", func(t *testing.T) {
		vol := f("P1", 3)
		if !vol.ApproximateEqual(0.1) {
			t.Errorf("got %v, expected %v", vol, 0.1)
		}
	})
	t.Run("4/5", func(t *testing.T) {
		vol := f("P1", 4)
		if !vol.ApproximateEqual(0.01) {
			t.Errorf("got %v, expected %v", vol, 0.01)
		}
	})
	t.Run("5/5", func(t *testing.T) {
		vol := f("P1", 5)
		if !vol.ApproximateEqual(0.001) {
			t.Errorf("got %v, expected %v", vol, 0.001)
		}
	})
	t.Run("6/5", func(t *testing.T) {
		vol := f("P1", 6)
		if !vol.ApproximateEqual(0.0001) {
			t.Errorf("got %v, expected %v", vol, 0)
		}
	})
}

func FakeReworkVolumeFunc(initVolumeFunc InitialVolumeFunc) ReworkVolumeFunc {
	return ExponentialReworkVolumeFunc(0.5, initVolumeFunc)
}
