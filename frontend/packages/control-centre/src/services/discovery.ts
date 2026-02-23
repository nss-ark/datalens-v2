import { api } from '@datalens/shared';
import type { ApiResponse } from '@datalens/shared';
import type {
    PIIClassification,
    SubmitFeedbackInput,
    FeedbackResponse,
    AccuracyStats,
    DetectionMethod,
} from '../types/discovery';
import type { ID } from '@datalens/shared';

export interface ClassificationFilters {
    status?: string;
    data_source_id?: string;
    detection_method?: string;
    page?: number;
    page_size?: number;
}

export interface PaginatedClassifications {
    items: PIIClassification[];
    total: number;
    page: number;
    page_size: number;
    total_pages: number;
}

export const discoveryService = {
    /**
     * List PII classifications for the current tenant.
     * Endpoint: GET /api/v2/discovery/classifications
     * Backend wraps response in { success, data: PaginatedResult }
     */
    async listClassifications(filters?: ClassificationFilters): Promise<PaginatedClassifications> {
        const params = new URLSearchParams();
        if (filters?.status) params.set('status', filters.status);
        if (filters?.data_source_id) params.set('data_source_id', filters.data_source_id);
        if (filters?.detection_method) params.set('detection_method', filters.detection_method);
        if (filters?.page) params.set('page', String(filters.page));
        if (filters?.page_size) params.set('page_size', String(filters.page_size));

        const res = await api.get<ApiResponse<PaginatedClassifications>>(
            `/discovery/classifications?${params.toString()}`
        );
        return res.data.data as PaginatedClassifications;
    },

    /** Get classifications for a specific data source */
    async getByDataSource(dataSourceId: ID): Promise<PIIClassification[]> {
        const res = await api.get<ApiResponse<PIIClassification[]>>(
            `/discovery/data-sources/${dataSourceId}/classifications`
        );
        return res.data.data as PIIClassification[];
    },

    /** Submit feedback (verify, correct, reject) */
    async submitFeedback(input: SubmitFeedbackInput): Promise<FeedbackResponse> {
        const res = await api.post<ApiResponse<FeedbackResponse>>('/discovery/feedback', input);
        return res.data.data as FeedbackResponse;
    },

    /** Get accuracy stats for a detection method */
    async getAccuracyStats(method: DetectionMethod): Promise<AccuracyStats> {
        const res = await api.get<ApiResponse<AccuracyStats>>(`/discovery/feedback/accuracy/${method}`);
        return res.data.data as AccuracyStats;
    },
};
