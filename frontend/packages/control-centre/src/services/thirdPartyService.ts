import { api } from '@datalens/shared';
import type { ApiResponse } from '@datalens/shared';

export interface ThirdParty {
    id: string;
    tenant_id: string;
    name: string;
    type: 'PROCESSOR' | 'CONTROLLER' | 'VENDOR';
    country: string;
    dpa_doc_path?: string;
    is_active: boolean;
    purpose_ids: string[];
    dpa_status: 'NONE' | 'PENDING' | 'SIGNED' | 'EXPIRED';
    dpa_signed_at?: string;
    dpa_expires_at?: string;
    dpa_notes?: string;
    contact_name?: string;
    contact_email?: string;
    created_at: string;
    updated_at: string;
}

export const thirdPartyService = {
    async list() {
        const res = await api.get<ApiResponse<ThirdParty[]>>('/v2/third-parties');
        return res.data.data;
    },
    async getById(id: string) {
        const res = await api.get<ApiResponse<ThirdParty>>(`/v2/third-parties/${id}`);
        return res.data.data;
    },
    async create(data: Partial<ThirdParty>) {
        const res = await api.post<ApiResponse<ThirdParty>>('/v2/third-parties', data);
        return res.data.data;
    },
    async update(id: string, data: Partial<ThirdParty>) {
        const res = await api.put<ApiResponse<ThirdParty>>(`/v2/third-parties/${id}`, data);
        return res.data.data;
    },
    async remove(id: string) {
        return api.delete(`/v2/third-parties/${id}`);
    },
};
