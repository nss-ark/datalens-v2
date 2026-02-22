package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/domain/governance"
	"github.com/complyark/datalens/pkg/types"
)

// RoPAService provides business logic for RoPA (Record of Processing Activities) management.
type RoPAService struct {
	ropaRepo       compliance.RoPARepository
	purposeRepo    governance.PurposeRepository
	dsRepo         discovery.DataSourceRepository
	retentionRepo  compliance.RetentionPolicyRepository
	thirdPartyRepo governance.ThirdPartyRepository
	auditSvc       *AuditService
	logger         *slog.Logger
}

// NewRoPAService creates a new RoPAService.
func NewRoPAService(
	ropaRepo compliance.RoPARepository,
	purposeRepo governance.PurposeRepository,
	dsRepo discovery.DataSourceRepository,
	retentionRepo compliance.RetentionPolicyRepository,
	thirdPartyRepo governance.ThirdPartyRepository,
	auditSvc *AuditService,
	logger *slog.Logger,
) *RoPAService {
	return &RoPAService{
		ropaRepo:       ropaRepo,
		purposeRepo:    purposeRepo,
		dsRepo:         dsRepo,
		retentionRepo:  retentionRepo,
		thirdPartyRepo: thirdPartyRepo,
		auditSvc:       auditSvc,
		logger:         logger.With("service", "ropa"),
	}
}

// SaveEditRequest holds input for user edits to RoPA.
type SaveEditRequest struct {
	Content       compliance.RoPAContent `json:"content"`
	ChangeSummary string                 `json:"change_summary"`
}

// PublishRequest holds input for publishing a RoPA version.
type PublishRequest struct {
	ID types.ID `json:"id"`
}

// Generate auto-generates a new RoPA version from live data.
func (s *RoPAService) Generate(ctx context.Context) (*compliance.RoPAVersion, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	// Query all live data
	purposes, err := s.purposeRepo.GetByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get purposes: %w", err)
	}

	dataSources, err := s.dsRepo.GetByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get data sources: %w", err)
	}

	retentionPolicies, err := s.retentionRepo.GetByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get retention policies: %w", err)
	}

	thirdParties, err := s.thirdPartyRepo.GetByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get third parties: %w", err)
	}

	// Build purpose name lookup for retention policies
	purposeNameMap := make(map[types.ID]string)
	for _, p := range purposes {
		purposeNameMap[p.ID] = p.Name
	}

	// Build RoPAContent
	content := compliance.RoPAContent{
		OrganizationName: "", // Populated by tenant info if available
		GeneratedAt:      time.Now(),
	}

	for _, p := range purposes {
		content.Purposes = append(content.Purposes, compliance.RoPAPurpose{
			ID:          p.ID,
			Name:        p.Name,
			Code:        p.Code,
			LegalBasis:  string(p.LegalBasis),
			Description: p.Description,
			IsActive:    p.IsActive,
		})
	}

	for _, ds := range dataSources {
		content.DataSources = append(content.DataSources, compliance.RoPADataSource{
			ID:       ds.ID,
			Name:     ds.Name,
			Type:     string(ds.Type),
			IsActive: ds.Status == discovery.ConnectionStatusConnected,
		})
	}

	// Collect unique data categories from retention policies
	categoriesSet := make(map[string]bool)
	for _, rp := range retentionPolicies {
		purposeName := purposeNameMap[rp.PurposeID]
		content.RetentionPolicies = append(content.RetentionPolicies, compliance.RoPARetention{
			ID:               rp.ID,
			PurposeName:      purposeName,
			MaxRetentionDays: rp.MaxRetentionDays,
			DataCategories:   rp.DataCategories,
			AutoErase:        rp.AutoErase,
		})
		for _, cat := range rp.DataCategories {
			categoriesSet[cat] = true
		}
	}

	for _, tp := range thirdParties {
		content.ThirdParties = append(content.ThirdParties, compliance.RoPAThirdParty{
			ID:      tp.ID,
			Name:    tp.Name,
			Type:    string(tp.Type),
			Country: tp.Country,
		})
	}

	// Flatten categories set to slice
	for cat := range categoriesSet {
		content.DataCategories = append(content.DataCategories, cat)
	}

	// Version logic: get latest, bump minor
	latest, err := s.ropaRepo.GetLatest(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get latest ropa: %w", err)
	}

	var versionStr string
	if latest == nil {
		versionStr = "1.0"
	} else {
		major, minor := parseVersion(latest.Version)
		versionStr = formatVersion(major, minor+1)
	}

	version := &compliance.RoPAVersion{
		TenantID:    tenantID,
		Version:     versionStr,
		GeneratedBy: "auto",
		Status:      compliance.RoPAStatusDraft,
		Content:     content,
	}

	if err := s.ropaRepo.Create(ctx, version); err != nil {
		return nil, fmt.Errorf("create ropa version: %w", err)
	}

	// Audit log
	s.auditSvc.Log(ctx, types.ID{}, "ROPA_GENERATE", "ROPA", version.ID, nil, map[string]any{"version": version.Version}, tenantID)

	s.logger.Info("RoPA version generated",
		slog.String("tenant_id", tenantID.String()),
		slog.String("version", versionStr),
	)

	return version, nil
}

// SaveEdit creates a new minor version with user-edited content.
func (s *RoPAService) SaveEdit(ctx context.Context, req SaveEditRequest) (*compliance.RoPAVersion, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	userID, _ := types.UserIDFromContext(ctx)

	// Get latest, bump minor version
	latest, err := s.ropaRepo.GetLatest(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get latest ropa: %w", err)
	}

	var versionStr string
	if latest == nil {
		versionStr = "1.0"
	} else {
		major, minor := parseVersion(latest.Version)
		versionStr = formatVersion(major, minor+1)
	}

	version := &compliance.RoPAVersion{
		TenantID:      tenantID,
		Version:       versionStr,
		GeneratedBy:   userID.String(),
		Status:        compliance.RoPAStatusDraft,
		Content:       req.Content,
		ChangeSummary: req.ChangeSummary,
	}

	if err := s.ropaRepo.Create(ctx, version); err != nil {
		return nil, fmt.Errorf("create ropa version: %w", err)
	}

	// Audit log
	s.auditSvc.Log(ctx, userID, "ROPA_EDIT", "ROPA", version.ID, nil, map[string]any{"version": version.Version, "change_summary": req.ChangeSummary}, tenantID)

	s.logger.Info("RoPA version edited",
		slog.String("tenant_id", tenantID.String()),
		slog.String("version", versionStr),
		slog.String("user_id", userID.String()),
	)

	return version, nil
}

// Publish marks a version as PUBLISHED and archives any existing published version.
func (s *RoPAService) Publish(ctx context.Context, id types.ID) error {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return types.NewForbiddenError("tenant context required")
	}
	userID, _ := types.UserIDFromContext(ctx)

	// Find and archive any existing PUBLISHED version for this tenant
	// We do this by getting the latest and checking; a more robust approach
	// would query by status, but for now we iterate versions
	// Simple approach: try to find the published one and archive it
	// Since we don't have a GetByStatus, we'll use ListVersions and filter
	allVersions, err := s.ropaRepo.ListVersions(ctx, tenantID, types.Pagination{Page: 1, PageSize: 1000})
	if err != nil {
		return fmt.Errorf("list ropa versions: %w", err)
	}

	for _, v := range allVersions.Items {
		if v.Status == compliance.RoPAStatusPublished {
			if err := s.ropaRepo.UpdateStatus(ctx, v.ID, compliance.RoPAStatusArchived); err != nil {
				return fmt.Errorf("archive existing published version: %w", err)
			}
		}
	}

	// Set target version to PUBLISHED
	if err := s.ropaRepo.UpdateStatus(ctx, id, compliance.RoPAStatusPublished); err != nil {
		return fmt.Errorf("publish ropa version: %w", err)
	}

	// Audit log
	s.auditSvc.Log(ctx, userID, "ROPA_PUBLISH", "ROPA", id, nil, map[string]any{"status": "PUBLISHED"}, tenantID)

	s.logger.Info("RoPA version published",
		slog.String("tenant_id", tenantID.String()),
		slog.String("version_id", id.String()),
	)

	return nil
}

// PromoteMajor creates a new major version by bumping the major number.
func (s *RoPAService) PromoteMajor(ctx context.Context) (*compliance.RoPAVersion, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}
	userID, _ := types.UserIDFromContext(ctx)

	latest, err := s.ropaRepo.GetLatest(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get latest ropa: %w", err)
	}
	if latest == nil {
		return nil, types.NewValidationError("no existing RoPA version to promote", nil)
	}

	major, _ := parseVersion(latest.Version)
	versionStr := formatVersion(major+1, 0)

	version := &compliance.RoPAVersion{
		TenantID:      tenantID,
		Version:       versionStr,
		GeneratedBy:   userID.String(),
		Status:        compliance.RoPAStatusDraft,
		Content:       latest.Content, // Copy content from latest
		ChangeSummary: fmt.Sprintf("Major version promotion from %s to %s", latest.Version, versionStr),
	}

	if err := s.ropaRepo.Create(ctx, version); err != nil {
		return nil, fmt.Errorf("create ropa version: %w", err)
	}

	// Audit log
	s.auditSvc.Log(ctx, userID, "ROPA_PROMOTE", "ROPA", version.ID, nil, map[string]any{"version": version.Version, "from_version": latest.Version}, tenantID)

	s.logger.Info("RoPA major version promoted",
		slog.String("tenant_id", tenantID.String()),
		slog.String("version", versionStr),
	)

	return version, nil
}

// GetLatest retrieves the latest RoPA version for the tenant.
func (s *RoPAService) GetLatest(ctx context.Context) (*compliance.RoPAVersion, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}
	return s.ropaRepo.GetLatest(ctx, tenantID)
}

// GetByVersion retrieves a specific RoPA version by version string.
func (s *RoPAService) GetByVersion(ctx context.Context, version string) (*compliance.RoPAVersion, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}
	return s.ropaRepo.GetByVersion(ctx, tenantID, version)
}

// ListVersions retrieves paginated RoPA versions for the tenant.
func (s *RoPAService) ListVersions(ctx context.Context, pagination types.Pagination) (*types.PaginatedResult[compliance.RoPAVersion], error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}
	return s.ropaRepo.ListVersions(ctx, tenantID, pagination)
}

// parseVersion extracts major and minor version numbers from a version string.
func parseVersion(v string) (major, minor int) {
	parts := strings.Split(v, ".")
	major, _ = strconv.Atoi(parts[0])
	if len(parts) > 1 {
		minor, _ = strconv.Atoi(parts[1])
	}
	return
}

// formatVersion creates a version string from major and minor numbers.
func formatVersion(major, minor int) string {
	return fmt.Sprintf("%d.%d", major, minor)
}
