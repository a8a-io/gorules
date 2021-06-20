package utils

import "time"

func InBetween(ts time.Time, fromTime time.Time, toTime time.Time) bool {
	return (ts.Equal(fromTime) || ts.After(fromTime)) &&
		(ts.Equal(toTime) || ts.Before(toTime))
}

func Contains(arr []string, val string) bool {
	for _, a := range arr {
		if a == val {
			return true
		}
	}
	return false
}

func MaxInt(x, y int) int {
	if x < y {
		return y
	}
	return x
}

// Min returns the smaller of x or y.
func MinInt(x, y int) int {
	if x > y {
		return y
	}
	return x
}
