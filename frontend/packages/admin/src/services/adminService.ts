import { api } from '@datalens/shared';
import type { ApiResponse, PaginatedResponse } from '@datalens/shared';
import type { Tenant, CreateTenantInput, AdminStats, AdminUser, AdminRole, AdminDSR } from '@/types/admin';

export const adminService = {
    // Tenants
    async getTenants(params?: { page?: number; limit?: number; search?: string }): Promise<PaginatedResponse<Tenant>> {
        const res = await api.get<ApiResponse<PaginatedResponse<Tenant>>>('/admin/tenants', { params });
        return res.data.data;
    },

    async createTenant(data: CreateTenantInput): Promise<{ tenant: Tenant; user: unknown }> {
        const res = await api.post<ApiResponse<{ tenant: Tenant; user: unknown }>>('/admin/tenants', data);
        return res.data.data;
    },

    // Stats
    async getStats(): Promise<AdminStats> {
        const res = await api.get<ApiResponse<AdminStats>>('/admin/stats');
        return res.data.data;
    },

    // Users
    async getUsers(params?: { page?: number; limit?: number; search?: string; tenant_id?: string; status?: string }): Promise<PaginatedResponse<AdminUser>> {
        const res = await api.get<ApiResponse<PaginatedResponse<AdminUser>>>('/admin/users', { params });
        return res.data.data;
    },

    async getUserById(id: string): Promise<AdminUser> {
        const res = await api.get<ApiResponse<AdminUser>>(`/admin/users/${id}`);
        return res.data.data;
    },

    async updateUserStatus(id: string, status: string): Promise<void> {
        await api.put(`/admin/users/${id}/status`, { status });
    },

    async assignRoles(id: string, roleIds: string[]): Promise<void> {
        await api.put(`/admin/users/${id}/roles`, { role_ids: roleIds });
    },

    async getRoles(): Promise<AdminRole[]> {
        const res = await api.get<ApiResponse<AdminRole[]>>('/admin/roles');
        return res.data.data;
    },

    // Compliance / DSRs
    // Note: This endpoint is assumed to be implemented in Backend Task #2 or future
    // If it returns 404, we might need to fallback to per-tenant DSRService if possible
    async getDSRs(params?: { page?: number; limit?: number; status?: string; type?: string; tenant_id?: string }): Promise<PaginatedResponse<AdminDSR>> {
        const res = await api.get<ApiResponse<PaginatedResponse<AdminDSR>>>('/admin/dsr', { params });
        return res.data.data;
    },

    async getDSRById(id: string): Promise<AdminDSR> {
        const res = await api.get<ApiResponse<AdminDSR>>(`/admin/dsr/${id}`);
        return res.data.data;
    },
};
