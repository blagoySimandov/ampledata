package user

import (
	"context"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/services"
)

type Service interface {
	GetOrCreate(ctx context.Context, userID, email, firstName, lastName string) (*models.User, error)
}

type UserService struct {
	repo    Repository
	billing services.BillingService
}

func NewUserService(repo Repository, billing services.BillingService) *UserService {
	return &UserService{
		repo:    repo,
		billing: billing,
	}
}

func (s *UserService) GetOrCreate(ctx context.Context, userID, email, firstName, lastName string) (*models.User, error) {
	user, err := s.repo.GetOrCreate(ctx, userID, email, firstName, lastName)
	if err != nil {
		return nil, err
	}

	if user.StripeCustomerID == nil {
		customer, err := s.billing.CreateCustomer(ctx, userID, email)
		if err != nil {
			return nil, err
		}
		if err := s.repo.UpdateStripeCustomerID(ctx, userID, customer.ID); err != nil {
			return nil, err
		}
		user.StripeCustomerID = &customer.ID
	}

	return user, nil
}
