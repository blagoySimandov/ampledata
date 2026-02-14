package services

import (
	"context"

	"github.com/stripe/stripe-go/v84"
)

type BillingService interface {
	ReportUsage(ctx context.Context, stripeCustomerID string, credits int) error
	CreateCustomer(ctx context.Context, userID, email string) (*stripe.Customer, error)
	CreateSubscriptionCheckout(ctx context.Context, customerID, tierID, successURL, cancelURL string) (*stripe.CheckoutSession, error)
	GetSubscription(ctx context.Context, subscriptionID string) (*stripe.Subscription, error)
	CancelSubscription(ctx context.Context, subscriptionID string) (*stripe.Subscription, error)
	CreateCreditGrant(ctx context.Context, customerID string, amountCents int64) (*stripe.BillingCreditGrant, error)
	GetCreditBalance(ctx context.Context, customerID string) (*stripe.BillingCreditBalanceSummary, error)
	VerifyWebhookSignature(payload []byte, signature string) (*stripe.Event, error)
	SyncStripeCatalog(ctx context.Context) error
}
