package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/config"
	"github.com/blagoySimandov/ampledata/go/internal/user"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/webhook"
)

type Billing struct {
	sc                      *stripe.Client
	enrichmentCostMeterName string
	webhookSecret           string
	userRepo                user.Repository
	meterID                 string
}

func NewBilling(userRepo user.Repository) *Billing {
	cfg := config.Load()
	sc := stripe.NewClient(cfg.StripeSecretKey)
	return &Billing{
		sc:                      sc,
		enrichmentCostMeterName: cfg.EnrichmentCostMeterName,
		webhookSecret:           cfg.StripeWebhookSecret,
		userRepo:                userRepo,
	}
}

func (b *Billing) ReportUsage(ctx context.Context, stripeCustomerID string, credits int) error {
	if credits <= 0 {
		return nil
	}

	if stripeCustomerID != "" && b.userRepo != nil {
		if err := b.userRepo.IncrementTokensUsed(ctx, stripeCustomerID, int64(credits)); err != nil {
			return fmt.Errorf("failed to increment tokens used: %w", err)
		}
	}

	if stripeCustomerID == "" {
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

func (b *Billing) CreateSubscriptionCheckout(ctx context.Context, customerID, tierID, successURL, cancelURL string) (*stripe.CheckoutSession, error) {
	tier := GetTier(tierID)
	if tier == nil {
		return nil, fmt.Errorf("unknown tier: %s", tierID)
	}

	if tier.BasePriceID == "" || tier.MeteredPriceID == "" {
		return nil, fmt.Errorf("tier %s not synced with Stripe (missing price IDs)", tierID)
	}

	params := &stripe.CheckoutSessionCreateParams{
		Customer:           stripe.String(customerID),
		PaymentMethodTypes: []*string{stripe.String("card")},
		LineItems: []*stripe.CheckoutSessionCreateLineItemParams{
			{
				Price:    stripe.String(tier.BasePriceID),
				Quantity: stripe.Int64(1),
			},
			{
				Price: stripe.String(tier.MeteredPriceID),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		SubscriptionData: &stripe.CheckoutSessionCreateSubscriptionDataParams{
			Metadata: map[string]string{
				"tier_id": tierID,
			},
		},
	}
	return b.sc.V1CheckoutSessions.Create(ctx, params)
}

func (b *Billing) GetSubscription(ctx context.Context, subscriptionID string) (*stripe.Subscription, error) {
	return b.sc.V1Subscriptions.Retrieve(ctx, subscriptionID, nil)
}

func (b *Billing) CancelSubscription(ctx context.Context, subscriptionID string) (*stripe.Subscription, error) {
	return b.sc.V1Subscriptions.Cancel(ctx, subscriptionID, nil)
}

func (b *Billing) CreateCreditGrant(ctx context.Context, customerID string, amountCents int64, idempotencyKey string) (*stripe.BillingCreditGrant, error) {
	params := &stripe.BillingCreditGrantCreateParams{
		Customer: stripe.String(customerID),
		Name:     stripe.String("Subscription Credits"),
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
	if idempotencyKey != "" {
		params.SetIdempotencyKey(idempotencyKey)
	}
	return b.sc.V1BillingCreditGrants.Create(ctx, params)
}

func (b *Billing) VerifyWebhookSignature(payload []byte, signature string) (*stripe.Event, error) {
	event, err := webhook.ConstructEvent(payload, signature, b.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("webhook signature verification failed: %w", err)
	}
	return &event, nil
}
