import { api } from '@datalens/shared';
import type { ApiResponse, ID, PaginatedResponse } from '@datalens/shared';
import type { ConsentWidget, CreateWidgetInput, UpdateWidgetInput, ConsentNotice, CreateNoticeInput, UpdateNoticeInput } from '../types/consent';

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

    // --- Notices ---
    async listNotices(): Promise<ConsentNotice[]> {
        const res = await api.get<ApiResponse<ConsentNotice[]>>('/consent/notices');
        return res.data.data;
    },

    async getNotice(id: ID): Promise<ConsentNotice> {
        const res = await api.get<ApiResponse<ConsentNotice>>(`/consent/notices/${id}`);
        return res.data.data;
    },

    async createNotice(data: CreateNoticeInput): Promise<ConsentNotice> {
        const res = await api.post<ApiResponse<ConsentNotice>>('/consent/notices', data);
        return res.data.data;
    },

    async updateNotice(id: ID, data: UpdateNoticeInput): Promise<void> {
        await api.put(`/consent/notices/${id}`, data);
    },

    async publishNotice(id: ID): Promise<ConsentNotice> {
        const res = await api.post<ApiResponse<ConsentNotice>>(`/consent/notices/${id}/publish`);
        return res.data.data;
    },

    async archiveNotice(id: ID): Promise<void> {
        await api.post(`/consent/notices/${id}/archive`);
    },

    async bindNotice(id: ID, widgetIds: ID[]): Promise<void> {
        await api.post(`/consent/notices/${id}/bind`, { widget_ids: widgetIds });
    },
};
