// Package types provides universal types, enums, and value objects
// used across all bounded contexts in DataLens.
//
// These types are regulation-agnostic and form the shared vocabulary
// of the entire system.
package types

import (
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Base Types
// =============================================================================

// ID is the universal identifier type used throughout the system.
type ID = uuid.UUID

// NewID generates a new unique identifier.
func NewID() ID {
	return uuid.New()
}

// ParseID parses a string into an ID.
func ParseID(s string) (ID, error) {
	return uuid.Parse(s)
}

// =============================================================================
// Enums — Data Classification
// =============================================================================

// PIICategory classifies personal data into universal categories.
type PIICategory string

const (
	PIICategoryIdentity     PIICategory = "IDENTITY"
	PIICategoryContact      PIICategory = "CONTACT"
	PIICategoryFinancial    PIICategory = "FINANCIAL"
	PIICategoryHealth       PIICategory = "HEALTH"
	PIICategoryBiometric    PIICategory = "BIOMETRIC"
	PIICategoryGenetic      PIICategory = "GENETIC"
	PIICategoryLocation     PIICategory = "LOCATION"
	PIICategoryBehavioral   PIICategory = "BEHAVIORAL"
	PIICategoryProfessional PIICategory = "PROFESSIONAL"
	PIICategoryGovernmentID PIICategory = "GOVERNMENT_ID"
	PIICategoryMinor        PIICategory = "MINOR"
)

// PIIType identifies specific types of personal data.
type PIIType string

const (
	PIITypeName          PIIType = "NAME"
	PIITypeEmail         PIIType = "EMAIL"
	PIITypePhone         PIIType = "PHONE"
	PIITypeAddress       PIIType = "ADDRESS"
	PIITypeAadhaar       PIIType = "AADHAAR"
	PIITypePAN           PIIType = "PAN"
	PIITypePassport      PIIType = "PASSPORT"
	PIITypeSSN           PIIType = "SSN"
	PIITypeNationalID    PIIType = "NATIONAL_ID"
	PIITypeDOB           PIIType = "DATE_OF_BIRTH"
	PIITypeGender        PIIType = "GENDER"
	PIITypeBankAccount   PIIType = "BANK_ACCOUNT"
	PIITypeCreditCard    PIIType = "CREDIT_CARD"
	PIITypeIPAddress     PIIType = "IP_ADDRESS"
	PIITypeMACAddress    PIIType = "MAC_ADDRESS"
	PIITypeDeviceID      PIIType = "DEVICE_ID"
	PIITypeBiometric     PIIType = "BIOMETRIC"
	PIITypeMedicalRecord PIIType = "MEDICAL_RECORD"
	PIITypePhoto         PIIType = "PHOTO"
	PIITypeSignature     PIIType = "SIGNATURE"
)

// SensitivityLevel classifies data sensitivity.
type SensitivityLevel string

const (
	SensitivityLow      SensitivityLevel = "LOW"
	SensitivityMedium   SensitivityLevel = "MEDIUM"
	SensitivityHigh     SensitivityLevel = "HIGH"
	SensitivityCritical SensitivityLevel = "CRITICAL"
)

// =============================================================================
// Enums — Detection
// =============================================================================

// DetectionMethod indicates how PII was identified.
type DetectionMethod string

const (
	DetectionMethodAI        DetectionMethod = "AI"
	DetectionMethodRegex     DetectionMethod = "REGEX"
	DetectionMethodHeuristic DetectionMethod = "HEURISTIC"
	DetectionMethodIndustry  DetectionMethod = "INDUSTRY"
	DetectionMethodManual    DetectionMethod = "MANUAL"
)

// VerificationStatus tracks human verification state.
type VerificationStatus string

const (
	VerificationPending  VerificationStatus = "PENDING"
	VerificationVerified VerificationStatus = "VERIFIED"
	VerificationRejected VerificationStatus = "REJECTED"
)

// =============================================================================
// Enums — Compliance
// =============================================================================

// DSRType defines universal data subject request types.
type DSRType string

const (
	DSRTypeAccess      DSRType = "ACCESS"
	DSRTypeErasure     DSRType = "ERASURE"
	DSRTypeCorrection  DSRType = "CORRECTION"
	DSRTypePortability DSRType = "PORTABILITY"
	DSRTypeObjection   DSRType = "OBJECTION"
	DSRTypeRestriction DSRType = "RESTRICTION"
	DSRTypeNomination  DSRType = "NOMINATION"
)

// LegalBasis defines the lawful basis for data processing.
type LegalBasis string

const (
	LegalBasisConsent            LegalBasis = "CONSENT"
	LegalBasisContract           LegalBasis = "CONTRACT"
	LegalBasisLegalObligation    LegalBasis = "LEGAL_OBLIGATION"
	LegalBasisVitalInterest      LegalBasis = "VITAL_INTEREST"
	LegalBasisPublicInterest     LegalBasis = "PUBLIC_INTEREST"
	LegalBasisLegitimateInterest LegalBasis = "LEGITIMATE_INTEREST"
	LegalBasisEmployment         LegalBasis = "EMPLOYMENT"
)

// ConsentMechanism describes how consent was obtained.
type ConsentMechanism string

const (
	ConsentExplicit ConsentMechanism = "EXPLICIT"
	ConsentImplicit ConsentMechanism = "IMPLICIT"
	ConsentOptIn    ConsentMechanism = "OPT_IN"
	ConsentOptOut   ConsentMechanism = "OPT_OUT"
	ConsentGranular ConsentMechanism = "GRANULAR"
)

// =============================================================================
// Enums — Status
// =============================================================================

// Status is a generic lifecycle status.
type Status string

const (
	StatusActive    Status = "ACTIVE"
	StatusInactive  Status = "INACTIVE"
	StatusPending   Status = "PENDING"
	StatusDeleted   Status = "DELETED"
	StatusSuspended Status = "SUSPENDED"
)

// Priority classifies urgency.
type Priority string

const (
	PriorityLow    Priority = "LOW"
	PriorityMedium Priority = "MEDIUM"
	PriorityHigh   Priority = "HIGH"
	PriorityUrgent Priority = "URGENT"
)

// Severity classifies impact level.
type Severity string

const (
	SeverityInfo     Severity = "INFO"
	SeverityWarning  Severity = "WARNING"
	SeverityCritical Severity = "CRITICAL"
)

// =============================================================================
// Enums — Data Sources
// =============================================================================

// DataSourceType identifies supported data source types.
type DataSourceType string

const (
	DataSourcePostgreSQL   DataSourceType = "POSTGRESQL"
	DataSourceMySQL        DataSourceType = "MYSQL"
	DataSourceMongoDB      DataSourceType = "MONGODB"
	DataSourceSQLServer    DataSourceType = "SQLSERVER"
	DataSourceSnowflake    DataSourceType = "SNOWFLAKE"
	DataSourceS3           DataSourceType = "S3"
	DataSourceRDS          DataSourceType = "RDS"
	DataSourceDynamoDB     DataSourceType = "DYNAMODB"
	DataSourceAzureBlob    DataSourceType = "AZURE_BLOB"
	DataSourceAzureSQL     DataSourceType = "AZURE_SQL"
	DataSourceGoogleDrive  DataSourceType = "GOOGLE_DRIVE"
	DataSourceOneDrive     DataSourceType = "ONEDRIVE"
	DataSourceSalesforce   DataSourceType = "SALESFORCE"
	DataSourceMicrosoft365 DataSourceType = "MICROSOFT_365"
	DataSourceIMAP         DataSourceType = "IMAP"
	DataSourceFileSystem   DataSourceType = "FILE_SYSTEM"
	DataSourceAPI          DataSourceType = "API"
)

// =============================================================================
// Value Objects
// =============================================================================

// TimeRange represents a bounded time period.
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Metadata holds arbitrary key-value pairs.
type Metadata map[string]any

// Pagination holds pagination parameters.
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// PaginatedResult wraps a paginated response.
type PaginatedResult[T any] struct {
	Items      []T `json:"items"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalPages int `json:"total_pages"`
}

// =============================================================================
// Base Entity
// =============================================================================

// BaseEntity provides common fields for all domain entities.
type BaseEntity struct {
	ID        ID         `json:"id" db:"id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// TenantEntity extends BaseEntity with tenant isolation.
type TenantEntity struct {
	BaseEntity
	TenantID ID `json:"tenant_id" db:"tenant_id"`
}

// Ptr returns a pointer to the given value.
func Ptr[T any](v T) *T {
	return &v
}

// PortalTokenResponse represents a portal authentication token response.
type PortalTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}
