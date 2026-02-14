export interface PIIStats {
    total_pii: number;
    pii_by_category: Record<string, number>;
}

export interface DashboardStats {
    total_data_sources: number;
    total_scans: number;
    total_pii_fields: number;
    pending_reviews: number;
    pii_by_category: Record<string, number>;
    recent_scans: ScanSummary[];
}

export interface ScanSummary {
    id: string;
    data_source_id: string;
    data_source_name: string;
    status: 'QUEUED' | 'RUNNING' | 'COMPLETED' | 'FAILED';
    tables_scanned: number;
    pii_found: number;
    started_at: string;
    completed_at: string | null;
    duration_seconds?: number;
}
