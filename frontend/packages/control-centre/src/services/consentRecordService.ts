import { api } from '@datalens/shared';
import type { PaginatedResponse, ApiResponse } from '@datalens/shared';

export interface ConsentSession {
    id: string;
    tenant_id: string;
    widget_id: string;
    subject_id: string;
    ip_address: string;
    user_agent: string;
    decisions: { purpose_id: string; granted: boolean }[];
    page_url: string;
    widget_version: number;
    notice_version: string;
    signature: string;
    status?: string;
    created_at: string;
    updated_at: string;
}

export const consentRecordService = {
    async list(params?: { page?: number; page_size?: number; status?: string; purpose_id?: string }): Promise<PaginatedResponse<ConsentSession>> {
        const res = await api.get<ApiResponse<PaginatedResponse<ConsentSession>>>('/consent/sessions', { params });
        return res.data.data;
    },
};
