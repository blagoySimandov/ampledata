package services

import (
	"context"

	"github.com/stripe/stripe-go/v84"
)

type BillingService interface {
	ReportUsage(ctx context.Context, stripeCustomerID string, credits int) error
	CreateCustomer(ctx context.Context, userID, email string) (*stripe.Customer, error)
	CreateCheckoutSession(ctx context.Context, customerID string, amountCents int64, successURL, cancelURL string) (*stripe.CheckoutSession, error)
	CreateCreditGrant(ctx context.Context, customerID string, amountCents int64) (*stripe.BillingCreditGrant, error)
	GetCreditBalance(ctx context.Context, customerID string) (*stripe.BillingCreditBalanceSummary, error)
	VerifyWebhookSignature(payload []byte, signature string) (*stripe.Event, error)
}
