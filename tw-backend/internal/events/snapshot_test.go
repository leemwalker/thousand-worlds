package events

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"testing"

	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"
)

func TestSerializeDeserializeHeightmap(t *testing.T) {
	// Create a small test heightmap
	topo := spatial.NewCubeSphereTopology(8) // 8x8 per face
	hm := geography.NewSphereHeightmap(topo)

	// Set some test values
	for face := 0; face < 6; face++ {
		faceHM := hm.GetFace(face)
		for i := range faceHM.Elevations {
			// Unique value for each cell: face*1000 + index
			faceHM.Elevations[i] = float64(face*1000 + i)
		}
	}
	hm.UpdateMinMax()

	// Serialize
	data, err := SerializeHeightmap(hm)
	if err != nil {
		t.Fatalf("SerializeHeightmap failed: %v", err)
	}

	// Verify compression is working (data should be smaller than raw)
	rawSize := EstimateSnapshotSize(8)
	if len(data) >= rawSize {
		t.Errorf("Compression not working: compressed=%d >= raw=%d", len(data), rawSize)
	}
	t.Logf("Compression ratio: %.1f%% (raw=%d, compressed=%d)",
		float64(len(data))/float64(rawSize)*100, rawSize, len(data))

	// Deserialize
	restored, err := DeserializeHeightmap(data, topo)
	if err != nil {
		t.Fatalf("DeserializeHeightmap failed: %v", err)
	}

	// Verify values match
	for face := 0; face < 6; face++ {
		origFace := hm.GetFace(face)
		restoredFace := restored.GetFace(face)

		for i := range origFace.Elevations {
			if origFace.Elevations[i] != restoredFace.Elevations[i] {
				t.Errorf("Face %d, index %d: got %f, want %f",
					face, i, restoredFace.Elevations[i], origFace.Elevations[i])
			}
		}
	}
}

func TestSerializeHeightmap_NilInput(t *testing.T) {
	_, err := SerializeHeightmap(nil)
	if err == nil {
		t.Error("Expected error for nil heightmap")
	}
}

func TestDeserializeHeightmap_EmptyData(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(8)
	_, err := DeserializeHeightmap([]byte{}, topo)
	if err == nil {
		t.Error("Expected error for empty data")
	}
}

func TestDeserializeHeightmap_ResolutionMismatch(t *testing.T) {
	// Create snapshot at resolution 8
	topo8 := spatial.NewCubeSphereTopology(8)
	hm := geography.NewSphereHeightmap(topo8)
	data, err := SerializeHeightmap(hm)
	if err != nil {
		t.Fatalf("SerializeHeightmap failed: %v", err)
	}

	// Try to deserialize with wrong resolution
	topo16 := spatial.NewCubeSphereTopology(16)
	_, err = DeserializeHeightmap(data, topo16)
	if err == nil {
		t.Error("Expected error for resolution mismatch")
	}
}

func TestDeserializeHeightmap_CorruptedData(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(8)

	// Create valid-looking but truncated gzip data
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	binary.Write(gw, binary.LittleEndian, int32(8)) // resolution
	binary.Write(gw, binary.LittleEndian, int32(6)) // face count
	// Write only partial data
	binary.Write(gw, binary.LittleEndian, float64(1.0))
	gw.Close()

	_, err := DeserializeHeightmap(buf.Bytes(), topo)
	if err == nil {
		t.Error("Expected error for truncated data")
	}
}

func TestEstimateSnapshotSize(t *testing.T) {
	tests := []struct {
		resolution int
		wantSize   int
	}{
		{8, 8 + 6*8*8*8},       // 3080 bytes
		{64, 8 + 6*64*64*8},    // ~200KB
		{256, 8 + 6*256*256*8}, // ~3MB
	}

	for _, tt := range tests {
		got := EstimateSnapshotSize(tt.resolution)
		if got != tt.wantSize {
			t.Errorf("EstimateSnapshotSize(%d) = %d, want %d", tt.resolution, got, tt.wantSize)
		}
	}
}

func BenchmarkSerializeHeightmap(b *testing.B) {
	topo := spatial.NewCubeSphereTopology(64) // Realistic size
	hm := geography.NewSphereHeightmap(topo)

	// Initialize with realistic elevation data
	for face := 0; face < 6; face++ {
		faceHM := hm.GetFace(face)
		for i := range faceHM.Elevations {
			faceHM.Elevations[i] = float64(i%1000) - 500 // -500 to 500
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := SerializeHeightmap(hm)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDeserializeHeightmap(b *testing.B) {
	topo := spatial.NewCubeSphereTopology(64)
	hm := geography.NewSphereHeightmap(topo)

	data, err := SerializeHeightmap(hm)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := DeserializeHeightmap(data, topo)
		if err != nil {
			b.Fatal(err)
		}
	}
}
