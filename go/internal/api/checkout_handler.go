package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/blagoySimandov/ampledata/go/internal/services"
	"github.com/blagoySimandov/ampledata/go/internal/user"
)

type CheckoutHandler struct {
	billing  services.BillingService
	userRepo user.Repository
}

func NewCheckoutHandler(billing services.BillingService, userRepo user.Repository) *CheckoutHandler {
	return &CheckoutHandler{billing: billing, userRepo: userRepo}
}

type CreateCheckoutRequest struct {
	AmountCents int64  `json:"amount_cents"`
	SuccessURL  string `json:"success_url"`
	CancelURL   string `json:"cancel_url"`
}

type CreateCheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
	SessionID   string `json:"session_id"`
}

type CreditBalanceResponse struct {
	BalanceCents int64 `json:"balance_cents"`
}

func (h *CheckoutHandler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	dbUser, ok := user.GetDBUserFromContext(r.Context())
	if !ok || dbUser.StripeCustomerID == nil {
		http.Error(w, "User not found or missing Stripe customer", http.StatusBadRequest)
		return
	}

	var req CreateCheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AmountCents <= 0 {
		http.Error(w, "amount_cents must be positive", http.StatusBadRequest)
		return
	}

	if req.SuccessURL == "" || req.CancelURL == "" {
		http.Error(w, "success_url and cancel_url are required", http.StatusBadRequest)
		return
	}

	session, err := h.billing.CreateCheckoutSession(r.Context(), *dbUser.StripeCustomerID, req.AmountCents, req.SuccessURL, req.CancelURL)
	if err != nil {
		log.Printf("Failed to create checkout session: %v", err)
		http.Error(w, "Failed to create checkout session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CreateCheckoutResponse{
		CheckoutURL: session.URL,
		SessionID:   session.ID,
	})
}

func (h *CheckoutHandler) GetCreditBalance(w http.ResponseWriter, r *http.Request) {
	dbUser, ok := user.GetDBUserFromContext(r.Context())
	if !ok || dbUser.StripeCustomerID == nil {
		http.Error(w, "User not found or missing Stripe customer", http.StatusBadRequest)
		return
	}

	balance, err := h.billing.GetCreditBalance(r.Context(), *dbUser.StripeCustomerID)
	if err != nil {
		log.Printf("Failed to get credit balance: %v", err)
		http.Error(w, "Failed to get credit balance", http.StatusInternalServerError)
		return
	}

	var balanceCents int64
	if len(balance.Balances) > 0 {
		if balance.Balances[0].AvailableBalance != nil {
			balanceCents = balance.Balances[0].AvailableBalance.Monetary.Value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CreditBalanceResponse{
		BalanceCents: balanceCents,
	})
}

func (h *CheckoutHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// TODO: better logging using span/wide event logg.
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

	if event.Type == "checkout.session.completed" {
		var session struct {
			ID       string `json:"id"`
			Customer string `json:"customer"`
			Metadata struct {
				Type        string `json:"type"`
				AmountCents string `json:"amount_cents"`
			} `json:"metadata"`
		}

		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			log.Printf("Failed to parse checkout session: %v", err)
			http.Error(w, "Failed to parse event", http.StatusBadRequest)
			return
		}

		if session.Metadata.Type == "credit_purchase" {
			amountCents, err := strconv.ParseInt(session.Metadata.AmountCents, 10, 64)
			if err != nil {
				log.Printf("Failed to parse amount: %v", err)
				http.Error(w, "Invalid amount", http.StatusBadRequest)
				return
			}

			_, err = h.billing.CreateCreditGrant(r.Context(), session.Customer, amountCents)
			if err != nil {
				log.Printf("Failed to create credit grant for customer %s: %v", session.Customer, err)
				http.Error(w, "Failed to grant credits", http.StatusInternalServerError)
				return
			}

			if h.userRepo != nil {
				usr, err := h.userRepo.GetByStripeCustomerID(r.Context(), session.Customer)
				if err != nil {
					log.Printf("Failed to find user for stripe customer %s: %v", session.Customer, err)
				} else {
					if err := h.userRepo.IncrementTokensPurchased(r.Context(), usr.ID, amountCents); err != nil {
						log.Printf("Failed to increment tokens purchased for user %s: %v", usr.ID, err)
					}
				}
			}

			log.Printf("Granted %d cents in credits to customer %s", amountCents, session.Customer)
		}
	}

	w.WriteHeader(http.StatusOK)
}
