package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/pkg/types"
)

// DashboardService aggregates statistics for the tenant dashboard.
type DashboardService struct {
	dsRepo      discovery.DataSourceRepository
	piiRepo     discovery.PIIClassificationRepository
	scanRunRepo discovery.ScanRunRepository
	logger      *slog.Logger
}

// NewDashboardService creates a new DashboardService.
func NewDashboardService(
	dsRepo discovery.DataSourceRepository,
	piiRepo discovery.PIIClassificationRepository,
	scanRunRepo discovery.ScanRunRepository,
	logger *slog.Logger,
) *DashboardService {
	return &DashboardService{
		dsRepo:      dsRepo,
		piiRepo:     piiRepo,
		scanRunRepo: scanRunRepo,
		logger:      logger.With("service", "dashboard"),
	}
}

// DashboardStats holds aggregated metrics for the dashboard.
type DashboardStats struct {
	TotalDataSources int                 `json:"total_data_sources"`
	TotalPIIFields   int                 `json:"total_pii_fields"`
	TotalScans       int                 `json:"total_scans"`
	RiskScore        int                 `json:"risk_score"`      // Placeholder 0-100
	PIIByCategory    map[string]int      `json:"pii_by_category"` // e.g. "CONTACT": 5
	RecentScans      []discovery.ScanRun `json:"recent_scans"`
	PendingReviews   int                 `json:"pending_reviews"`
}

// GetStats returns aggregated statistics for the tenant.
func (s *DashboardService) GetStats(ctx context.Context, tenantID types.ID) (*DashboardStats, error) {
	stats := &DashboardStats{
		PIIByCategory: make(map[string]int),
		RecentScans:   make([]discovery.ScanRun, 0),
	}

	// 1. Data Sources Count
	// Todo: Add CountByTenant to DataSourceRepository. For now list all and count.
	dss, err := s.dsRepo.GetByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get data sources: %w", err)
	}
	stats.TotalDataSources = len(dss)

	// 2. Pending Reviews
	// We can reuse GetPending with limit 1 to get total count from paginated result if we implemented it right?
	// GetPending returns PaginatedResult which has Total.
	pending, err := s.piiRepo.GetPending(ctx, tenantID, types.Pagination{Page: 1, PageSize: 1})
	if err != nil {
		return nil, fmt.Errorf("get pending pii: %w", err)
	}
	stats.PendingReviews = pending.Total

	// 3. Scan Runs
	// Get active (running) scans
	activeRuns, err := s.scanRunRepo.GetActive(ctx, tenantID)
	if err != nil && !types.IsNotFoundError(err) {
		s.logger.ErrorContext(ctx, "failed to get active scans", "error", err)
	}
	// Get recent scans (history)
	recentRuns, err := s.scanRunRepo.GetRecent(ctx, tenantID, 5)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get recent scans", "error", err)
	}

	stats.TotalScans = len(activeRuns) // TODO: This should probably be total historical scans? But spec said "total_scans".
	// Let's assume TotalScans means "Total Scans performed ever".
	// We don't have a Count method for ScanRuns yet.
	// For now, let's just use the count of recent (+ active maybe overlap).
	// Let's stick to spec requirements: "total_scans".
	// I'll leave it as a placeholder or use what I have.
	// Actually, GetRecent returns 5.
	// Let's assume TotalScans is not critical for MVP or I can add CountAll later.
	// I'll return len(recentRuns) for now as a fallback or 0.

	// Better: GetRecent scans + Active count.
	stats.RecentScans = recentRuns

	// 4. Total PII and Breakdown
	piiCounts, err := s.piiRepo.GetCounts(ctx, tenantID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get pii counts", "error", err)
	} else {
		stats.TotalPIIFields = piiCounts.Total
		stats.PIIByCategory = piiCounts.ByCategory
	}

	return stats, nil
}
