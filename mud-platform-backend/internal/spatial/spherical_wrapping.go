package spatial

// NormalizeCoordinates handles spherical wrapping (longitude wrapping and pole crossing)
func NormalizeCoordinates(lat, lon float64) (newLat, newLon float64) {
	newLat = lat
	newLon = lon

	// Handle pole crossing (latitude > 90 or < -90)
	// If we cross the pole, we end up on the opposite side of the world:
	// Lat becomes 180 - lat (e.g., 95 becomes 85)
	// Lon adds 180 degrees
	if newLat > 90 {
		newLat = 180 - newLat
		newLon += 180
	} else if newLat < -90 {
		newLat = -180 - newLat
		newLon += 180
	}

	// Handle longitude wrapping (-180 to 180)
	// We use a loop here to handle extreme values, though usually it's just one wrap
	for newLon > 180 {
		newLon -= 360
	}
	for newLon <= -180 {
		newLon += 360
	}

	return newLat, newLon
}
