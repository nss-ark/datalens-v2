import { api } from '@datalens/shared';
import type { ApiResponse } from '@datalens/shared';

export interface Department {
    id: string;
    tenant_id: string;
    name: string;
    description?: string;
    owner_id?: string;
    owner_name?: string;
    owner_email?: string;
    responsibilities: string[];
    notification_enabled: boolean;
    is_active: boolean;
    created_at: string;
    updated_at: string;
}

export const departmentService = {
    async list() {
        const res = await api.get<ApiResponse<Department[]>>('/v2/departments');
        return res.data.data;
    },
    async getById(id: string) {
        const res = await api.get<ApiResponse<Department>>(`/v2/departments/${id}`);
        return res.data.data;
    },
    async create(data: Partial<Department>) {
        const res = await api.post<ApiResponse<Department>>('/v2/departments', data);
        return res.data.data;
    },
    async update(id: string, data: Partial<Department>) {
        const res = await api.put<ApiResponse<Department>>(`/v2/departments/${id}`, data);
        return res.data.data;
    },
    async remove(id: string) {
        return api.delete(`/v2/departments/${id}`);
    },
    async notify(id: string, subject: string, body: string) {
        return api.post(`/v2/departments/${id}/notify`, { subject, body });
    },
};
