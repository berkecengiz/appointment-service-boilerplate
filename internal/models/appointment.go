package models

import (
	"time"

	"github.com/uptrace/bun"
)

// Appointment captures persisted appointment information.
type Appointment struct {
	bun.BaseModel `bun:"table:appointments"`

	// ID is the unique identifier of the appointment.
	ID string `bun:"id,pk" json:"id"`
	// CustomerID identifies the customer associated with the appointment.
	CustomerID string `bun:"customerid" json:"customer_id"`
	// ProviderID identifies the service provider assigned to the appointment.
	ProviderID string `bun:"providerid" json:"provider_id"`
	// Branch denotes the location or branch where the appointment takes place.
	Branch string `bun:"branch" json:"branch"`
	// StartTime marks when the appointment begins.
	StartTime time.Time `bun:"starttime" json:"start_time"`
	// EndTime marks when the appointment is expected to finish.
	EndTime time.Time `bun:"endtime" json:"end_time"`
	// Status represents the current status of the appointment (e.g., scheduled, cancelled).
	Status string `bun:"status" json:"status"`
	// Notes contains optional remarks about the appointment.
	Notes *string `bun:"notes,nullzero" json:"notes,omitempty"`
}

// AppointmentFilter gathers query parameters used when listing appointments.
type AppointmentFilter struct {
	// Date optionally filters appointments by date in YYYY-MM-DD format.
	Date string
	// CustomerID optionally filters appointments by customer identifier.
	CustomerID string
	// ProviderID optionally filters appointments by provider identifier.
	ProviderID string
	// Branch optionally filters appointments by branch.
	Branch string
}

// CreateAppointmentRequest defines the request body for creating an appointment.
type CreateAppointmentRequest struct {
	// CustomerID identifies the customer this appointment is for.
	CustomerID string `json:"customer_id"`
	// ProviderID identifies the provider fulfilling the appointment.
	ProviderID string `json:"provider_id"`
	// Branch specifies where the appointment will happen.
	Branch string `json:"branch"`
	// StartTime indicates when the appointment should begin.
	StartTime time.Time `json:"start_time"`
	// EndTime indicates when the appointment should end.
	EndTime time.Time `json:"end_time"`
	// Notes optionally contains additional information about the appointment.
	Notes *string `json:"notes,omitempty"`
}
