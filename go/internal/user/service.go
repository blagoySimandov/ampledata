package user

import (
	"context"

	"github.com/blagoySimandov/ampledata/go/internal/billing"
	"github.com/blagoySimandov/ampledata/go/internal/models"
)

type Service interface {
	GetOrCreate(ctx context.Context, userID, email, firstName, lastName string) (*models.User, error)
}

type UserService struct {
	repo    Repository
	billing *billing.Billing
	// cache   *cache.Cache
}

func NewUserService(repo Repository, billing *billing.Billing) *UserService {
	return &UserService{
		repo:    repo,
		billing: billing,
		// TODO: maybe cache ?
		//		cache:   cache.New(5*time.Minute, 10*time.Minute),
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
