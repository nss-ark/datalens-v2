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

func newTestDataSourceService() (*DataSourceService, *mockDataSourceRepo, *mockEventBus) {
	repo := newMockDataSourceRepo()
	eb := newMockEventBus()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := NewDataSourceService(repo, eb, logger)
	return svc, repo, eb
}

func TestDataSourceService_Create_Success(t *testing.T) {
	svc, _, eb := newTestDataSourceService()
	ctx := context.Background()
	tid := types.NewID()

	ds, err := svc.Create(ctx, CreateDataSourceInput{
		TenantID:    tid,
		Name:        "Production DB",
		Type:        "POSTGRESQL",
		Description: "Main production database",
		Host:        "db.acme.local",
		Port:        5432,
		Database:    "prod",
	})

	require.NoError(t, err)
	assert.NotEqual(t, types.ID{}, ds.ID)
	assert.Equal(t, "Production DB", ds.Name)
	assert.Equal(t, types.DataSourceType("POSTGRESQL"), ds.Type)
	assert.Len(t, eb.Events, 1, "should publish datasource.created event")
	assert.Equal(t, "datasource.created", eb.Events[0].Type)
}

func TestDataSourceService_Create_MissingName(t *testing.T) {
	svc, _, _ := newTestDataSourceService()
	ctx := context.Background()

	_, err := svc.Create(ctx, CreateDataSourceInput{
		TenantID: types.NewID(),
		Name:     "",
		Type:     "POSTGRESQL",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestDataSourceService_Create_MissingType(t *testing.T) {
	svc, _, _ := newTestDataSourceService()
	ctx := context.Background()

	_, err := svc.Create(ctx, CreateDataSourceInput{
		TenantID: types.NewID(),
		Name:     "My DB",
		Type:     "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "type")
}

func TestDataSourceService_GetByID(t *testing.T) {
	svc, _, _ := newTestDataSourceService()
	ctx := context.Background()

	ds, _ := svc.Create(ctx, CreateDataSourceInput{
		TenantID: types.NewID(),
		Name:     "Test DB",
		Type:     "MYSQL",
	})

	fetched, err := svc.GetByID(ctx, ds.ID)
	require.NoError(t, err)
	assert.Equal(t, "Test DB", fetched.Name)
}

func TestDataSourceService_ListByTenant(t *testing.T) {
	svc, _, _ := newTestDataSourceService()
	ctx := context.Background()
	tid := types.NewID()

	_, _ = svc.Create(ctx, CreateDataSourceInput{TenantID: tid, Name: "DB 1", Type: "POSTGRESQL"})
	_, _ = svc.Create(ctx, CreateDataSourceInput{TenantID: tid, Name: "DB 2", Type: "MYSQL"})
	_, _ = svc.Create(ctx, CreateDataSourceInput{TenantID: types.NewID(), Name: "Other Tenant DB", Type: "MONGODB"})

	result, err := svc.ListByTenant(ctx, tid)
	require.NoError(t, err)
	assert.Len(t, result, 2, "should only return sources for the specified tenant")
}

func TestDataSourceService_Update(t *testing.T) {
	svc, _, eb := newTestDataSourceService()
	ctx := context.Background()

	ds, _ := svc.Create(ctx, CreateDataSourceInput{
		TenantID: types.NewID(),
		Name:     "Original Name",
		Type:     "POSTGRESQL",
	})

	updated, err := svc.Update(ctx, UpdateDataSourceInput{
		ID:   ds.ID,
		Name: "Updated Name",
	})

	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Len(t, eb.Events, 2, "should have created + updated events")
}

func TestDataSourceService_Delete(t *testing.T) {
	svc, repo, eb := newTestDataSourceService()
	ctx := context.Background()

	ds, _ := svc.Create(ctx, CreateDataSourceInput{
		TenantID: types.NewID(),
		Name:     "To Delete",
		Type:     "POSTGRESQL",
	})

	err := svc.Delete(ctx, ds.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, ds.ID)
	assert.Error(t, err, "should not find deleted data source")

	assert.Len(t, eb.Events, 2, "should have created + deleted events")
	assert.Equal(t, "datasource.deleted", eb.Events[1].Type)
}

func TestDataSourceService_Delete_NotFound(t *testing.T) {
	svc, _, _ := newTestDataSourceService()
	ctx := context.Background()

	err := svc.Delete(ctx, types.NewID())
	require.Error(t, err)
}
