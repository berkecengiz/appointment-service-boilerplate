package models

import "github.com/uptrace/bun"

// Provider represents a service provider fulfilling appointments.
type Provider struct {
	bun.BaseModel `bun:"table:providers" swaggerignore:"true"`

	// ID is the unique provider identifier.
	ID string `bun:"id,pk" json:"id"`
	// Name is the provider's display name.
	Name string `bun:"name" json:"name"`
	// Email is the provider's contact email.
	Email string `bun:"email" json:"email"`
	// Phone is the provider's contact phone number.
	Phone string `bun:"phone" json:"phone"`
	// Address is the provider's business address.
	Address string `bun:"address" json:"address"`
}

// CreateProviderRequest defines the payload for adding a new provider.
type CreateProviderRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}
