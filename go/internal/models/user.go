package models

import "time"

type User struct {
	ID               string    `json:"id"`
	Email            string    `json:"email"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	StripeCustomerID *string   `json:"stripe_customer_id,omitempty"`
	TokensUsed       int64     `json:"tokens_used"`
	TokensPurchased  int64     `json:"tokens_purchased"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
