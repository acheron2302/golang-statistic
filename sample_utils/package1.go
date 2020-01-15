package sample_utils

import (
	"math"
)

func LinearVariance(alpha, beta float64, point_x, point_y []float64) float64 {
	var variance_square float64

	for index := range point_x {
		variance_square += math.Pow(point_y[index]-beta*point_x[index]-alpha, 2)
	}

	variance_square = variance_square / float64(len(point_x)-2)
	return math.Sqrt(variance_square)
}

func SumSquare(x []float64) float64 {
	var sum float64

	for _, each_element := range x {
		sum += each_element * each_element
	}

	return sum
}
