package execmodel

import "math"

type Time float64

const MinimumTime = 0.01

func (t Time) ApproximateEqual(b Time) bool {
	return math.Abs(float64(t-b)) < MinimumTime
}
