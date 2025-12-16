// Package pathogen provides disease simulation and management for the ecosystem.
package pathogen

import (
	"math/rand"

	"github.com/google/uuid"
)

// DiseaseSystem manages all pathogens and outbreaks in a world
type DiseaseSystem struct {
	WorldID       uuid.UUID               `json:"world_id"`
	Pathogens     map[uuid.UUID]*Pathogen `json:"pathogens"`
	Outbreaks     map[uuid.UUID]*Outbreak `json:"outbreaks"`
	PastOutbreaks []*Outbreak             `json:"past_outbreaks"` // History for lore
	CurrentYear   int64                   `json:"current_year"`
	rng           *rand.Rand

	// Configuration
	OutbreakBaseChance float64 `json:"outbreak_base_chance"` // Per species per year
	ZoonoticChance     float64 `json:"zoonotic_chance"`      // Cross-species jump chance
	MaxActiveOutbreaks int     `json:"max_active_outbreaks"` // Limit concurrent outbreaks
}

// NewDiseaseSystem creates a new disease management system
func NewDiseaseSystem(worldID uuid.UUID, seed int64) *DiseaseSystem {
	return &DiseaseSystem{
		WorldID:            worldID,
		Pathogens:          make(map[uuid.UUID]*Pathogen),
		Outbreaks:          make(map[uuid.UUID]*Outbreak),
		PastOutbreaks:      make([]*Outbreak, 0),
		rng:                rand.New(rand.NewSource(seed)),
		OutbreakBaseChance: 0.0001, // 0.01% chance per species per year
		ZoonoticChance:     0.001,  // 0.1% chance for cross-species transmission
		MaxActiveOutbreaks: 10,
	}
}

// AddPathogen adds a pathogen to the system
func (ds *DiseaseSystem) AddPathogen(p *Pathogen) {
	ds.Pathogens[p.ID] = p
}

// GetPathogen retrieves a pathogen by ID
func (ds *DiseaseSystem) GetPathogen(id uuid.UUID) *Pathogen {
	return ds.Pathogens[id]
}

// GetActiveOutbreaks returns all currently active outbreaks
func (ds *DiseaseSystem) GetActiveOutbreaks() []*Outbreak {
	active := make([]*Outbreak, 0)
	for _, o := range ds.Outbreaks {
		if o.IsActive {
			active = append(active, o)
		}
	}
	return active
}

// CreateNovelPathogen creates a new pathogen that emerges from a species
func (ds *DiseaseSystem) CreateNovelPathogen(
	speciesID uuid.UUID,
	speciesName string,
	pType PathogenType,
) *Pathogen {
	name := generatePathogenName(speciesName, pType, ds.rng)
	p := NewPathogen(name, pType, speciesID, ds.CurrentYear, ds.rng)
	ds.AddPathogen(p)
	return p
}

// CheckSpontaneousOutbreak checks if a new outbreak should spontaneously occur
// Returns the new pathogen and outbreak if one starts
func (ds *DiseaseSystem) CheckSpontaneousOutbreak(
	speciesID uuid.UUID,
	speciesName string,
	population int64,
	densityFactor float64, // 0-1, how dense the population is
) (*Pathogen, *Outbreak) {
	if len(ds.GetActiveOutbreaks()) >= ds.MaxActiveOutbreaks {
		return nil, nil
	}

	// Higher density = higher chance
	chance := ds.OutbreakBaseChance * (1 + densityFactor*2)

	// Higher population = slightly higher chance
	if population > 10000 {
		chance *= 1.2
	}
	if population > 100000 {
		chance *= 1.3
	}

	if ds.rng.Float64() > chance {
		return nil, nil
	}

	// Determine pathogen type (weighted probabilities)
	var pType PathogenType
	roll := ds.rng.Float64()
	switch {
	case roll < 0.4:
		pType = PathogenVirus // 40% - most common
	case roll < 0.7:
		pType = PathogenBacteria // 30%
	case roll < 0.85:
		pType = PathogenFungus // 15%
	case roll < 0.98:
		pType = PathogenParasite // 13%
	default:
		pType = PathogenPrion // 2% - rare
	}

	// Create pathogen
	pathogen := ds.CreateNovelPathogen(speciesID, speciesName, pType)

	// Start outbreak
	initialInfected := int64(1 + ds.rng.Intn(10))
	outbreak := NewOutbreak(pathogen.ID, speciesID, uuid.Nil, ds.CurrentYear, initialInfected)
	ds.Outbreaks[outbreak.ID] = outbreak

	pathogen.ActiveOutbreaks++

	return pathogen, outbreak
}

// CheckZoonoticTransfer checks if a pathogen jumps to a new host species
func (ds *DiseaseSystem) CheckZoonoticTransfer(
	pathogen *Pathogen,
	sourceSpeciesID uuid.UUID,
	targetSpeciesID uuid.UUID,
	targetDietType string,
	targetPopulation int64,
	targetResistance float32,
	contactRate float64, // 0-1, how much the species interact
) *Outbreak {
	if len(ds.GetActiveOutbreaks()) >= ds.MaxActiveOutbreaks {
		return nil
	}

	// Base chance modified by contact rate and host specificity
	chance := ds.ZoonoticChance * contactRate * float64(1-pathogen.HostSpecificity)

	if ds.rng.Float64() > chance {
		return nil
	}

	// Check if pathogen can actually infect the target
	if !pathogen.CanInfectHost(targetSpeciesID, targetDietType, targetResistance) {
		return nil
	}

	// Zoonotic jump occurs - pathogen may mutate
	if ds.rng.Float64() < float64(pathogen.MutationRate) {
		pathogen.Mutate(ds.rng)
	}

	// Add the new diet to susceptible diets
	pathogen.SusceptibleDiets = append(pathogen.SusceptibleDiets, targetDietType)

	// Start outbreak in new species
	initialInfected := int64(1)
	outbreak := NewOutbreak(pathogen.ID, targetSpeciesID, uuid.Nil, ds.CurrentYear, initialInfected)
	ds.Outbreaks[outbreak.ID] = outbreak

	pathogen.ActiveOutbreaks++

	return outbreak
}

// Update advances all outbreaks and pathogens by one year
func (ds *DiseaseSystem) Update(
	year int64,
	speciesData map[uuid.UUID]SpeciesInfo, // Species ID -> population info
) {
	ds.CurrentYear = year

	// Update each active outbreak
	for id, outbreak := range ds.Outbreaks {
		if !outbreak.IsActive {
			continue
		}

		pathogen := ds.Pathogens[outbreak.PathogenID]
		if pathogen == nil {
			outbreak.IsActive = false
			continue
		}

		// Get species info
		species, exists := speciesData[outbreak.SpeciesID]
		if !exists || species.Population <= 0 {
			outbreak.IsActive = false
			outbreak.EndYear = year
			continue
		}

		// Update the outbreak
		outbreak.Update(pathogen, species.Population, species.DiseaseResistance, ds.rng)

		// Check if outbreak ended
		if !outbreak.IsActive {
			outbreak.EndYear = year
			ds.PastOutbreaks = append(ds.PastOutbreaks, outbreak)
			delete(ds.Outbreaks, id)
			pathogen.ActiveOutbreaks--

			// Update pathogen totals
			pathogen.TotalInfected += outbreak.TotalInfected
			pathogen.TotalDeaths += outbreak.TotalDeaths

			// Check for endemic evolution
			if pathogen.IsBecomingEndemic() && !pathogen.IsEndemic {
				pathogen.IsEndemic = true
			}
		}
	}

	// Mutate pathogens periodically
	for _, pathogen := range ds.Pathogens {
		if !pathogen.IsEradicated && ds.rng.Float64() < float64(pathogen.MutationRate)*0.1 {
			pathogen.Mutate(ds.rng)
		}
	}
}

// GetImpact returns the total population impact of active outbreaks for a species
func (ds *DiseaseSystem) GetImpact(speciesID uuid.UUID) (currentInfected, currentDeaths int64) {
	for _, outbreak := range ds.Outbreaks {
		if outbreak.SpeciesID == speciesID && outbreak.IsActive {
			currentInfected += outbreak.CurrentInfected
			currentDeaths += outbreak.TotalDeaths
		}
	}
	return
}

// GetHistoricalOutbreaks returns past outbreaks affecting a species
func (ds *DiseaseSystem) GetHistoricalOutbreaks(speciesID uuid.UUID) []*Outbreak {
	outbreaks := make([]*Outbreak, 0)
	for _, o := range ds.PastOutbreaks {
		if o.SpeciesID == speciesID {
			outbreaks = append(outbreaks, o)
		}
	}
	return outbreaks
}

// GetPandemics returns all outbreaks that reached pandemic severity
func (ds *DiseaseSystem) GetPandemics() []*Outbreak {
	pandemics := make([]*Outbreak, 0)
	for _, o := range ds.PastOutbreaks {
		if o.Severity == SeverityPandemic {
			pandemics = append(pandemics, o)
		}
	}
	for _, o := range ds.Outbreaks {
		if o.Severity == SeverityPandemic {
			pandemics = append(pandemics, o)
		}
	}
	return pandemics
}

// EradicatePathogen marks a pathogen as eradicated (no more outbreaks possible)
func (ds *DiseaseSystem) EradicatePathogen(pathogenID uuid.UUID) {
	pathogen := ds.Pathogens[pathogenID]
	if pathogen != nil {
		pathogen.IsEradicated = true
		pathogen.ActiveOutbreaks = 0

		// End any active outbreaks
		for _, o := range ds.Outbreaks {
			if o.PathogenID == pathogenID {
				o.IsActive = false
				o.EndYear = ds.CurrentYear
			}
		}
	}
}

// SpeciesInfo provides population data for the simulation
type SpeciesInfo struct {
	Population        int64
	DiseaseResistance float32
	DietType          string
	Density           float64 // 0-1
}

// generatePathogenName creates a name for a new pathogen
func generatePathogenName(speciesName string, pType PathogenType, rng *rand.Rand) string {
	prefixes := map[PathogenType][]string{
		PathogenVirus:    {"Grippe", "Pox", "Fever", "Flu", "Wasting"},
		PathogenBacteria: {"Black", "Spotted", "Bloody", "Sweating", "Rotting"},
		PathogenFungus:   {"Black", "White", "Creeping", "Grey", "Withering"},
		PathogenPrion:    {"Shaking", "Wasting", "Mad", "Trembling", "Hollow"},
		PathogenParasite: {"Blood", "Gut", "Skin", "Eye", "Brain"},
	}

	suffixes := map[PathogenType][]string{
		PathogenVirus:    {"Virus", "Plague", "Fever", "Sickness", "Blight"},
		PathogenBacteria: {"Plague", "Rot", "Disease", "Sickness", "Murrain"},
		PathogenFungus:   {"Mold", "Blight", "Rot", "Fungus", "Corruption"},
		PathogenPrion:    {"Disease", "Sickness", "Wasting", "Madness", "Curse"},
		PathogenParasite: {"Worm", "Parasite", "Infestation", "Plague", "Curse"},
	}

	prefix := prefixes[pType][rng.Intn(len(prefixes[pType]))]
	suffix := suffixes[pType][rng.Intn(len(suffixes[pType]))]

	return prefix + " " + suffix
}
