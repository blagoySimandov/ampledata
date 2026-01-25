package services

import (
	"context"

	"github.com/stripe/stripe-go/v84"
)

type BillingService interface {
	ReportUsage(ctx context.Context, stripeCustomerID string, credits int) error
	CreateCustomer(ctx context.Context, userID, email string) (*stripe.Customer, error)
}
