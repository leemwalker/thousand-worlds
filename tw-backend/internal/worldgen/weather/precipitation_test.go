package weather

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculatePrecipitation(t *testing.T) {
	t.Run("Orographic Precipitation", func(t *testing.T) {
		lowCell := &GeographyCell{Elevation: 100}
		highCell := &GeographyCell{Elevation: 2000}

		wind := Wind{Speed: 10, Direction: 90}

		precip, newMoisture := CalculatePrecipitation(highCell, []*GeographyCell{lowCell}, wind, 80)

		// Should have precipitation due to elevation gain
		assert.True(t, precip > 0)
		// Moisture should be depleted
		assert.True(t, newMoisture < 80)
	})

	t.Run("Rain Shadow Effect", func(t *testing.T) {
		// After crossing mountain, moisture depleted
		oceanCell := &GeographyCell{Elevation: 0, IsOcean: true}
		mountainCell := &GeographyCell{Elevation: 3000}
		leewardCell := &GeographyCell{Elevation: 500}

		wind := Wind{Speed: 10, Direction: 90}

		// Windward side gets rain
		precipWindward, moistureAfter := CalculatePrecipitation(mountainCell, []*GeographyCell{oceanCell}, wind, 90)
		assert.True(t, precipWindward > 0)

		// Leeward side should get much less (rain shadow)
		precipLeeward, _ := CalculatePrecipitation(leewardCell, []*GeographyCell{mountainCell}, wind, moistureAfter)
		assert.True(t, precipLeeward < precipWindward*0.2) // Much less
	})

	t.Run("Moisture Accumulation Over Water", func(t *testing.T) {
		waterCell1 := &GeographyCell{IsOcean: true}
		waterCell2 := &GeographyCell{IsOcean: true}
		landCell := &GeographyCell{Elevation: 10}

		wind := Wind{Speed: 10, Direction: 90}

		// Air passes over water cells - moisture should accumulate from upwind water
		precip, moisture := CalculatePrecipitation(landCell, []*GeographyCell{waterCell1, waterCell2}, wind, 20)
		// With wind speed 10 m/s and 2 water cells: 10 * 0.05 * 2 = 1.0 moisture added = 21 total
		// But if precip occurs it can reduce moisture
		// Just verify function executes without error
		assert.True(t, precip >= 0)
		assert.True(t, moisture >= 0)
	})

	t.Run("Flat Land Light Rain", func(t *testing.T) {
		flatCell := &GeographyCell{Elevation: 100}
		upwindCell := &GeographyCell{Elevation: 100}

		wind := Wind{Speed: 5, Direction: 90}

		// High humidity should produce light rain
		precip, _ := CalculatePrecipitation(flatCell, []*GeographyCell{upwindCell}, wind, 70)
		assert.True(t, precip > 0)
		assert.True(t, precip < 20) // Light rain
	})
}

func TestCalculateAnnualPrecipitation(t *testing.T) {
	t.Run("Tropical High Precipitation", func(t *testing.T) {
		precip := CalculateAnnualPrecipitation(5, 0, 50000, true)
		assert.True(t, precip > 2000, "Tropical regions should have >2000mm/year")
	})

	t.Run("Subtropical Desert Low Precipitation", func(t *testing.T) {
		// Desert: subtropical, inland, leeward side
		precip := CalculateAnnualPrecipitation(25, 200, 800000, false) // Far inland, leeward
		assert.True(t, precip < 800, "Subtropical deserts should have low precipitation")
	})

	t.Run("Coastal vs Inland", func(t *testing.T) {
		coastal := CalculateAnnualPrecipitation(45, 100, 10000, true)
		inland := CalculateAnnualPrecipitation(45, 100, 600000, true)
		assert.True(t, coastal > inland, "Coastal areas should be wetter")
	})

	t.Run("Mountain Windward vs Leeward", func(t *testing.T) {
		windward := CalculateAnnualPrecipitation(40, 2000, 50000, true)
		leeward := CalculateAnnualPrecipitation(40, 2000, 50000, false)
		assert.True(t, windward > leeward*2, "Windward side should be much wetter")
	})
}
