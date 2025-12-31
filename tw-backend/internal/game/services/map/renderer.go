package gamemap

import (
	"bytes"
	"context"
	"errors"
	"image"
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
								// Water coloring - varies by depth and sediment
								depth := geo.SeaLevel - elev
								maxDepth := math.Max(geo.SeaLevel-minElev, 1.0)
								depthFactor := depth / maxDepth
								if depthFactor > 1.0 {
									depthFactor = 1.0
								}

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
									// Deep ocean gradient
									f := 1.0 - depthFactor // 0 = Deep, 1 = Shallow
									r = uint8(5.0 + f*25.0)
									g = uint8(10.0 + f*50.0)
									b = uint8(25.0 + f*65.0)
								}

							} else if cellData.IsLake {
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
							} else {
								// Land coloring
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

							// ===========================================
							// Phase B: Climate-Based Coloring Override
							// ===========================================
							// Apply ice caps, polar regions, and temperature effects

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
							// Every 1000m reduces temp by 0.1 (lapse rate effect)
							if elev > geo.SeaLevel {
								elevationEffect := heightAboveSea / 10000.0 // Max 0.3 effect at 3000m
								if elevationEffect > 0.3 {
									elevationEffect = 0.3
								}
								temperature -= elevationEffect
							}

							// Ice/Snow caps: Cold temperature + high elevation OR polar
							if temperature < 0.2 || (temperature < 0.4 && elevHeightFactor > 0.6) {
								if elev > geo.SeaLevel {
									// Ice cap / glacier (white with blue tint)
									r, g, b = 235, 240, 250
								} else {
									// Sea ice (light blue-white)
									r, g, b = 200, 220, 240
								}
							} else if temperature < 0.35 && elev > geo.SeaLevel {
								// Tundra / frozen ground (grey-brown)
								r, g, b = 130, 125, 115
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

							// Get neighbor coordinates from topology (graph navigation)
							// Coordinate is an integer ID that maps to faces/grid
							// Note: Topology neighbors might wrap around cube faces automatically

							leftCoord := geo.Topology.GetNeighbor(coord, spatial.West)
							rightCoord := geo.Topology.GetNeighbor(coord, spatial.East)
							upCoord := geo.Topology.GetNeighbor(coord, spatial.North)
							downCoord := geo.Topology.GetNeighbor(coord, spatial.South)

							// If neighbor lookup returns invalid coordinate (if implemented), check face index?
							// Actually coord 0 is valid (Face 0, 0,0). GetNeighbor should always return valid on sphere.

							leftElev := geo.SphereHeightmap.Get(leftCoord)
							rightElev := geo.SphereHeightmap.Get(rightCoord)
							upElev := geo.SphereHeightmap.Get(upCoord)
							downElev := geo.SphereHeightmap.Get(downCoord)

							// Calculate gradients
							// Slope X = (Right - Left)
							// Slope Y = (Up - Down)

							dzdx := (rightElev - leftElev) * 4.0 // Scale factor for more visible relief
							dzdy := (upElev - downElev) * 4.0

							// Calculate gradients (elevation change per pixel)
							// Scale factor to exaggerate relief for visual effect
							reliefScale := 0.0001 // Adjust for visual effect
							dzdx = (rightElev - leftElev) * reliefScale
							dzdy = (upElev - downElev) * reliefScale

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
							lighting = 0.4 + lighting*0.8

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

							// Apply lighting to color
							r = uint8(math.Min(255, math.Max(0, float64(r)*lighting)))
							g = uint8(math.Min(255, math.Max(0, float64(g)*lighting)))
							b = uint8(math.Min(255, math.Max(0, float64(b)*lighting)))

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
