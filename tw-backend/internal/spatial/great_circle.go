package spatial

import "math"

// GreatCircleDistance calculates the shortest distance between two points on a sphere
func GreatCircleDistance(lat1, lon1, lat2, lon2, radius float64) float64 {
	lat1Rad := degToRad(lat1)
	lon1Rad := degToRad(lon1)
	lat2Rad := degToRad(lat2)
	lon2Rad := degToRad(lon2)

	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	// Haversine formula:
	// a = sin²(Δlat/2) + cos(lat1) * cos(lat2) * sin²(Δlon/2)
	// c = 2 * atan2(√a, √(1-a))
	// d = R * c

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return radius * c
}
