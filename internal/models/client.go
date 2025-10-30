package models

import "github.com/uptrace/bun"

// Client represents a customer receiving services.
type Client struct {
	bun.BaseModel `bun:"table:clients" swaggerignore:"true"`

	// ID is the unique client identifier.
	ID string `bun:"id,pk" json:"id"`
	// Name is the client's full name.
	Name string `bun:"name" json:"name"`
	// Email is the client's preferred contact email.
	Email string `bun:"email" json:"email"`
	// Phone is the client's contact phone number.
	Phone string `bun:"phone" json:"phone"`
	// Address is the client's mailing address.
	Address string `bun:"address" json:"address"`
}

// CreateClientRequest defines the payload for adding a new client.
type CreateClientRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}
