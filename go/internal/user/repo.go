package user

import (
	"context"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/uptrace/bun"
)

type Repository interface {
	InitializeDatabase(ctx context.Context) error
	GetByID(ctx context.Context, userID string) (*models.User, error)
	GetByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	GetOrCreate(ctx context.Context, userID, email, firstName, lastName string) (*models.User, error)
	UpdateStripeCustomerID(ctx context.Context, userID, stripeCustomerID string) error
	GetAvailableCredits(ctx context.Context, userID string) (int64, error)
	IncrementTokensUsed(ctx context.Context, stripeCustomerID string, amount int64) error
	IncrementTokensPurchased(ctx context.Context, stripeCustomerID string, amount int64) error
}

type UserRepository struct {
	db *bun.DB
}

func NewUserRepository(db *bun.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) InitializeDatabase(ctx context.Context) error {
	_, err := r.db.NewCreateTable().
		Model((*models.UserDB)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = r.db.NewCreateIndex().
		Model((*models.UserDB)(nil)).
		Index("idx_users_email").
		Column("email").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = r.db.NewCreateIndex().
		Model((*models.UserDB)(nil)).
		Index("idx_users_stripe_customer_id").
		Column("stripe_customer_id").
		IfNotExists().
		Exec(ctx)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (*models.User, error) {
	userDB := new(models.UserDB)
	err := r.db.NewSelect().
		Model(userDB).
		Where("id = ?", userID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return userDB.ToUser(), nil
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	userDB := models.UserFromDomain(user)
	userDB.CreatedAt = time.Now()
	userDB.UpdatedAt = time.Now()
	_, err := r.db.NewInsert().Model(userDB).Exec(ctx)
	return err
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	userDB := models.UserFromDomain(user)
	userDB.UpdatedAt = time.Now()
	_, err := r.db.NewUpdate().
		Model(userDB).
		WherePK().
		Exec(ctx)
	return err
}

func (r *UserRepository) GetOrCreate(ctx context.Context, userID, email, firstName, lastName string) (*models.User, error) {
	user, err := r.GetByID(ctx, userID)
	if err == nil {
		return user, nil
	}

	newUser := &models.User{
		ID:        userID,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}

	if err := r.Create(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

func (r *UserRepository) UpdateStripeCustomerID(ctx context.Context, userID, stripeCustomerID string) error {
	_, err := r.db.NewUpdate().
		Model((*models.UserDB)(nil)).
		Set("stripe_customer_id = ?", stripeCustomerID).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", userID).
		Exec(ctx)
	return err
}

func (r *UserRepository) GetByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*models.User, error) {
	userDB := new(models.UserDB)
	err := r.db.NewSelect().
		Model(userDB).
		Where("stripe_customer_id = ?", stripeCustomerID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return userDB.ToUser(), nil
}

func (r *UserRepository) GetAvailableCredits(ctx context.Context, userID string) (int64, error) {
	userDB := new(models.UserDB)
	err := r.db.NewSelect().
		Model(userDB).
		Column("tokens_purchased", "tokens_used").
		Where("id = ?", userID).
		Scan(ctx)
	if err != nil {
		return 0, err
	}
	return userDB.TokensPurchased - userDB.TokensUsed, nil
}

func (r *UserRepository) IncrementTokensUsed(ctx context.Context, stripeCustomerID string, amount int64) error {
	_, err := r.db.NewUpdate().
		Model((*models.UserDB)(nil)).
		Set("tokens_used = tokens_used + ?", amount).
		Set("updated_at = ?", time.Now()).
		Where("stripe_customer_id = ?", stripeCustomerID).
		Exec(ctx)
	return err
}

func (r *UserRepository) IncrementTokensPurchased(ctx context.Context, stripeCustomerID string, amount int64) error {
	_, err := r.db.NewUpdate().
		Model((*models.UserDB)(nil)).
		Set("tokens_purchased = tokens_purchased + ?", amount).
		Set("updated_at = ?", time.Now()).
		Where("stripe_customer_id = ?", stripeCustomerID).
		Exec(ctx)
	return err
}
