package main

import (
	"math"
)

func Square(x float64) float64 {
	return x * x
}

func RadiansFromDegrees(Degrees float64) float64 {
	return 0.01745329251994329577 * Degrees
}

func ReferenceHaversine(X0 float64, Y0 float64, X1 float64, Y1 float64, EarthRadius float64) float64 {
	/* NOTE(casey): This is not meant to be a "good" way to calculate the Haversine distance.
	   Instead, it attempts to follow, as closely as possible, the formula used in the real-world
	   question on which these homework exercises are loosely based.
	*/
	lat1 := Y0
	lat2 := Y1
	lon1 := X0
	lon2 := X1

	dLat := RadiansFromDegrees(lat2 - lat1)
	dLon := RadiansFromDegrees(lon2 - lon1)
	lat1 = RadiansFromDegrees(lat1)
	lat2 = RadiansFromDegrees(lat2)

	a := Square(math.Sin(dLat/2.0)) + math.Cos(lat1)*math.Cos(lat2)*Square(math.Sin(dLon/2))
	c := 2.0 * math.Asin(math.Sqrt(a))

	return EarthRadius * c
}
