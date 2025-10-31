package handlers

import (
	"context"

	"github.com/berkecengiz/appointment-service-boilerplate/internal/models"
)

// AppointmentService defines the interface for appointment business logic.
type AppointmentService interface {
	ListAppointments(ctx context.Context, filter models.AppointmentFilter) ([]models.Appointment, int, error)
	GetAppointmentByID(ctx context.Context, id string) (*models.Appointment, error)
	CreateAppointment(ctx context.Context, req models.CreateAppointmentRequest) (*models.Appointment, error)
	UpdateAppointment(ctx context.Context, id string, req models.UpdateAppointmentRequest) (*models.Appointment, error)
}

// ClientService defines the interface for client business logic.
type ClientService interface {
	ListClients(ctx context.Context) ([]models.Client, error)
	GetClientByID(ctx context.Context, id string) (*models.Client, error)
	CreateClient(ctx context.Context, req models.CreateClientRequest) (*models.Client, error)
}

// ProviderService defines the interface for provider business logic.
type ProviderService interface {
	ListProviders(ctx context.Context) ([]models.Provider, error)
	GetProviderByID(ctx context.Context, id string) (*models.Provider, error)
	CreateProvider(ctx context.Context, req models.CreateProviderRequest) (*models.Provider, error)
}

// DatabasePinger defines the interface for database health checks.
type DatabasePinger interface {
	PingContext(ctx context.Context) error
}
