package service

import (
	"context"
	"testing"
	"time"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNoticeService_ValidateSchema(t *testing.T) {
	// Setup mocks
	repo := newMockNoticeRepo()
	widgetRepo := newMockWidgetRepo()
	eb := newMockEventBus()
	svc := NewNoticeService(repo, widgetRepo, eb, nil)

	t.Run("valid notice", func(t *testing.T) {
		n := &consent.ConsentNotice{
			Schema: consent.NoticeSchemaFields{
				DataTypesCollected:   []string{"Name"},
				Purposes:             []string{"Marketing"},
				FiduciaryName:        "Acme Corp",
				FiduciaryContact:     "privacy@acme.com",
				RightsWithdraw:       true,
				RightsAccess:         true,
				RightsCorrection:     true,
				RightsGrievance:      true,
				RightsNomination:     true,
				ComplaintMethod:      "Email",
				DPOName:              "John Doe",
				DPOContact:           "dpo@acme.com",
				BoardComplaintMethod: "Website",
				SharingCategories:    []string{"None"},
				CrossBorderTransfer:  "US",
				RetentionPeriod:      "1 Year",
			},
		}
		missing := svc.ValidateSchema(n)
		assert.Empty(t, missing)
	})

	t.Run("missing fields", func(t *testing.T) {
		n := &consent.ConsentNotice{
			Schema: consent.NoticeSchemaFields{
				// Empty schema
			},
		}
		missing := svc.ValidateSchema(n)
		assert.Contains(t, missing, "data_types_collected")
		assert.Contains(t, missing, "purposes")
		assert.Contains(t, missing, "fiduciary_name")
		assert.Contains(t, missing, "rights_withdraw")
		assert.Contains(t, missing, "dpo_contact")
	})
}

func TestNoticeService_Publish_Validation(t *testing.T) {
	repo := newMockNoticeRepo()
	widgetRepo := newMockWidgetRepo()
	eb := newMockEventBus()
	svc := NewNoticeService(repo, widgetRepo, eb, nil)

	tenantID := types.NewID()
	ctx := context.WithValue(context.Background(), types.ContextKeyTenantID, tenantID)

	t.Run("publish fails validation", func(t *testing.T) {
		noticeID := types.NewID()
		notice := &consent.ConsentNotice{
			TenantEntity: types.TenantEntity{
				BaseEntity: types.BaseEntity{
					ID:        noticeID,
					CreatedAt: time.Now().UTC(),
				},
				TenantID: tenantID,
			},
			Status: consent.NoticeStatusDraft,
			// Empty schema
		}
		repo.Create(ctx, notice)

		published, err := svc.Publish(ctx, noticeID)
		assert.Error(t, err)
		assert.Nil(t, published)
		assert.Contains(t, err.Error(), "schema validation failed")
	})

	t.Run("publish succeeds validation", func(t *testing.T) {
		noticeID := types.NewID()
		notice := &consent.ConsentNotice{
			TenantEntity: types.TenantEntity{
				BaseEntity: types.BaseEntity{
					ID:        noticeID,
					CreatedAt: time.Now().UTC(),
				},
				TenantID: tenantID,
			},
			Status: consent.NoticeStatusDraft,
			Schema: consent.NoticeSchemaFields{
				DataTypesCollected:   []string{"Name"},
				Purposes:             []string{"Marketing"},
				FiduciaryName:        "Acme Corp",
				FiduciaryContact:     "privacy@acme.com",
				RightsWithdraw:       true,
				RightsAccess:         true,
				RightsCorrection:     true,
				RightsGrievance:      true,
				RightsNomination:     true,
				ComplaintMethod:      "Email",
				DPOName:              "John Doe",
				DPOContact:           "dpo@acme.com",
				BoardComplaintMethod: "Website",
				SharingCategories:    []string{"None"},
				CrossBorderTransfer:  "US",
				RetentionPeriod:      "1 Year",
			},
		}
		repo.Create(ctx, notice)

		published, err := svc.Publish(ctx, noticeID)
		assert.NoError(t, err)
		assert.NotNil(t, published)
		assert.Equal(t, consent.NoticeStatusPublished, published.Status)

		// Check side effects
		updatedNotice, _ := repo.GetByID(ctx, noticeID)
		assert.Equal(t, consent.NoticeStatusPublished, updatedNotice.Status)
		assert.Equal(t, 1, updatedNotice.Version)

		// EventBus check (Events is exported field in mockEventBus)
		assert.NotEmpty(t, eb.Events)
	})
}
