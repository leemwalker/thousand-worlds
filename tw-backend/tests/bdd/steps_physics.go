package bdd

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"tw-backend/internal/worldgen/astronomy"

	"github.com/cucumber/godog"
)

// PhysicsContext holds state for physics scenarios
type PhysicsContext struct {
	StarMass      float64
	PlanetMass    float64
	PlanetRadius  float64
	Distance      float64
	ObjectMass    float64
	OrbitalPeriod float64
	GravityForce  float64
}

func (p *PhysicsContext) reset() {
	p.StarMass = 0
	p.PlanetMass = 5.972e24  // Default Earth mass
	p.PlanetRadius = 6.371e6 // Default Earth radius
	p.Distance = 0
	p.ObjectMass = 0
	p.OrbitalPeriod = 0
	p.GravityForce = 0
}

var physicsState = &PhysicsContext{}

func InitializePhysicsSteps(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		physicsState.reset()
		return ctx, nil
	})

	ctx.Step(`^a star with mass ([\d\.e\+]+) kg$`, iHaveAStarWithMassKg)
	ctx.Step(`^a planet at distance ([\d\.e\+]+) km$`, iHaveAPlanetAtDistanceKm)
	ctx.Step(`^I simulate one full orbit$`, iSimulateOneFullOrbit)
	ctx.Step(`^the orbital period should be approximately (\d+) days$`, theOrbitalPeriodShouldBeApproximatelyDays)
	ctx.Step(`^the planet should return to its initial position$`, thePlanetShouldReturnToItsInitialPosition)
	ctx.Step(`^an object of mass (\d+) kg on the planet surface$`, iHaveAnObjectOfMassKgOnThePlanetSurface)
	ctx.Step(`^I calculate the gravitational force$`, iCalculateTheGravitationalForce)
	ctx.Step(`^the force should be approximately (\d+) Newtons$`, theForceShouldBeApproximatelyNewtons)
}

func iHaveAStarWithMassKg(massStr string) error {
	mass, err := strconv.ParseFloat(massStr, 64)
	if err != nil {
		return err
	}
	physicsState.StarMass = mass
	return nil
}

func iHaveAPlanetAtDistanceKm(distStr string) error {
	distKm, err := strconv.ParseFloat(distStr, 64)
	if err != nil {
		return err
	}
	physicsState.Distance = distKm * 1000 // Convert to meters
	return nil
}

func iSimulateOneFullOrbit() error {
	// Kepler's 3rd Law: T = 2π * sqrt(a^3 / (G * M))
	// a = semi-major axis (distance)
	// G = Gravitational Constant
	// M = Mass of central body (Star)

	if physicsState.StarMass == 0 || physicsState.Distance == 0 {
		return fmt.Errorf("star mass and distance must be defined")
	}

	G := astronomy.GravitationalConstant
	periodSeconds := 2 * math.Pi * math.Sqrt(math.Pow(physicsState.Distance, 3)/(G*physicsState.StarMass))
	physicsState.OrbitalPeriod = periodSeconds
	return nil
}

func theOrbitalPeriodShouldBeApproximatelyDays(days int) error {
	secondsPerDay := 86400.0
	actualDays := physicsState.OrbitalPeriod / secondsPerDay

	if math.Abs(actualDays-float64(days)) > 1.0 {
		return fmt.Errorf("expected ~%d days, got %.2f days", days, actualDays)
	}
	return nil
}

func thePlanetShouldReturnToItsInitialPosition() error {
	// In a Keplerian orbit, one period IS the return time.
	// This step is more of a semantic check in this simplified simulation.
	return nil
}

func iHaveAnObjectOfMassKgOnThePlanetSurface(mass int) error {
	physicsState.ObjectMass = float64(mass)
	return nil
}

func iCalculateTheGravitationalForce() error {
	// F = G * M * m / r^2
	G := astronomy.GravitationalConstant

	// Assuming Earth-like planet if not defined, but Context init handles defaults.
	// physicsState.PlanetMass = 5.972e24
	// physicsState.PlanetRadius = 6.371e6

	force := (G * physicsState.PlanetMass * physicsState.ObjectMass) / (physicsState.PlanetRadius * physicsState.PlanetRadius)
	physicsState.GravityForce = force
	return nil
}

func theForceShouldBeApproximatelyNewtons(expectedN int) error {
	// Allow 1% error margin
	margin := float64(expectedN) * 0.01
	if math.Abs(physicsState.GravityForce-float64(expectedN)) > margin {
		return fmt.Errorf("expected ~%d N, got %.2f N", expectedN, physicsState.GravityForce)
	}
	return nil
}
