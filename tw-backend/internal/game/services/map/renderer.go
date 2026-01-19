package gamemap

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"runtime"
	"sync"
	"time"

	"tw-backend/internal/ecosystem"
	"tw-backend/internal/spatial"

	"github.com/chai2010/webp"
)

var (
	// ErrRenderTimeout is returned when rendering takes too long
	ErrRenderTimeout = errors.New("map render timed out")
	// ErrWorldNotInitialized is returned when geology is missing
	ErrWorldNotInitialized = errors.New("world geology not initialized")
)

// Renderer handles the generation of map images
type Renderer struct {
	config RenderConfig
	pool   *RendererPool
	cache  *MapCache

	// Metrics are handled globally in metrics.go for now, but could be injected
}

// NewRenderer creates a new map renderer
func NewRenderer(config RenderConfig, pool *RendererPool, cache *MapCache) *Renderer {
	return &Renderer{
		config: config,
		pool:   pool,
		cache:  cache,
	}
}

// RenderWorldMap generates a high-resolution map image for the given world.
// It enforces concurrency limits and timeouts.
func (r *Renderer) RenderWorldMap(ctx context.Context, worldID string, geo *ecosystem.WorldGeology, width, height int) ([]byte, error) {
	// 1. Validation & Defaulting
	if width <= 0 {
		width = r.config.DefaultWidth
	}
	if height <= 0 {
		height = r.config.DefaultHeight
	}
	if width > r.config.MaxWidth {
		width = r.config.MaxWidth
	}
	if height > r.config.MaxHeight {
		height = r.config.MaxHeight
	}

	// 2. Cache Check
	// Use simplified version 1 for now (Sprint 1)
	cacheKey := GetCacheKey(worldID, width, height, int64(geo.TotalYearsSimulated))
	if cached, ok := r.cache.Get(cacheKey); ok {
		metricRenderDuration.WithLabelValues("cache_hit").Observe(0)
		return cached, nil
	}

	// 3. Concurrency Check (Immediate Reject)
	if !r.pool.AcquireConcurrencySlot() {
		metricRenderRejected.Inc()
		return nil, ErrConcurrencyLimitExceeded
	}
	defer r.pool.ReleaseConcurrencySlot()

	// 4. Render Execution with Timeout
	// Create a child context with the configured timeout
	ctx, cancel := context.WithTimeout(ctx, r.config.RenderTimeout)
	defer cancel()

	start := time.Now()

	// We run the CPU-intensive render in a goroutine to allow context cancellation checks
	type result struct {
		data []byte
		err  error
	}
	resChan := make(chan result, 1)

	go func() {
		data, err := r.renderInternal(ctx, geo, width, height)
		resChan <- result{data, err}
	}()

	select {
	case res := <-resChan:
		duration := time.Since(start).Seconds()
		if res.err != nil {
			metricRenderDuration.WithLabelValues("failure").Observe(duration)
			return nil, res.err
		}

		// Success
		metricRenderDuration.WithLabelValues("success").Observe(duration)
		metricImageSize.Observe(float64(len(res.data)))

		// Update Cache
		r.cache.Set(cacheKey, res.data)

		return res.data, nil

	case <-ctx.Done():
		// Timeout
		metricRenderDuration.WithLabelValues("timeout").Observe(r.config.RenderTimeout.Seconds())
		return nil, ErrRenderTimeout
	}
}

// RenderHeightmapPNG generates a 16-bit grayscale PNG of the world's elevation data.
// This format is GPU-friendly for displacement mapping in 3D renderers.
func (r *Renderer) RenderHeightmapPNG(ctx context.Context, worldID string, geo *ecosystem.WorldGeology, width, height int) ([]byte, error) {
	// Validation
	if width <= 0 {
		width = r.config.DefaultWidth
	}
	if height <= 0 {
		height = r.config.DefaultHeight
	}
	if width > r.config.MaxWidth {
		width = r.config.MaxWidth
	}
	if height > r.config.MaxHeight {
		height = r.config.MaxHeight
	}

	// Require either SphereHeightmap or flat Heightmap
	var minElev, maxElev float64
	var getHeight func(x, y int) float64

	if geo.SphereHeightmap != nil {
		// Use spherical heightmap (lat/lon projection)
		topo := geo.SphereHeightmap.Topology()
		minElev, maxElev = geo.SphereHeightmap.MinMax()
		getHeight = func(x, y int) float64 {
			tY := float64(y) / float64(height)
			lat := (0.5 - tY) * math.Pi
			tX := float64(x) / float64(width)
			lon := (tX - 0.5) * 2 * math.Pi

			// Convert lat/lon to unit vector
			cosLat := math.Cos(lat)
			sinLat := math.Sin(lat)
			cosLon := math.Cos(lon)
			sinLon := math.Sin(lon)

			vx := cosLat * cosLon
			vy := sinLat
			vz := cosLat * sinLon

			coord := topo.FromVector(vx, vy, vz)
			return geo.SphereHeightmap.Get(coord)
		}
	} else if geo.Heightmap != nil {
		// Fallback to flat heightmap
		hm := geo.Heightmap
		minElev, maxElev = float64(hm.Elevations[0]), float64(hm.Elevations[0])
		for _, e := range hm.Elevations {
			val := float64(e)
			if val < minElev {
				minElev = val
			}
			if val > maxElev {
				maxElev = val
			}
		}
		getHeight = func(x, y int) float64 {
			// Map render coordinates to heightmap coordinates
			hmX := x * hm.Width / width
			hmY := y * hm.Height / height
			if hmX >= hm.Width {
				hmX = hm.Width - 1
			}
			if hmY >= hm.Height {
				hmY = hm.Height - 1
			}
			return float64(hm.Elevations[hmY*hm.Width+hmX])
		}
	} else {
		return nil, ErrWorldNotInitialized
	}

	elevRange := maxElev - minElev
	if elevRange <= 0 {
		elevRange = 1.0
	}

	// Create 16-bit grayscale image
	img := image.NewGray16(image.Rect(0, 0, width, height))

	// Fill image with normalized elevation values
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			elev := getHeight(x, y)
			// Normalize to 0-65535 range
			normalized := (elev - minElev) / elevRange
			val := uint16(normalized * 65535)
			img.SetGray16(x, y, color.Gray16{Y: val})
		}
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// RenderMaterialPNG generates a PNG with terrain material data in RGB channels.
// This enables data-driven terrain coloring in the 3D globe renderer.
// R: RockHardness (0-255, from Province data: 0=soft sediment, 255=hard granite)
// G: IsContinental (0=oceanic/basalt, 255=continental/granite)
// B: Sediment depth (normalized 0-255, 255=max sediment)
func (r *Renderer) RenderMaterialPNG(ctx context.Context, worldID string, geo *ecosystem.WorldGeology, width, height int) ([]byte, error) {
	// Validation
	if width <= 0 {
		width = r.config.DefaultWidth
	}
	if height <= 0 {
		height = r.config.DefaultHeight
	}
	if width > r.config.MaxWidth {
		width = r.config.MaxWidth
	}
	if height > r.config.MaxHeight {
		height = r.config.MaxHeight
	}

	if geo.SphereHeightmap == nil {
		return nil, ErrWorldNotInitialized
	}

	topo := geo.SphereHeightmap.Topology()

	// Create RGBA image for material data
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Max sediment for normalization (200m is typical max from GetSedimentMap)
	const maxSediment = 200.0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Convert pixel to lat/lon to sphere coord
			tY := float64(y) / float64(height)
			lat := (0.5 - tY) * math.Pi
			tX := float64(x) / float64(width)
			lon := (tX - 0.5) * 2 * math.Pi

			cosLat := math.Cos(lat)
			sinLat := math.Sin(lat)
			cosLon := math.Cos(lon)
			sinLon := math.Sin(lon)

			vx := cosLat * cosLon
			vy := sinLat
			vz := cosLat * sinLon

			coord := topo.FromVector(vx, vy, vz)
			cellData := geo.SphereHeightmap.GetCellData(coord)

			// R: Rock hardness (0-1 scaled to 0-255)
			hardness := uint8(cellData.RockHardness * 255)

			// G: IsContinental
			var isContinental uint8
			if cellData.IsContinental {
				isContinental = 255
			}

			// B: Sediment depth (normalized)
			sediment := cellData.Sediment / maxSediment
			if sediment > 1.0 {
				sediment = 1.0
			}
			sedimentVal := uint8(sediment * 255)

			img.SetRGBA(x, y, color.RGBA{R: hardness, G: isContinental, B: sedimentVal, A: 255})
		}
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// RenderIcePNG generates a grayscale PNG of ice sheet coverage.
// 0 = no ice, 255 = maximum glacier thickness
func (r *Renderer) RenderIcePNG(ctx context.Context, worldID string, geo *ecosystem.WorldGeology, width, height int) ([]byte, error) {
	// Validation
	if width <= 0 {
		width = r.config.DefaultWidth
	}
	if height <= 0 {
		height = r.config.DefaultHeight
	}
	if width > r.config.MaxWidth {
		width = r.config.MaxWidth
	}
	if height > r.config.MaxHeight {
		height = r.config.MaxHeight
	}

	if geo.SphereHeightmap == nil {
		return nil, ErrWorldNotInitialized
	}

	topo := geo.SphereHeightmap.Topology()

	// Create grayscale image
	img := image.NewGray(image.Rect(0, 0, width, height))

	// Max ice thickness for normalization (3000m is ~Antarctic max)
	const maxIce = 3000.0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Convert pixel to sphere coord
			tY := float64(y) / float64(height)
			lat := (0.5 - tY) * math.Pi
			tX := float64(x) / float64(width)
			lon := (tX - 0.5) * 2 * math.Pi

			cosLat := math.Cos(lat)
			sinLat := math.Sin(lat)
			cosLon := math.Cos(lon)
			sinLon := math.Sin(lon)

			vx := cosLat * cosLon
			vy := sinLat
			vz := cosLat * sinLon

			coord := topo.FromVector(vx, vy, vz)

			var iceThickness float64
			if geo.IceSheet != nil && len(geo.IceSheet.Ice) > 0 {
				// Use IceSheet slice indexing
				res := geo.IceSheet.Resolution
				idx := coord.Face*res*res + coord.Y*res + coord.X
				if idx >= 0 && idx < len(geo.IceSheet.Ice) {
					iceThickness = geo.IceSheet.Ice[idx].Thickness
				}
			}

			// Normalize ice thickness
			normalized := iceThickness / maxIce
			if normalized > 1.0 {
				normalized = 1.0
			}
			val := uint8(normalized * 255)

			img.SetGray(x, y, color.Gray{Y: val})
		}
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// renderInternal performs the actual pixel generation and encoding
func (r *Renderer) renderInternal(ctx context.Context, geo *ecosystem.WorldGeology, width, height int) ([]byte, error) {
	// Use SphereHeightmap if available (Phase 2+ / "Satellite" view)
	// Fallback to Flat Heightmap only if sphere is missing (shouldn't happen for valid worlds)
	if geo.SphereHeightmap == nil {
		return nil, errors.New("sphere heightmap required for satellite view")
	}

	// Acquire buffer for raw pixels (RGBA)
	// 4 bytes per pixel
	rawSize := width * height * 4
	rawBuf := r.pool.GetBuffer(rawSize)
	// Ensure it has correct length (GetBuffer might have cap >= size but len 0)
	rawBuf = rawBuf[:rawSize]
	defer r.pool.ReturnBuffer(rawBuf)

	// Create image.RGBA wrapper around the buffer
	img := &image.RGBA{
		Pix:    rawBuf,
		Stride: 4 * width,
		Rect:   image.Rect(0, 0, width, height),
	}

	// Parallel Rendering
	// Split height into chunks for parallel processing
	numWorkers := runtime.NumCPU()
	rowsPerWorker := height / numWorkers
	if rowsPerWorker == 0 {
		rowsPerWorker = 1
	}

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	topo := geo.SphereHeightmap.Topology()

	// Precompute min/max for normalization
	minElev, maxElev := geo.SphereHeightmap.MinMax()
	elevRange := maxElev - minElev
	if elevRange <= 0 {
		elevRange = 1.0
	}
	log.Printf("[RENDER] Min: %.2f, Max: %.2f, Range: %.2f", minElev, maxElev, elevRange)

	// Era-based rendering: Determine visual style based on planetary age
	// - hadean_molten (< 10M years): Magma ocean, black crust with orange fissures
	// - hadean_cooling (10M - 100M years): Cooling basalt, volcanic activity
	// - eoarchean (100M - 500M years): Grey rock, early oceans forming
	// - archean+ (> 500M years): Full color system with biomes
	type eraType int
	const (
		eraHadeanMolten eraType = iota
		eraHadeanCooling
		eraEoarchean
		eraArchean
	)
	var era eraType
	yearsSimulated := geo.TotalYearsSimulated
	switch {
	case yearsSimulated < 10_000_000:
		era = eraHadeanMolten
		log.Printf("[RENDER] Era: Hadean Molten (< 10M years)")
	case yearsSimulated < 100_000_000:
		era = eraHadeanCooling
		log.Printf("[RENDER] Era: Hadean Cooling (10M-100M years)")
	case yearsSimulated < 500_000_000:
		era = eraEoarchean
		log.Printf("[RENDER] Era: Eoarchean (100M-500M years)")
	default:
		era = eraArchean
		log.Printf("[RENDER] Era: Archean+ (> 500M years)")
	}

	// chunkErr := make(chan error, numWorkers) // Unused for now

	for w := 0; w < numWorkers; w++ {
		startRow := w * rowsPerWorker
		endRow := startRow + rowsPerWorker
		if w == numWorkers-1 {
			endRow = height
		}

		go func(yStart, yEnd int) {
			defer wg.Done()

			// Check context occasionally
			if ctx.Err() != nil {
				return
			}

			// reused vector variables to avoid allocation?
			// No, simple scalar math is fine.

			for y := yStart; y < yEnd; y++ {
				// Normalized Y (0.0 at top, 1.0 at bottom)
				// Latitude: +PI/2 (North) to -PI/2 (South)
				// T = y / height
				// Lat = (0.5 - T) * PI
				tY := float64(y) / float64(height)
				lat := (0.5 - tY) * math.Pi

				// Adaptive Sampling: Supersample near poles (DISABLED)
				// Distortion increases as |lat| approaches PI/2
				// Threshold: 60 degrees (approx PI/3)
				// supersample := math.Abs(lat) > (math.Pi / 3.0)
				_ = lat // Prevent lint error for unused variable
				supersample := false
				samples := 1
				if supersample {
					samples = 4 // 2x2 grid
				}

				for x := 0; x < width; x++ {
					var rSum, gSum, bSum, count float64

					// Sample offsets
					// If samples=1: offset (0.5, 0.5) center of pixel
					// If samples=4: offsets (0.25, 0.25), (0.75, 0.25), (0.25, 0.75), (0.75, 0.75)

					step := 1.0 / math.Sqrt(float64(samples))

					for sy := 0; sy < int(math.Sqrt(float64(samples))); sy++ {
						for sx := 0; sx < int(math.Sqrt(float64(samples))); sx++ {
							subX := float64(x) + (float64(sx)+0.5)*step
							subY := float64(y) + (float64(sy)+0.5)*step

							// Recalculate lat/lon for sub-pixel
							sLat := (0.5 - (subY / float64(height))) * math.Pi
							sLon := ((subX / float64(width)) - 0.5) * 2 * math.Pi

							// Lat/Lon to Unit Vector
							cosLat := math.Cos(sLat)
							sinLat := math.Sin(sLat)
							cosLon := math.Cos(sLon)
							sinLon := math.Sin(sLon)

							vx := cosLat * cosLon
							vy := sinLat // Y is up in this engine? Checked topology: ToVector uses same math
							vz := cosLat * sinLon

							// Sample from Sphere
							// Vector Sampling: Map vector to face coordinate
							// Uses Nearest Neighbor for Sprint 1
							coord := topo.FromVector(vx, vy, vz)
							elev := geo.SphereHeightmap.Get(coord)

							// Color Mapping (Photorealistic - Lifeless Protoplanet)
							// Phase 3-4: Enhanced with coastal features
							var r, g, b uint8

							// Get cell data for sediment info
							cellData := geo.SphereHeightmap.GetCellData(coord)
							sediment := cellData.Sediment
							flux := cellData.Flux

							if elev < geo.SeaLevel {
								// Special Era Handling for Oceans (Molten/Empty)
								if era == eraHadeanMolten {
									// Magma Ocean
									// Glowing orange/red based on depth (hotter deep down)
									depth := geo.SeaLevel - elev
									maxDepth := math.Max(geo.SeaLevel-minElev, 1.0)
									heatFactor := depth / maxDepth // 0=surface, 1=deep

									// Deep magma: Bright yellow/white
									// Surface magma: Red/Orange cooling crust
									r = uint8(255)
									g = uint8(50 + heatFactor*200)
									b = uint8(10 + heatFactor*50)
								} else if era == eraHadeanCooling {
									// Cooling Basaltic Basin (Empty or shallow acidic water)
									// Dark grey/black basalt
									r, g, b = 25, 25, 30

									// Cooling cracks (gold/orange) based on noise/gradient
									// Use flux/erosion patterns to simulate cracks
									if flux > 50 {
										// Active fissure
										r, g, b = 200, 80, 20
									}
								} else if era == eraEoarchean {
									// Early Oceans - Dark, iron-rich (greenish/black)
									depth := geo.SeaLevel - elev
									maxDepth := math.Max(geo.SeaLevel-minElev, 1.0)
									depthFactor := depth / maxDepth

									f := 1.0 - depthFactor
									r = uint8(10 + f*20)
									g = uint8(20 + f*30)
									b = uint8(20 + f*40)
								} else {
									// Standard Archean+ Ocean Coloring
									// Water coloring - varies by depth and sediment
									depth := geo.SeaLevel - elev
									maxDepth := math.Max(geo.SeaLevel-minElev, 1.0)
									depthFactor := depth / maxDepth
									if depthFactor > 1.0 {
										depthFactor = 1.0
									}
									// ... standard ocean logic ...
									// Re-implement standard logic to avoid deep nesting or use `goto` (Go allows it but discouraged)
									// Better: Use flag or `else` block
									// Expanding standard logic here for clarity:

									// Phase 4: Estuary coloring (brackish water)
									if cellData.IsEstuary {
										// Brownish-green brackish water
										r, g, b = 55, 50, 40
									} else if cellData.IsSpit && depthFactor < 0.1 {
										// Shallow water over spit - very light blue
										r, g, b = 40, 55, 65
									} else if flux > 100 && depthFactor < 0.1 {
										// Muddy brown water at river deltas
										r, g, b = 60, 45, 30
									} else if depthFactor < 0.05 {
										// Very shallow water - may show through to sediment
										if sediment > 0.5 {
											// Tidal flat / sandy shallow (tan-blue mix)
											r, g, b = 50, 55, 60
										} else {
											// Clear shallow
											r, g, b = 25, 50, 70
										}
									} else {
										// Deep ocean gradient - brightened to match lakes
										f := 1.0 - depthFactor   // 0 = Deep, 1 = Shallow
										r = uint8(15.0 + f*40.0) // 15-55
										g = uint8(35.0 + f*60.0) // 35-95
										b = uint8(65.0 + f*70.0) // 65-135
									}
								}

							} else if cellData.IsLake {
								// Special Era Handling for Lakes
								if era == eraHadeanMolten {
									// Lava Lake
									r, g, b = 255, 80, 20
								} else if era == eraHadeanCooling {
									// Solidifying Lava Lake (Dark crust)
									r, g, b = 40, 35, 30
								} else {
									// Standard Lake
									// Phase 6: Lake Rendering
									// Render as water based on depth
									depth := cellData.LakeDepth
									maxDepth := 100.0 // Assumed max depth for visual gradient
									depthFactor := depth / maxDepth
									if depthFactor > 1.0 {
										depthFactor = 1.0
									}

									// Freshwater Blue (Lighter than ocean)
									// Shallow: 100, 150, 200
									// Deep: 20, 40, 80
									f := 1.0 - depthFactor
									r = uint8(20.0 + f*80.0)
									g = uint8(40.0 + f*110.0)
									b = uint8(80.0 + f*120.0)
								}
							} else {
								// Land coloring
								if era == eraHadeanMolten {
									// Molten Land: Black crust patches floating on magma
									// Use rock hardness to simulate crust thickness
									if cellData.RockHardness > 0.8 {
										// Thick crust (black/grey)
										r, g, b = 30, 30, 35
									} else {
										// Thin crust / oozing lava
										r, g, b = 180, 60, 20
									}
								} else if era == eraHadeanCooling {
									// Cooling Land: Dark Basalt everywhere
									// No life, no sediment, just rock
									r, g, b = 40, 40, 45 // Dark Grey Basalt

									// Highlight high peaks (volcanic cones)
									height := elev - geo.SeaLevel
									if height > 4000 {
										r, g, b = 20, 20, 20 // Black obsidian/glass
									}
								} else if era == eraEoarchean {
									// Proto-Continents: Grey/Brown barren rock
									r, g, b = 100, 95, 90
								} else {
									// Standard Archean+ Land Coloring
									height := elev - geo.SeaLevel
									maxHeight := math.Max(maxElev-geo.SeaLevel, 1.0)
									heightFactor := height / maxHeight
									if heightFactor > 1.0 {
										heightFactor = 1.0
									}

									// Base Land Color
									// Phase 5: Province Tinting
									// Cratons (Hardness > 0.8) -> Reddish/Pink (Granite)
									// FoldBelts (Hardness 0.5-0.8) -> Grey/Purple
									// Basins (Hardness < 0.4) -> Yellowish/Tan (Sediment)

									hardness := cellData.RockHardness
									if hardness > 0.8 {
										// Craton - Granite Shield
										r, g, b = 140, 120, 110
									} else if hardness > 0.4 {
										// Orogen / Fold Belt
										r, g, b = 110, 110, 100
									} else {
										// Basin / Sediment
										r, g, b = 130, 125, 100
									}

									// Phase 4: Intertidal zone coloring (wet rock/sand)
									if cellData.IsIntertidal {
										// Wet dark rock/sand exposed at low tide
										if sediment > 0.5 {
											// Wet sand
											r, g, b = 120, 110, 90
										} else {
											// Wet rock
											r, g, b = 45, 45, 50
										}
									} else if cellData.IsSpit {
										// Sandy spit/bar (light tan)
										r, g, b = 170, 155, 130
									} else if cellData.IsEstuary {
										// Marshy estuary (darker, muddy)
										r, g, b = 70, 65, 50
									} else if heightFactor < 0.02 && sediment > 0.5 {
										// Beach sand color (tan)
										r, g, b = 180, 160, 130
									} else if heightFactor < 0.1 {
										// Lowlands - check for sediment (alluvial plains)
										if sediment > 1.0 {
											// Silty lowlands (lighter brown)
											r, g, b = 90, 80, 65
										} else {
											// Dark Basalt
											r, g, b = 60, 55, 50
										}
									} else if heightFactor < 0.3 {
										// Hills - Phase E: Check for volcanic basalt
										rockHardness := cellData.RockHardness
										if rockHardness > 0.9 && sediment < 0.2 {
											// Fresh volcanic basalt (very dark, almost black)
											r, g, b = 35, 32, 30
										} else {
											// Regular hills (Reddish/Brown Rock)
											r, g, b = 100, 80, 70
										}
									} else if heightFactor < 0.6 {
										// Mountains - Phase E: Volcanic vs regular
										rockHardness := cellData.RockHardness
										if rockHardness > 0.85 && sediment < 0.3 {
											// Volcanic cone (dark grey-black)
											r, g, b = 50, 48, 45
										} else {
											// Regular Mountain (Grey Stone)
											r, g, b = 120, 115, 115
										}
									} else if heightFactor < 0.8 {
										// High Mountains - volcanic vs regular
										rockHardness := cellData.RockHardness
										if rockHardness > 0.85 && sediment < 0.2 {
											// Volcanic high peak (dark)
											r, g, b = 80, 75, 70
										} else {
											// High Mountains (Light Grey)
											r, g, b = 160, 160, 160
										}
									} else {
										// Peaks (White/Snow)
										r, g, b = 240, 240, 250
									}
								}
							}

							// ===========================================
							// Phase B: Climate-Based Coloring Override
							// ===========================================
							// Apply ice caps, polar regions, and temperature effects
							// Using actual simulation data for organic appearance

							// Calculate approximate temperature from latitude
							// sLat ranges from -PI/2 (south) to +PI/2 (north)
							// distFromEquator: 0 at equator, 1 at poles
							distFromEquator := math.Abs(sLat) / (math.Pi / 2.0)
							temperature := 1.0 - distFromEquator // 1 = hot (equator), 0 = cold (poles)

							// Calculate height above sea level for climate effects
							heightAboveSea := 0.0
							elevHeightFactor := 0.0
							if elev > geo.SeaLevel {
								heightAboveSea = elev - geo.SeaLevel
								maxH := math.Max(maxElev-geo.SeaLevel, 1.0)
								elevHeightFactor = heightAboveSea / maxH
								if elevHeightFactor > 1.0 {
									elevHeightFactor = 1.0
								}
							}

							// Adjust temperature for elevation (higher = colder)
							// Lapse rate: ~6.5°C per 1000m, normalized to our 0-1 scale
							if elev > geo.SeaLevel {
								elevationEffect := heightAboveSea / 8000.0 // Max effect at 8000m
								if elevationEffect > 0.4 {
									elevationEffect = 0.4
								}
								temperature -= elevationEffect
							}

							// Calculate slope from neighbors for snow shedding
							// Get neighbor elevations (also used later for hillshading)
							leftCoord := geo.Topology.GetNeighbor(coord, spatial.West)
							rightCoord := geo.Topology.GetNeighbor(coord, spatial.East)
							upCoord := geo.Topology.GetNeighbor(coord, spatial.North)
							downCoord := geo.Topology.GetNeighbor(coord, spatial.South)

							leftElev := geo.SphereHeightmap.Get(leftCoord)
							rightElev := geo.SphereHeightmap.Get(rightCoord)
							upElev := geo.SphereHeightmap.Get(upCoord)
							downElev := geo.SphereHeightmap.Get(downCoord)

							// Compute slope magnitude (meters per cell)
							slopeX := (rightElev - leftElev) / 2.0
							slopeY := (upElev - downElev) / 2.0
							slopeMag := math.Sqrt(slopeX*slopeX + slopeY*slopeY)
							slopeFactor := math.Min(slopeMag/500.0, 1.0) // 500m rise = max slope

							// Snow/Ice Rendering using simulation data
							// Cold enough for snow? (temp threshold + modifiers from real data)
							snowThreshold := 0.25 // Base threshold

							// Sediment makes snow accumulate easier (soft surfaces hold snow)
							if sediment > 0.3 {
								snowThreshold += 0.05
							}

							// High flux areas (rivers) don't freeze as easily
							if flux > 50 {
								snowThreshold -= 0.1
							}

							// Steep slopes shed snow
							snowThreshold -= slopeFactor * 0.15

							if temperature < snowThreshold {
								// In snow zone - blend based on how cold it is
								coldness := (snowThreshold - temperature) / snowThreshold // 0 = edge, 1 = deep cold

								if elev > geo.SeaLevel {
									// Land snow/ice
									// Use sediment to vary texture: sediment = smooth snow, no sediment = rocky ice
									if sediment > 0.5 {
										// Deep snow on soft ground (pure white)
										snowR := uint8(245 + coldness*10)
										snowG := uint8(248 + coldness*7)
										snowB := uint8(255)
										r, g, b = snowR, snowG, snowB
									} else if sediment > 0.1 {
										// Patchy snow (white with underlying rock showing)
										// Blend existing rock color with snow
										blendFactor := coldness * (0.5 + sediment)
										if blendFactor > 0.9 {
											blendFactor = 0.9
										}
										r = uint8(float64(r)*(1-blendFactor) + 240*blendFactor)
										g = uint8(float64(g)*(1-blendFactor) + 245*blendFactor)
										b = uint8(float64(b)*(1-blendFactor) + 255*blendFactor)
									} else {
										// Bare rock/ice (blue-grey glacial ice)
										// Less snow sticks to hard rock
										if coldness > 0.6 {
											// Very cold: glacial ice
											r, g, b = 200, 215, 235
										}
										// Otherwise: keep existing rock color (too steep/rocky for snow)
									}
								} else {
									// Sea ice - use depth for color variation
									depth := geo.SeaLevel - elev
									maxDepth := math.Max(geo.SeaLevel-minElev, 1.0)
									depthNorm := math.Min(depth/maxDepth, 1.0)

									if coldness > 0.5 {
										// Thick pack ice
										r = uint8(220 - depthNorm*20)
										g = uint8(235 - depthNorm*15)
										b = uint8(250 - depthNorm*10)
									} else {
										// Thin ice / ice edge (more blue, translucent look)
										r = uint8(180 + coldness*40)
										g = uint8(210 + coldness*25)
										b = uint8(240 + coldness*15)
									}
								}
							} else if temperature < 0.35 && elev > geo.SeaLevel {
								// Tundra / subpolar zone
								// Color varies based on sediment (vegetated vs bare)
								if sediment > 0.3 {
									// Some vegetation possible - brownish-green
									r, g, b = 115, 115, 95
								} else {
									// Bare permafrost - grey-brown rock
									r, g, b = 130, 125, 115
								}
							}

							// Note: Desert coloring would require moisture data
							// which we don't have per-pixel access to in current renderer

							// ===========================================
							// Hillshading: Add 3D depth perception
							// ===========================================
							// Calculate surface normal from neighbor elevations
							// and apply directional lighting

							// Sun direction (from northwest, elevated 45 degrees)
							// In screen space: light comes from top-left
							sunAzimuth := math.Pi * 0.75  // 135 degrees (NW)
							sunAltitude := math.Pi * 0.25 // 45 degrees elevation

							// Sun vector (normalized)
							sunX := math.Cos(sunAltitude) * math.Sin(sunAzimuth)
							sunY := math.Sin(sunAltitude) // Up component
							sunZ := math.Cos(sunAltitude) * math.Cos(sunAzimuth)

							// Optimize: Use topology neighbors instead of vector re-projection
							// This avoids expensive sin/cos/atan calls for every neighbor

							// Neighbor coordinates were already computed above for snow calculations
							// Reusing: leftCoord, rightCoord, upCoord, downCoord
							// Reusing: leftElev, rightElev, upElev, downElev

							// Calculate gradients
							// Slope X = (Right - Left)
							// Slope Y = (Up - Down)

							dzdx := (rightElev - leftElev) * 4.0 // Scale factor for more visible relief
							dzdy := (upElev - downElev) * 4.0

							// Calculate gradients (elevation change per pixel)
							// Scale factor to exaggerate relief for visual effect
							reliefScale := 0.001 // 10x increase for more visible shadows
							dzdx = (rightElev - leftElev) * reliefScale
							dzdy = (upElev - downElev) * reliefScale

							// Cap extreme slope values to prevent black/white artifacts at plate boundaries
							// Plate boundaries can have 10000+ meter elevation changes between neighbors
							maxSlope := 1.5 // Reasonable max slope for visual shading
							if dzdx > maxSlope {
								dzdx = maxSlope
							} else if dzdx < -maxSlope {
								dzdx = -maxSlope
							}
							if dzdy > maxSlope {
								dzdy = maxSlope
							} else if dzdy < -maxSlope {
								dzdy = -maxSlope
							}

							// Calculate surface normal from gradient
							// Normal = (-dz/dx, 1, -dz/dy) normalized
							nx := -dzdx
							ny := 1.0
							nz := -dzdy
							nLen := math.Sqrt(nx*nx + ny*ny + nz*nz)
							if nLen > 0 {
								nx /= nLen
								ny /= nLen
								nz /= nLen
							}

							// Calculate lighting (dot product of normal and sun direction)
							lighting := nx*sunX + ny*sunY + nz*sunZ

							// Clamp and remap lighting to useful range
							// Range: 0.4 (shadow) to 1.2 (highlight)
							if lighting < 0 {
								lighting = 0
							}
							lighting = 0.5 + lighting*0.7 // Brightened: was 0.4 + 0.8

							// Add ambient occlusion for valleys (lower elevations get darker)
							// Only apply to land, not water (oceans or lakes)
							if elev > geo.SeaLevel && !cellData.IsLake {
								// Check if surrounded by higher terrain (valley)
								avgNeighbor := (leftElev + rightElev + upElev + downElev) / 4.0
								if avgNeighbor > elev {
									// In a valley - darken slightly
									valleyDepth := (avgNeighbor - elev) / 500.0 // 500m = max valley effect
									if valleyDepth > 0.2 {
										valleyDepth = 0.2
									}
									lighting -= valleyDepth
								}
							}

							// Apply lighting to color - but SKIP for underwater pixels
							// Water should appear smooth without seafloor topology showing through
							if elev >= geo.SeaLevel && !cellData.IsLake {
								r = uint8(math.Min(255, math.Max(0, float64(r)*lighting)))
								g = uint8(math.Min(255, math.Max(0, float64(g)*lighting)))
								b = uint8(math.Min(255, math.Max(0, float64(b)*lighting)))
							}
							// For water/lakes: keep original colors without hillshading

							// Accumulate
							rSum += float64(r)
							gSum += float64(g)
							bSum += float64(b)
							count++
						}
					}

					// Set Pixel
					offset := (y*width + x) * 4
					rawBuf[offset] = uint8(rSum / count)
					rawBuf[offset+1] = uint8(gSum / count)
					rawBuf[offset+2] = uint8(bSum / count)
					rawBuf[offset+3] = 255 // Alpha
				}
			}
		}(startRow, endRow)
	}

	wg.Wait()

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Encode to WebP
	var buf bytes.Buffer
	// Pre-allocate buffer estimate?
	// New Buffer doesn't support that easily, but Encode writes streams.

	// Create options
	opts := &webp.Options{
		Lossless: false,
		Quality:  float32(r.config.WebPQuality),
	}

	// Use pool for the output buffer?
	// The output needs to be a []byte returned to caller.
	// Caller assumes ownership. So we allocate a standard slice.
	// Or we use a pool and copy?
	// For simplicity in Sprint 1, just allocate.
	if err := webp.Encode(&buf, img, opts); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// RenderNormalMapPNG generates a tangent-space normal map from heightmap data.
// R = X (+1) * 0.5
// G = Y (+1) * 0.5
// B = Z
func (r *Renderer) RenderNormalMapPNG(ctx context.Context, worldID string, geo *ecosystem.WorldGeology, width, height int) ([]byte, error) {
	// Validation
	if width <= 0 {
		width = r.config.DefaultWidth
	}
	if height <= 0 {
		height = r.config.DefaultHeight
	}
	if width > r.config.MaxWidth {
		width = r.config.MaxWidth
	}
	if height > r.config.MaxHeight {
		height = r.config.MaxHeight
	}

	if geo.SphereHeightmap == nil {
		return nil, ErrWorldNotInitialized
	}

	topo := geo.SphereHeightmap.Topology()
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Precompute normalized height per neighbor for gradient calculation
	// We operate in spherical space but map to 2D texture
	minElev, maxElev := geo.SphereHeightmap.MinMax()
	elevRange := maxElev - minElev
	if elevRange <= 0 {
		elevRange = 1.0
	}

	// Use parallel processing as this is expensive
	numWorkers := runtime.NumCPU()
	rowsPerWorker := height / numWorkers
	if rowsPerWorker == 0 {
		rowsPerWorker = 1
	}

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for w := 0; w < numWorkers; w++ {
		startRow := w * rowsPerWorker
		endRow := startRow + rowsPerWorker
		if w == numWorkers-1 {
			endRow = height
		}

		go func(yStart, yEnd int) {
			defer wg.Done()
			for y := yStart; y < yEnd; y++ {
				for x := 0; x < width; x++ {
					// Get spherical coord
					u, v := float64(x)/float64(width), float64(y)/float64(height)

					// We need 4 neighbors to compute gradient
					// Wrap x (longitude), clamp y (latitude)

					// Center
					lat := (0.5 - v) * math.Pi
					lon := (u - 0.5) * 2 * math.Pi

					vx := math.Cos(lat) * math.Cos(lon)
					vy := math.Sin(lat)
					vz := math.Cos(lat) * math.Sin(lon)
					centerCoord := topo.FromVector(vx, vy, vz)
					// hCenter := geo.SphereHeightmap.Get(centerCoord) // Unused optimization

					// Neighbors for Sobel-like filter
					// Due to spherical distortion, pixel-based neighbors are tricky.
					// We'll use the topology neighbors from the face coordinate.
					leftCoord := topo.GetNeighbor(centerCoord, spatial.West)
					rightCoord := topo.GetNeighbor(centerCoord, spatial.East)
					upCoord := topo.GetNeighbor(centerCoord, spatial.North)
					downCoord := topo.GetNeighbor(centerCoord, spatial.South)

					hLeft := geo.SphereHeightmap.Get(leftCoord)
					hRight := geo.SphereHeightmap.Get(rightCoord)
					hUp := geo.SphereHeightmap.Get(upCoord)
					hDown := geo.SphereHeightmap.Get(downCoord)

					// Calculate slope vectors
					// Scale strength creates deeper looking normals
					strength := 4.0
					dX := (hRight - hLeft) * strength / elevRange
					dY := (hUp - hDown) * strength / elevRange

					// Construct Normal Vector (Z is up)
					// N = normalize(-dX, -dY, 1.0)
					// Note: BabylonJS expects tangent space normals
					nx, ny, nz := -dX, -dY, 1.0
					len := math.Sqrt(nx*nx + ny*ny + nz*nz)
					if len > 0 {
						nx /= len
						ny /= len
						nz /= len
					}

					// Encode to RGB [0, 255]
					// R = (x + 1) * 0.5 * 255
					// G = (y + 1) * 0.5 * 255
					// B = z * 255
					r := uint8((nx + 1.0) * 0.5 * 255)
					g := uint8((ny + 1.0) * 0.5 * 255)
					b := uint8(nz * 255)

					img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
				}
			}
		}(startRow, endRow)
	}

	wg.Wait()

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
