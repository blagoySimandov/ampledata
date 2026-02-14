package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/billing"
	"github.com/blagoySimandov/ampledata/go/internal/services"
	"github.com/blagoySimandov/ampledata/go/internal/user"
	"github.com/stripe/stripe-go/v84"
)

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

type CheckoutHandler struct {
	billing  services.BillingService
	userRepo user.Repository
}

func NewCheckoutHandler(billing services.BillingService, userRepo user.Repository) *CheckoutHandler {
	return &CheckoutHandler{billing: billing, userRepo: userRepo}
}

type CreateSubscriptionRequest struct {
	TierID     string `json:"tier_id"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
}

type CreateCheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
	SessionID   string `json:"session_id"`
}

type SubscriptionStatusResponse struct {
	Tier               *string    `json:"tier"`
	TokensUsed         int64      `json:"tokens_used"`
	TokensIncluded     int64      `json:"tokens_included"`
	CurrentPeriodStart *time.Time `json:"current_period_start,omitempty"`
	CurrentPeriodEnd   *time.Time `json:"current_period_end,omitempty"`
}

type TierResponse struct {
	ID                string `json:"id"`
	DisplayName       string `json:"display_name"`
	MonthlyPriceCents int64  `json:"monthly_price_cents"`
	IncludedTokens    int64  `json:"included_tokens"`
	OveragePrice      string `json:"overage_price_cents_decimal"`
}

func (h *CheckoutHandler) CreateSubscriptionCheckout(w http.ResponseWriter, r *http.Request) {
	dbUser, ok := user.GetDBUserFromContext(r.Context())
	if !ok || dbUser.StripeCustomerID == nil {
		http.Error(w, "User not found or missing Stripe customer", http.StatusBadRequest)
		return
	}

	var req CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TierID == "" {
		http.Error(w, "tier_id is required", http.StatusBadRequest)
		return
	}

	if billing.GetTier(req.TierID) == nil {
		http.Error(w, "Invalid tier_id", http.StatusBadRequest)
		return
	}

	if req.SuccessURL == "" || req.CancelURL == "" {
		http.Error(w, "success_url and cancel_url are required", http.StatusBadRequest)
		return
	}

	session, err := h.billing.CreateSubscriptionCheckout(r.Context(), *dbUser.StripeCustomerID, req.TierID, req.SuccessURL, req.CancelURL)
	if err != nil {
		log.Printf("Failed to create subscription checkout: %v", err)
		http.Error(w, "Failed to create checkout session", http.StatusInternalServerError)
		return
	}

	writeJSON(w, CreateCheckoutResponse{
		CheckoutURL: session.URL,
		SessionID:   session.ID,
	})
}

func (h *CheckoutHandler) GetSubscriptionStatus(w http.ResponseWriter, r *http.Request) {
	dbUser, ok := user.GetDBUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	writeJSON(w, SubscriptionStatusResponse{
		Tier:               dbUser.SubscriptionTier,
		TokensUsed:         dbUser.TokensUsed,
		TokensIncluded:     dbUser.TokensIncluded,
		CurrentPeriodStart: dbUser.CurrentPeriodStart,
		CurrentPeriodEnd:   dbUser.CurrentPeriodEnd,
	})
}

func (h *CheckoutHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	dbUser, ok := user.GetDBUserFromContext(r.Context())
	if !ok || dbUser.StripeSubscriptionID == nil {
		http.Error(w, "No active subscription found", http.StatusBadRequest)
		return
	}

	_, err := h.billing.CancelSubscription(r.Context(), *dbUser.StripeSubscriptionID)
	if err != nil {
		log.Printf("Failed to cancel subscription: %v", err)
		http.Error(w, "Failed to cancel subscription", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"message": "Subscription cancelled"})
}

func (h *CheckoutHandler) ListTiers(w http.ResponseWriter, r *http.Request) {
	tiers := make([]TierResponse, 0, len(billing.TierOrder))
	for _, id := range billing.TierOrder {
		t := billing.Tiers[id]
		tiers = append(tiers, TierResponse{
			ID:                t.ID,
			DisplayName:       t.DisplayName,
			MonthlyPriceCents: t.MonthlyPriceCents,
			IncludedTokens:    t.IncludedTokens,
			OveragePrice:      t.OveragePriceCentsDecimal,
		})
	}

	writeJSON(w, tiers)
}

func (h *CheckoutHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read webhook body: %v", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	signature := r.Header.Get("Stripe-Signature")
	event, err := h.billing.VerifyWebhookSignature(payload, signature)
	if err != nil {
		log.Printf("Webhook signature verification failed: %v", err)
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	var handleErr error
	switch event.Type {
	case "checkout.session.completed":
		handleErr = h.handleCheckoutCompleted(r.Context(), event)
	case "invoice.paid":
		handleErr = h.handleInvoicePaid(r.Context(), event)
	case "customer.subscription.deleted":
		handleErr = h.handleSubscriptionDeleted(r.Context(), event)
	}

	if handleErr != nil {
		log.Printf("Webhook %s handling failed: %v", event.Type, handleErr)
		http.Error(w, "Webhook handling failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *CheckoutHandler) handleCheckoutCompleted(ctx context.Context, event *stripe.Event) error {
	session, err := parseEventData[checkoutSession](event)
	if err != nil {
		return fmt.Errorf("failed to parse checkout session: %w", err)
	}

	if session.Subscription == "" {
		return nil
	}

	sub, err := h.billing.GetSubscription(ctx, session.Subscription)
	if err != nil {
		return fmt.Errorf("failed to retrieve subscription %s: %w", session.Subscription, err)
	}

	tierID := sub.Metadata["tier_id"]
	if tierID == "" {
		return fmt.Errorf("no tier_id in subscription %s metadata", session.Subscription)
	}

	tier := billing.GetTier(tierID)
	if tier == nil {
		return fmt.Errorf("unknown tier %s in checkout session %s", tierID, session.ID)
	}

	periodStart, periodEnd, err := subscriptionPeriod(sub)
	if err != nil {
		return fmt.Errorf("subscription %s: %w", session.Subscription, err)
	}

	if err := h.userRepo.UpdateSubscription(ctx, session.Customer, tierID, session.Subscription, tier.IncludedTokens, periodStart, periodEnd); err != nil {
		return fmt.Errorf("failed to update subscription for customer %s: %w", session.Customer, err)
	}

	h.issueCreditGrant(ctx, session.Customer, tier, event.ID)

	log.Printf("Subscription created for customer %s: tier=%s, subscription=%s", session.Customer, tierID, session.Subscription)
	return nil
}

func (h *CheckoutHandler) handleInvoicePaid(ctx context.Context, event *stripe.Event) error {
	invoice, err := parseEventData[invoiceEvent](event)
	if err != nil {
		return fmt.Errorf("failed to parse invoice: %w", err)
	}

	if invoice.Subscription == "" || invoice.BillingReason == "subscription_create" {
		return nil
	}

	sub, err := h.billing.GetSubscription(ctx, invoice.Subscription)
	if err != nil {
		return fmt.Errorf("failed to retrieve subscription %s: %w", invoice.Subscription, err)
	}

	tier, err := h.lookupUserTier(ctx, invoice.Customer)
	if err != nil {
		return err
	}

	periodStart, periodEnd, err := subscriptionPeriod(sub)
	if err != nil {
		return fmt.Errorf("subscription %s: %w", invoice.Subscription, err)
	}

	if err := h.userRepo.ResetBillingCycle(ctx, invoice.Customer, periodStart, periodEnd); err != nil {
		return fmt.Errorf("failed to reset billing cycle for customer %s: %w", invoice.Customer, err)
	}

	h.issueCreditGrant(ctx, invoice.Customer, tier, event.ID)

	log.Printf("Billing cycle reset for customer %s: period %s to %s", invoice.Customer, periodStart, periodEnd)
	return nil
}

func (h *CheckoutHandler) handleSubscriptionDeleted(ctx context.Context, event *stripe.Event) error {
	sub, err := parseEventData[subscriptionEvent](event)
	if err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	if err := h.userRepo.ClearSubscription(ctx, sub.Customer); err != nil {
		return fmt.Errorf("failed to clear subscription for customer %s: %w", sub.Customer, err)
	}

	log.Printf("Subscription %s deleted for customer %s", sub.ID, sub.Customer)
	return nil
}

func (h *CheckoutHandler) lookupUserTier(ctx context.Context, stripeCustomerID string) (*billing.SubscriptionTier, error) {
	usr, err := h.userRepo.GetByStripeCustomerID(ctx, stripeCustomerID)
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

func (h *CheckoutHandler) issueCreditGrant(ctx context.Context, customerID string, tier *billing.SubscriptionTier, eventID string) {
	amount := creditGrantCentsForTier(tier)
	if amount <= 0 {
		return
	}
	idempotencyKey := fmt.Sprintf("credit_grant_%s", eventID)
	if _, err := h.billing.CreateCreditGrant(ctx, customerID, amount, idempotencyKey); err != nil {
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
