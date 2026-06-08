package poisson

import (
	"math"
	"math/rand"
)

func Sample(rng *rand.Rand, lambda float64) int {
	if lambda <= 0 {
		return 0
	}

	limit := math.Exp(-lambda)
	product := 1.0
	k := 0

	for product > limit {
		k++
		product *= rng.Float64()
	}

	return k - 1
}
