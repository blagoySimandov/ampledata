package services

import "context"

type BillingService interface {
	ReportUsage(ctx context.Context, userID string, credits int) error
}
