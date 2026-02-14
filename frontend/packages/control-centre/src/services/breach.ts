import { api } from '@datalens/shared';
import type { ApiResponse, PaginatedResponse, ID } from '@datalens/shared';
import type {
    BreachIncident,
    CreateIncidentInput,
    UpdateIncidentInput,
    BreachFilter,
    IncidentDetailResponse
} from '../types/breach';

export const breachService = {
    async list(params?: { page?: number; page_size?: number } & BreachFilter): Promise<PaginatedResponse<BreachIncident>> {
        const queryParams = new URLSearchParams();
        if (params?.page) queryParams.append('page', params.page.toString());
        if (params?.page_size) queryParams.append('page_size', params.page_size.toString());
        if (params?.status) queryParams.append('status', params.status);
        if (params?.severity) queryParams.append('severity', params.severity);

        const res = await api.get<ApiResponse<PaginatedResponse<BreachIncident>>>(`/breach?${queryParams.toString()}`);
        return res.data.data;
    },

    async getById(id: ID): Promise<IncidentDetailResponse> {
        // The backend returns { incident: ..., sla: ... }
        // We cast it to our IncidentDetailResponse type
        const res = await api.get<ApiResponse<IncidentDetailResponse>>(`/breach/${id}`);
        return res.data.data;
    },

    async create(data: CreateIncidentInput): Promise<BreachIncident> {
        const res = await api.post<ApiResponse<BreachIncident>>('/breach', data);
        return res.data.data;
    },

    async update(id: ID, data: UpdateIncidentInput): Promise<BreachIncident> {
        const res = await api.put<ApiResponse<BreachIncident>>(`/breach/${id}`, data);
        return res.data.data;
    },

    async generateCertInReport(id: ID): Promise<Record<string, unknown>> {
        const res = await api.get<ApiResponse<Record<string, unknown>>>(`/breach/${id}/report/cert-in`);
        return res.data.data;
    }
};
