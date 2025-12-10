package minerals

import (
	"math"
	"math/rand"

	"tw-backend/internal/worldgen/geography"
)

// GenerateCluster generates a cluster of mineral veins around a primary vein
func GenerateCluster(
	primary *MineralDeposit,
	context *TectonicContext,
) []*MineralDeposit {
	deposits := []*MineralDeposit{primary}

	// 60% chance of secondary veins
	if rand.Float64() < 0.6 {
		numSecondary := 1 + rand.Intn(3) // 1-3 secondary veins
		for i := 0; i < numSecondary; i++ {
			secondary := generateSatelliteVein(primary, context, 0.5, 0.8, 50, 500)
			deposits = append(deposits, secondary)
		}
	}

	// 30% chance of tertiary veins
	if rand.Float64() < 0.3 {
		numTertiary := 1 + rand.Intn(4) // 1-4 tertiary veins
		for i := 0; i < numTertiary; i++ {
			tertiary := generateSatelliteVein(primary, context, 0.2, 0.5, 500, 2000)
			deposits = append(deposits, tertiary)
		}
	}

	return deposits
}

func generateSatelliteVein(
	primary *MineralDeposit,
	context *TectonicContext,
	sizeMinScale, sizeMaxScale float64,
	distMin, distMax float64,
) *MineralDeposit {
	// Calculate location offset
	angle := rand.Float64() * 2 * math.Pi
	dist := distMin + rand.Float64()*(distMax-distMin)
	offsetX := math.Cos(angle) * dist
	offsetY := math.Sin(angle) * dist

	// New location (simplified, assuming flat plane for offset)
	// In a real spherical world, this would need projection logic, but for local clusters it's fine
	newLoc := geography.Point{
		X: primary.Location.X + offsetX,
		Y: primary.Location.Y + offsetY,
	}

	// Scale quantity and size
	scale := sizeMinScale + rand.Float64()*(sizeMaxScale-sizeMinScale)

	// Create a copy of the primary mineral type
	mineralType := primary.MineralType

	// Generate the vein
	// We manually construct it to enforce the scaling relative to primary

	// Determine size category based on scale relative to primary
	// This is a simplification; ideally we'd map scale to VeinSize enum
	var newSize VeinSize
	if scale > 0.8 {
		newSize = primary.VeinSize
	} else if scale > 0.5 {
		// Downgrade size
		newSize = downgradeSize(primary.VeinSize)
	} else {
		newSize = downgradeSize(downgradeSize(primary.VeinSize))
	}

	vein := GenerateMineralVein(context, mineralType, newLoc)
	vein.VeinSize = newSize

	// Recalculate quantity based on new size
	baseQty := GetBaseQuantity(mineralType, newSize)
	vein.Quantity = int(float64(baseQty) * vein.Concentration)

	return vein
}

func downgradeSize(size VeinSize) VeinSize {
	switch size {
	case VeinSizeMassive:
		return VeinSizeLarge
	case VeinSizeLarge:
		return VeinSizeMedium
	case VeinSizeMedium:
		return VeinSizeSmall
	case VeinSizeSmall:
		return VeinSizeSmall
	}
	return VeinSizeSmall
}
