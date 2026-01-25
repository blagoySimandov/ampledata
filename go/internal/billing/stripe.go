package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/config"
	"github.com/stripe/stripe-go/v84"
)

type Billing struct {
	sc                      *stripe.Client
	enrichmentCostMeterName string
}

func NewBilling() *Billing {
	cfg := config.Load()
	sc := stripe.NewClient(cfg.StripeSecretKey)
	return &Billing{
		sc:                      sc,
		enrichmentCostMeterName: cfg.EnrichmentCostMeterName,
	}
}

func (b *Billing) ReportUsage(ctx context.Context, stripeCustomerID string, credits int) error {
	if stripeCustomerID == "" || credits <= 0 {
		return nil
	}

	params := &stripe.BillingMeterEventCreateParams{
		EventName: stripe.String(b.enrichmentCostMeterName),
		Payload: map[string]string{
			"stripe_customer_id": stripeCustomerID,
			"value":              fmt.Sprintf("%d", credits),
		},
		Timestamp: stripe.Int64(time.Now().Unix()),
	}

	_, err := b.sc.V1BillingMeterEvents.Create(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to report meter event: %w", err)
	}

	return nil
}

func (b *Billing) CreateCustomer(ctx context.Context, userID, email string) (*stripe.Customer, error) {
	params := &stripe.CustomerCreateParams{
		Email:    stripe.String(email),
		Metadata: map[string]string{"user_id": userID},
	}
	return b.sc.V1Customers.Create(ctx, params)
}
