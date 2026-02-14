package billing

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/stripe/stripe-go/v84"
)

// SyncStripeCatalog idempotently ensures that the Stripe meter, products, and prices
// exist for all subscription tiers. It stores the resulting Stripe IDs back on the
// tier structs so they can be used at runtime.
func (b *Billing) SyncStripeCatalog(ctx context.Context) error {
	meterID, err := b.ensureMeter(ctx)
	if err != nil {
		return fmt.Errorf("failed to ensure meter: %w", err)
	}
	b.meterID = meterID
	log.Printf("Stripe meter synced: %s", meterID)

	for _, tierID := range TierOrder {
		tier := Tiers[tierID]
		if err := b.syncTier(ctx, tier, meterID); err != nil {
			return fmt.Errorf("failed to sync tier %s: %w", tierID, err)
		}
		log.Printf("Stripe tier synced: %s (product=%s, base_price=%s, metered_price=%s)",
			tierID, tier.ProductID, tier.BasePriceID, tier.MeteredPriceID)
	}

	return nil
}

func (b *Billing) ensureMeter(ctx context.Context) (string, error) {
	// List meters and find one matching our event name
	params := &stripe.BillingMeterListParams{}
	iter := b.sc.V1BillingMeters.List(ctx, params)
	for iter.Next() {
		meter := iter.Current()
		if meter.EventName == b.enrichmentCostMeterName {
			return meter.ID, nil
		}
	}
	if err := iter.Err(); err != nil {
		return "", fmt.Errorf("failed to list meters: %w", err)
	}

	// Create meter
	createParams := &stripe.BillingMeterCreateParams{
		DisplayName: stripe.String("Enrichment Credits"),
		EventName:   stripe.String(b.enrichmentCostMeterName),
		DefaultAggregation: &stripe.BillingMeterCreateDefaultAggregationParams{
			Formula: stripe.String(string(stripe.BillingMeterDefaultAggregationFormulaSum)),
		},
		CustomerMapping: &stripe.BillingMeterCreateCustomerMappingParams{
			EventPayloadKey: stripe.String("stripe_customer_id"),
			Type:            stripe.String(string(stripe.BillingMeterCustomerMappingTypeByID)),
		},
		ValueSettings: &stripe.BillingMeterCreateValueSettingsParams{
			EventPayloadKey: stripe.String("value"),
		},
	}
	meter, err := b.sc.V1BillingMeters.Create(ctx, createParams)
	if err != nil {
		return "", fmt.Errorf("failed to create meter: %w", err)
	}
	return meter.ID, nil
}

func (b *Billing) syncTier(ctx context.Context, tier *SubscriptionTier, meterID string) error {
	// Find or create product
	productID, err := b.ensureProduct(ctx, tier)
	if err != nil {
		return err
	}
	tier.ProductID = productID

	// Find or create base price (flat monthly)
	basePriceID, err := b.ensureBasePrice(ctx, tier, productID)
	if err != nil {
		return err
	}
	tier.BasePriceID = basePriceID

	// Find or create metered price
	meteredPriceID, err := b.ensureMeteredPrice(ctx, tier, productID, meterID)
	if err != nil {
		return err
	}
	tier.MeteredPriceID = meteredPriceID

	return nil
}

func (b *Billing) ensureProduct(ctx context.Context, tier *SubscriptionTier) (string, error) {
	params := &stripe.ProductSearchParams{
		SearchParams: stripe.SearchParams{
			Query: fmt.Sprintf("metadata['ampledata_tier']:'%s'", tier.ID),
		},
	}
	iter := b.sc.V1Products.Search(ctx, params)
	for iter.Next() {
		return iter.Current().ID, nil
	}
	if err := iter.Err(); err != nil {
		return "", fmt.Errorf("failed to search products: %w", err)
	}

	// Create product
	createParams := &stripe.ProductCreateParams{
		Name: stripe.String(fmt.Sprintf("AmpleData %s", tier.DisplayName)),
		Metadata: map[string]string{
			"ampledata_tier": tier.ID,
		},
	}
	product, err := b.sc.V1Products.Create(ctx, createParams)
	if err != nil {
		return "", fmt.Errorf("failed to create product: %w", err)
	}
	return product.ID, nil
}

func (b *Billing) ensureBasePrice(ctx context.Context, tier *SubscriptionTier, productID string) (string, error) {
	params := &stripe.PriceSearchParams{
		SearchParams: stripe.SearchParams{
			Query: fmt.Sprintf("product:'%s' AND metadata['ampledata_price_type']:'base'", productID),
		},
	}
	iter := b.sc.V1Prices.Search(ctx, params)
	for iter.Next() {
		p := iter.Current()
		if p.Active {
			return p.ID, nil
		}
	}
	if err := iter.Err(); err != nil {
		return "", fmt.Errorf("failed to search prices: %w", err)
	}

	createParams := &stripe.PriceCreateParams{
		Product:    stripe.String(productID),
		Currency:   stripe.String(string(stripe.CurrencyUSD)),
		UnitAmount: stripe.Int64(tier.MonthlyPriceCents),
		Recurring: &stripe.PriceCreateRecurringParams{
			Interval: stripe.String(string(stripe.PriceRecurringIntervalMonth)),
		},
		Metadata: map[string]string{
			"ampledata_price_type": "base",
			"ampledata_tier":       tier.ID,
		},
	}
	price, err := b.sc.V1Prices.Create(ctx, createParams)
	if err != nil {
		return "", fmt.Errorf("failed to create base price: %w", err)
	}
	return price.ID, nil
}

func (b *Billing) ensureMeteredPrice(ctx context.Context, tier *SubscriptionTier, productID, meterID string) (string, error) {
	params := &stripe.PriceSearchParams{
		SearchParams: stripe.SearchParams{
			Query: fmt.Sprintf("product:'%s' AND metadata['ampledata_price_type']:'metered'", productID),
		},
	}
	iter := b.sc.V1Prices.Search(ctx, params)
	for iter.Next() {
		p := iter.Current()
		if p.Active {
			return p.ID, nil
		}
	}
	if err := iter.Err(); err != nil {
		return "", fmt.Errorf("failed to search metered prices: %w", err)
	}

	overageDecimal, err := strconv.ParseFloat(tier.OveragePriceCentsDecimal, 64)
	if err != nil {
		return "", fmt.Errorf("invalid overage price decimal %q: %w", tier.OveragePriceCentsDecimal, err)
	}

	createParams := &stripe.PriceCreateParams{
		Product:           stripe.String(productID),
		Currency:          stripe.String(string(stripe.CurrencyUSD)),
		UnitAmountDecimal: stripe.Float64(overageDecimal),
		Recurring: &stripe.PriceCreateRecurringParams{
			Interval:  stripe.String(string(stripe.PriceRecurringIntervalMonth)),
			UsageType: stripe.String(string(stripe.PriceRecurringUsageTypeMetered)),
			Meter:     stripe.String(meterID),
		},
		Metadata: map[string]string{
			"ampledata_price_type": "metered",
			"ampledata_tier":       tier.ID,
		},
	}
	price, err := b.sc.V1Prices.Create(ctx, createParams)
	if err != nil {
		return "", fmt.Errorf("failed to create metered price: %w", err)
	}
	return price.ID, nil
}
