package service

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestPurposeService() (*PurposeService, *mockPurposeRepo, *mockEventBus) {
	repo := newMockPurposeRepo()
	eb := newMockEventBus()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := NewPurposeService(repo, eb, logger)
	return svc, repo, eb
}

func TestPurposeService_Create_Success(t *testing.T) {
	svc, _, eb := newTestPurposeService()
	ctx := context.Background()
	tid := types.NewID()

	p, err := svc.Create(ctx, CreatePurposeInput{
		TenantID:        tid,
		Code:            "MARKETING",
		Name:            "Marketing Analytics",
		Description:     "Data processing for marketing campaigns",
		LegalBasis:      "CONSENT",
		RetentionDays:   180,
		RequiresConsent: true,
	})

	require.NoError(t, err)
	assert.NotEqual(t, types.ID{}, p.ID)
	assert.Equal(t, "MARKETING", p.Code)
	assert.Equal(t, 180, p.RetentionDays)
	assert.True(t, p.IsActive)
	assert.Len(t, eb.Events, 1)
}

func TestPurposeService_Create_MissingCode(t *testing.T) {
	svc, _, _ := newTestPurposeService()
	ctx := context.Background()

	_, err := svc.Create(ctx, CreatePurposeInput{
		TenantID:   types.NewID(),
		Code:       "",
		Name:       "Marketing",
		LegalBasis: "CONSENT",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "code")
}

func TestPurposeService_Create_MissingName(t *testing.T) {
	svc, _, _ := newTestPurposeService()
	ctx := context.Background()

	_, err := svc.Create(ctx, CreatePurposeInput{
		TenantID:   types.NewID(),
		Code:       "MKT",
		Name:       "",
		LegalBasis: "CONSENT",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestPurposeService_Create_MissingLegalBasis(t *testing.T) {
	svc, _, _ := newTestPurposeService()
	ctx := context.Background()

	_, err := svc.Create(ctx, CreatePurposeInput{
		TenantID:   types.NewID(),
		Code:       "MKT",
		Name:       "Marketing",
		LegalBasis: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "legal_basis")
}

func TestPurposeService_Create_DefaultRetention(t *testing.T) {
	svc, _, _ := newTestPurposeService()
	ctx := context.Background()

	p, err := svc.Create(ctx, CreatePurposeInput{
		TenantID:      types.NewID(),
		Code:          "SVC",
		Name:          "Service Delivery",
		LegalBasis:    "CONTRACT",
		RetentionDays: 0, // should default to 365
	})

	require.NoError(t, err)
	assert.Equal(t, 365, p.RetentionDays, "should default to 365 days")
}

func TestPurposeService_Create_DuplicateCode(t *testing.T) {
	svc, _, _ := newTestPurposeService()
	ctx := context.Background()
	tid := types.NewID()

	_, _ = svc.Create(ctx, CreatePurposeInput{
		TenantID: tid, Code: "MKT", Name: "Marketing", LegalBasis: "CONSENT",
	})

	_, err := svc.Create(ctx, CreatePurposeInput{
		TenantID: tid, Code: "MKT", Name: "Marketing Duplicate", LegalBasis: "CONSENT",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflict")
}

func TestPurposeService_GetByID(t *testing.T) {
	svc, _, _ := newTestPurposeService()
	ctx := context.Background()

	p, _ := svc.Create(ctx, CreatePurposeInput{
		TenantID: types.NewID(), Code: "SVC", Name: "Service", LegalBasis: "CONTRACT",
	})

	fetched, err := svc.GetByID(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, "SVC", fetched.Code)
}

func TestPurposeService_Update(t *testing.T) {
	svc, _, eb := newTestPurposeService()
	ctx := context.Background()

	p, _ := svc.Create(ctx, CreatePurposeInput{
		TenantID: types.NewID(), Code: "MKT", Name: "Marketing", LegalBasis: "CONSENT",
	})

	isActive := false
	updated, err := svc.Update(ctx, UpdatePurposeInput{
		ID:       p.ID,
		Name:     "Updated Marketing",
		IsActive: &isActive,
	})

	require.NoError(t, err)
	assert.Equal(t, "Updated Marketing", updated.Name)
	assert.False(t, updated.IsActive)
	assert.Len(t, eb.Events, 2)
}

func TestPurposeService_Delete(t *testing.T) {
	svc, repo, eb := newTestPurposeService()
	ctx := context.Background()

	p, _ := svc.Create(ctx, CreatePurposeInput{
		TenantID: types.NewID(), Code: "MKT", Name: "Marketing", LegalBasis: "CONSENT",
	})

	err := svc.Delete(ctx, p.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, p.ID)
	assert.Error(t, err)
	assert.Len(t, eb.Events, 2)
}

func TestPurposeService_ListByTenant(t *testing.T) {
	svc, _, _ := newTestPurposeService()
	ctx := context.Background()
	tid := types.NewID()

	_, _ = svc.Create(ctx, CreatePurposeInput{TenantID: tid, Code: "MKT", Name: "Marketing", LegalBasis: "CONSENT"})
	_, _ = svc.Create(ctx, CreatePurposeInput{TenantID: tid, Code: "SVC", Name: "Service", LegalBasis: "CONTRACT"})
	_, _ = svc.Create(ctx, CreatePurposeInput{TenantID: types.NewID(), Code: "OTHER", Name: "Other", LegalBasis: "CONSENT"})

	result, err := svc.ListByTenant(ctx, tid)
	require.NoError(t, err)
	assert.Len(t, result, 2)
}
