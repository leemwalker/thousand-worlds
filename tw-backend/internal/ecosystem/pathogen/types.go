// Package pathogen provides disease and pathogen simulation for the ecosystem.
// This implements various pathogen types with realistic epidemiological properties.
package pathogen

import (
	"math/rand"

	"github.com/google/uuid"
)

// PathogenType represents the category of pathogen
type PathogenType string

const (
	PathogenVirus    PathogenType = "virus"    // Fast-mutating, high transmission
	PathogenBacteria PathogenType = "bacteria" // Variable, can be treated
	PathogenFungus   PathogenType = "fungus"   // Environmental persistence, slow
	PathogenPrion    PathogenType = "prion"    // No mutation, always fatal, rare
	PathogenParasite PathogenType = "parasite" // Complex lifecycle, chronic
)

// TransmissionMode describes how the pathogen spreads
type TransmissionMode string

const (
	TransmissionAirborne TransmissionMode = "airborne" // Respiratory
	TransmissionContact  TransmissionMode = "contact"  // Direct contact
	TransmissionVector   TransmissionMode = "vector"   // Via another species
	TransmissionWater    TransmissionMode = "water"    // Waterborne
	TransmissionFood     TransmissionMode = "food"     // Foodborne
	TransmissionVertical TransmissionMode = "vertical" // Parent to offspring
)

// Pathogen represents a disease-causing agent
type Pathogen struct {
	ID              uuid.UUID        `json:"id"`
	Name            string           `json:"name"`
	Type            PathogenType     `json:"pathogen_type"`
	Transmission    TransmissionMode `json:"transmission"`
	OriginSpeciesID uuid.UUID        `json:"origin_species_id"` // Original host

	// Epidemiological properties (0.0 - 1.0 scale)
	Virulence        float32 `json:"virulence"`        // Damage to host (high = more deadly)
	Transmissibility float32 `json:"transmissibility"` // Ease of spread (R0 factor)
	Latency          float32 `json:"latency"`          // Time before symptoms (high = longer)
	Persistence      float32 `json:"persistence"`      // Environmental survival
	MutationRate     float32 `json:"mutation_rate"`    // How fast it evolves

	// Host interactions
	HostSpecificity  float32  `json:"host_specificity"`  // 0=broad, 1=narrow host range
	SusceptibleDiets []string `json:"susceptible_diets"` // Which diets are affected
	ResistanceGene   int      `json:"resistance_gene"`   // Which gene provides resistance

	// Evolution tracking
	Generation     int64 `json:"generation"`
	OriginYear     int64 `json:"origin_year"`
	MutationsCount int   `json:"mutations_count"`
	IsEndemic      bool  `json:"is_endemic"` // True if long-established
	IsEradicated   bool  `json:"is_eradicated"`

	// Current state
	ActiveOutbreaks int   `json:"active_outbreaks"`
	TotalInfected   int64 `json:"total_infected"`
	TotalDeaths     int64 `json:"total_deaths"`
}

// NewPathogen creates a new pathogen with random properties based on type
func NewPathogen(name string, pType PathogenType, originSpeciesID uuid.UUID, originYear int64, rng *rand.Rand) *Pathogen {
	p := &Pathogen{
		ID:               uuid.New(),
		Name:             name,
		Type:             pType,
		OriginSpeciesID:  originSpeciesID,
		OriginYear:       originYear,
		Generation:       0,
		SusceptibleDiets: make([]string, 0),
	}

	// Set properties based on pathogen type
	switch pType {
	case PathogenVirus:
		p.Virulence = 0.3 + rng.Float32()*0.5        // 0.3-0.8
		p.Transmissibility = 0.4 + rng.Float32()*0.5 // 0.4-0.9 (high)
		p.Latency = rng.Float32() * 0.5              // 0.0-0.5 (short-medium)
		p.Persistence = rng.Float32() * 0.3          // 0.0-0.3 (low)
		p.MutationRate = 0.3 + rng.Float32()*0.5     // 0.3-0.8 (high)
		p.HostSpecificity = 0.3 + rng.Float32()*0.5  // 0.3-0.8
		p.Transmission = TransmissionAirborne

	case PathogenBacteria:
		p.Virulence = 0.2 + rng.Float32()*0.6        // 0.2-0.8
		p.Transmissibility = 0.2 + rng.Float32()*0.5 // 0.2-0.7
		p.Latency = 0.1 + rng.Float32()*0.4          // 0.1-0.5
		p.Persistence = 0.2 + rng.Float32()*0.5      // 0.2-0.7
		p.MutationRate = 0.1 + rng.Float32()*0.3     // 0.1-0.4
		p.HostSpecificity = 0.2 + rng.Float32()*0.6
		p.Transmission = TransmissionContact

	case PathogenFungus:
		p.Virulence = 0.1 + rng.Float32()*0.4        // 0.1-0.5 (lower)
		p.Transmissibility = 0.1 + rng.Float32()*0.3 // 0.1-0.4 (slow)
		p.Latency = 0.3 + rng.Float32()*0.5          // 0.3-0.8 (long)
		p.Persistence = 0.5 + rng.Float32()*0.5      // 0.5-1.0 (high - spores)
		p.MutationRate = 0.05 + rng.Float32()*0.15   // 0.05-0.2 (slow)
		p.HostSpecificity = 0.4 + rng.Float32()*0.5
		p.Transmission = TransmissionContact

	case PathogenPrion:
		p.Virulence = 0.95 + rng.Float32()*0.05       // 0.95-1.0 (always fatal)
		p.Transmissibility = 0.05 + rng.Float32()*0.1 // 0.05-0.15 (very rare)
		p.Latency = 0.7 + rng.Float32()*0.3           // 0.7-1.0 (very long)
		p.Persistence = 0.9 + rng.Float32()*0.1       // 0.9-1.0 (indestructible)
		p.MutationRate = 0                            // Prions don't mutate
		p.HostSpecificity = 0.8 + rng.Float32()*0.2
		p.Transmission = TransmissionFood

	case PathogenParasite:
		p.Virulence = 0.1 + rng.Float32()*0.3        // 0.1-0.4 (chronic, not acute)
		p.Transmissibility = 0.1 + rng.Float32()*0.4 // 0.1-0.5
		p.Latency = 0.2 + rng.Float32()*0.3          // 0.2-0.5
		p.Persistence = 0.3 + rng.Float32()*0.4      // 0.3-0.7
		p.MutationRate = 0.05 + rng.Float32()*0.15
		p.HostSpecificity = 0.5 + rng.Float32()*0.4
		p.Transmission = TransmissionVector
	}

	// Pick a random resistance gene (from disease resistance trait genes)
	p.ResistanceGene = 50 + rng.Intn(10) // Genes 50-59 affect disease resistance

	return p
}

// Clone creates a mutated descendant of this pathogen
func (p *Pathogen) Clone(rng *rand.Rand) *Pathogen {
	clone := &Pathogen{
		ID:               uuid.New(),
		Name:             p.Name,
		Type:             p.Type,
		Transmission:     p.Transmission,
		OriginSpeciesID:  p.OriginSpeciesID,
		OriginYear:       p.OriginYear,
		Virulence:        p.Virulence,
		Transmissibility: p.Transmissibility,
		Latency:          p.Latency,
		Persistence:      p.Persistence,
		MutationRate:     p.MutationRate,
		HostSpecificity:  p.HostSpecificity,
		ResistanceGene:   p.ResistanceGene,
		Generation:       p.Generation + 1,
		MutationsCount:   p.MutationsCount,
		SusceptibleDiets: make([]string, len(p.SusceptibleDiets)),
	}
	copy(clone.SusceptibleDiets, p.SusceptibleDiets)
	return clone
}

// Mutate applies random mutations to the pathogen
func (p *Pathogen) Mutate(rng *rand.Rand) {
	if p.Type == PathogenPrion {
		return // Prions don't mutate
	}

	mutationChance := float64(p.MutationRate)

	// Virulence-transmissibility tradeoff
	// Most pathogens evolve toward lower virulence over time (endemic evolution)
	if rng.Float64() < mutationChance {
		// 70% chance to decrease virulence, 30% to increase
		if rng.Float64() < 0.7 {
			p.Virulence = clamp32(p.Virulence-0.02-rng.Float32()*0.03, 0, 1)
		} else {
			p.Virulence = clamp32(p.Virulence+0.01+rng.Float32()*0.02, 0, 1)
		}
		p.MutationsCount++
	}

	// Transmissibility can increase or decrease
	if rng.Float64() < mutationChance {
		delta := (rng.Float32() - 0.5) * 0.05
		p.Transmissibility = clamp32(p.Transmissibility+delta, 0.01, 1)
		p.MutationsCount++
	}

	// Host specificity can broaden (zoonotic jumps)
	if rng.Float64() < mutationChance*0.1 { // Rare
		p.HostSpecificity = clamp32(p.HostSpecificity-0.05, 0, 1)
		p.MutationsCount++
	}
}

// CalculateR0 returns the basic reproduction number
// R0 > 1 means epidemic potential, R0 < 1 means dying out
func (p *Pathogen) CalculateR0(populationDensity, diseaseResistance float32) float32 {
	// R0 = transmissibility * contact_rate * duration / recovery
	baseR0 := p.Transmissibility * 3.0 // Scale to realistic R0 range

	// Density increases transmission
	densityFactor := 0.5 + populationDensity*0.5

	// Host resistance decreases R0
	resistanceFactor := 1.0 - diseaseResistance*0.7

	// Latency period increases infectious period
	durationFactor := 0.5 + p.Latency*0.5

	return baseR0 * densityFactor * resistanceFactor * durationFactor
}

// CalculateMortality returns death rate for infected hosts
func (p *Pathogen) CalculateMortality(hostDiseaseResistance float32) float32 {
	// Base mortality from virulence
	baseMortality := p.Virulence

	// Host resistance reduces mortality
	mortality := baseMortality * (1.0 - hostDiseaseResistance*0.8)

	// Minimum and maximum mortality
	if mortality < 0.01 {
		mortality = 0.01
	}
	if mortality > 0.95 {
		mortality = 0.95
	}

	return mortality
}

// CanInfectHost checks if this pathogen can infect a host based on specificity
func (p *Pathogen) CanInfectHost(hostSpeciesID uuid.UUID, hostDietType string, hostResistanceLevel float32) bool {
	// Origin species is always susceptible
	if hostSpeciesID == p.OriginSpeciesID {
		return true
	}

	// Check diet compatibility if specified
	if len(p.SusceptibleDiets) > 0 {
		dietMatch := false
		for _, diet := range p.SusceptibleDiets {
			if diet == hostDietType {
				dietMatch = true
				break
			}
		}
		if !dietMatch {
			return false
		}
	}

	// Host specificity determines cross-species infection probability
	// Low specificity (0) = broad host range, high specificity (1) = narrow
	crossSpeciesChance := 1.0 - p.HostSpecificity

	// Host resistance provides protection
	protectionChance := hostResistanceLevel * 0.5

	return crossSpeciesChance > protectionChance
}

// IsBecomingEndemic returns true if the pathogen is evolving toward endemic state
func (p *Pathogen) IsBecomingEndemic() bool {
	// Pathogens become endemic when:
	// 1. Virulence has dropped significantly
	// 2. Many mutations have occurred
	// 3. Long time since origin
	return p.Virulence < 0.3 && p.MutationsCount > 10
}

// OutbreakSeverity represents the intensity of an outbreak
type OutbreakSeverity string

const (
	SeverityMinor    OutbreakSeverity = "minor"    // < 1% population affected
	SeverityModerate OutbreakSeverity = "moderate" // 1-5% population
	SeveritySevere   OutbreakSeverity = "severe"   // 5-20% population
	SeverityPandemic OutbreakSeverity = "pandemic" // > 20% population
)

// Outbreak represents an active disease outbreak in a population
type Outbreak struct {
	ID              uuid.UUID        `json:"id"`
	PathogenID      uuid.UUID        `json:"pathogen_id"`
	SpeciesID       uuid.UUID        `json:"species_id"`
	BiomeID         uuid.UUID        `json:"biome_id"`
	StartYear       int64            `json:"start_year"`
	EndYear         int64            `json:"end_year,omitempty"` // 0 if ongoing
	Severity        OutbreakSeverity `json:"severity"`
	PeakInfected    int64            `json:"peak_infected"`
	TotalInfected   int64            `json:"total_infected"`
	TotalDeaths     int64            `json:"total_deaths"`
	CurrentInfected int64            `json:"current_infected"`
	RecoveredCount  int64            `json:"recovered_count"`
	IsActive        bool             `json:"is_active"`
}

// NewOutbreak creates a new disease outbreak
func NewOutbreak(pathogenID, speciesID, biomeID uuid.UUID, startYear int64, initialInfected int64) *Outbreak {
	severity := SeverityMinor
	return &Outbreak{
		ID:              uuid.New(),
		PathogenID:      pathogenID,
		SpeciesID:       speciesID,
		BiomeID:         biomeID,
		StartYear:       startYear,
		Severity:        severity,
		CurrentInfected: initialInfected,
		TotalInfected:   initialInfected,
		IsActive:        true,
	}
}

// Update advances the outbreak by one time step
func (o *Outbreak) Update(
	pathogen *Pathogen,
	susceptiblePopulation int64,
	diseaseResistance float32,
	rng *rand.Rand,
) {
	if !o.IsActive {
		return
	}

	// Calculate transmission and mortality
	r0 := pathogen.CalculateR0(float32(o.CurrentInfected)/float32(susceptiblePopulation+1), diseaseResistance)
	mortality := pathogen.CalculateMortality(diseaseResistance)

	// New infections
	newInfections := int64(float32(o.CurrentInfected) * r0 * (float32(susceptiblePopulation) / float32(susceptiblePopulation+o.RecoveredCount+1)))
	if newInfections > susceptiblePopulation/10 {
		newInfections = susceptiblePopulation / 10 // Cap at 10% per tick
	}
	if newInfections < 0 {
		newInfections = 0
	}

	// Deaths from current infected
	deaths := int64(float32(o.CurrentInfected) * mortality * 0.1) // 10% of mortality per tick
	if deaths > o.CurrentInfected {
		deaths = o.CurrentInfected
	}

	// Recoveries (those who survive)
	recoveries := int64(float32(o.CurrentInfected) * (1 - mortality) * 0.15) // 15% recover per tick
	if recoveries > o.CurrentInfected-deaths {
		recoveries = o.CurrentInfected - deaths
	}

	// Update counts
	o.CurrentInfected = o.CurrentInfected + newInfections - deaths - recoveries
	o.TotalInfected += newInfections
	o.TotalDeaths += deaths
	o.RecoveredCount += recoveries

	if o.CurrentInfected > o.PeakInfected {
		o.PeakInfected = o.CurrentInfected
	}

	// Update severity
	infectionRate := float64(o.TotalInfected) / float64(susceptiblePopulation+1)
	switch {
	case infectionRate > 0.20:
		o.Severity = SeverityPandemic
	case infectionRate > 0.05:
		o.Severity = SeveritySevere
	case infectionRate > 0.01:
		o.Severity = SeverityModerate
	default:
		o.Severity = SeverityMinor
	}

	// Check if outbreak has ended
	if o.CurrentInfected <= 0 {
		o.IsActive = false
		o.CurrentInfected = 0
	}
}

func clamp32(val, min, max float32) float32 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
