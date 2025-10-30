package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/berkecengiz/appointment-service-boilerplate/internal/models"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ProviderService encapsulates business logic for providers.
type ProviderService struct {
	db *bun.DB
}

// NewProviderService constructs a provider service using the given database connection.
func NewProviderService(db *bun.DB) *ProviderService {
	return &ProviderService{db: db}
}

// ListProviders returns all providers ordered by name.
func (s *ProviderService) ListProviders(ctx context.Context) ([]models.Provider, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var providers []models.Provider
	if err := s.db.NewSelect().Model(&providers).OrderExpr("name ASC").Scan(ctx); err != nil {
		return nil, fmt.Errorf("query providers: %w", err)
	}
	return providers, nil
}

// GetProviderByID finds a provider by identifier.
func (s *ProviderService) GetProviderByID(ctx context.Context, id string) (*models.Provider, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var provider models.Provider
	if err := s.db.NewSelect().Model(&provider).Where("id = ?", id).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query provider by id: %w", err)
	}
	return &provider, nil
}

// CreateProvider inserts a new provider record.
func (s *ProviderService) CreateProvider(ctx context.Context, req models.CreateProviderRequest) (*models.Provider, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	provider := models.Provider{
		ID:      uuid.NewString(),
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.Phone,
		Address: req.Address,
	}

	if _, err := s.db.NewInsert().Model(&provider).Returning("").Exec(ctx); err != nil {
		return nil, fmt.Errorf("insert provider: %w", err)
	}
	return &provider, nil
}
