import type { BaseEntity, ID } from './common';
import type { DSR } from './dsr';

export interface Tenant extends BaseEntity {
    id: ID;
    name: string;
    domain: string;
    status: 'ACTIVE' | 'INACTIVE' | 'SUSPENDED';
    plan: 'FREE' | 'STARTER' | 'PROFESSIONAL' | 'ENTERPRISE';
    log_retention_days: number;
    platform_fee_percent: number;
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
