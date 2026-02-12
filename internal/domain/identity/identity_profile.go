package identity

import (
	"context"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// IdentityProfile — Identity Assurance Level (IAL) profile
// =============================================================================

// IdentityProfile represents the identity verification status of a Data Principal.
// It links a Subject (Compliance) to their verified identity attributes.
type IdentityProfile struct {
	types.TenantEntity
	SubjectID           types.ID           `json:"subject_id" db:"subject_id"`
	AssuranceLevel      AssuranceLevel     `json:"assurance_level" db:"assurance_level"`
	VerificationStatus  VerificationStatus `json:"verification_status" db:"verification_status"`
	Documents           []IdentityDocument `json:"documents,omitempty" db:"documents"`
	LastVerifiedAt      *time.Time         `json:"last_verified_at,omitempty" db:"last_verified_at"`
	NextVerificationDue *time.Time         `json:"next_verification_due,omitempty" db:"next_verification_due"`
}

// AssuranceLevel based on NIST 800-63-3 / eIDAS / India Stack
type AssuranceLevel string

const (
	AssuranceLevelNone        AssuranceLevel = "NONE"        // No verification
	AssuranceLevelBasic       AssuranceLevel = "BASIC"       // Email/Phone OTP (IAL1)
	AssuranceLevelSubstantial AssuranceLevel = "SUBSTANTIAL" // Gov ID / DigiLocker (IAL2)
	AssuranceLevelHigh        AssuranceLevel = "HIGH"        // Biometric / In-person (IAL3)
)

// VerificationStatus tracks the overall profile status
type VerificationStatus string

const (
	VerificationStatusPending  VerificationStatus = "PENDING"
	VerificationStatusVerified VerificationStatus = "VERIFIED"
	VerificationStatusFailed   VerificationStatus = "FAILED"
	VerificationStatusExpired  VerificationStatus = "EXPIRED"
)

// IdentityDocument represents a verified document linked to the profile.
type IdentityDocument struct {
	Type        DocumentType `json:"type"`         // AADHAAR, PAN, DRIVING_LICENSE, PASSPORT
	ReferenceID string       `json:"reference_id"` // Last 4 digits or masked ID
	Issuer      string       `json:"issuer"`       // "DigiLocker", "UIDAI", "ITD"
	VerifiedAt  time.Time    `json:"verified_at"`
	Metadata    types.JSON   `json:"metadata"` // Provider-specific metadata
}

// DocumentType classifies the identity document
type DocumentType string

const (
	DocumentTypeAadhaar        DocumentType = "AADHAAR"
	DocumentTypePAN            DocumentType = "PAN"
	DocumentTypeDrivingLicense DocumentType = "DRIVING_LICENSE"
	DocumentTypePassport       DocumentType = "PASSPORT"
	DocumentTypeVoterID        DocumentType = "VOTER_ID"
)

// =============================================================================
// Interfaces
// =============================================================================

// IdentityProfileRepository defines persistence for identity profiles.
type IdentityProfileRepository interface {
	Create(ctx context.Context, profile *IdentityProfile) error
	GetBySubject(ctx context.Context, tenantID, subjectID types.ID) (*IdentityProfile, error)
	Update(ctx context.Context, profile *IdentityProfile) error
}

// IdentityProvider defines the interface for external identity providers (e.g., DigiLocker).
type IdentityProvider interface {
	// Name returns the provider name (e.g., "DigiLocker")
	Name() string

	// GetAuthorizationURL returns the URL to start the OAuth flow
	GetAuthorizationURL(state string) string

	// ExchangeToken exchanges the authorization code for an access token
	ExchangeToken(ctx context.Context, code string) (*TokenResponse, error)

	// GetUserProfile fetches the basic user profile from the provider
	GetUserProfile(ctx context.Context, token string) (*UserProfile, error)

	// FetchDocuments retrieves available documents from the provider
	FetchDocuments(ctx context.Context, token string) ([]IdentityDocument, error)
}

// TokenResponse represents the OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// UserProfile represents the normalized user profile from a provider
type UserProfile struct {
	ProviderID  string `json:"provider_id"`
	Name        string `json:"name"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	DateOfBirth string `json:"date_of_birth,omitempty"`
	Gender      string `json:"gender,omitempty"`
}

// VerificationPriorities defines tenant-specific verification rules
type VerificationPriorities struct {
	DSRRequirement   AssuranceLevel `json:"dsr_requirement"`   // Min IAL for DSRs
	LoginRequirement AssuranceLevel `json:"login_requirement"` // Min IAL for Portal Login
}
