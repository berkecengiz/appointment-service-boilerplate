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

// Service-level errors
var (
	ErrAppointmentConflict         = errors.New("appointment overlaps with existing booking")
	ErrClientAlreadyBooked         = errors.New("client already has an appointment on this date")
	ErrAppointmentNotFound         = errors.New("appointment not found")
	ErrAppointmentAlreadyCancelled = errors.New("appointment already cancelled")
	ErrAppointmentForbidden        = errors.New("appointment not owned by client")
	ErrClientNotFound              = errors.New("client not found")
	ErrProviderNotFound            = errors.New("provider not found")
	ErrInvalidStatus               = errors.New("invalid appointment status")
	ErrInvalidStatusTransition     = errors.New("invalid status transition")
)

// NewAppointmentService creates a new appointment service with the given database connection.
func NewAppointmentService(db *bun.DB) *AppointmentService {
	return &AppointmentService{db: db}
}

// ListAppointments retrieves appointments from the database with optional filtering.
// Filters can be applied by date, customer ID, provider ID, branch, status, and date range.
// Returns appointments and total count for pagination.
func (s *AppointmentService) ListAppointments(ctx context.Context, f models.AppointmentFilter) ([]models.Appointment, int, error) {
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
	if f.Status != "" {
		query = query.Where("status = ?", f.Status)
	}
	if f.Date != "" {
		// Expecting YYYY-MM-DD - validate format
		start, err := time.Parse("2006-01-02", f.Date)
		if err == nil {
			end := start.Add(24 * time.Hour)
			query = query.Where("starttime >= ?", start).Where("starttime < ?", end)
		}
	}
	// Handle date range filters
	if f.StartDate != "" {
		start, err := time.Parse("2006-01-02", f.StartDate)
		if err == nil {
			query = query.Where("starttime >= ?", start)
		}
	}
	if f.EndDate != "" {
		end, err := time.Parse("2006-01-02", f.EndDate)
		if err == nil {
			query = query.Where("starttime < ?", end)
		}
	}

	// Get total count before applying pagination
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count appointments: %w", err)
	}

	// Apply pagination
	if f.Limit > 0 {
		query = query.Limit(f.Limit)
	}
	if f.Offset > 0 {
		query = query.Offset(f.Offset)
	}

	query = query.OrderExpr("starttime ASC")

	if err := query.Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("query appointments: %w", err)
	}

	return list, total, nil
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
		// Check if client exists
		clientExists, err := tx.NewSelect().Model((*models.Client)(nil)).
			Where("id = ?", req.ClientID).
			Exists(ctx)
		if err != nil {
			return fmt.Errorf("check client existence: %w", err)
		}
		if !clientExists {
			return ErrClientNotFound
		}

		// Check if provider exists
		providerExists, err := tx.NewSelect().Model((*models.Provider)(nil)).
			Where("id = ?", req.ProviderID).
			Exists(ctx)
		if err != nil {
			return fmt.Errorf("check provider existence: %w", err)
		}
		if !providerExists {
			return ErrProviderNotFound
		}

		startOfDay := time.Date(req.StartTime.Year(), req.StartTime.Month(), req.StartTime.Day(), 0, 0, 0, 0, req.StartTime.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)

		clientBooked, err := tx.NewSelect().Model((*models.Appointment)(nil)).
			Where("clientid = ?", req.ClientID).
			Where("status <> 'cancelled'").
			Where("starttime >= ?", startOfDay).
			Where("starttime < ?", endOfDay).
			Exists(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				clientBooked = false
			} else {
				return fmt.Errorf("check client availability: %w", err)
			}
		}
		if clientBooked {
			return ErrClientAlreadyBooked
		}

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
			Status:     models.StatusScheduled,
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
	if errors.Is(err, ErrClientNotFound) {
		return nil, ErrClientNotFound
	}
	if errors.Is(err, ErrProviderNotFound) {
		return nil, ErrProviderNotFound
	}
	if errors.Is(err, ErrAppointmentConflict) {
		return nil, ErrAppointmentConflict
	}
	if errors.Is(err, ErrClientAlreadyBooked) {
		return nil, ErrClientAlreadyBooked
	}
	if err != nil {
		return nil, err
	}

	return &appointment, nil
}

// UpdateAppointment updates an existing appointment's time and/or notes.
func (s *AppointmentService) UpdateAppointment(ctx context.Context, id string, req models.UpdateAppointmentRequest) (*models.Appointment, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	var appointment models.Appointment

	err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Fetch current appointment
		if err := tx.NewSelect().Model(&appointment).Where("id = ?", id).For("UPDATE").Scan(ctx); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrAppointmentNotFound
			}
			return fmt.Errorf("query appointment for update: %w", err)
		}

		// Check ownership
		if appointment.ClientID != req.ClientID {
			return ErrAppointmentForbidden
		}

		// Track what's being updated
		columnsToUpdate := []string{}

		// Handle status update
		if req.Status != nil {
			if !models.ValidateStatus(*req.Status) {
				return ErrInvalidStatus
			}
			// Validate status transitions
			if appointment.Status == models.StatusCancelled && *req.Status != models.StatusCancelled {
				return ErrInvalidStatusTransition
			}
			if appointment.Status == models.StatusCompleted && *req.Status == models.StatusScheduled {
				return ErrInvalidStatusTransition
			}
			appointment.Status = *req.Status
			columnsToUpdate = append(columnsToUpdate, "status")
		}

		// Handle time updates
		newStartTime := appointment.StartTime
		newEndTime := appointment.EndTime
		timesChanged := false

		if req.StartTime != nil {
			newStartTime = *req.StartTime
			timesChanged = true
			columnsToUpdate = append(columnsToUpdate, "starttime")
		}
		if req.EndTime != nil {
			newEndTime = *req.EndTime
			timesChanged = true
			columnsToUpdate = append(columnsToUpdate, "endtime")
		}

		// Validate time range if times are changing
		if timesChanged {
			if !newEndTime.After(newStartTime) {
				return fmt.Errorf("end time must be after start time")
			}

			// Check if the appointment date has changed (either time field could cause this)
			oldDate := time.Date(appointment.StartTime.Year(), appointment.StartTime.Month(), appointment.StartTime.Day(), 0, 0, 0, 0, appointment.StartTime.Location())
			newDate := time.Date(newStartTime.Year(), newStartTime.Month(), newStartTime.Day(), 0, 0, 0, 0, newStartTime.Location())

			// Check if client already has another appointment on the new date if date changed
			if !oldDate.Equal(newDate) {
				startOfDay := newDate
				endOfDay := startOfDay.Add(24 * time.Hour)

				clientBooked, err := tx.NewSelect().Model((*models.Appointment)(nil)).
					Where("clientid = ?", appointment.ClientID).
					Where("id <> ?", id).
					Where("status <> ?", models.StatusCancelled).
					Where("starttime >= ?", startOfDay).
					Where("starttime < ?", endOfDay).
					Exists(ctx)
				if err != nil {
					return fmt.Errorf("check client availability: %w", err)
				}
				if clientBooked {
					return ErrClientAlreadyBooked
				}
			}

			// Check provider availability for new time slot
			exists, err := tx.NewSelect().Model((*models.Appointment)(nil)).
				Where("providerid = ?", appointment.ProviderID).
				Where("id <> ?", id).
				Where("status <> ?", models.StatusCancelled).
				Where("endtime > ?", newStartTime).
				Where("starttime < ?", newEndTime).
				Exists(ctx)
			if err != nil {
				return fmt.Errorf("check provider availability: %w", err)
			}
			if exists {
				return ErrAppointmentConflict
			}

			appointment.StartTime = newStartTime
			appointment.EndTime = newEndTime
		}

		// Handle notes update
		if req.Notes != nil {
			appointment.Notes = req.Notes
			columnsToUpdate = append(columnsToUpdate, "notes")
		}

		// Only update if there are changes
		if len(columnsToUpdate) > 0 {
			if _, err := tx.NewUpdate().Model(&appointment).
				Column(columnsToUpdate...).
				Where("id = ?", id).
				Exec(ctx); err != nil {
				return fmt.Errorf("update appointment: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, ErrAppointmentNotFound) ||
			errors.Is(err, ErrAppointmentForbidden) ||
			errors.Is(err, ErrClientAlreadyBooked) ||
			errors.Is(err, ErrAppointmentConflict) ||
			errors.Is(err, ErrInvalidStatus) ||
			errors.Is(err, ErrInvalidStatusTransition) {
			return nil, err
		}
		return nil, fmt.Errorf("update appointment: %w", err)
	}

	return &appointment, nil
}
