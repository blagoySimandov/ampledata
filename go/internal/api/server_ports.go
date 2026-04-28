package api

import (
	"context"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/gcs"
	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/services"
	"github.com/blagoySimandov/ampledata/go/internal/state"
	"github.com/google/uuid"
	stripe "github.com/stripe/stripe-go/v84"
)

type ITemplateRepo interface {
	ListTemplates(ctx context.Context, userId string) ([]*models.TemplateDB, error)
}

type IEnricher interface {
	Enrich(ctx context.Context, jobID, userID, stripeCustomerID string, rowKeys []string, columnsMetadata []*models.ColumnMetadata, keyColumnDescription *string) error
	GetProgress(ctx context.Context, jobID string) (*models.JobProgress, error)
	Cancel(ctx context.Context, jobID string) error
	GetResults(ctx context.Context, jobID string, offset, limit int) ([]*models.EnrichmentResult, error)
	GetRowsProgress(ctx context.Context, jobID string, params state.RowsQueryParams) (*models.RowsProgressResponse, error)
}

type Store interface {
	CreateSource(ctx context.Context, source *models.SourceDB) error
	GetSource(ctx context.Context, sourceID uuid.UUID) (*models.Source, error)
	GetSourcesByUser(ctx context.Context, userID string, offset, limit int) ([]*models.Source, error)
	GetJobsBySource(ctx context.Context, sourceID uuid.UUID) ([]*models.Job, error)
	CreatePendingJob(ctx context.Context, jobID, userID string, sourceID uuid.UUID, templateID *uuid.UUID) error
	GetJob(ctx context.Context, jobID string) (*models.Job, error)
	UpdateJobConfiguration(ctx context.Context, jobID string, keyColumns []string, columnsMetadata []*models.ColumnMetadata, keyColumnDescription *string) error
	StartJob(ctx context.Context, jobID string, totalRows int) error
	GetJobsByUser(ctx context.Context, userID string, offset, limit int) ([]*models.Job, error)
	BulkCreateRows(ctx context.Context, jobID string, rowKeys []string) error
	SaveRowState(ctx context.Context, jobID string, rowState *models.RowState) error
	GetRowState(ctx context.Context, jobID string, key string) (*models.RowState, error)
	GetRowsAtStage(ctx context.Context, jobID string, stage models.RowStage, offset, limit int) ([]*models.RowState, error)
	GetRowsPaginated(ctx context.Context, jobID string, params state.RowsQueryParams) (*state.PaginatedRows, error)
	SetJobStatus(ctx context.Context, jobID string, status models.JobStatus) error
	GetJobStatus(ctx context.Context, jobID string) (models.JobStatus, error)
	GetJobProgress(ctx context.Context, jobID string) (*models.JobProgress, error)
	IncrementJobCost(ctx context.Context, jobID string, costDollars, costCredits int) error
	Close() error
}

type UserRepo interface {
	InitializeDatabase(ctx context.Context) error
	GetByID(ctx context.Context, userID string) (*models.User, error)
	GetByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	GetOrCreate(ctx context.Context, userID, email, firstName, lastName, profilePictureURL string) (*models.User, error)
	UpdateStripeCustomerID(ctx context.Context, userID, stripeCustomerID string) error
	GetAvailableCredits(ctx context.Context, userID string) (int64, error)
	IncrementTokensUsed(ctx context.Context, stripeCustomerID string, amount int64) error
	UpdateSubscription(ctx context.Context, stripeCustomerID, tier, subscriptionID string, tokensIncluded int64, periodStart, periodEnd time.Time) error
	ResetBillingCycle(ctx context.Context, stripeCustomerID string, periodStart, periodEnd time.Time) error
	ClearSubscription(ctx context.Context, stripeCustomerID string) error
	SetCancelAtPeriodEnd(ctx context.Context, stripeCustomerID string, value bool) error
	UpdateSubscriptionTier(ctx context.Context, stripeCustomerID, tier string, tokensIncluded int64) error
}

type BillingService interface {
	ReportUsage(ctx context.Context, stripeCustomerID string, credits int) error
	GetOrCreateCustomer(ctx context.Context, userID, email string) (*stripe.Customer, error)
	CreateSubscriptionCheckout(ctx context.Context, customerID, tierID, successURL, cancelURL string) (*stripe.CheckoutSession, error)
	GetSubscription(ctx context.Context, subscriptionID string) (*stripe.Subscription, error)
	CancelSubscription(ctx context.Context, subscriptionID string) (*stripe.Subscription, error)
	CancelSubscriptionImmediately(ctx context.Context, subscriptionID string) (*stripe.Subscription, error)
	UpgradeSubscription(ctx context.Context, subscriptionID, newTierID string) (*stripe.Subscription, error)
	CreatePortalSession(ctx context.Context, customerID, returnURL string) (*stripe.BillingPortalSession, error)
	CreateCreditGrant(ctx context.Context, customerID string, amountCents int64, idempotencyKey string) (*stripe.BillingCreditGrant, error)
	VerifyWebhookSignature(payload []byte, signature string) (*stripe.Event, error)
}

type KeySelector interface {
	SelectBestKey(ctx context.Context, headers []string, columnsMetadata []*models.ColumnMetadata) (*services.KeySelectorResult, error)
}

type SourcesService interface {
	ListSources(ctx context.Context, userID string, offset, limit int) ([]*services.SourceWithJobs, error)
	GetSource(ctx context.Context, sourceID uuid.UUID, userID string) (*services.SourceWithJobs, error)
	GetSourceData(ctx context.Context, sourceID uuid.UUID, userID string) (*gcs.CSVResult, error)
	EnrichSource(ctx context.Context, input services.EnrichSourceInput) (string, error)
	CreateUploadSource(ctx context.Context, userID, contentType string, headers []string) (uuid.UUID, string, error)
}
