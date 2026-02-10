// Package dpdpa implements the ComplianceAdapter for the Digital Personal
// Data Protection Act, 2023 (India).
package dpdpa

import (
	"context"
	"time"

	"github.com/complyark/datalens/internal/adapter"
	"github.com/complyark/datalens/pkg/types"
)

// Adapter implements the ComplianceAdapter interface for DPDPA.
type Adapter struct{}

// New creates a new DPDPA compliance adapter.
func New() *Adapter {
	return &Adapter{}
}

func (a *Adapter) Name() string { return "Digital Personal Data Protection Act, 2023" }
func (a *Adapter) Code() string { return "DPDPA" }

// --- DSR Configuration ---

func (a *Adapter) SupportedDSRTypes() []types.DSRType {
	return []types.DSRType{
		types.DSRTypeAccess,
		types.DSRTypeErasure,
		types.DSRTypeCorrection,
		types.DSRTypeNomination,
	}
}

func (a *Adapter) DSRDeadline(dsrType types.DSRType, receivedAt time.Time) time.Time {
	// DPDPA mandates response within 30 days
	return receivedAt.AddDate(0, 0, 30)
}

func (a *Adapter) DSRRequiresVerification(dsrType types.DSRType) bool {
	// All DPDPA requests require identity verification
	return true
}

// --- Consent Configuration ---

func (a *Adapter) ConsentRequirements() adapter.ConsentConfig {
	return adapter.ConsentConfig{
		RequiresExplicit: true,
		GranularRequired: true,
		MaxAgeMonths:     0, // No explicit max, but should be reasonable
		WithdrawalMustBe: "EASY",
		MinAge:           18,
		GuardianRequired: true,
	}
}

// --- Breach Configuration ---

func (a *Adapter) BreachNotificationDeadline(detectedAt time.Time) time.Time {
	// CERT-In requires notification within 6 hours
	return detectedAt.Add(6 * time.Hour)
}

func (a *Adapter) BreachRequiresSubjectNotification(severity types.Severity) bool {
	// Notify subjects for all breaches involving personal data
	return true
}

// --- Classification ---

func (a *Adapter) PIICategories() []types.PIICategory {
	return []types.PIICategory{
		types.PIICategoryIdentity,
		types.PIICategoryContact,
		types.PIICategoryFinancial,
		types.PIICategoryHealth,
		types.PIICategoryBiometric,
		types.PIICategoryGenetic,
		types.PIICategoryLocation,
		types.PIICategoryGovernmentID,
		types.PIICategoryMinor,
	}
}

func (a *Adapter) SensitiveCategories() []types.PIICategory {
	return []types.PIICategory{
		types.PIICategoryHealth,
		types.PIICategoryBiometric,
		types.PIICategoryGenetic,
		types.PIICategoryFinancial,
		types.PIICategoryMinor,
	}
}

// --- Rights ---

func (a *Adapter) DataSubjectRights() []adapter.Right {
	return []adapter.Right{
		{Code: "ACCESS", Name: "Right to Access", Description: "Right to obtain confirmation and access to personal data", Mandatory: true},
		{Code: "CORRECTION", Name: "Right to Correction", Description: "Right to correct inaccurate or misleading personal data", Mandatory: true},
		{Code: "ERASURE", Name: "Right to Erasure", Description: "Right to have personal data erased when no longer necessary", Mandatory: true},
		{Code: "NOMINATION", Name: "Right to Nominate", Description: "Right to nominate a person to exercise rights in case of death/incapacity", Mandatory: true},
		{Code: "GRIEVANCE", Name: "Right to Grievance Redressal", Description: "Right to have grievances addressed by the Data Fiduciary", Mandatory: true},
	}
}

// --- Validation ---

func (a *Adapter) ValidateCompliance(ctx context.Context, tenantID types.ID) (*adapter.ComplianceReport, error) {
	// TODO: Implement comprehensive compliance validation
	return &adapter.ComplianceReport{
		Regulation:   a.Code(),
		OverallScore: 0.0,
		Status:       adapter.CompliancePartial,
		Issues:       []adapter.ComplianceIssue{},
		GeneratedAt:  time.Now().UTC(),
	}, nil
}

// Ensure Adapter implements ComplianceAdapter at compile time.
var _ adapter.ComplianceAdapter = (*Adapter)(nil)
