package analytics

import (
	"context"
	"time"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/types"
)

// ConsentAnalyticsService aggregates consent data for dashboards.
type ConsentAnalyticsService struct {
	repo consent.ConsentSessionRepository
}

// NewConsentAnalyticsService creates a new analytics service.
func NewConsentAnalyticsService(repo consent.ConsentSessionRepository) *ConsentAnalyticsService {
	return &ConsentAnalyticsService{repo: repo}
}

// GetConversionStats returns opt-in/opt-out trends over time.
func (s *ConsentAnalyticsService) GetConversionStats(ctx context.Context, from, to time.Time, interval string) ([]consent.ConversionStat, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}
	return s.repo.GetConversionStats(ctx, tenantID, from, to, interval)
}

// GetPurposeStats returns aggregate counts per purpose.
func (s *ConsentAnalyticsService) GetPurposeStats(ctx context.Context, from, to time.Time) ([]consent.PurposeStat, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}
	return s.repo.GetPurposeStats(ctx, tenantID, from, to)
}
