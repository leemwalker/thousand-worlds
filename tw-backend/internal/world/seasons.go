package world

import "time"

// Season represents the current season
type Season string

const (
	SeasonSpring Season = "Spring"
	SeasonSummer Season = "Summer"
	SeasonAutumn Season = "Autumn"
	SeasonWinter Season = "Winter"
)

// DefaultSeasonLength is 90 game-days
const DefaultSeasonLength = 90 * 24 * time.Hour

// CalculateSeason calculates the current season and progress (0.0-1.0)
func CalculateSeason(gameTime time.Duration, seasonLength time.Duration) (Season, float64) {
	if seasonLength <= 0 {
		return SeasonSpring, 0.0
	}

	yearLength := seasonLength * 4
	yearProgress := gameTime % yearLength

	seasonIndex := int(yearProgress / seasonLength)
	seasonProgress := float64(yearProgress%seasonLength) / float64(seasonLength)

	switch seasonIndex {
	case 0:
		return SeasonSpring, seasonProgress
	case 1:
		return SeasonSummer, seasonProgress
	case 2:
		return SeasonAutumn, seasonProgress
	case 3:
		return SeasonWinter, seasonProgress
	default:
		return SeasonSpring, seasonProgress
	}
}
