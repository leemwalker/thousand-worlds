// Package main provides a CLI tool to generate time-lapse GIFs of world formation.
//
// Usage:
//
//	gif-generator --seed 42 --years 1000000000 --width 1024 --out evolution.gif --interval 10000000
package main

import (
	"flag"
	"image"
	"image/color/palette"
	"image/gif"
	"log"
	"os"
	"time"

	"github.com/google/uuid"

	"tw-backend/internal/ecosystem"
	"tw-backend/internal/graphics"
)

func main() {
	// Parse command line flags
	seed := flag.Int64("seed", time.Now().UnixNano(), "Random seed for simulation")
	years := flag.Int64("years", 1_000_000_000, "Total simulation duration in years")
	width := flag.Int("width", 1024, "Output image width (height = width/2)")
	out := flag.String("out", "evolution.gif", "Output GIF filename")
	interval := flag.Int64("interval", 10_000_000, "Years between GIF frames")
	flag.Parse()

	if *years <= 0 || *interval <= 0 || *width <= 0 {
		log.Fatal("Invalid arguments: years, interval, and width must be positive")
	}

	log.Printf("GIF Generator - Starting simulation")
	log.Printf("  Seed: %d", *seed)
	log.Printf("  Duration: %d years", *years)
	log.Printf("  Frame interval: %d years", *interval)
	log.Printf("  Output: %s (%dx%d)", *out, *width, *width/2)

	// Initialize world geology
	worldID := uuid.New()
	circumference := 40_075_000.0 // Earth-like circumference in meters

	geology := ecosystem.NewWorldGeology(worldID, *seed, circumference)
	geology.InitializeGeology()

	log.Printf("World initialized with %d plates", len(geology.Plates))

	// Create renderer
	renderer := graphics.NewRenderer(*width)

	// Collect frames
	var frames []*image.Paletted
	var delays []int
	frameDelay := 10 // 100ms per frame (10 * 10ms)

	// Capture initial state
	captureFrame(renderer, geology, &frames, &delays, frameDelay, 0)

	// Run simulation loop
	startTime := time.Now()
	stepsPerInterval := *interval / 100_000 // Simulate in 100k year steps
	if stepsPerInterval < 1 {
		stepsPerInterval = 1
	}
	stepSize := *interval / stepsPerInterval

	totalFrames := int(*years / *interval)
	frameCount := 1

	for simYear := int64(0); simYear < *years; simYear += stepSize {
		// Advance simulation
		geology.SimulateGeology(stepSize, 0.0)

		// Capture frame at each interval
		if (simYear+stepSize)%(*interval) == 0 || simYear+stepSize >= *years {
			captureFrame(renderer, geology, &frames, &delays, frameDelay, simYear+stepSize)
			frameCount++

			// Progress update
			elapsed := time.Since(startTime)
			progress := float64(simYear+stepSize) / float64(*years) * 100
			log.Printf("Frame %d/%d (%.1f%%) - Year %dM - Elapsed: %v",
				frameCount, totalFrames, progress,
				(simYear+stepSize)/1_000_000, elapsed.Round(time.Second))
		}
	}

	// Encode GIF
	log.Printf("Encoding %d frames to GIF...", len(frames))

	f, err := os.Create(*out)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer f.Close()

	err = gif.EncodeAll(f, &gif.GIF{
		Image: frames,
		Delay: delays,
	})
	if err != nil {
		log.Fatalf("Failed to encode GIF: %v", err)
	}

	elapsed := time.Since(startTime)
	log.Printf("Done! Generated %s with %d frames in %v", *out, len(frames), elapsed.Round(time.Second))
}

// captureFrame renders a frame and adds it to the GIF
func captureFrame(
	renderer *graphics.Renderer,
	geology *ecosystem.WorldGeology,
	frames *[]*image.Paletted,
	delays *[]int,
	delay int,
	year int64,
) {
	// Build events string from recent events
	events := make([]string, 0)
	if len(geology.EventBuffer) > 0 {
		events = append(events, geology.EventBuffer[len(geology.EventBuffer)-1])
	}

	info := graphics.FrameInfo{
		Year:       year,
		PlateCount: len(geology.Plates),
		Events:     events,
	}

	// Render frame
	img := renderer.RenderFrame(geology.Heightmap, info)

	// Convert to paletted image for GIF
	bounds := img.Bounds()
	paletted := image.NewPaletted(bounds, palette.Plan9)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			paletted.Set(x, y, img.At(x, y))
		}
	}

	*frames = append(*frames, paletted)
	*delays = append(*delays, delay)
}
