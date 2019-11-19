package utils

// Abs - Просто потому что я могу так сделать
func Abs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}

// LimitMaxValue - Я просто линивый
func LimitMaxValue(n int64, max int64) int64 {
	if n > max {
		return max
	}
	return n
}

// LimitMinValue - Крайне ленивый
func LimitMinValue(n int64, min int64) int64 {
	if n < min {
		return min
	}
	return n
}
