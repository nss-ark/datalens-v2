export type ID = string;

export interface BaseEntity {
    id: ID;
    created_at?: string;
    updated_at?: string;
}

export interface Auditable {
    created_by?: string;
    updated_by?: string;
}

export interface TenantEntity extends BaseEntity {
    tenant_id: ID;
}

export interface PaginationParams {
    page: number;
    pageSize: number;
    sortBy?: string;
    sortOrder?: 'asc' | 'desc';
}

export interface PaginatedResponse<T> {
    data: T[];
    meta: {
        total: number;
        page: number;
        pageSize: number;
        totalPages: number;
    };
}

export interface ApiError {
    code: string;
    message: string;
    details?: Record<string, unknown>;
}
