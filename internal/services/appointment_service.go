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

// AppointmentService provides business logic for appointment operations.
type AppointmentService struct {
	db *bun.DB
}

// ErrAppointmentConflict indicates the requested slot overlaps an existing appointment.
var ErrAppointmentConflict = errors.New("appointment overlaps with existing booking")

// NewAppointmentService creates a new appointment service with the given database connection.
func NewAppointmentService(db *bun.DB) *AppointmentService {
	return &AppointmentService{db: db}
}

// ListAppointments retrieves appointments from the database with optional filtering.
// Filters can be applied by date, customer ID, provider ID, and branch.
func (s *AppointmentService) ListAppointments(ctx context.Context, f models.AppointmentFilter) ([]models.Appointment, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Build query with proper parameterization to prevent SQL injection
	var list []models.Appointment

	query := s.db.NewSelect().Model(&list)

	if f.ClientID != "" {
		query = query.Where("clientid = ?", f.ClientID)
	}
	if f.ProviderID != "" {
		query = query.Where("providerid = ?", f.ProviderID)
	}
	if f.Branch != "" {
		query = query.Where("branch = ?", f.Branch)
	}
	if f.Date != "" {
		// Expecting YYYY-MM-DD - validate format
		start, err := time.Parse("2006-01-02", f.Date)
		if err == nil {
			end := start.Add(24 * time.Hour)
			query = query.Where("starttime >= ?", start).Where("starttime < ?", end)
		}
	}

	query = query.OrderExpr("starttime ASC")

	if err := query.Scan(ctx); err != nil {
		return nil, fmt.Errorf("query appointments: %w", err)
	}

	return list, nil
}

// GetAppointmentByID retrieves a single appointment by its ID.
// Returns nil if the appointment is not found.
func (s *AppointmentService) GetAppointmentByID(ctx context.Context, id string) (*models.Appointment, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var a models.Appointment
	err := s.db.NewSelect().Model(&a).Where("id = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query appointment by id: %w", err)
	}
	return &a, nil
}

// CreateAppointment creates a new appointment in the database.
// The appointment is created with status 'scheduled'.
func (s *AppointmentService) CreateAppointment(ctx context.Context, req models.CreateAppointmentRequest) (*models.Appointment, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	var appointment models.Appointment

	err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		exists, err := tx.NewSelect().Model((*models.Appointment)(nil)).
			Where("providerid = ?", req.ProviderID).
			Where("status <> 'cancelled'").
			Where("endtime > ?", req.StartTime).
			Where("starttime < ?", req.EndTime).
			Exists(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				exists = false
			} else {
				return fmt.Errorf("check provider availability: %w", err)
			}
		}
		if exists {
			return ErrAppointmentConflict
		}

		appointment = models.Appointment{
			ClientID:   req.ClientID,
			ProviderID: req.ProviderID,
			Branch:     req.Branch,
			StartTime:  req.StartTime,
			EndTime:    req.EndTime,
			Status:     "scheduled",
			Notes:      req.Notes,
		}

		if appointment.ID == "" {
			appointment.ID = uuid.NewString()
		}

		if _, err := tx.NewInsert().Model(&appointment).Returning("").Exec(ctx); err != nil {
			return fmt.Errorf("insert appointment: %w", err)
		}
		return nil
	})
	if errors.Is(err, ErrAppointmentConflict) {
		return nil, ErrAppointmentConflict
	}
	if err != nil {
		return nil, err
	}

	return &appointment, nil
}
