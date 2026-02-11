import { api } from './api';
import type { ApiResponse, ID, PaginatedResponse } from '../types/common';
import type { ConsentWidget, CreateWidgetInput, UpdateWidgetInput } from '../types/consent';

export const consentService = {
    // Widgets
    async listWidgets(params?: { page?: number; page_size?: number }): Promise<PaginatedResponse<ConsentWidget>> {
        const res = await api.get<ApiResponse<PaginatedResponse<ConsentWidget>>>('/consent/widgets', { params });
        return res.data.data;
    },

    async getWidget(id: ID): Promise<ConsentWidget> {
        const res = await api.get<ApiResponse<ConsentWidget>>(`/consent/widgets/${id}`);
        return res.data.data;
    },

    async createWidget(data: CreateWidgetInput): Promise<ConsentWidget> {
        const res = await api.post<ApiResponse<ConsentWidget>>('/consent/widgets', data);
        return res.data.data;
    },

    async updateWidget(id: ID, data: UpdateWidgetInput): Promise<ConsentWidget> {
        const res = await api.put<ApiResponse<ConsentWidget>>(`/consent/widgets/${id}`, data);
        return res.data.data;
    },

    async deleteWidget(id: ID): Promise<void> {
        await api.delete(`/consent/widgets/${id}`);
    },

    async activateWidget(id: ID): Promise<ConsentWidget> {
        const res = await api.put<ApiResponse<ConsentWidget>>(`/consent/widgets/${id}/activate`);
        return res.data.data;
    },

    async pauseWidget(id: ID): Promise<ConsentWidget> {
        const res = await api.put<ApiResponse<ConsentWidget>>(`/consent/widgets/${id}/pause`);
        return res.data.data;
    },

    // Public / Widget Config (Authenticated via API Key usually, but for preview we might use internal)
    // For now, let's stick to the internal management APIs.
};
