package domain

import "time"

// Product represents an injectable product in the registry.
type Product struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Manufacturer  string    `json:"manufacturer"`
	ProductType   string    `json:"product_type"`
	Concentration *string   `json:"concentration"`
	Active        bool      `json:"active"`
	CreatedAt     time.Time `json:"created_at"`
}
