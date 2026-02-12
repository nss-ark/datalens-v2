//go:build integration

package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/complyark/datalens/internal/domain/breach"
	"github.com/complyark/datalens/internal/repository"
	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBreachService_Integration(t *testing.T) {
	pool := setupPostgres(t)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	eventBus := newMockEventBus()

	// Repositories & Services
	breachRepo := repository.NewPostgresBreachRepository(pool)
	auditRepo := repository.NewPostgresAuditRepository(pool)
	auditService := NewAuditService(auditRepo, logger)
	breachService := NewBreachService(breachRepo, auditService, eventBus, logger)

	// Test Data
	tenantID := types.NewID()
	userID := types.NewID()
	ctx := context.Background()
	ctx = context.WithValue(ctx, types.ContextKeyTenantID, tenantID)
	ctx = context.WithValue(ctx, types.ContextKeyUserID, userID)

	t.Run("Full Breach Lifecycle", func(t *testing.T) {
		// 1. Create Incident
		req := CreateIncidentRequest{
			Title:                    "Test Data Leak",
			Description:              "Sensitive data found on public share",
			Type:                     "Data Leak",
			Severity:                 breach.SeverityHigh,
			DetectedAt:               time.Now().UTC(),
			OccurredAt:               time.Now().UTC().Add(-1 * time.Hour),
			AffectedSystems:          []string{"SharePoint", "Email"},
			AffectedDataSubjectCount: 100,
			PiiCategories:            []string{"Financial", "Identity"},
			PoCName:                  "John Security",
			PoCRole:                  "CISO",
			PoCEmail:                 "ciso@example.com",
		}

		created, err := breachService.CreateIncident(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, req.Title, created.Title)
		assert.Equal(t, breach.StatusOpen, created.Status)
		assert.True(t, created.IsReportableToCertIn) // High severity

		// 2. Verify Audit Log (Async, so wait/poll)
		assert.Eventually(t, func() bool {
			logs, err := auditRepo.GetByTenant(ctx, tenantID, 10)
			if err != nil {
				return false
			}
			for _, l := range logs {
				if l.Action == "CREATE_INCIDENT" && l.ResourceID == created.ID {
					// Verify content
					return l.UserID == userID && l.ResourceType == "BREACH_INCIDENT"
				}
			}
			return false
		}, 5*time.Second, 100*time.Millisecond, "Audit log for creation not found")

		// 3. Get with SLA
		fetched, sla, err := breachService.GetIncident(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, fetched.ID)
		assert.NotNil(t, sla)
		assert.Contains(t, sla, "cert_in_deadline")
		assert.Contains(t, sla, "time_remaining_cert_in")

		// 4. Update Incident
		newStatus := breach.StatusInvestigating
		newDesc := "Updated description after investigation"
		updateReq := UpdateIncidentRequest{
			Status:      &newStatus,
			Description: &newDesc,
		}

		updated, err := breachService.UpdateIncident(ctx, created.ID, updateReq)
		require.NoError(t, err)
		assert.Equal(t, breach.StatusInvestigating, updated.Status)
		assert.Equal(t, newDesc, updated.Description)

		// 5. Verify Update Audit Log
		assert.Eventually(t, func() bool {
			logs, err := auditRepo.GetByTenant(ctx, tenantID, 10)
			if err != nil {
				return false
			}
			for _, l := range logs {
				if l.Action == "UPDATE_INCIDENT" && l.ResourceID == created.ID {
					// Verify diff capture
					oldVal, ok := l.OldValues["status"]
					if !ok {
						return false
					}
					return oldVal == "OPEN"
				}
			}
			return false
		}, 5*time.Second, 100*time.Millisecond, "Audit log for update not found")
	})

	t.Run("Create Invalid Incident", func(t *testing.T) {
		req := CreateIncidentRequest{
			Title: "", // Missing title
		}
		_, err := breachService.CreateIncident(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "title is required")
	})
}
