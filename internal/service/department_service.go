package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"os"
	"time"

	"github.com/complyark/datalens/internal/domain/governance"
	"github.com/complyark/datalens/pkg/types"
)

// DepartmentService provides business logic for department management.
type DepartmentService struct {
	repo     governance.DepartmentRepository
	auditSvc *AuditService
	logger   *slog.Logger
}

// NewDepartmentService creates a new DepartmentService.
func NewDepartmentService(repo governance.DepartmentRepository, auditSvc *AuditService, logger *slog.Logger) *DepartmentService {
	return &DepartmentService{
		repo:     repo,
		auditSvc: auditSvc,
		logger:   logger.With("service", "department"),
	}
}

// CreateDepartmentRequest holds input for creating a department.
type CreateDepartmentRequest struct {
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	OwnerName           string   `json:"owner_name"`
	OwnerEmail          string   `json:"owner_email"`
	Responsibilities    []string `json:"responsibilities"`
	NotificationEnabled bool     `json:"notification_enabled"`
}

// UpdateDepartmentRequest holds input for updating a department.
type UpdateDepartmentRequest struct {
	Name                *string  `json:"name,omitempty"`
	Description         *string  `json:"description,omitempty"`
	OwnerName           *string  `json:"owner_name,omitempty"`
	OwnerEmail          *string  `json:"owner_email,omitempty"`
	Responsibilities    []string `json:"responsibilities,omitempty"`
	NotificationEnabled *bool    `json:"notification_enabled,omitempty"`
	IsActive            *bool    `json:"is_active,omitempty"`
}

// NotifyDepartmentRequest holds input for sending a notification to a department owner.
type NotifyDepartmentRequest struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// Create creates a new department for the tenant.
func (s *DepartmentService) Create(ctx context.Context, req CreateDepartmentRequest) (*governance.Department, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	if req.Name == "" {
		return nil, types.NewValidationError("name is required", nil)
	}

	dept := &governance.Department{
		TenantID:            tenantID,
		Name:                req.Name,
		Description:         req.Description,
		OwnerName:           req.OwnerName,
		OwnerEmail:          req.OwnerEmail,
		Responsibilities:    req.Responsibilities,
		NotificationEnabled: req.NotificationEnabled,
		IsActive:            true,
	}

	if err := s.repo.Create(ctx, dept); err != nil {
		return nil, fmt.Errorf("create department: %w", err)
	}

	// Audit log (fire-and-forget)
	userID, _ := types.UserIDFromContext(ctx)
	s.auditSvc.Log(ctx, userID, "DEPARTMENT_CREATE", "DEPARTMENT", dept.ID, nil,
		map[string]any{"name": dept.Name}, tenantID)

	s.logger.Info("department created",
		slog.String("tenant_id", tenantID.String()),
		slog.String("department_id", dept.ID.String()),
	)

	return dept, nil
}

// GetByID retrieves a department by ID.
func (s *DepartmentService) GetByID(ctx context.Context, id types.ID) (*governance.Department, error) {
	return s.repo.GetByID(ctx, id)
}

// List retrieves all departments for the tenant.
func (s *DepartmentService) List(ctx context.Context) ([]governance.Department, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}
	return s.repo.GetByTenant(ctx, tenantID)
}

// Update updates an existing department.
func (s *DepartmentService) Update(ctx context.Context, id types.ID, req UpdateDepartmentRequest) (*governance.Department, error) {
	dept, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldName := dept.Name

	if req.Name != nil {
		dept.Name = *req.Name
	}
	if req.Description != nil {
		dept.Description = *req.Description
	}
	if req.OwnerName != nil {
		dept.OwnerName = *req.OwnerName
	}
	if req.OwnerEmail != nil {
		dept.OwnerEmail = *req.OwnerEmail
	}
	if req.Responsibilities != nil {
		dept.Responsibilities = req.Responsibilities
	}
	if req.NotificationEnabled != nil {
		dept.NotificationEnabled = *req.NotificationEnabled
	}
	if req.IsActive != nil {
		dept.IsActive = *req.IsActive
	}

	dept.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, dept); err != nil {
		return nil, fmt.Errorf("update department: %w", err)
	}

	// Audit log
	userID, _ := types.UserIDFromContext(ctx)
	tenantID, _ := types.TenantIDFromContext(ctx)
	s.auditSvc.Log(ctx, userID, "DEPARTMENT_UPDATE", "DEPARTMENT", dept.ID,
		map[string]any{"name": oldName},
		map[string]any{"name": dept.Name}, tenantID)

	s.logger.Info("department updated",
		slog.String("department_id", id.String()),
	)

	return dept, nil
}

// Delete removes a department by ID.
func (s *DepartmentService) Delete(ctx context.Context, id types.ID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	userID, _ := types.UserIDFromContext(ctx)
	tenantID, _ := types.TenantIDFromContext(ctx)
	s.auditSvc.Log(ctx, userID, "DEPARTMENT_DELETE", "DEPARTMENT", id, nil, nil, tenantID)

	s.logger.Info("department deleted",
		slog.String("department_id", id.String()),
	)

	return nil
}

// Notify sends an email notification to the department owner.
func (s *DepartmentService) Notify(ctx context.Context, id types.ID, req NotifyDepartmentRequest) error {
	dept, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !dept.NotificationEnabled {
		return types.NewValidationError("notifications are disabled for this department", nil)
	}
	if dept.OwnerEmail == "" {
		return types.NewValidationError("department has no owner email configured", nil)
	}
	if req.Subject == "" {
		return types.NewValidationError("subject is required", nil)
	}
	if req.Body == "" {
		return types.NewValidationError("body is required", nil)
	}

	if err := s.sendEmail(dept.OwnerEmail, req.Subject, req.Body); err != nil {
		s.logger.Error("failed to send department notification",
			slog.String("department_id", id.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("send email: %w", err)
	}

	s.logger.Info("department notification sent",
		slog.String("department_id", id.String()),
	)

	return nil
}

// sendEmail sends an email using SMTP configuration from environment variables.
// This duplicates the pattern from NotificationService.sendEmail (which is unexported).
func (s *DepartmentService) sendEmail(to, subject, body string) error {
	from := os.Getenv("SMTP_FROM_EMAIL")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html\r\n\r\n%s",
		from, to, subject, body)

	auth := smtp.PlainAuth("", user, pass, host)
	return smtp.SendMail(host+":"+port, auth, from, []string{to}, []byte(msg))
}
