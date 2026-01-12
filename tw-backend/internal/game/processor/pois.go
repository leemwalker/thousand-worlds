package processor

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"
	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
)

// handleGetPOIs handles the request for points of interest
func (p *GameProcessor) handleGetPOIs(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	// Instead of relying on world geology POIs, we now generate them dynamically from system stats
	// This matches the requirement: "highest cpu/memory utilization... generated off of the related data"

	sysPOIs, err := p.generateSystemPOIs()
	if err != nil {
		log.Printf("Failed to generate system POIs: %v", err)
		client.SendGameMessage("error", "Failed to generate system data", nil)
		return nil
	}

	// Name them using Ollama if available
	// doing this concurrently
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // Concurrency limit

	for i := range sysPOIs {
		// Only name if not already named (though generated ones usually are raw)
		// We want to add flavor: "Scottish or any other ethnicity"
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Generate ethnic flavor name
			name := p.generateEthnicPOIName(sysPOIs[idx])
			if name != "" {
				sysPOIs[idx].Name = name
			}
		}(i)
	}

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		log.Printf("System POI naming timed out, returning partial results")
	}

	client.SendGameMessage("points_of_interest", "", map[string]interface{}{
		"pois": sysPOIs,
	})
	return nil
}

// generateSystemPOIs fetches top processes and converts them to POIs
// generateSystemPOIs fetches top processes and converts them to POIs
func (p *GameProcessor) generateSystemPOIs() ([]geography.PointOfInterest, error) {
	var pois []geography.PointOfInterest

	// 1. Get Top CPU Processes
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	// Sort by CPU usage (requires gathering stats)
	type procStat struct {
		p    *process.Process
		cpu  float64
		mem  float32
		name string
	}
	var stats []procStat

	for _, proc := range procs {
		// Limit number of checks to avoid lag
		c, _ := proc.CPUPercent()
		m, _ := proc.MemoryPercent()
		n, _ := proc.Name()
		if c > 0.1 || m > 0.1 {
			stats = append(stats, procStat{p: proc, cpu: c, mem: m, name: n})
		}
	}

	// Sort by CPU
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].cpu > stats[j].cpu
	})

	// Top 10 CPU -> "Volcanoes" or "Peaks"
	for i := 0; i < 10 && i < len(stats); i++ {
		s := stats[i]
		// Deterministic UUID based on PID and Name
		ns := uuid.Must(uuid.Parse("00000000-0000-0000-0000-000000000000"))
		id := uuid.NewSHA1(ns, []byte(fmt.Sprintf("cpu_%d_%s", s.p.Pid, s.name)))

		pois = append(pois, geography.PointOfInterest{
			ID:        id,
			Name:      s.name,
			Type:      "volcano",
			Elevation: s.cpu * 100,
			Coordinates: geography.Coordinates{
				Lat: (rand.Float64() * 140) - 70,
				Lon: (rand.Float64() * 360) - 180,
			},
		})
	}

	// Sort by Mem
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].mem > stats[j].mem
	})

	// Top 10 Mem -> "Deep Oceans" or "Lakes"
	for i := 0; i < 10 && i < len(stats); i++ {
		s := stats[i]
		ns := uuid.Must(uuid.Parse("00000000-0000-0000-0000-000000000000"))
		id := uuid.NewSHA1(ns, []byte(fmt.Sprintf("mem_%d_%s", s.p.Pid, s.name)))

		// Avoid duplicates by ID
		exists := false
		for _, ex := range pois {
			if ex.ID == id {
				exists = true
				break
			}
		}
		if exists {
			continue
		}

		pois = append(pois, geography.PointOfInterest{
			ID:        id,
			Name:      s.name,
			Type:      "deep_ocean",
			Elevation: -float64(s.mem) * 100,
			Coordinates: geography.Coordinates{
				Lat: (rand.Float64() * 140) - 70,
				Lon: (rand.Float64() * 360) - 180,
			},
		})
	}

	// System Stats as "Capitals"
	vm, _ := mem.VirtualMemory()
	cpuStats, _ := cpu.Percent(0, false)

	sysRamID := uuid.NewSHA1(uuid.Must(uuid.Parse("00000000-0000-0000-0000-000000000000")), []byte("sys_ram"))
	pois = append(pois, geography.PointOfInterest{
		ID:          sysRamID,
		Name:        "RAM Capital",
		Type:        "mountain_peak",
		Elevation:   vm.UsedPercent * 10,
		Coordinates: geography.Coordinates{Lat: 20, Lon: 20},
	})

	if len(cpuStats) > 0 {
		sysCpuID := uuid.NewSHA1(uuid.Must(uuid.Parse("00000000-0000-0000-0000-000000000000")), []byte("sys_cpu"))
		pois = append(pois, geography.PointOfInterest{
			ID:          sysCpuID,
			Name:        "CPU Core",
			Type:        "mountain_peak",
			Elevation:   cpuStats[0] * 10,
			Coordinates: geography.Coordinates{Lat: -20, Lon: -20},
		})
	}

	return pois, nil
}

func (p *GameProcessor) generateEthnicPOIName(poi geography.PointOfInterest) string {
	if p.ollamaClient == nil {
		return poi.Name // detailed system name
	}

	// Random ethnicity flavor
	ethnicities := []string{"Scottish", "Japanese", "Norse", "Swahili", "Incan", "Slavic", "Celtic", "Polynesian"}
	flavor := ethnicities[time.Now().UnixNano()%int64(len(ethnicities))]

	prompt := fmt.Sprintf("Rename this system process '%s' (Type: %s, Usage: %.0f) as a %s fantasy location. Keep it short. Example: 'Mount Chrome' or 'River of Docker'. Just the name.",
		poi.Name, poi.Type, poi.Elevation, flavor)

	resp, err := p.ollamaClient.Generate(prompt)
	if err != nil {
		log.Printf("Ollama naming failed: %v", err)
		return poi.Name
	}
	return resp
}
