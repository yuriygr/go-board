package utils

// Abs - Просто потому что я могу так сделать
func Abs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}
