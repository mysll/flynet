package util

import (
	"math/rand"
)

func RandRangeI32(min int32, max int32) int32 {
	if max < min || max-min == 0 {
		return 0
	}
	return min + rand.Int31n(max-min)
}

func RandRange(min int, max int) int {
	if max < min || max-min == 0 {
		return 0
	}
	return min + rand.Intn(max-min)
}

func RandRangef(min float32, max float32) float32 {
	return min + (max-min)*rand.Float32()
}
