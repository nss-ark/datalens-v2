import { api, type PaginatedResponse, type ApiResponse, type LoginResponse, type ID } from '@datalens/shared';
import type {
    Tenant,
    AdminUser,
    AdminRole,
    AdminDSR,
    Subscription,
    ModuleAccess,
    ModuleAccessInput,
    RetentionPolicy,
    PlatformSettings,
    CreateTenantInput,
    AdminStats
} from '@/types/admin';

export const adminService = {
    // Auth - Public Route
    async login(email: string, password: string): Promise<LoginResponse> {
        const res = await api.post<ApiResponse<LoginResponse>>('/superadmin/login', { email, password });
        return res.data.data;
    },

    async getCurrentUser(): Promise<AdminUser> {
        const res = await api.get<ApiResponse<AdminUser>>('/admin/me');
        return res.data.data;
    },

    // Tenants
    async getTenants(params?: { page?: number; limit?: number; search?: string }): Promise<PaginatedResponse<Tenant>> {
        const res = await api.get<ApiResponse<PaginatedResponse<Tenant>>>('/admin/tenants', { params });
        return res.data.data;
    },

    async createTenant(data: CreateTenantInput): Promise<{ tenant: Tenant; user: unknown }> {
        const res = await api.post<ApiResponse<{ tenant: Tenant; user: unknown }>>('/admin/tenants', data);
        return res.data.data;
    },

    async getTenant(id: ID): Promise<Tenant> {
        const res = await api.get<ApiResponse<Tenant>>(`/admin/tenants/${id}`);
        return res.data.data;
    },

    async updateTenant(id: ID, data: Partial<Tenant>): Promise<Tenant> {
        const res = await api.patch<ApiResponse<Tenant>>(`/admin/tenants/${id}`, data);
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
        await api.patch<ApiResponse<void>>(`/admin/users/${id}/status`, { status });
    },

    async assignRoles(id: string, roleIds: string[]): Promise<void> {
        await api.put<ApiResponse<void>>(`/admin/users/${id}/roles`, { role_ids: roleIds });
    },

    async getRoles(): Promise<AdminRole[]> {
        const res = await api.get<ApiResponse<AdminRole[]>>('/admin/roles');
        return res.data.data;
    },

    // Compliance / DSRs
    async getDSRs(params?: { page?: number; limit?: number; status?: string; type?: string; tenant_id?: string }): Promise<PaginatedResponse<AdminDSR>> {
        const res = await api.get<ApiResponse<PaginatedResponse<AdminDSR>>>('/admin/dsr', { params });
        return res.data.data;
    },

    async getDSRById(id: string): Promise<AdminDSR> {
        const res = await api.get<ApiResponse<AdminDSR>>(`/admin/dsr/${id}`);
        return res.data.data;
    },

    // Subscriptions
    async getSubscription(tenantId: ID): Promise<Subscription> {
        const res = await api.get<ApiResponse<Subscription>>(`/admin/tenants/${tenantId}/subscription`);
        return res.data.data;
    },

    async updateSubscription(tenantId: ID, data: Partial<Subscription>): Promise<Subscription> {
        const res = await api.put<ApiResponse<Subscription>>(`/admin/tenants/${tenantId}/subscription`, data);
        return res.data.data;
    },

    // Module Access
    getModuleAccess: async (tenantId: string) => {
        const { data } = await api.get<ModuleAccess[]>(`/admin/tenants/${tenantId}/modules`);
        return data;
    },

    updateModuleAccess: async (tenantId: string, modules: ModuleAccessInput[]) => {
        const { data } = await api.put<ModuleAccess[]>(`/admin/tenants/${tenantId}/modules`, modules);
        return data;
    },

    // Retention Policies
    getRetentionPolicies: async (tenantId: string) => {
        const { data } = await api.get<RetentionPolicy[]>(`/admin/retention-policies`, { params: { tenant_id: tenantId } });
        return data;
    },

    createRetentionPolicy: async (policy: Partial<RetentionPolicy>) => {
        const { data } = await api.post<RetentionPolicy>(`/admin/retention-policies`, policy);
        return data;
    },

    updateRetentionPolicy: async (id: string, policy: Partial<RetentionPolicy>) => {
        const { data } = await api.put<RetentionPolicy>(`/admin/retention-policies/${id}`, policy);
        return data;
    },

    deleteRetentionPolicy: async (id: string) => {
        await api.delete(`/admin/retention-policies/${id}`);
    },

    // Platform Settings
    getPlatformSettings: async () => {
        const { data } = await api.get<PlatformSettings>(`/admin/settings`);
        return data;
    },

    updatePlatformSettings: async (settings: Partial<PlatformSettings>) => {
        const { data } = await api.put<any>(`/admin/settings`, settings); // Backend accepts map[string]any
        return data;
    },
};
