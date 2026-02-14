import { api } from '@datalens/shared';
import type { ApiResponse, PaginatedResponse } from '@datalens/shared';
import type { ConsentNotification, NotificationFilter } from '../types/notification';

export const notificationService = {
    async listNotifications(params?: NotificationFilter & { page?: number; page_size?: number }): Promise<PaginatedResponse<ConsentNotification>> {
        const res = await api.get<ApiResponse<PaginatedResponse<ConsentNotification>>>('/compliance/notifications', { params });
        return res.data.data;
    },

    async getNotification(id: string): Promise<ConsentNotification> {
        const res = await api.get<ApiResponse<ConsentNotification>>(`/compliance/notifications/${id}`);
        return res.data.data;
    },

    async resendNotification(id: string): Promise<void> {
        await api.post(`/compliance/notifications/${id}/resend`);
    }
};
