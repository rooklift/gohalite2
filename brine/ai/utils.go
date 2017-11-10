package ai

import (
	"math"
)

func StringSliceIndex(slice []string, s string) int {
	for i, item := range slice {
		if item == s {
			return i
		}
	}
	return -1
}

func StringSliceContains(slice []string, s string) bool {
	return StringSliceIndex(slice, s) != -1
}

func Dist(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx * dx + dy * dy)
}

func Max(a, b int) int {
	if a > b { return a }
	return b
}

func Min(a, b int) int {
	if a < b { return a }
	return b
}
