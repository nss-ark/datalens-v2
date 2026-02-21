import { api } from '@datalens/shared';
import type { ApiResponse, PaginatedResponse, ID } from '@datalens/shared';

export interface DataPrincipalProfile {
    id: ID;
    tenant_id: ID;
    email: string;
    phone: string | null;
    verification_status: 'PENDING' | 'VERIFIED' | 'FAILED';
    verified_at: string | null;
    verification_method: 'EMAIL' | 'SMS' | 'OAUTH' | 'DIGILOCKER' | null;
    subject_id: string; // The opaque ID given to the principal
    last_access_at: string;
    preferred_lang: string;
    is_minor: boolean;
    guardian_verified: boolean;
    created_at: string;
    updated_at: string;
}

export const dataSubjectService = {
    async listSubjects(params?: {
        page?: number;
        page_size?: number;
        search?: string
    }): Promise<PaginatedResponse<DataPrincipalProfile>> {
        const res = await api.get<ApiResponse<PaginatedResponse<DataPrincipalProfile>>>('/subjects', { params });
        return res.data.data;
    }
};
