package models

import "time"

type User struct {
	UserID           string    `json:"user_id"`
	Email            string    `json:"email"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	StripeCustomerID *string   `json:"stripe_customer_id,omitempty"`
	Credits          int       `json:"credits"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
