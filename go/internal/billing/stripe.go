package billing

import (
	"context"

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

func (b *Billing) AddEnrichmentCost(cost int) error {
	return nil
}

func (b *Billing) CreateCustomer(ctx context.Context, userID, email string) (*stripe.Customer, error) {
	params := &stripe.CustomerCreateParams{
		Email:    stripe.String(email),
		Metadata: map[string]string{"user_id": userID},
	}
	return b.sc.V1Customers.Create(ctx, params)
}
