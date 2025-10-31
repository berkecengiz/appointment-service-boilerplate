package models

import (
	"time"

	"github.com/uptrace/bun"
)

// Appointment status constants
const (
	StatusScheduled  = "scheduled"
	StatusCancelled  = "cancelled"
	StatusCompleted  = "completed"
	StatusNoShow     = "no_show"
	StatusInProgress = "in_progress"
)

// ValidStatuses is a map of all valid appointment statuses
var ValidStatuses = map[string]bool{
	StatusScheduled:  true,
	StatusCancelled:  true,
	StatusCompleted:  true,
	StatusNoShow:     true,
	StatusInProgress: true,
}

// ValidateStatus checks if a status string is valid
func ValidateStatus(s string) bool {
	return ValidStatuses[s]
}

// Appointment captures persisted appointment information.
type Appointment struct {
	bun.BaseModel `bun:"table:appointments" swaggerignore:"true"`

	// ID is the unique identifier of the appointment.
	ID string `bun:"id,pk" json:"id"`
	// ClientID identifies the client associated with the appointment.
	ClientID string `bun:"clientid" json:"client_id"`
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
	// ClientID optionally filters appointments by client identifier.
	ClientID string
	// ProviderID optionally filters appointments by provider identifier.
	ProviderID string
	// Branch optionally filters appointments by branch.
	Branch string
	// Status optionally filters appointments by status.
	Status string
	// StartDate optionally filters appointments starting from this date (YYYY-MM-DD).
	StartDate string
	// EndDate optionally filters appointments before this date (YYYY-MM-DD).
	EndDate string
	// Limit specifies the maximum number of results to return.
	Limit int
	// Offset specifies the number of results to skip.
	Offset int
}

// CreateAppointmentRequest defines the request body for creating an appointment.
type CreateAppointmentRequest struct {
	// ClientID identifies the client this appointment is for.
	ClientID string `json:"client_id"`
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

// UpdateAppointmentRequest defines the request body for updating an appointment.
type UpdateAppointmentRequest struct {
	// StartTime optionally updates when the appointment should begin.
	StartTime *time.Time `json:"start_time,omitempty"`
	// EndTime optionally updates when the appointment should end.
	EndTime *time.Time `json:"end_time,omitempty"`
	// Notes optionally updates additional information about the appointment.
	Notes *string `json:"notes,omitempty"`
	// Status optionally updates the appointment status.
	Status *string `json:"status,omitempty"`
	// ClientID ensures only the owner can update status-related fields.
	ClientID string `json:"client_id"`
}
