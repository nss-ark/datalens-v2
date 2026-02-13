import { api } from './api';
import type { ApiResponse, PaginatedResponse } from '../types/common';
import type { Tenant, CreateTenantInput, AdminStats } from '../types/admin';

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
};
