package cmp2

func CompareBool(a, b bool) int {
	if a {
		if !b {
			return 1
		}
	} else {
		if b {
			return -1
		}
	}
	return 0
}
