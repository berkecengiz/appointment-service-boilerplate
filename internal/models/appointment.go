package models

import "time"

type Appointment struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	ProviderID string    `json:"provider_id"`
	Branch     string    `json:"branch"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Status     string    `json:"status"`
	Notes      *string   `json:"notes,omitempty"`
}

type AppointmentFilter struct {
	Date       string
	CustomerID string
	ProviderID string
	Branch     string
}

type CreateAppointmentRequest struct {
	CustomerID string    `json:"customer_id"`
	ProviderID string    `json:"provider_id"`
	Branch     string    `json:"branch"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Notes      *string   `json:"notes,omitempty"`
}
