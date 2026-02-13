// Package consent defines the domain entities for the embeddable
// Consent Management widget, Data Principal Portal, consent history
// timeline, and Data Principal Rights (DPR) flows.
//
// This context provides the public-facing consent collection and
// rights management capability that companies embed into their
// digital touchpoints (websites, apps, kiosks, etc.).
package consent

import (
	"context"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// ConsentWidget — Embeddable consent collection widget configuration
// =============================================================================

// ConsentWidget is the configuration for an embeddable consent
// collection widget that companies deploy on their digital touchpoints.
type ConsentWidget struct {
	types.TenantEntity
	Name           string       `json:"name" db:"name"`     // e.g., "Website Banner", "Mobile App"
	Type           WidgetType   `json:"type" db:"type"`     // BANNER, PREFERENCE_CENTER, PORTAL, INLINE
	Domain         string       `json:"domain" db:"domain"` // Authorized domain: "*.company.com"
	Status         WidgetStatus `json:"status" db:"status"`
	Config         WidgetConfig `json:"config" db:"config"`                   // Visual + behavioral configuration
	EmbedCode      string       `json:"embed_code" db:"embed_code"`           // Generated <script> snippet
	APIKey         string       `json:"-" db:"api_key"`                       // Public key for this widget (hidden)
	AllowedOrigins []string     `json:"allowed_origins" db:"allowed_origins"` // CORS origins
	Version        int          `json:"version" db:"version"`                 // Config version (auto-incremented)
}

// WidgetType classifies the embed form factor.
type WidgetType string

const (
	WidgetTypeBanner           WidgetType = "BANNER"
	WidgetTypePreferenceCenter WidgetType = "PREFERENCE_CENTER"
	WidgetTypePortal           WidgetType = "PORTAL"      // Full Data Principal Portal (iframe)
	WidgetTypeInlineForm       WidgetType = "INLINE_FORM" // Inline consent form
)

// WidgetStatus tracks widget lifecycle.
type WidgetStatus string

const (
	WidgetStatusDraft  WidgetStatus = "DRAFT"
	WidgetStatusActive WidgetStatus = "ACTIVE"
	WidgetStatusPaused WidgetStatus = "PAUSED"
)

// =============================================================================
// WidgetConfig — Visual + behavioral configuration
// =============================================================================

// WidgetConfig holds the full configuration for a consent widget.
type WidgetConfig struct {
	// Visual
	Theme     ThemeConfig `json:"theme"`
	Layout    LayoutType  `json:"layout"`
	CustomCSS *string     `json:"custom_css,omitempty"`

	// Behavior
	PurposeIDs        []types.ID `json:"purpose_ids"`
	DefaultState      string     `json:"default_state"`       // "OPT_IN" or "OPT_OUT"
	ShowCategories    bool       `json:"show_categories"`     // Group purposes by category
	GranularToggle    bool       `json:"granular_toggle"`     // Per-purpose toggles
	BlockUntilConsent bool       `json:"block_until_consent"` // Block page access

	// Content
	Languages       []string                     `json:"languages"`        // ["en", "hi", "ta"]
	DefaultLanguage string                       `json:"default_language"` // "en"
	Translations    map[string]map[string]string `json:"translations"`     // lang → key → text

	// Compliance
	RegulationRef     string `json:"regulation_ref"`
	RequireExplicit   bool   `json:"require_explicit"`
	ConsentExpiryDays int    `json:"consent_expiry_days"` // Days until re-consent

	// Script blocking (SDK)
	BlockedScriptPatterns []BlockedScriptPattern `json:"blocked_script_patterns,omitempty"`
	Purposes              []PurposeRef           `json:"purposes,omitempty"`
}

// ThemeConfig holds visual styling for the widget.
type ThemeConfig struct {
	PrimaryColor    string  `json:"primary_color"`
	BackgroundColor string  `json:"background_color"`
	TextColor       string  `json:"text_color"`
	FontFamily      string  `json:"font_family"`
	LogoURL         *string `json:"logo_url,omitempty"`
	BorderRadius    string  `json:"border_radius"`
}

// BlockedScriptPattern maps a URL pattern to a purpose for script blocking.
type BlockedScriptPattern struct {
	Pattern   string `json:"pattern"`    // URL substring match, e.g. "google-analytics.com"
	PurposeID string `json:"purpose_id"` // Purpose this script belongs to
}

// PurposeRef is a lightweight purpose reference included in the widget config.
type PurposeRef struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsEssential bool   `json:"is_essential"`
}

// LayoutType defines how the widget appears.
type LayoutType string

const (
	LayoutBottomBar LayoutType = "BOTTOM_BAR"
	LayoutTopBar    LayoutType = "TOP_BAR"
	LayoutModal     LayoutType = "MODAL"
	LayoutSidebar   LayoutType = "SIDEBAR"
	LayoutFullPage  LayoutType = "FULL_PAGE" // For portal/iframe
)

// =============================================================================
// ConsentSession — A single consent interaction
// =============================================================================

// ConsentSession captures a single interaction with a consent widget.
// This is the lightweight, immutable record stored in the Control Centre.
type ConsentSession struct {
	types.BaseEntity
	TenantID  types.ID  `json:"tenant_id" db:"tenant_id"`
	WidgetID  types.ID  `json:"widget_id" db:"widget_id"`
	SubjectID *types.ID `json:"subject_id,omitempty" db:"subject_id"`

	// Consent decisions
	Decisions []ConsentDecision `json:"decisions" db:"decisions"`

	// Context
	IPAddress     string `json:"ip_address" db:"ip_address"`
	UserAgent     string `json:"user_agent" db:"user_agent"`
	PageURL       string `json:"page_url" db:"page_url"`
	WidgetVersion int    `json:"widget_version" db:"widget_version"`
	NoticeVersion string `json:"notice_version" db:"notice_version"`

	// Integrity — immutable proof
	Signature string `json:"signature" db:"signature"`
}

// ConsentDecision is a single purpose-level consent choice.
type ConsentDecision struct {
	PurposeID types.ID `json:"purpose_id"`
	Granted   bool     `json:"granted"`
}

// =============================================================================
// DataPrincipalProfile — Portal identity for data subjects
// =============================================================================

// DataPrincipalProfile represents a data principal's self-service
// portal identity, linking consent history and DPR requests.
type DataPrincipalProfile struct {
	types.BaseEntity
	TenantID           types.ID           `json:"tenant_id" db:"tenant_id"`
	Email              string             `json:"email" db:"email"`
	Phone              *string            `json:"phone,omitempty" db:"phone"`
	VerificationStatus VerificationStatus `json:"verification_status" db:"verification_status"`
	VerifiedAt         *time.Time         `json:"verified_at,omitempty" db:"verified_at"`
	VerificationMethod *string            `json:"verification_method,omitempty" db:"verification_method"` // EMAIL_OTP, PHONE_OTP

	// Links
	SubjectID *types.ID `json:"subject_id,omitempty" db:"subject_id"` // Links to compliance.DataSubject

	// Portal state
	LastAccessAt  *time.Time `json:"last_access_at,omitempty" db:"last_access_at"`
	PreferredLang string     `json:"preferred_lang" db:"preferred_lang"`
}

// VerificationStatus tracks identity verification for portal access.
type VerificationStatus string

const (
	VerificationStatusPending  VerificationStatus = "PENDING"
	VerificationStatusVerified VerificationStatus = "VERIFIED"
	VerificationStatusExpired  VerificationStatus = "EXPIRED"
)

// =============================================================================
// ConsentHistoryEntry — Immutable consent timeline entry
// =============================================================================

// ConsentHistoryEntry captures a single change in consent state,
// forming a chronological, immutable timeline of all consent activity.
type ConsentHistoryEntry struct {
	types.BaseEntity
	TenantID  types.ID  `json:"tenant_id" db:"tenant_id"`
	SubjectID types.ID  `json:"subject_id" db:"subject_id"`
	WidgetID  *types.ID `json:"widget_id,omitempty" db:"widget_id"`

	// What changed
	PurposeID      types.ID `json:"purpose_id" db:"purpose_id"`
	PurposeName    string   `json:"purpose_name" db:"purpose_name"` // Denormalized
	PreviousStatus *string  `json:"previous_status,omitempty" db:"previous_status"`
	NewStatus      string   `json:"new_status" db:"new_status"` // GRANTED, WITHDRAWN, EXPIRED

	// Context
	Source        string `json:"source" db:"source"` // BANNER, PORTAL, API, FORM
	IPAddress     string `json:"ip_address" db:"ip_address"`
	UserAgent     string `json:"user_agent" db:"user_agent"`
	NoticeVersion string `json:"notice_version" db:"notice_version"`

	// Integrity
	Signature string `json:"signature" db:"signature"`
}

// =============================================================================
// DPR — Data Principal Rights Request (portal-initiated)
// =============================================================================

// DPRRequest is a Data Principal Rights request submitted through
// the portal. It wraps a compliance-level DSR with portal-specific
// fields (verification, appeal, guardian consent).
type DPRRequest struct {
	types.BaseEntity
	TenantID  types.ID  `json:"tenant_id" db:"tenant_id"`
	ProfileID types.ID  `json:"profile_id" db:"profile_id"`   // DataPrincipalProfile
	DSRID     *types.ID `json:"dsr_id,omitempty" db:"dsr_id"` // Links to compliance.DSR once created

	// Request
	Type        string     `json:"type" db:"type"` // ACCESS, ERASURE, CORRECTION, NOMINATION, PORTABILITY
	Description string     `json:"description,omitempty" db:"description"`
	Status      DPRStatus  `json:"status" db:"status"`
	SubmittedAt time.Time  `json:"submitted_at" db:"submitted_at"`
	Deadline    *time.Time `json:"deadline,omitempty" db:"deadline"`

	// Identity verification
	VerifiedAt      *time.Time `json:"verified_at,omitempty" db:"verified_at"`
	VerificationRef *string    `json:"verification_ref,omitempty" db:"verification_ref"`

	// Guardian (for minors — DPDPA Section 9)
	IsMinor          bool    `json:"is_minor" db:"is_minor"`
	GuardianName     *string `json:"guardian_name,omitempty" db:"guardian_name"`
	GuardianEmail    *string `json:"guardian_email,omitempty" db:"guardian_email"`
	GuardianRelation *string `json:"guardian_relation,omitempty" db:"guardian_relation"` // PARENT, GUARDIAN
	GuardianVerified bool    `json:"guardian_verified" db:"guardian_verified"`

	// Resolution
	CompletedAt     *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	ResponseSummary *string    `json:"response_summary,omitempty" db:"response_summary"`
	DownloadURL     *string    `json:"download_url,omitempty" db:"download_url"` // For ACCESS requests

	// Appeal (DPDPA Section 18 — right to appeal)
	AppealOf     *types.ID `json:"appeal_of,omitempty" db:"appeal_of"` // Original DPR if this is an appeal
	AppealReason *string   `json:"appeal_reason,omitempty" db:"appeal_reason"`
	IsEscalated  bool      `json:"is_escalated" db:"is_escalated"`
	EscalatedTo  *string   `json:"escalated_to,omitempty" db:"escalated_to"` // DPA authority
}

// DPRStatus tracks portal-side request lifecycle.
type DPRStatus string

const (
	DPRStatusSubmitted       DPRStatus = "SUBMITTED"
	DPRStatusPendingVerify   DPRStatus = "PENDING_VERIFICATION"
	DPRStatusVerified        DPRStatus = "VERIFIED"
	DPRStatusInProgress      DPRStatus = "IN_PROGRESS"
	DPRStatusCompleted       DPRStatus = "COMPLETED"
	DPRStatusRejected        DPRStatus = "REJECTED"
	DPRStatusAppealed        DPRStatus = "APPEALED"
	DPRStatusEscalated       DPRStatus = "ESCALATED"
	DPRStatusGuardianPending DPRStatus = "GUARDIAN_PENDING" // Awaiting guardian verification
)

// =============================================================================
// Repository Interfaces
// =============================================================================

// ConsentWidgetRepository defines persistence for consent widgets.
type ConsentWidgetRepository interface {
	Create(ctx context.Context, w *ConsentWidget) error
	GetByID(ctx context.Context, id types.ID) (*ConsentWidget, error)
	GetByTenant(ctx context.Context, tenantID types.ID) ([]ConsentWidget, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*ConsentWidget, error)
	Update(ctx context.Context, w *ConsentWidget) error
	Delete(ctx context.Context, id types.ID) error
}

// ConsentSessionRepository defines persistence for consent sessions.
type ConsentSessionRepository interface {
	Create(ctx context.Context, s *ConsentSession) error
	GetBySubject(ctx context.Context, tenantID, subjectID types.ID) ([]ConsentSession, error)
	GetConversionStats(ctx context.Context, tenantID types.ID, from, to time.Time, interval string) ([]ConversionStat, error)
	GetPurposeStats(ctx context.Context, tenantID types.ID, from, to time.Time) ([]PurposeStat, error)
	GetExpiringSessions(ctx context.Context, withinDays int) ([]ConsentSession, error)
}

// ConversionStat represents consent conversion metrics over time.
type ConversionStat struct {
	Date           time.Time `json:"date"`
	TotalSessions  int       `json:"total_sessions"`
	OptInCount     int       `json:"opt_in_count"`
	OptOutCount    int       `json:"opt_out_count"`
	ConversionRate float64   `json:"conversion_rate"`
}

// PurposeStat represents grant/deny counts per purpose.
type PurposeStat struct {
	PurposeID    types.ID `json:"purpose_id"`
	GrantedCount int      `json:"granted_count"`
	DeniedCount  int      `json:"denied_count"`
}

// DataPrincipalProfileRepository defines persistence for portal profiles.
type DataPrincipalProfileRepository interface {
	Create(ctx context.Context, p *DataPrincipalProfile) error
	GetByID(ctx context.Context, id types.ID) (*DataPrincipalProfile, error)
	GetByEmail(ctx context.Context, tenantID types.ID, email string) (*DataPrincipalProfile, error)
	Update(ctx context.Context, p *DataPrincipalProfile) error
}

// ConsentHistoryRepository defines persistence for the consent timeline.
type ConsentHistoryRepository interface {
	Create(ctx context.Context, entry *ConsentHistoryEntry) error
	GetBySubject(ctx context.Context, tenantID, subjectID types.ID, pagination types.Pagination) (*types.PaginatedResult[ConsentHistoryEntry], error)
	GetByPurpose(ctx context.Context, tenantID, purposeID types.ID) ([]ConsentHistoryEntry, error)
	GetLatestState(ctx context.Context, tenantID, subjectID, purposeID types.ID) (*ConsentHistoryEntry, error)
}

// DPRRequestRepository defines persistence for portal DPR requests.
type DPRRequestRepository interface {
	Create(ctx context.Context, r *DPRRequest) error
	GetByID(ctx context.Context, id types.ID) (*DPRRequest, error)
	GetByProfile(ctx context.Context, profileID types.ID) ([]DPRRequest, error)
	GetByTenant(ctx context.Context, tenantID types.ID, pagination types.Pagination) (*types.PaginatedResult[DPRRequest], error)
	Update(ctx context.Context, r *DPRRequest) error
}

// =============================================================================
// Consent Renewal
// =============================================================================

// ConsentRenewalLog tracks the lifecycle of consent renewal requests.
type ConsentRenewalLog struct {
	types.BaseEntity
	TenantID       types.ID   `json:"tenant_id" db:"tenant_id"`
	SubjectID      types.ID   `json:"subject_id" db:"subject_id"`
	PurposeID      types.ID   `json:"purpose_id" db:"purpose_id"`
	OriginalExpiry time.Time  `json:"original_expiry" db:"original_expiry"`
	Status         string     `json:"status" db:"status"` // PENDING, RENEWED, LAPSED, IGNORED
	ReminderSentAt *time.Time `json:"reminder_sent_at,omitempty" db:"reminder_sent_at"`
	RenewedAt      *time.Time `json:"renewed_at,omitempty" db:"renewed_at"`
}

// ConsentRenewalRepository defines persistence for consent renewal logs.
type ConsentRenewalRepository interface {
	Create(ctx context.Context, l *ConsentRenewalLog) error
	GetBySubject(ctx context.Context, tenantID, subjectID types.ID) ([]ConsentRenewalLog, error)
	Update(ctx context.Context, l *ConsentRenewalLog) error
}

// =============================================================================
// ConsentNotice — Privacy Notice Management
// =============================================================================

// ConsentNotice represents a versioned privacy notice displayed to users.
type ConsentNotice struct {
	types.TenantEntity
	SeriesID    types.ID     `json:"series_id" db:"series_id"` // Groups versions of the same notice
	Title       string       `json:"title" db:"title"`
	Content     string       `json:"content" db:"content"`       // Rich text (markdown/html)
	Version     int          `json:"version" db:"version"`       // Auto-incremented on publish
	Status      NoticeStatus `json:"status" db:"status"`         // DRAFT, PUBLISHED, ARCHIVED
	Purposes    []types.ID   `json:"purposes" db:"purposes"`     // Linked purpose IDs
	WidgetIDs   []types.ID   `json:"widget_ids" db:"widget_ids"` // Bound widgets
	Regulation  string       `json:"regulation" db:"regulation"` // e.g., "DPDPA_2023"
	PublishedAt *time.Time   `json:"published_at,omitempty" db:"published_at"`
}

// NoticeStatus tracks the lifecycle of a privacy notice.
type NoticeStatus string

const (
	NoticeStatusDraft     NoticeStatus = "DRAFT"
	NoticeStatusPublished NoticeStatus = "PUBLISHED"
	NoticeStatusArchived  NoticeStatus = "ARCHIVED"
)

// ConsentNoticeTranslation holds translations for a specific notice version.
type ConsentNoticeTranslation struct {
	types.BaseEntity
	NoticeID          types.ID   `json:"notice_id" db:"notice_id"`
	NoticeVersion     int        `json:"notice_version" db:"notice_version"`
	LanguageCode      string     `json:"language_code" db:"language_code"`           // ISO 639-1 (en, hi, ta, etc.)
	TranslatedText    string     `json:"translated_text" db:"translated_text"`       // The actual content
	TranslationSource string     `json:"translation_source" db:"translation_source"` // INDICTRANS2 | MANUAL | UNSUPPORTED
	IsRTL             bool       `json:"is_rtl" db:"is_rtl"`                         // Text direction
	TranslatedAt      time.Time  `json:"translated_at" db:"translated_at"`
	ReviewedBy        *types.ID  `json:"reviewed_by,omitempty" db:"reviewed_by"`
	ReviewedAt        *time.Time `json:"reviewed_at,omitempty" db:"reviewed_at"`
}

// ConsentNoticeTranslationRepository defines persistence for notice translations.
type ConsentNoticeTranslationRepository interface {
	SaveTranslation(ctx context.Context, t *ConsentNoticeTranslation) error
	GetByNoticeAndVersion(ctx context.Context, noticeID types.ID, version int) ([]ConsentNoticeTranslation, error)
	GetByNoticeAndLang(ctx context.Context, noticeID types.ID, version int, lang string) (*ConsentNoticeTranslation, error)
	Upsert(ctx context.Context, t *ConsentNoticeTranslation) error // For overrides
}

// ConsentNoticeRepository defines persistence for privacy notices.
type ConsentNoticeRepository interface {
	Create(ctx context.Context, n *ConsentNotice) error
	GetByID(ctx context.Context, id types.ID) (*ConsentNotice, error)
	GetByTenant(ctx context.Context, tenantID types.ID) ([]ConsentNotice, error)
	Update(ctx context.Context, n *ConsentNotice) error
	// Publish promotes a draft to published status and increments version.
	Publish(ctx context.Context, id types.ID) (int, error) // Returns new version number
	Archive(ctx context.Context, id types.ID) error
	// BindToWidgets links a notice to specific widgets.
	BindToWidgets(ctx context.Context, noticeID types.ID, widgetIDs []types.ID) error

	// Translation methods
	GetLatestVersion(ctx context.Context, seriesID types.ID) (int, error)
}
