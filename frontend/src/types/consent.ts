import type { ID, TenantEntity } from './common';

export interface ConsentWidget extends TenantEntity {
    name: string;
    type: 'BANNER' | 'PREFERENCE_CENTER' | 'PORTAL' | 'INLINE_FORM';
    domain: string;
    status: 'DRAFT' | 'ACTIVE' | 'PAUSED';
    config: WidgetConfig;
    embed_code: string;
    allowed_origins: string[];
    version: number;
    api_key?: string;
}

export interface WidgetConfig {
    theme: ThemeConfig;
    layout: 'BOTTOM_BAR' | 'TOP_BAR' | 'MODAL' | 'SIDEBAR' | 'FULL_PAGE';
    custom_css?: string;
    purpose_ids: ID[];
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
}

export interface ThemeConfig {
    primary_color: string;
    background_color: string;
    text_color: string;
    font_family: string;
    logo_url?: string;
    border_radius: string;
}

export interface CreateWidgetInput {
    name: string;
    type: ConsentWidget['type'];
    domain: string;
    allowed_origins: string[];
    config: WidgetConfig;
}

export interface UpdateWidgetInput {
    name?: string;
    domain?: string;
    allowed_origins?: string[];
    status?: ConsentWidget['status'];
    config?: WidgetConfig;
}

export interface RecordConsentRequest {
    widget_id: ID;
    decisions: { purpose_id: ID; granted: boolean }[];
    page_url: string;
}

export interface WithdrawConsentRequest {
    purpose_id: ID;
}

export interface ConsentNotice extends TenantEntity {
    series_id: ID;
    title: string;
    content: string;
    version: number;
    status: 'DRAFT' | 'PUBLISHED' | 'ARCHIVED';
    purposes: ID[];
    widget_ids: ID[];
    regulation: string;
    published_at?: string;
}

export interface CreateNoticeInput {
    title: string;
    content: string;
    purposes: ID[];
    regulation: string;
    series_id?: ID;
}

export interface UpdateNoticeInput {
    id: ID;
    title: string;
    content: string;
    purposes: ID[];
    regulation: string;
}

export interface BindNoticeInput {
    widget_ids: ID[];
}
