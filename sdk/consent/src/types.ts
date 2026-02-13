// =============================================================================
// DataLens Consent SDK — Type Definitions
// =============================================================================

/** Widget configuration returned by the API */
export interface WidgetConfig {
    id: string;
    name: string;
    type: WidgetType;
    domain: string;
    status: string;
    version: number;
    config: WidgetSettings;
}

export type WidgetType = 'BANNER' | 'PREFERENCE_CENTER' | 'INLINE_FORM' | 'PORTAL';
export type LayoutType = 'BOTTOM_BAR' | 'TOP_BAR' | 'MODAL' | 'SIDEBAR' | 'FULL_PAGE';

export interface WidgetSettings {
    theme: ThemeConfig;
    layout: LayoutType;
    custom_css?: string;

    // Behavior
    purpose_ids: string[];
    default_state: 'OPT_IN' | 'OPT_OUT';
    show_categories: boolean;
    granular_toggle: boolean;
    block_until_consent: boolean;

    // Content / i18n
    languages: string[];
    default_language: string;
    translations: Record<string, Record<string, string>>; // lang → key → text

    // Compliance
    regulation_ref: string;
    require_explicit: boolean;
    consent_expiry_days: number;

    // Script blocking
    blocked_script_patterns?: BlockedScriptPattern[];

    // Purposes (enriched by API)
    purposes?: PurposeConfig[];
}

export interface ThemeConfig {
    primary_color: string;
    background_color: string;
    text_color: string;
    font_family: string;
    logo_url?: string;
    border_radius: string;
}

export interface BlockedScriptPattern {
    pattern: string;
    purpose_id: string;
}

export interface PurposeConfig {
    id: string;
    name: string;
    description: string;
    is_essential: boolean;
}

/** Consent decision for a single purpose */
export interface ConsentDecision {
    purpose_id: string;
    granted: boolean;
}

/** Stored consent state */
export interface StoredConsent {
    widget_id: string;
    widget_version: number;
    decisions: ConsentDecision[];
    timestamp: string;
    expires_at: string;
}

/** Consent session payload sent to API */
export interface ConsentSessionPayload {
    widget_id: string;
    widget_version: number;
    decisions: ConsentDecision[];
    ip_address?: string;
    user_agent: string;
    page_url: string;
}

/** SDK initialization options */
export interface SDKOptions {
    widgetId: string;
    apiKey: string;
    apiUrl: string;
}
