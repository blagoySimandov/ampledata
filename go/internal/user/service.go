package user

import (
	"context"

	stripe "github.com/stripe/stripe-go/v84"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"golang.org/x/sync/singleflight"
)

type Service interface {
	GetOrCreate(ctx context.Context, userID, email, firstName, lastName, profilePictureURL string) (*models.User, error)
}

type billingClient interface {
	GetOrCreateCustomer(ctx context.Context, userID, email string) (*stripe.Customer, error)
}

type UserService struct {
	repo    Repository
	billing billingClient
	sf      singleflight.Group
}

func NewUserService(repo Repository, billing billingClient) *UserService {
	return &UserService{
		repo:    repo,
		billing: billing,
	}
}

func (s *UserService) GetOrCreate(ctx context.Context, userID, email, firstName, lastName, profilePictureURL string) (*models.User, error) {
	user, err := s.repo.GetOrCreate(ctx, userID, email, firstName, lastName, profilePictureURL)
	if err != nil {
		return nil, err
	}

	if user.StripeCustomerID != nil {
		return user, nil
	}

	// NOTE: Use a singleflight group to ensure that only one user is created at a time (db and stripe also use idempotency keys so this
	// is really just to prevent error logs.)
	_, err, _ = s.sf.Do(userID, func() (any, error) {
		return nil, s.ensureStripeCustomer(ctx, userID, email)
	})
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, userID)
}

func (s *UserService) ensureStripeCustomer(ctx context.Context, userID, email string) error {
	customer, err := s.billing.GetOrCreateCustomer(ctx, userID, email)
	if err != nil {
		return err
	}
	return s.repo.UpdateStripeCustomerID(ctx, userID, customer.ID)
}
