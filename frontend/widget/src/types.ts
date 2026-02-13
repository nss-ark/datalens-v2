export interface ConsentWidget {
    id: string;
    tenant_id: string;
    name: string;
    type: WidgetType;
    domain: string;
    status: WidgetStatus;
    config: WidgetConfig;
    embed_code: string;
    api_key: string;
    allowed_origins: string[];
    version: number;
}

export type WidgetType = 'BANNER' | 'PREFERENCE_CENTER' | 'PORTAL' | 'INLINE_FORM';

export type WidgetStatus = 'DRAFT' | 'ACTIVE' | 'PAUSED';

export interface WidgetConfig {
    theme: ThemeConfig;
    layout: LayoutType;
    custom_css?: string;
    purpose_ids: string[];
    default_state: 'OPT_IN' | 'OPT_OUT';
    show_categories: boolean;
    granular_toggle: boolean;
    block_until_consent: boolean;
    languages: string[];
    default_language: string;
    translations: Record<string, Record<string, string>>;
    regulation_ref: string;
    require_explicit: boolean;
    consent_expiry_days: number;
    purposes?: PurposeRef[]; // Enriched by API
}

export interface ThemeConfig {
    primary_color: string;
    background_color: string;
    text_color: string;
    font_family: string;
    logo_url?: string;
    border_radius: string;
}

export type LayoutType = 'BOTTOM_BAR' | 'TOP_BAR' | 'MODAL' | 'SIDEBAR' | 'FULL_PAGE';

export interface PurposeRef {
    id: string;
    name: string;
    description: string;
    is_essential: boolean;
}

export interface ConsentSession {
    widget_id: string;
    subject_id?: string; // If known/persisted
    decisions: ConsentDecision[];
    url: string;
    user_agent: string;
}

export interface ConsentDecision {
    purpose_id: string;
    granted: boolean;
}

export interface WidgetOptions {
    widgetId: string;
    apiBase?: string;
    debug?: boolean;
}
