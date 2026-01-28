package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/config"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/webhook"
)

type Billing struct {
	sc                      *stripe.Client
	enrichmentCostMeterName string
	webhookSecret           string
}

func NewBilling() *Billing {
	cfg := config.Load()
	sc := stripe.NewClient(cfg.StripeSecretKey)
	return &Billing{
		sc:                      sc,
		enrichmentCostMeterName: cfg.EnrichmentCostMeterName,
		webhookSecret:           cfg.StripeWebhookSecret,
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

func (b *Billing) CreateCheckoutSession(ctx context.Context, customerID string, amountCents int64, successURL, cancelURL string) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionCreateParams{
		Customer:           stripe.String(customerID),
		PaymentMethodTypes: []*string{stripe.String("card")},
		LineItems: []*stripe.CheckoutSessionCreateLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionCreateLineItemPriceDataParams{
					Currency: stripe.String(string(stripe.CurrencyUSD)),
					ProductData: &stripe.CheckoutSessionCreateLineItemPriceDataProductDataParams{
						Name:        stripe.String("Credit Package"),
						Description: stripe.String(fmt.Sprintf("$%.2f credits", float64(amountCents)/100)),
					},
					UnitAmount: stripe.Int64(amountCents),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata: map[string]string{
			"type":         "credit_purchase",
			"amount_cents": fmt.Sprintf("%d", amountCents),
		},
	}
	return b.sc.V1CheckoutSessions.Create(ctx, params)
}

func (b *Billing) CreateCreditGrant(ctx context.Context, customerID string, amountCents int64) (*stripe.BillingCreditGrant, error) {
	params := &stripe.BillingCreditGrantCreateParams{
		Customer: stripe.String(customerID),
		Name:     stripe.String("Prepaid Credits"),
		Category: stripe.String(string(stripe.BillingCreditGrantCategoryPaid)),
		ApplicabilityConfig: &stripe.BillingCreditGrantCreateApplicabilityConfigParams{
			Scope: &stripe.BillingCreditGrantCreateApplicabilityConfigScopeParams{
				PriceType: stripe.String(string(stripe.BillingCreditGrantApplicabilityConfigScopePriceTypeMetered)),
			},
		},
		Amount: &stripe.BillingCreditGrantCreateAmountParams{
			Type: stripe.String(string(stripe.BillingCreditGrantAmountTypeMonetary)),
			Monetary: &stripe.BillingCreditGrantCreateAmountMonetaryParams{
				Value:    stripe.Int64(amountCents),
				Currency: stripe.String(string(stripe.CurrencyUSD)),
			},
		},
	}
	return b.sc.V1BillingCreditGrants.Create(ctx, params)
}

func (b *Billing) GetCreditBalance(ctx context.Context, customerID string) (*stripe.BillingCreditBalanceSummary, error) {
	params := &stripe.BillingCreditBalanceSummaryRetrieveParams{
		Customer: stripe.String(customerID),
		Filter: &stripe.BillingCreditBalanceSummaryRetrieveFilterParams{
			Type: stripe.String("applicability_scope"),
			ApplicabilityScope: &stripe.BillingCreditBalanceSummaryRetrieveFilterApplicabilityScopeParams{
				PriceType: stripe.String(string(stripe.BillingCreditGrantApplicabilityConfigScopePriceTypeMetered)),
			},
		},
	}
	return b.sc.V1BillingCreditBalanceSummary.Retrieve(ctx, params)
}

func (b *Billing) VerifyWebhookSignature(payload []byte, signature string) (*stripe.Event, error) {
	event, err := webhook.ConstructEvent(payload, signature, b.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("webhook signature verification failed: %w", err)
	}
	return &event, nil
}
