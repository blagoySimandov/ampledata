package api

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/billing"
	"github.com/blagoySimandov/ampledata/go/internal/services"
	"github.com/blagoySimandov/ampledata/go/internal/user"
	"github.com/stripe/stripe-go/v84"
)

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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CreateCheckoutResponse{
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SubscriptionStatusResponse{
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Subscription cancelled"})
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tiers)
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

	switch event.Type {
	case "checkout.session.completed":
		h.handleCheckoutCompleted(r.Context(), event, w)
	case "invoice.paid":
		h.handleInvoicePaid(r.Context(), event, w)
	case "customer.subscription.deleted":
		h.handleSubscriptionDeleted(r.Context(), event, w)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *CheckoutHandler) handleCheckoutCompleted(ctx context.Context, event *stripe.Event, w http.ResponseWriter) {
	var session struct {
		ID           string `json:"id"`
		Customer     string `json:"customer"`
		Subscription string `json:"subscription"`
		Metadata     struct {
			TierID string `json:"tier_id"`
		} `json:"metadata"`
	}

	// The metadata is on the subscription_data, which gets copied to the subscription.
	// But checkout.session.completed also has subscription ID. Let's get the tier from
	// the subscription metadata by retrieving the subscription object, or we can put it
	// in session metadata too. For simplicity, we parse the session and then look up
	// subscription metadata from the raw event.

	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		log.Printf("Failed to parse checkout session: %v", err)
		return
	}

	if session.Subscription == "" {
		return // not a subscription checkout
	}

	// Parse subscription metadata from the event - we need to fetch it from subscription_data
	// The tier_id was set in subscription_data.metadata, which is on the subscription object
	// We'll extract it from the session display_items or we use a simpler approach:
	// re-parse to get nested subscription data
	var fullSession struct {
		Customer     string `json:"customer"`
		Subscription string `json:"subscription"`
	}
	json.Unmarshal(event.Data.Raw, &fullSession)

	// We need to find the tier. Since the metadata was put on subscription_data,
	// it ends up on the subscription. We need to look at the subscription's metadata.
	// Unfortunately, the checkout.session.completed event only has the subscription ID.
	// Let's parse the raw event more carefully for subscription metadata.
	var rawMap map[string]interface{}
	json.Unmarshal(event.Data.Raw, &rawMap)

	tierID := ""
	// Try to get tier from subscription metadata (if expanded in the event)
	if subData, ok := rawMap["subscription_data"].(map[string]interface{}); ok {
		if md, ok := subData["metadata"].(map[string]interface{}); ok {
			if tid, ok := md["tier_id"].(string); ok {
				tierID = tid
			}
		}
	}

	// Fallback: session metadata might also have it
	if tierID == "" {
		if md, ok := rawMap["metadata"].(map[string]interface{}); ok {
			if tid, ok := md["tier_id"].(string); ok {
				tierID = tid
			}
		}
	}

	if tierID == "" {
		log.Printf("No tier_id found in checkout session %s metadata", session.ID)
		return
	}

	tier := billing.GetTier(tierID)
	if tier == nil {
		log.Printf("Unknown tier %s in checkout session %s", tierID, session.ID)
		return
	}

	now := time.Now()
	periodEnd := now.AddDate(0, 1, 0)
	if err := h.userRepo.UpdateSubscription(ctx, fullSession.Customer, tierID, fullSession.Subscription, tier.IncludedTokens, now, periodEnd); err != nil {
		log.Printf("Failed to update subscription for customer %s: %v", fullSession.Customer, err)
		return
	}

	// Issue initial credit grant
	creditAmountCents := creditGrantCentsForTier(tier)
	if creditAmountCents > 0 {
		if _, err := h.billing.CreateCreditGrant(ctx, fullSession.Customer, creditAmountCents); err != nil {
			log.Printf("Failed to create initial credit grant for customer %s: %v", fullSession.Customer, err)
		}
	}

	log.Printf("Subscription created for customer %s: tier=%s, subscription=%s", fullSession.Customer, tierID, fullSession.Subscription)
}

func (h *CheckoutHandler) handleInvoicePaid(ctx context.Context, event *stripe.Event, w http.ResponseWriter) {
	var invoice struct {
		Customer     string `json:"customer"`
		Subscription string `json:"subscription"`
		PeriodStart  int64  `json:"period_start"`
		PeriodEnd    int64  `json:"period_end"`
		BillingReason string `json:"billing_reason"`
	}

	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("Failed to parse invoice: %v", err)
		return
	}

	if invoice.Subscription == "" {
		return
	}

	// Skip the first invoice (handled by checkout.session.completed)
	if invoice.BillingReason == "subscription_create" {
		return
	}

	// Find user to determine tier
	usr, err := h.userRepo.GetByStripeCustomerID(ctx, invoice.Customer)
	if err != nil {
		log.Printf("Failed to find user for customer %s: %v", invoice.Customer, err)
		return
	}

	if usr.SubscriptionTier == nil {
		log.Printf("User for customer %s has no subscription tier", invoice.Customer)
		return
	}

	tier := billing.GetTier(*usr.SubscriptionTier)
	if tier == nil {
		log.Printf("Unknown tier %s for customer %s", *usr.SubscriptionTier, invoice.Customer)
		return
	}

	periodStart := time.Unix(invoice.PeriodStart, 0)
	periodEnd := time.Unix(invoice.PeriodEnd, 0)

	// Reset billing cycle
	if err := h.userRepo.ResetBillingCycle(ctx, invoice.Customer, periodStart, periodEnd); err != nil {
		log.Printf("Failed to reset billing cycle for customer %s: %v", invoice.Customer, err)
		return
	}

	// Issue credit grant for new period
	creditAmountCents := creditGrantCentsForTier(tier)
	if creditAmountCents > 0 {
		if _, err := h.billing.CreateCreditGrant(ctx, invoice.Customer, creditAmountCents); err != nil {
			log.Printf("Failed to create credit grant for customer %s: %v", invoice.Customer, err)
		}
	}

	log.Printf("Billing cycle reset for customer %s: period %s to %s", invoice.Customer, periodStart, periodEnd)
}

func (h *CheckoutHandler) handleSubscriptionDeleted(ctx context.Context, event *stripe.Event, w http.ResponseWriter) {
	var subscription struct {
		ID       string `json:"id"`
		Customer string `json:"customer"`
	}

	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		log.Printf("Failed to parse subscription: %v", err)
		return
	}

	if err := h.userRepo.ClearSubscription(ctx, subscription.Customer); err != nil {
		log.Printf("Failed to clear subscription for customer %s: %v", subscription.Customer, err)
		return
	}

	log.Printf("Subscription %s deleted for customer %s", subscription.ID, subscription.Customer)
}

// creditGrantCentsForTier computes the credit grant value in cents for a tier's included tokens.
// credit_amount = included_tokens * overage_price_per_token_in_cents
func creditGrantCentsForTier(tier *billing.SubscriptionTier) int64 {
	overageCents, err := strconv.ParseFloat(tier.OveragePriceCentsDecimal, 64)
	if err != nil {
		return 0
	}
	return int64(float64(tier.IncludedTokens) * overageCents)
}
