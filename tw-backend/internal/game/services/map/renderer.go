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
							var r, g, b uint8

							if elev < geo.SeaLevel {
								// Water - deep, dark oceans (less "tropical" more "primordial")
								// Deep: 5, 10, 25 (Very Dark)
								// Shallow: 30, 60, 90 (Dark Blue-Grey)
								depth := geo.SeaLevel - elev
								maxDepth := math.Max(geo.SeaLevel-minElev, 1.0)
								depthFactor := depth / maxDepth
								if depthFactor > 1.0 {
									depthFactor = 1.0
								}

								f := 1.0 - depthFactor // 0 = Deep, 1 = Shallow
								r = uint8(5.0 + f*25.0)
								g = uint8(10.0 + f*50.0)
								b = uint8(25.0 + f*65.0)

							} else {
								// Land - Lifeless Rock (Greys, Browns, Rusty Reds)
								// No greens/vegetation colors
								height := elev - geo.SeaLevel
								maxHeight := math.Max(maxElev-geo.SeaLevel, 1.0)
								heightFactor := height / maxHeight
								if heightFactor > 1.0 {
									heightFactor = 1.0
								}

								if heightFactor < 0.1 {
									// Lowlands (Dark Basalt/Sediment)
									// 50, 45, 40
									r, g, b = 60, 55, 50
								} else if heightFactor < 0.3 {
									// Hills (Reddish/Brown Rock)
									// 100, 80, 70
									r, g, b = 100, 80, 70
								} else if heightFactor < 0.6 {
									// Mountains (Grey Stone)
									// 120, 115, 115
									r, g, b = 120, 115, 115
								} else if heightFactor < 0.8 {
									// High Mountains (Light Grey)
									// 160, 160, 160
									r, g, b = 160, 160, 160
								} else {
									// Peaks (White/Snow)
									// 240, 240, 250
									r, g, b = 240, 240, 250
								}
							}

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
