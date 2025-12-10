package spatial

import "math"

// ToCartesian converts spherical coordinates (lat/lon in degrees) to Cartesian (x, y, z)
// Z axis is North/South pole
func ToCartesian(lat, lon, radius float64) (x, y, z float64) {
	latRad := degToRad(lat)
	lonRad := degToRad(lon)

	// Formula:
	// X = R * cos(lat) * cos(lon)
	// Y = R * cos(lat) * sin(lon)
	// Z = R * sin(lat)

	cosLat := math.Cos(latRad)

	x = radius * cosLat * math.Cos(lonRad)
	y = radius * cosLat * math.Sin(lonRad)
	z = radius * math.Sin(latRad)

	return x, y, z
}

// ToLatLon converts Cartesian coordinates (x, y, z) to spherical (lat/lon in degrees)
func ToLatLon(x, y, z, radius float64) (lat, lon float64) {
	// Formula:
	// lat = arcsin(Z / R)
	// lon = arctan2(Y, X)

	latRad := math.Asin(z / radius)
	lonRad := math.Atan2(y, x)

	return radToDeg(latRad), radToDeg(lonRad)
}

func degToRad(deg float64) float64 {
	return deg * math.Pi / 180
}

func radToDeg(rad float64) float64 {
	return rad * 180 / math.Pi
}
