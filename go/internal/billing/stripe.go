package billing

import (
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

func (b *Billing) CreateCustomer(userID string) (*stripe.Customer, error) {
	///
	//retunr customer, err
	return nil, nil
}
