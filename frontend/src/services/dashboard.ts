import { api } from './api';
import type { DashboardStats } from '../types/dashboard';

export const dashboardService = {
    /**
     * Get aggregated stats for the dashboard.
     * Endpoint: GET /api/v2/dashboard/stats
     */
    async getStats(): Promise<DashboardStats> {
        const res = await api.get<DashboardStats>('/dashboard/stats');
        return res.data;
    },
};
