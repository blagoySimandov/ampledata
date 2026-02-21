package billing

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/blagoySimandov/ampledata/go/internal/config"
	"github.com/stripe/stripe-go/v84"
)

func (b *Billing) SyncStripeCatalog(ctx context.Context) error {
	meterID, err := b.ensureMeter(ctx)
	if err != nil {
		return fmt.Errorf("failed to ensure meter: %w", err)
	}
	b.meterID = meterID
	log.Printf("Stripe meter synced: %s", meterID)

	products, err := b.listActiveProducts(ctx)
	if err != nil {
		return fmt.Errorf("failed to list products: %w", err)
	}

	prices, err := b.listActivePrices(ctx)
	if err != nil {
		return fmt.Errorf("failed to list prices: %w", err)
	}

	for _, tierID := range TierOrder {
		tier := Tiers[tierID]
		if err := b.syncTier(ctx, tier, meterID, products, prices); err != nil {
			return fmt.Errorf("failed to sync tier %s: %w", tierID, err)
		}
		log.Printf("Stripe tier synced: %s (base_product=%s, metered_product=%s, base_price=%s, metered_price=%s)",
			tierID, tier.BaseProductID, tier.MeteredProductID, tier.BasePriceID, tier.MeteredPriceID)
	}

	return nil
}

func (b *Billing) listActiveProducts(ctx context.Context) ([]*stripe.Product, error) {
	var products []*stripe.Product
	for p, err := range b.sc.V1Products.List(ctx, &stripe.ProductListParams{Active: stripe.Bool(true)}) {
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (b *Billing) listActivePrices(ctx context.Context) ([]*stripe.Price, error) {
	var prices []*stripe.Price
	for p, err := range b.sc.V1Prices.List(ctx, &stripe.PriceListParams{Active: stripe.Bool(true)}) {
		if err != nil {
			return nil, err
		}
		prices = append(prices, p)
	}
	return prices, nil
}

func findProduct(products []*stripe.Product, tierID, productType string) string {
	for _, p := range products {
		if p.Metadata[config.StripeMetadataTier] == tierID && p.Metadata[config.StripeMetadataProductType] == productType {
			return p.ID
		}
	}
	return ""
}

func findPrice(prices []*stripe.Price, productID, priceType string) string {
	for _, p := range prices {
		if p.Product != nil && p.Product.ID == productID && p.Metadata[config.StripeMetadataPriceType] == priceType {
			return p.ID
		}
	}
	return ""
}

func (b *Billing) ensureMeter(ctx context.Context) (string, error) {
	for meter, err := range b.sc.V1BillingMeters.List(ctx, &stripe.BillingMeterListParams{}) {
		if err != nil {
			return "", fmt.Errorf("failed to list meters: %w", err)
		}
		if meter.EventName == b.enrichmentCostMeterName {
			return meter.ID, nil
		}
	}
	return b.createMeter(ctx)
}

func (b *Billing) createMeter(ctx context.Context) (string, error) {
	params := &stripe.BillingMeterCreateParams{
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
	meter, err := b.sc.V1BillingMeters.Create(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to create meter: %w", err)
	}
	return meter.ID, nil
}

func (b *Billing) syncTier(ctx context.Context, tier *SubscriptionTier, meterID string, products []*stripe.Product, prices []*stripe.Price) error {
	baseProductID, err := b.ensureProduct(ctx, tier, config.StripePriceTypeBase, products)
	if err != nil {
		return err
	}
	tier.BaseProductID = baseProductID

	meteredProductID, err := b.ensureProduct(ctx, tier, config.StripePriceTypeMetered, products)
	if err != nil {
		return err
	}
	tier.MeteredProductID = meteredProductID

	basePriceID, err := b.ensureBasePrice(ctx, tier, baseProductID, prices)
	if err != nil {
		return err
	}
	tier.BasePriceID = basePriceID

	meteredPriceID, err := b.ensureMeteredPrice(ctx, tier, meteredProductID, meterID, prices)
	if err != nil {
		return err
	}
	tier.MeteredPriceID = meteredPriceID

	return nil
}

func (b *Billing) ensureProduct(ctx context.Context, tier *SubscriptionTier, productType string, products []*stripe.Product) (string, error) {
	if id := findProduct(products, tier.ID, productType); id != "" {
		return id, nil
	}
	return b.createProduct(ctx, tier, productType)
}

func (b *Billing) createProduct(ctx context.Context, tier *SubscriptionTier, productType string) (string, error) {
	name, description := productNameAndDescription(tier, productType)
	params := &stripe.ProductCreateParams{
		Name:        stripe.String(name),
		Description: stripe.String(description),
		Metadata: map[string]string{
			config.StripeMetadataTier:        tier.ID,
			config.StripeMetadataProductType: productType,
		},
	}
	product, err := b.sc.V1Products.Create(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to create product: %w", err)
	}
	return product.ID, nil
}

func productNameAndDescription(tier *SubscriptionTier, productType string) (string, string) {
	if productType == config.StripePriceTypeBase {
		return fmt.Sprintf("AmpleData %s", tier.DisplayName),
			fmt.Sprintf("Includes %d enrichment credits per month", tier.IncludedTokens)
	}
	return fmt.Sprintf("AmpleData %s — Overage", tier.DisplayName),
		"Usage-based billing for credits beyond your included allowance"
}

func (b *Billing) ensureBasePrice(ctx context.Context, tier *SubscriptionTier, productID string, prices []*stripe.Price) (string, error) {
	if id := findPrice(prices, productID, config.StripePriceTypeBase); id != "" {
		return id, nil
	}
	return b.createBasePrice(ctx, tier, productID)
}

func (b *Billing) createBasePrice(ctx context.Context, tier *SubscriptionTier, productID string) (string, error) {
	params := &stripe.PriceCreateParams{
		Product:    stripe.String(productID),
		Currency:   stripe.String(string(stripe.CurrencyUSD)),
		UnitAmount: stripe.Int64(tier.MonthlyPriceCents),
		Recurring: &stripe.PriceCreateRecurringParams{
			Interval: stripe.String(string(stripe.PriceRecurringIntervalMonth)),
		},
		Metadata: map[string]string{
			config.StripeMetadataPriceType: config.StripePriceTypeBase,
			config.StripeMetadataTier:      tier.ID,
		},
	}
	price, err := b.sc.V1Prices.Create(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to create base price: %w", err)
	}
	return price.ID, nil
}

func (b *Billing) ensureMeteredPrice(ctx context.Context, tier *SubscriptionTier, productID, meterID string, prices []*stripe.Price) (string, error) {
	if id := findPrice(prices, productID, config.StripePriceTypeMetered); id != "" {
		return id, nil
	}
	return b.createMeteredPrice(ctx, tier, productID, meterID)
}

func (b *Billing) createMeteredPrice(ctx context.Context, tier *SubscriptionTier, productID, meterID string) (string, error) {
	overageDecimal, err := strconv.ParseFloat(tier.OveragePriceCentsDecimal, 64)
	if err != nil {
		return "", fmt.Errorf("invalid overage price decimal %q: %w", tier.OveragePriceCentsDecimal, err)
	}

	params := &stripe.PriceCreateParams{
		Product:           stripe.String(productID),
		Currency:          stripe.String(string(stripe.CurrencyUSD)),
		UnitAmountDecimal: stripe.Float64(overageDecimal),
		Recurring: &stripe.PriceCreateRecurringParams{
			Interval:  stripe.String(string(stripe.PriceRecurringIntervalMonth)),
			UsageType: stripe.String(string(stripe.PriceRecurringUsageTypeMetered)),
			Meter:     stripe.String(meterID),
		},
		Metadata: map[string]string{
			config.StripeMetadataPriceType: config.StripePriceTypeMetered,
			config.StripeMetadataTier:      tier.ID,
		},
	}
	price, err := b.sc.V1Prices.Create(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to create metered price: %w", err)
	}
	return price.ID, nil
}
