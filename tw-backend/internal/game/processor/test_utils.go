package processor

import (
	"context"
	"tw-backend/internal/auth"
	"tw-backend/internal/ecosystem"
	gamemap "tw-backend/internal/game/services/map"

	"github.com/google/uuid"
)

// mockMapService implements necessary methods for testing
type mockMapService struct {
	geologyMap map[uuid.UUID]*ecosystem.WorldGeology
}

func (m *mockMapService) RenderMap(ctx context.Context, worldID uuid.UUID, geo *ecosystem.WorldGeology, width, height int) ([]byte, error) {
	return []byte("fake-image"), nil
}

func (m *mockMapService) RenderHeightmapPNG(ctx context.Context, worldID uuid.UUID, geo *ecosystem.WorldGeology, width, height int) ([]byte, error) {
	return []byte("fake-heightmap"), nil
}

func (m *mockMapService) RenderMaterialPNG(ctx context.Context, worldID uuid.UUID, geo *ecosystem.WorldGeology, width, height int) ([]byte, error) {
	return []byte("fake-material"), nil
}

func (m *mockMapService) RenderIcePNG(ctx context.Context, worldID uuid.UUID, geo *ecosystem.WorldGeology, width, height int) ([]byte, error) {
	return []byte("fake-ice"), nil
}

func (m *mockMapService) RenderNormalMapPNG(ctx context.Context, worldID uuid.UUID, geo *ecosystem.WorldGeology, width, height int) ([]byte, error) {
	return []byte("fake-normal"), nil
}

func (m *mockMapService) BuildBinaryGrid(geo *ecosystem.WorldGeology, width, height int) *gamemap.BinaryGrid {
	return nil
}

func (m *mockMapService) SetWorldGeology(worldID uuid.UUID, geo *ecosystem.WorldGeology) {
	if m.geologyMap == nil {
		m.geologyMap = make(map[uuid.UUID]*ecosystem.WorldGeology)
	}
	if geo == nil {
		delete(m.geologyMap, worldID)
	} else {
		m.geologyMap[worldID] = geo
	}
}

func (m *mockMapService) GetWorldGeology(worldID uuid.UUID) *ecosystem.WorldGeology {
	if m.geologyMap == nil {
		return nil
	}
	return m.geologyMap[worldID]
}

func (m *mockMapService) GetWorldMapData(ctx context.Context, char *auth.Character, gridSize int) (*gamemap.WorldMapData, error) {
	return &gamemap.WorldMapData{}, nil
}

func (m *mockMapService) GetMapData(ctx context.Context, char *auth.Character) (*gamemap.MapData, error) {
	return &gamemap.MapData{}, nil
}

func (m *mockMapService) TileRenderer() *gamemap.TileRenderer {
	return nil
}
