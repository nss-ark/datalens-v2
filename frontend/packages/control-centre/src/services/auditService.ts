import { api } from '@datalens/shared';
import type { ApiResponse } from '@datalens/shared';

// --- Types ---

export interface AuditLog {
    id: string;
    tenant_id: string;
    user_id: string;
    action: string;
    resource_type: string;
    resource_id: string;
    old_values: Record<string, unknown> | null;
    new_values: Record<string, unknown> | null;
    ip_address: string;
    user_agent: string;
    created_at: string;
}

export interface AuditLogFilters {
    entity_type?: string;
    action?: string;
    start_date?: string;
    end_date?: string;
    page?: number;
    page_size?: number;
}

export interface AuditLogListResponse {
    items: AuditLog[];
    total: number;
    page: number;
    page_size: number;
    total_pages: number;
}

// --- Service ---

/**
 * The backend returns: { success, data: AuditLog[], meta: { page, page_size, total, total_pages } }
 * We normalise this into AuditLogListResponse for the UI.
 */
export const auditService = {
    async list(params?: AuditLogFilters): Promise<AuditLogListResponse> {
        const res = await api.get<ApiResponse<AuditLog[]>>('/audit-logs', { params });
        const items = res.data.data;
        const meta = res.data.meta as { page: number; page_size: number; total: number; total_pages: number } | undefined;
        return {
            items: Array.isArray(items) ? items : [],
            total: meta?.total ?? 0,
            page: meta?.page ?? params?.page ?? 1,
            page_size: meta?.page_size ?? params?.page_size ?? 20,
            total_pages: meta?.total_pages ?? 1,
        };
    },
};
