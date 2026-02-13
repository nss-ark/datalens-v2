import { api } from './api';
import type { ApiResponse, PaginatedResponse } from '../types/common';
import type { Tenant, CreateTenantInput, AdminStats, AdminUser, AdminRole } from '../types/admin';

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
};
