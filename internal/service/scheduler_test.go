package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// Mocks
// =============================================================================

type MockScanOrchestrator struct {
	mock.Mock
}

func (m *MockScanOrchestrator) EnqueueScan(ctx context.Context, dataSourceID types.ID, tenantID types.ID, scanType discovery.ScanType) (*discovery.ScanRun, error) {
	args := m.Called(ctx, dataSourceID, tenantID, scanType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*discovery.ScanRun), args.Error(1)
}

func (m *MockScanOrchestrator) GetScan(ctx context.Context, id types.ID) (*discovery.ScanRun, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*discovery.ScanRun), args.Error(1)
}

func (m *MockScanOrchestrator) GetHistory(ctx context.Context, dataSourceID types.ID) ([]discovery.ScanRun, error) {
	args := m.Called(ctx, dataSourceID)
	return args.Get(0).([]discovery.ScanRun), args.Error(1)
}

// =============================================================================
// Tests
// =============================================================================

func TestScheduler_IsDue(t *testing.T) {
	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	svc := NewSchedulerService(nil, nil, nil, nil, nil, logger)

	// 1. Cron: Every minute (* * * * *)
	// Last run: 2 minutes ago -> Due
	cronExpr := "* * * * *"
	lastRun := time.Now().Add(-2 * time.Minute)
	due, err := svc.IsDue(cronExpr, &lastRun)
	require.NoError(t, err)
	assert.True(t, due, "Should be due")

	// 2. Cron: Daily at midnight (0 0 * * *)
	// Last run: 1 hour ago (11pm) -> Not due until next midnight
	// Note: Scheduler uses robfig/cron, check syntax support. Standard is 5 fields.
	cronExprDaily := "0 0 * * *"
	// Mock that it ran today at 00:00. Now is 10:00. Next run is tomorrow 00:00.
	// So it should NOT be due.

	// Better test:
	// Set LastRun to yesterday 23:00. Next run today 00:00. Now is today 10:00. -> Due.
	lastRunYesterday := time.Now().Add(-24 * time.Hour)
	// We need to be careful with "Now". implementation uses time.Now() internally.
	// This makes IsDue slightly widely-scoped.
	// We'll trust the logic: Next(lastRun) <= Now

	due, err = svc.IsDue(cronExprDaily, &lastRunYesterday)
	require.NoError(t, err)
	assert.True(t, due)

	// 3. Not due case
	// Every hour (0 * * * *)
	// Last run: Now. Next run: Next hour boundary. -> Not Due.
	cronHourly := "0 * * * *"
	lastRunJustNow := time.Now()
	due, err = svc.IsDue(cronHourly, &lastRunJustNow)
	require.NoError(t, err)
	// Depending on current time, this is tricky.
	// If I run this at 10:59, and last run was 10:54, next is 11:00. Not due.
	// Safe assumption.
	assert.False(t, due)
}

func TestScheduler_ValidateCron(t *testing.T) {
	svc := NewSchedulerService(nil, nil, nil, nil, nil, slog.Default())

	assert.NoError(t, svc.ValidateCron("* * * * *"))
	assert.NoError(t, svc.ValidateCron("0 0 * * *"))
	assert.Error(t, svc.ValidateCron("invalid cron"))
}

func TestScheduler_Lifecycle(t *testing.T) {
	// Just test Start/Stop doesn't panic
	svc := NewSchedulerService(nil, nil, nil, nil, nil, slog.Default())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := svc.Start(ctx)
	assert.NoError(t, err)

	svc.Stop()
}
