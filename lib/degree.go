package lib

import "math"

func Sin(dir int16) float64 {
	r := float64(dir) / 180 * math.Pi
	return math.Sin(r)
}

func Cos(dir int16) float64 {
	r := float64(dir) / 180 * math.Pi
	return math.Cos(r)
}
