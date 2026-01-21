package api

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	gamemap "tw-backend/internal/game/services/map"
	"tw-backend/internal/repository"

	"github.com/go-chi/chi/v5"
)

// MapHandler handles world map related requests
type MapHandler struct {
	worldRepo  repository.WorldRepository
	mapService *gamemap.Service
}

// NewMapHandler creates a new map handler
func NewMapHandler(worldRepo repository.WorldRepository, mapService *gamemap.Service) *MapHandler {
	return &MapHandler{
		worldRepo:  worldRepo,
		mapService: mapService,
	}
}

// GetTile returns a binary tile for the given face, level, x, y coordinates
// Route: /api/v1/world/tiles/{face}/{lod}/{x}/{y}
func (h *MapHandler) GetTile(w http.ResponseWriter, r *http.Request) {
	// Parse parameters
	faceStr := chi.URLParam(r, "face")
	lodStr := chi.URLParam(r, "lod")
	xStr := chi.URLParam(r, "x")
	yStr := chi.URLParam(r, "y")

	face, err := strconv.Atoi(faceStr)
	if err != nil || face < 0 || face > 5 {
		http.Error(w, "Invalid face", http.StatusBadRequest)
		return
	}

	lod, err := strconv.Atoi(lodStr)
	if err != nil || lod < 0 {
		http.Error(w, "Invalid lod", http.StatusBadRequest)
		return
	}

	x, err := strconv.Atoi(xStr)
	if err != nil || x < 0 {
		http.Error(w, "Invalid x", http.StatusBadRequest)
		return
	}

	y, err := strconv.Atoi(yStr)
	if err != nil || y < 0 {
		http.Error(w, "Invalid y", http.StatusBadRequest)
		return
	}

	// Assuming WorldID is available from context or authentication
	// For now, we might need to pass it or get it from the user.
	// Since the route is protected, we can get the character and then the world.
	// However, to keep it simple and stateless for this endpoint (typically these are public or generic),
	// let's assume we might need a query param for worldID or use a default active one.
	// But `RenderRawTile` needs `worldGeo`.
	// Let's use `h.mapService.GetWorldGeology(worldID)` (we need to expose this or similar).

	// For the MVP, we'll try to get the worldID from a header or query param, or check if `TileRenderer` context has it.
	// Wait, `TileRenderer.RenderTile` takes `worldGeo`. The handler needs to fetch it.
	// The previous implementation in `world_commands.go` got it from context/processor.
	// `h.mapService` should probably handle `GetWorldGeology`.

	// NOTE: This handler needs access to the WorldGeology. I'll need to update `MapHandler` to have access to a way to fetch it.
	// `gamemap.Service` (which is in `internal/game/services/map`) might not have a public `GetWorldGeology`.
	// I need to check `internal/game/services/map/service.go`.

	// Assuming we can get it:
	tileRenderer := h.mapService.TileRenderer()
	if tileRenderer == nil {
		http.Error(w, "Tile renderer not available", http.StatusServiceUnavailable)
		return
	}

	// TODO: Need real WorldGeology. passing nil will generate placeholder (which is fine for initial test)
	// In a real scenario, we'll likely need to extract World ID from the request context (user session)
	// and fetch the geology from a shared service or cache.
	// For now, passing nil to demonstrate the binary format.

	req := gamemap.TileRequest{
		Face:  gamemap.CubeFace(face),
		Level: lod,
		X:     x,
		Y:     y,
		Size:  256, // Fixed size for now, or make it query param
	}

	// We need to resolve `worldGeo`.
	// The `world_commands.go` used `p.worldGeology[char.WorldID]`.
	// We should probably inject a `GeologyProvider` interface into this handler.
	// For now, to unblock, I will pass nil which `RenderRawTile` handles by generating a placeholder.
	// This allows testing the pipeline structure.

	data, err := tileRenderer.RenderRawTile(r.Context(), req, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to render tile: %v", err), http.StatusInternalServerError)
		return
	}

	// SERIALIZATION
	// Header: Magic `0xWT` (WorldTile), Version `0x01`
	buf := new(bytes.Buffer)
	buf.Write([]byte("WT")) // Magic
	buf.WriteByte(0x01)     // Version

	// Metadata JSON
	meta := map[string]interface{}{
		"f": face,
		"l": lod,
		"x": x,
		"y": y,
		"w": data.Width,
		"h": data.Height,
	}
	metaBytes, _ := json.Marshal(meta)
	binary.Write(buf, binary.LittleEndian, uint32(len(metaBytes)))
	buf.Write(metaBytes)

	// Payload Construction (Uncompressed first)
	payloadBuf := new(bytes.Buffer)

	// Write Heightmap (floats)
	for _, h := range data.Heightmap {
		binary.Write(payloadBuf, binary.LittleEndian, h)
	}
	// Write Biomes (uint8)
	payloadBuf.Write(data.Biomes)
	// Write Water (floats)
	for _, w := range data.Water {
		binary.Write(payloadBuf, binary.LittleEndian, w)
	}

	// Compress Payload (GZIP)
	var compressedBuf bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedBuf)
	if _, err := gzipWriter.Write(payloadBuf.Bytes()); err != nil {
		http.Error(w, "Failed to compress data", http.StatusInternalServerError)
		return
	}
	if err := gzipWriter.Close(); err != nil {
		http.Error(w, "Failed to close gzip writer", http.StatusInternalServerError)
		return
	}
	compressedData := compressedBuf.Bytes()

	// Write Payload Length and Compressed Data
	binary.Write(buf, binary.LittleEndian, uint32(len(compressedData)))
	buf.Write(compressedData)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}
