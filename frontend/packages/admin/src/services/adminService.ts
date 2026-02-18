import { api } from '@datalens/shared';
import type { ApiResponse, PaginatedResponse } from '@datalens/shared';
import type { Tenant, CreateTenantInput, AdminStats, AdminUser, AdminRole, AdminDSR } from '@/types/admin';

export const adminService = {
    // Tenants
    async getTenants(params?: { page?: number; limit?: number; search?: string }): Promise<PaginatedResponse<Tenant>> {
        const items: Tenant[] = [
            {
                id: '1',
                name: 'Acme Corp',
                domain: 'acme',
                status: 'ACTIVE',
                plan: 'PROFESSIONAL',
                log_retention_days: 90,
                platform_fee_percent: 1.5,
                created_at: new Date(Date.now() - 86400000 * 5).toISOString(),
                updated_at: new Date().toISOString(),
                tenant_id: '1'
            },
            {
                id: '2',
                name: 'Globex Inc',
                domain: 'globex',
                status: 'ACTIVE',
                plan: 'ENTERPRISE',
                log_retention_days: 365,
                platform_fee_percent: 0.5,
                created_at: new Date(Date.now() - 86400000 * 15).toISOString(),
                updated_at: new Date().toISOString(),
                tenant_id: '2'
            },
            {
                id: '3',
                name: 'Soylent Corp',
                domain: 'soylent',
                status: 'SUSPENDED',
                plan: 'STARTER',
                log_retention_days: 30,
                platform_fee_percent: 2.5,
                created_at: new Date(Date.now() - 86400000 * 45).toISOString(),
                updated_at: new Date().toISOString(),
                tenant_id: '3'
            },
            {
                id: '4',
                name: 'Initech',
                domain: 'initech',
                status: 'ACTIVE',
                plan: 'FREE',
                log_retention_days: 7,
                platform_fee_percent: 5.0,
                created_at: new Date(Date.now() - 86400000 * 2).toISOString(),
                updated_at: new Date().toISOString(),
                tenant_id: '4'
            },
            {
                id: '5',
                name: 'Umbrella Corp',
                domain: 'umbrella',
                status: 'ACTIVE',
                plan: 'ENTERPRISE',
                log_retention_days: 365,
                platform_fee_percent: 1.0,
                created_at: new Date(Date.now() - 86400000 * 60).toISOString(),
                updated_at: new Date().toISOString(),
                tenant_id: '5'
            }
        ];
        return {
            items,
            total: 5,
            page: params?.page || 1,
            page_size: params?.limit || 10,
            total_pages: 1
        };
    },

    async createTenant(data: CreateTenantInput): Promise<{ tenant: Tenant; user: unknown }> {
        // Mock delay
        await new Promise(resolve => setTimeout(resolve, 800));
        return {
            tenant: {
                id: `new-${Date.now()}`,
                name: data.name,
                domain: data.domain,
                status: 'ACTIVE',
                plan: data.plan,
                log_retention_days: 30,
                platform_fee_percent: 2.0,
                created_at: new Date().toISOString(),
                updated_at: new Date().toISOString(),
                tenant_id: `new-${Date.now()}`
            },
            user: {}
        };
    },

    // Stats
    async getStats(): Promise<AdminStats> {
        return {
            total_tenants: 15,
            active_tenants: 12,
            total_users: 1250,
            total_dsr_requests: 45
        };
    },

    // Users
    async getUsers(params?: { page?: number; limit?: number; search?: string; tenant_id?: string; status?: string }): Promise<PaginatedResponse<AdminUser>> {
        return {
            items: [],
            total: 0,
            page: 1,
            page_size: 10,
            total_pages: 1
        };
    },

    async getUserById(id: string): Promise<AdminUser> {
        return {
            id,
            tenant_id: '1',
            email: 'user@example.com',
            name: 'Mock User',
            status: 'ACTIVE',
            role_ids: [],
            mfa_enabled: false,
            last_login_at: new Date().toISOString(),
            created_at: new Date().toISOString()
        };
    },

    async updateUserStatus(id: string, status: string): Promise<void> {
    },

    async assignRoles(id: string, roleIds: string[]): Promise<void> {
    },

    async getRoles(): Promise<AdminRole[]> {
        return [];
    },

    // Compliance / DSRs
    async getDSRs(params?: { page?: number; limit?: number; status?: string; type?: string; tenant_id?: string }): Promise<PaginatedResponse<AdminDSR>> {
        return {
            items: [],
            total: 0,
            page: 1,
            page_size: 10,
            total_pages: 1
        };
    },

    async getDSRById(id: string): Promise<AdminDSR> {
        throw new Error("Not implemented in mock");
    },
};
