package handlers

import (
	"context"

	"github.com/berkecengiz/appointment-service-boilerplate/internal/models"
)

// AppointmentService defines the interface for appointment business logic.
type AppointmentService interface {
	ListAppointments(ctx context.Context, filter models.AppointmentFilter) ([]models.Appointment, error)
	GetAppointmentByID(ctx context.Context, id string) (*models.Appointment, error)
	CreateAppointment(ctx context.Context, req models.CreateAppointmentRequest) (*models.Appointment, error)
}

// DatabasePinger defines the interface for database health checks.
type DatabasePinger interface {
	PingContext(ctx context.Context) error
}
