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

// Generic API Response wrapper
export interface ApiResponse<T> {
    success: boolean;
    data: T;
    error?: ApiError;
    meta?: any;
}

// Matches pkg/types/types.go PaginatedResult
export interface PaginatedResponse<T> {
    items: T[];
    total: number;
    page: number;
    page_size: number;
    total_pages: number;
}

export interface ApiError {
    code: string;
    message: string;
    details?: Record<string, unknown>;
}
