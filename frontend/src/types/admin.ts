import type { BaseEntity, ID } from './common';

export interface Tenant extends BaseEntity {
    tenant_id: ID;
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
