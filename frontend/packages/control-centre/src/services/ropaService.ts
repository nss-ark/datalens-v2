import { api } from '@datalens/shared';
import type { ApiResponse } from '@datalens/shared';

// --- Types ---

export interface RoPAPurpose {
    id: string;
    name: string;
    code: string;
    legal_basis: string;
    description: string;
    is_active: boolean;
}

export interface RoPADataSource {
    id: string;
    name: string;
    type: string;
    is_active: boolean;
}

export interface RoPARetention {
    id: string;
    purpose_name: string;
    max_retention_days: number;
    data_categories: string[];
    auto_erase: boolean;
}

export interface RoPAThirdParty {
    id: string;
    name: string;
    type: string;
    country: string;
}

export interface RoPAContent {
    organization_name: string;
    generated_at: string;
    purposes: RoPAPurpose[];
    data_sources: RoPADataSource[];
    retention_policies: RoPARetention[];
    third_parties: RoPAThirdParty[];
    data_categories: string[];
}

export interface RoPAVersion {
    id: string;
    tenant_id: string;
    version: string;
    generated_by: string;
    status: 'DRAFT' | 'PUBLISHED' | 'ARCHIVED';
    content: RoPAContent;
    change_summary?: string;
    created_at: string;
}

// --- Service ---

export const ropaService = {
    async getLatest(): Promise<RoPAVersion | null> {
        const res = await api.get<ApiResponse<RoPAVersion | null>>('/ropa');
        return res.data.data;
    },

    async generate(): Promise<RoPAVersion> {
        const res = await api.post<ApiResponse<RoPAVersion>>('/ropa');
        return res.data.data;
    },

    async listVersions(page = 1, pageSize = 10) {
        const res = await api.get<ApiResponse<{ items: RoPAVersion[]; total: number; page: number; page_size: number; total_pages: number }>>('/ropa/versions', {
            params: { page, page_size: pageSize },
        });
        return res.data.data;
    },

    async getVersion(version: string): Promise<RoPAVersion> {
        const res = await api.get<ApiResponse<RoPAVersion>>(`/ropa/versions/${version}`);
        return res.data.data;
    },

    async saveEdit(content: RoPAContent, changeSummary: string): Promise<RoPAVersion> {
        const res = await api.put<ApiResponse<RoPAVersion>>('/ropa', {
            content,
            change_summary: changeSummary,
        });
        return res.data.data;
    },

    async publish(id: string): Promise<void> {
        await api.post('/ropa/publish', { id });
    },

    async promote(): Promise<RoPAVersion> {
        const res = await api.post<ApiResponse<RoPAVersion>>('/ropa/promote');
        return res.data.data;
    },
};
