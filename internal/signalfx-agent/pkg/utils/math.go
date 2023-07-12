package utils

// MaxInt returns the greater of x and y
func MaxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// MinInt returns the lesser of x and y
func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}
