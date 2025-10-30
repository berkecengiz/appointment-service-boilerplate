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

// ClientService provides business logic for client operations.
type ClientService struct {
	db *bun.DB
}

// NewClientService creates a client service with the provided database connection.
func NewClientService(db *bun.DB) *ClientService {
	return &ClientService{db: db}
}

// ListClients fetches all clients sorted by name.
func (s *ClientService) ListClients(ctx context.Context) ([]models.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var clients []models.Client
	if err := s.db.NewSelect().Model(&clients).OrderExpr("name ASC").Scan(ctx); err != nil {
		return nil, fmt.Errorf("query clients: %w", err)
	}
	return clients, nil
}

// GetClientByID retrieves a client by its identifier.
func (s *ClientService) GetClientByID(ctx context.Context, id string) (*models.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var client models.Client
	if err := s.db.NewSelect().Model(&client).Where("id = ?", id).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query client by id: %w", err)
	}
	return &client, nil
}

// CreateClient inserts a new client record.
func (s *ClientService) CreateClient(ctx context.Context, req models.CreateClientRequest) (*models.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	client := models.Client{
		ID:      uuid.NewString(),
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.Phone,
		Address: req.Address,
	}

	if _, err := s.db.NewInsert().Model(&client).Returning("").Exec(ctx); err != nil {
		return nil, fmt.Errorf("insert client: %w", err)
	}
	return &client, nil
}
