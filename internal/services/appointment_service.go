package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/berkecengiz/appointment-service-boilerplate/internal/models"
)

// AppointmentService provides business logic for appointment operations.
type AppointmentService struct {
	db *sql.DB
}

// NewAppointmentService creates a new appointment service with the given database connection.
func NewAppointmentService(db *sql.DB) *AppointmentService {
	return &AppointmentService{db: db}
}

// ListAppointments retrieves appointments from the database with optional filtering.
// Filters can be applied by date, customer ID, provider ID, and branch.
func (s *AppointmentService) ListAppointments(ctx context.Context, f models.AppointmentFilter) ([]models.Appointment, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Build query with proper parameterization to prevent SQL injection
	base := "SELECT Id, CustomerId, ProviderId, Branch, StartTime, EndTime, Status, Notes FROM Appointments"
	var conditions []string
	var args []any
	argIndex := 1

	if f.CustomerID != "" {
		conditions = append(conditions, fmt.Sprintf("CustomerId = $%d", argIndex))
		args = append(args, f.CustomerID)
		argIndex++
	}
	if f.ProviderID != "" {
		conditions = append(conditions, fmt.Sprintf("ProviderId = $%d", argIndex))
		args = append(args, f.ProviderID)
		argIndex++
	}
	if f.Branch != "" {
		conditions = append(conditions, fmt.Sprintf("Branch = $%d", argIndex))
		args = append(args, f.Branch)
		argIndex++
	}
	if f.Date != "" {
		// Expecting YYYY-MM-DD - validate format
		start, err := time.Parse("2006-01-02", f.Date)
		if err == nil {
			end := start.Add(24 * time.Hour)
			conditions = append(conditions, fmt.Sprintf("StartTime >= $%d AND StartTime < $%d", argIndex, argIndex+1))
			args = append(args, start)
			args = append(args, end)
			argIndex += 2
		}
	}

	query := base
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY StartTime ASC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query appointments: %w", err)
	}
	defer rows.Close()

	var list []models.Appointment
	for rows.Next() {
		var a models.Appointment
		var notes sql.NullString
		if err := rows.Scan(&a.ID, &a.CustomerID, &a.ProviderID, &a.Branch, &a.StartTime, &a.EndTime, &a.Status, &notes); err != nil {
			return nil, fmt.Errorf("scan appointment: %w", err)
		}
		if notes.Valid {
			a.Notes = &notes.String
		}
		list = append(list, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate appointments: %w", err)
	}
	return list, nil
}

// GetAppointmentByID retrieves a single appointment by its ID.
// Returns nil if the appointment is not found.
func (s *AppointmentService) GetAppointmentByID(ctx context.Context, id string) (*models.Appointment, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = "SELECT Id, CustomerId, ProviderId, Branch, StartTime, EndTime, Status, Notes FROM Appointments WHERE Id = $1"
	var a models.Appointment
	var notes sql.NullString
	err := s.db.QueryRowContext(ctx, q, id).Scan(&a.ID, &a.CustomerID, &a.ProviderID, &a.Branch, &a.StartTime, &a.EndTime, &a.Status, &notes)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query appointment by id: %w", err)
	}
	if notes.Valid {
		a.Notes = &notes.String
	}
	return &a, nil
}

// CreateAppointment creates a new appointment in the database.
// The appointment is created with status 'scheduled'.
func (s *AppointmentService) CreateAppointment(ctx context.Context, req models.CreateAppointmentRequest) (*models.Appointment, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	const q = `INSERT INTO Appointments (CustomerId, ProviderId, Branch, StartTime, EndTime, Status, Notes)
               VALUES ($1, $2, $3, $4, $5, 'scheduled', $6)
               RETURNING Id`

	var id string
	if err := tx.QueryRowContext(ctx, q,
		req.CustomerID,
		req.ProviderID,
		req.Branch,
		req.StartTime,
		req.EndTime,
		req.Notes,
	).Scan(&id); err != nil {
		return nil, fmt.Errorf("insert appointment: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.Appointment{
		ID:         id,
		CustomerID: req.CustomerID,
		ProviderID: req.ProviderID,
		Branch:     req.Branch,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Status:     "scheduled",
		Notes:      req.Notes,
	}, nil
}
