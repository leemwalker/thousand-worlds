package service

import (
	"context"
	"fmt"

	"tw-backend/internal/repository"

	"github.com/google/uuid"
)

type SpatialService struct {
	repo repository.SpatialRepository
}

func NewSpatialService(repo repository.SpatialRepository) *SpatialService {
	return &SpatialService{
		repo: repo,
	}
}

func (s *SpatialService) UpdateLocation(ctx context.Context, entityID uuid.UUID, x, y, z float64) error {
	if err := s.repo.UpdateEntityLocation(ctx, entityID, x, y, z); err != nil {
		return fmt.Errorf("spatialService.UpdateLocation: %w", err)
	}
	return nil
}
