package service

import (
	"context"
	"fmt"

	"mud-platform-backend/internal/repository"
)

type SpatialService struct {
	repo *repository.SpatialRepository
}

func NewSpatialService(repo *repository.SpatialRepository) *SpatialService {
	return &SpatialService{
		repo: repo,
	}
}

func (s *SpatialService) UpdateLocation(ctx context.Context, entityID string, worldID string, x, y, z float64) error {
	if err := s.repo.UpdateEntityLocation(ctx, entityID, worldID, x, y, z); err != nil {
		return fmt.Errorf("spatialService.UpdateLocation: %w", err)
	}
	return nil
}
