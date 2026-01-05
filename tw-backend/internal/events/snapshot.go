package events

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"

	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"
)

// SerializeHeightmap creates a gzip-compressed binary snapshot of a SphereHeightmap.
// Format:
//   - int32: resolution (per face)
//   - int32: face count (always 6)
//   - For each face: resolution*resolution float64 elevation values
//
// Returns the compressed bytes suitable for HeightmapSnapshot.Data field.
func SerializeHeightmap(hm *geography.SphereHeightmap) ([]byte, error) {
	if hm == nil {
		return nil, fmt.Errorf("nil heightmap")
	}

	var buf bytes.Buffer
	gw, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		return nil, fmt.Errorf("gzip writer: %w", err)
	}

	res := hm.Resolution()

	// Write header
	if err := binary.Write(gw, binary.LittleEndian, int32(res)); err != nil {
		return nil, fmt.Errorf("write resolution: %w", err)
	}
	if err := binary.Write(gw, binary.LittleEndian, int32(6)); err != nil {
		return nil, fmt.Errorf("write face count: %w", err)
	}

	// Write all 6 faces
	for face := 0; face < 6; face++ {
		faceHM := hm.GetFace(face)
		if faceHM == nil {
			return nil, fmt.Errorf("nil face %d", face)
		}
		for _, elev := range faceHM.Elevations {
			if err := binary.Write(gw, binary.LittleEndian, elev); err != nil {
				return nil, fmt.Errorf("write elevation: %w", err)
			}
		}
	}

	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("gzip close: %w", err)
	}

	return buf.Bytes(), nil
}

// DeserializeHeightmap reconstructs a SphereHeightmap from gzip-compressed binary data.
// Requires a Topology to reconstruct the spatial structure.
func DeserializeHeightmap(data []byte, topo spatial.Topology) (*geography.SphereHeightmap, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("gzip reader: %w", err)
	}
	defer gr.Close()

	// Read header
	var res, faceCount int32
	if err := binary.Read(gr, binary.LittleEndian, &res); err != nil {
		return nil, fmt.Errorf("read resolution: %w", err)
	}
	if err := binary.Read(gr, binary.LittleEndian, &faceCount); err != nil {
		return nil, fmt.Errorf("read face count: %w", err)
	}

	if faceCount != 6 {
		return nil, fmt.Errorf("expected 6 faces, got %d", faceCount)
	}

	// Verify topology resolution matches
	if topo.Resolution() != int(res) {
		return nil, fmt.Errorf("topology resolution %d != snapshot resolution %d", topo.Resolution(), res)
	}

	// Create new heightmap and populate
	hm := geography.NewSphereHeightmap(topo)

	for face := 0; face < 6; face++ {
		faceHM := hm.GetFace(face)
		for i := range faceHM.Elevations {
			if err := binary.Read(gr, binary.LittleEndian, &faceHM.Elevations[i]); err != nil {
				if err == io.EOF {
					return nil, fmt.Errorf("unexpected EOF at face %d, index %d", face, i)
				}
				return nil, fmt.Errorf("read elevation: %w", err)
			}
		}
	}

	// Update min/max
	hm.UpdateMinMax()

	return hm, nil
}

// EstimateSnapshotSize returns the approximate uncompressed size of a heightmap snapshot.
// Formula: 8 bytes (header) + 6 faces * resolution^2 * 8 bytes per float64
func EstimateSnapshotSize(resolution int) int {
	return 8 + 6*resolution*resolution*8
}

// EstimateCompressedSize returns a rough estimate of compressed snapshot size.
// Typically gzip achieves ~60-70% compression on elevation data.
func EstimateCompressedSize(resolution int) int {
	return int(float64(EstimateSnapshotSize(resolution)) * 0.35)
}
