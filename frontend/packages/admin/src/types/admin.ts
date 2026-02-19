import type { BaseEntity, ID } from '@datalens/shared';
import type { DSR } from './dsr';

export interface TenantSettings {
    default_regulation: string;
    enabled_regulations: string[];
    retention_days: number;
    enable_ai: boolean;
    ai_provider?: string;
}

export interface Tenant extends BaseEntity {
    id: ID;
    name: string;
    domain: string;
    industry: string;
    country: string;
    status: 'ACTIVE' | 'SUSPENDED' | 'DELETED';
    plan: 'FREE' | 'STARTER' | 'PROFESSIONAL' | 'ENTERPRISE';
    settings: TenantSettings;
    created_at: string;
    updated_at: string;
}

export interface CreateTenantInput {
    name: string;
    domain: string;
    admin_email: string;
    plan: 'FREE' | 'STARTER' | 'PROFESSIONAL' | 'ENTERPRISE';
}

export interface AdminStats {
    total_tenants: number;
    active_tenants: number;
    total_users: number;
    total_dsr_requests: number; // Placeholder for future
}

export interface AdminUser {
    id: string;
    tenant_id: string;
    email: string;
    name: string;
    status: 'ACTIVE' | 'INVITED' | 'SUSPENDED';
    role_ids: string[];
    mfa_enabled: boolean;
    last_login_at: string | null;
    created_at: string;
}

export interface AdminRole {
    id: string;
    name: string;
    description: string;
    is_system: boolean;
}

// Admin DSR type (extending DSR with Tenant info if needed, though DSR already has tenant_id)
// We might need a specific response type if the admin API returns it differently
export interface AdminDSR extends DSR {
    tenant_name?: string; // If enriched
}

// Subscription tracks a tenant's billing lifecycle
export interface Subscription {
    id: string;
    tenant_id: string;
    plan: 'FREE' | 'STARTER' | 'PROFESSIONAL' | 'ENTERPRISE';
    billing_start: string | null;
    billing_end: string | null;
    auto_revoke: boolean;
    status: 'ACTIVE' | 'EXPIRED' | 'CANCELLED';
    created_at: string;
    updated_at: string;
}

// ModuleAccess tracks per-tenant module toggles
export interface ModuleAccess {
    id: string;
    tenant_id: string;
    module_name: ModuleName;
    enabled: boolean;
    created_at: string;
    updated_at: string;
}

export type ModuleName =
    | 'PII_DISCOVERY'
    | 'DSR_MANAGEMENT'
    | 'CONSENT_MANAGER'
    | 'BREACH_TRACKER'
    | 'DATA_GOVERNANCE'
    | 'AI_CLASSIFICATION'
    | 'ADVANCED_ANALYTICS'
    | 'AUDIT_TRAIL';

export const MODULE_NAMES = [
    'PII_DISCOVERY',
    'DSR_MANAGEMENT',
    'CONSENT_MANAGER',
    'BREACH_TRACKER',
    'DATA_GOVERNANCE',
    'AI_CLASSIFICATION',
    'ADVANCED_ANALYTICS',
    'AUDIT_TRAIL',
] as const;

export interface RetentionPolicy {
    id: string;
    tenant_id: string;
    purpose_id: string;
    max_retention_days: number;
    data_categories: string[]; // e.g., ["contact", "financial"]
    status: 'ACTIVE' | 'PAUSED';
    auto_erase: boolean;
    description?: string;
    created_at: string;
    updated_at: string;
}

export interface PlatformSettings {
    branding: {
        logo_url: string;
        primary_color: string;
        company_name: string;
    };
    maintenance: {
        enabled: boolean;
        message: string;
    };
    security: {
        mfa_required: boolean;
        session_timeout_minutes: number;
    };
    [key: string]: any; // Allow other keys
}

export interface ModuleAccessInput {
    module_name: ModuleName;
    enabled: boolean;
}
