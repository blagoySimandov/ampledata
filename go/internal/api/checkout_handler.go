package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/billing"
	"github.com/blagoySimandov/ampledata/go/internal/config"
	"github.com/blagoySimandov/ampledata/go/internal/user"
	"github.com/stripe/stripe-go/v84"
)

func (s *Server) HandleStripeWebhook(ctx context.Context, req HandleStripeWebhookRequestObject) (HandleStripeWebhookResponseObject, error) {
	payload, err := io.ReadAll(req.Body)
	if err != nil {
		return HandleStripeWebhook400JSONResponse{Message: "Failed to read body"}, nil
	}
	event, err := s.billing.VerifyWebhookSignature(payload, req.Params.StripeSignature)
	if err != nil {
		return HandleStripeWebhook401JSONResponse{Message: "Invalid signature"}, nil
	}
	if err := s.handleWebhookEvent(ctx, event); err != nil {
		log.Printf("Webhook %s handling failed: %v", event.Type, err)
		return HandleStripeWebhook500JSONResponse{Message: "Webhook handling failed"}, nil
	}
	return HandleStripeWebhook200Response{}, nil
}

func (s *Server) handleWebhookEvent(ctx context.Context, event *stripe.Event) error {
	switch event.Type {
	case "checkout.session.completed":
		return s.handleCheckoutCompleted(ctx, event)
	case "invoice.paid":
		return s.handleInvoicePaid(ctx, event)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(ctx, event)
	case "customer.subscription.updated":
		return s.handleSubscriptionUpdated(ctx, event)
	}
	return nil
}

func (s *Server) CreateSubscriptionCheckout(ctx context.Context, req CreateSubscriptionCheckoutRequestObject) (CreateSubscriptionCheckoutResponseObject, error) {
	dbUser, ok := user.GetDBUserFromContext(ctx)
	if !ok || dbUser.StripeCustomerID == nil {
		return CreateSubscriptionCheckout400JSONResponse{Message: "User not found or missing Stripe customer"}, nil
	}
	if resp := validateSubscriptionRequest(req.Body); resp != nil {
		return resp, nil
	}
	session, err := s.billing.CreateSubscriptionCheckout(ctx, *dbUser.StripeCustomerID, req.Body.TierId, req.Body.SuccessUrl, req.Body.CancelUrl)
	if err != nil {
		log.Printf("Failed to create subscription checkout: %v", err)
		return CreateSubscriptionCheckout500JSONResponse{Message: "Failed to create checkout session"}, nil
	}
	return CreateSubscriptionCheckout200JSONResponse{CheckoutUrl: session.URL, SessionId: session.ID}, nil
}

func validateSubscriptionRequest(body *CreateSubscriptionCheckoutJSONRequestBody) CreateSubscriptionCheckoutResponseObject {
	if body.TierId == "" {
		return CreateSubscriptionCheckout400JSONResponse{Message: "tier_id is required"}
	}
	if billing.GetTier(body.TierId) == nil {
		return CreateSubscriptionCheckout400JSONResponse{Message: "Invalid tier_id"}
	}
	if body.SuccessUrl == "" || body.CancelUrl == "" {
		return CreateSubscriptionCheckout400JSONResponse{Message: "success_url and cancel_url are required"}
	}
	return nil
}

func (s *Server) CreatePortalSession(ctx context.Context, req CreatePortalSessionRequestObject) (CreatePortalSessionResponseObject, error) {
	dbUser, ok := user.GetDBUserFromContext(ctx)
	if !ok || dbUser.StripeCustomerID == nil {
		return CreatePortalSession400JSONResponse{Message: "User not found or missing Stripe customer"}, nil
	}
	returnURL := req.Params.ReturnUrl
	session, err := s.billing.CreatePortalSession(ctx, *dbUser.StripeCustomerID, returnURL)
	if err != nil {
		log.Printf("Failed to create portal session for customer %s: %v", *dbUser.StripeCustomerID, err)
		return CreatePortalSession500JSONResponse{Message: "Failed to create portal session"}, nil
	}
	return CreatePortalSession200JSONResponse{Url: session.URL}, nil
}

func (s *Server) GetSubscriptionStatus(ctx context.Context, req GetSubscriptionStatusRequestObject) (GetSubscriptionStatusResponseObject, error) {
	dbUser, ok := user.GetDBUserFromContext(ctx)
	if !ok {
		return GetSubscriptionStatus400JSONResponse{Message: "User not found"}, nil
	}
	return GetSubscriptionStatus200JSONResponse{
		Tier:               dbUser.SubscriptionTier,
		TokensUsed:         dbUser.TokensUsed,
		TokensIncluded:     dbUser.TokensIncluded,
		CancelAtPeriodEnd:  dbUser.CancelAtPeriodEnd,
		CurrentPeriodStart: dbUser.CurrentPeriodStart,
		CurrentPeriodEnd:   dbUser.CurrentPeriodEnd,
	}, nil
}

func (s *Server) CancelSubscription(ctx context.Context, req CancelSubscriptionRequestObject) (CancelSubscriptionResponseObject, error) {
	dbUser, ok := user.GetDBUserFromContext(ctx)
	if !ok || dbUser.StripeSubscriptionID == nil || dbUser.StripeCustomerID == nil {
		return CancelSubscription400JSONResponse{Message: "No active subscription found"}, nil
	}
	if _, err := s.billing.CancelSubscription(ctx, *dbUser.StripeSubscriptionID); err != nil {
		log.Printf("Failed to schedule subscription cancellation: %v", err)
		return CancelSubscription500JSONResponse{Message: "Failed to cancel subscription"}, nil
	}
	if err := s.userRepo.SetCancelAtPeriodEnd(ctx, *dbUser.StripeCustomerID, true); err != nil {
		log.Printf("Failed to set cancel_at_period_end for customer %s: %v", *dbUser.StripeCustomerID, err)
	}
	return CancelSubscription200JSONResponse{Message: "Subscription will be cancelled at the end of the billing period"}, nil
}

func (s *Server) ListTiers(ctx context.Context, req ListTiersRequestObject) (ListTiersResponseObject, error) {
	tiers := make(ListTiers200JSONResponse, 0, len(billing.TierOrder))
	for _, id := range billing.TierOrder {
		t := billing.Tiers[id]
		tiers = append(tiers, TierResponse{
			Id:                       t.ID,
			DisplayName:              t.DisplayName,
			MonthlyPriceCents:        t.MonthlyPriceCents,
			IncludedTokens:           t.IncludedTokens,
			OveragePriceCentsDecimal: t.OveragePriceCentsDecimal,
		})
	}
	return tiers, nil
}

func (s *Server) handleCheckoutCompleted(ctx context.Context, event *stripe.Event) error {
	session, err := parseEventData[checkoutSession](event)
	if err != nil {
		return fmt.Errorf("failed to parse checkout session: %w", err)
	}
	if session.Subscription == "" {
		return nil
	}
	return s.applySubscription(ctx, session, event.ID)
}

func (s *Server) applySubscription(ctx context.Context, session *checkoutSession, eventID string) error {
	sub, tier, err := s.getSubscriptionAndTier(ctx, session.Subscription, session.ID)
	if err != nil {
		return err
	}
	periodStart, periodEnd, err := subscriptionPeriod(sub)
	if err != nil {
		return fmt.Errorf("subscription %s: %w", session.Subscription, err)
	}
	if err := s.userRepo.UpdateSubscription(ctx, session.Customer, tier.ID, session.Subscription, tier.IncludedTokens, periodStart, periodEnd); err != nil {
		return fmt.Errorf("failed to update subscription for customer %s: %w", session.Customer, err)
	}
	s.issueCreditGrant(ctx, session.Customer, tier, eventID)
	log.Printf("Subscription created for customer %s: tier=%s", session.Customer, tier.ID)
	return nil
}

func (s *Server) getSubscriptionAndTier(ctx context.Context, subscriptionID, sessionID string) (*stripe.Subscription, *billing.SubscriptionTier, error) {
	sub, err := s.billing.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve subscription %s: %w", subscriptionID, err)
	}
	tierID := sub.Metadata[config.StripeTierIDKey]
	if tierID == "" {
		return nil, nil, fmt.Errorf("no tier_id in subscription %s metadata", subscriptionID)
	}
	tier := billing.GetTier(tierID)
	if tier == nil {
		return nil, nil, fmt.Errorf("unknown tier %s in checkout session %s", tierID, sessionID)
	}
	return sub, tier, nil
}

func (s *Server) handleInvoicePaid(ctx context.Context, event *stripe.Event) error {
	invoice, err := parseEventData[invoiceEvent](event)
	if err != nil {
		return fmt.Errorf("failed to parse invoice: %w", err)
	}
	if invoice.Subscription == "" || invoice.BillingReason == "subscription_create" {
		return nil
	}
	return s.applyBillingCycleReset(ctx, invoice, event.ID)
}

func (s *Server) applyBillingCycleReset(ctx context.Context, invoice *invoiceEvent, eventID string) error {
	sub, err := s.billing.GetSubscription(ctx, invoice.Subscription)
	if err != nil {
		return fmt.Errorf("failed to retrieve subscription %s: %w", invoice.Subscription, err)
	}
	tier, err := s.lookupUserTier(ctx, invoice.Customer)
	if err != nil {
		return err
	}
	periodStart, periodEnd, err := subscriptionPeriod(sub)
	if err != nil {
		return fmt.Errorf("subscription %s: %w", invoice.Subscription, err)
	}
	if err := s.userRepo.ResetBillingCycle(ctx, invoice.Customer, periodStart, periodEnd); err != nil {
		return fmt.Errorf("failed to reset billing cycle for customer %s: %w", invoice.Customer, err)
	}
	s.issueCreditGrant(ctx, invoice.Customer, tier, eventID)
	return nil
}

func (s *Server) handleSubscriptionUpdated(ctx context.Context, event *stripe.Event) error {
	sub, err := parseEventData[subscriptionUpdatedEvent](event)
	if err != nil {
		return fmt.Errorf("failed to parse subscription updated: %w", err)
	}
	if err := s.userRepo.SetCancelAtPeriodEnd(ctx, sub.Customer, sub.isCancellationScheduled()); err != nil {
		return fmt.Errorf("failed to sync cancel_at_period_end for customer %s: %w", sub.Customer, err)
	}
	return nil
}

func (s *Server) handleSubscriptionDeleted(ctx context.Context, event *stripe.Event) error {
	sub, err := parseEventData[subscriptionEvent](event)
	if err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}
	if err := s.userRepo.ClearSubscription(ctx, sub.Customer); err != nil {
		return fmt.Errorf("failed to clear subscription for customer %s: %w", sub.Customer, err)
	}
	log.Printf("Subscription %s deleted for customer %s", sub.ID, sub.Customer)
	return nil
}

func (s *Server) lookupUserTier(ctx context.Context, stripeCustomerID string) (*billing.SubscriptionTier, error) {
	usr, err := s.userRepo.GetByStripeCustomerID(ctx, stripeCustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user for customer %s: %w", stripeCustomerID, err)
	}
	if usr.SubscriptionTier == nil {
		return nil, fmt.Errorf("user for customer %s has no subscription tier", stripeCustomerID)
	}
	tier := billing.GetTier(*usr.SubscriptionTier)
	if tier == nil {
		return nil, fmt.Errorf("unknown tier %s for customer %s", *usr.SubscriptionTier, stripeCustomerID)
	}
	return tier, nil
}

func (s *Server) issueCreditGrant(ctx context.Context, customerID string, tier *billing.SubscriptionTier, eventID string) {
	amount := creditGrantCentsForTier(tier)
	if amount <= 0 {
		return
	}
	idempotencyKey := fmt.Sprintf("credit_grant_%s", eventID)
	if _, err := s.billing.CreateCreditGrant(ctx, customerID, amount, idempotencyKey); err != nil {
		log.Printf("Failed to create credit grant for customer %s: %v", customerID, err)
	}
}

func subscriptionPeriod(sub *stripe.Subscription) (time.Time, time.Time, error) {
	if sub.Items == nil || len(sub.Items.Data) == 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("no subscription items found")
	}
	item := sub.Items.Data[0]
	return time.Unix(item.CurrentPeriodStart, 0), time.Unix(item.CurrentPeriodEnd, 0), nil
}

func creditGrantCentsForTier(tier *billing.SubscriptionTier) int64 {
	overageCents, err := strconv.ParseFloat(tier.OveragePriceCentsDecimal, 64)
	if err != nil {
		return 0
	}
	return int64(math.Round(float64(tier.IncludedTokens) * overageCents))
}

func parseEventData[T any](event *stripe.Event) (*T, error) {
	var data T
	if err := json.Unmarshal(event.Data.Raw, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

type checkoutSession struct {
	ID           string `json:"id"`
	Customer     string `json:"customer"`
	Subscription string `json:"subscription"`
}

type invoiceEvent struct {
	Customer      string `json:"customer"`
	Subscription  string `json:"subscription"`
	BillingReason string `json:"billing_reason"`
}

type subscriptionEvent struct {
	ID       string `json:"id"`
	Customer string `json:"customer"`
}

type subscriptionUpdatedEvent struct {
	Customer          string `json:"customer"`
	CancelAtPeriodEnd bool   `json:"cancel_at_period_end"`
	CancelAt          int64  `json:"cancel_at"`
}

func (e *subscriptionUpdatedEvent) isCancellationScheduled() bool {
	return e.CancelAtPeriodEnd || e.CancelAt != 0
}
