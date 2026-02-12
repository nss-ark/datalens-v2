import { api } from './api';
import type { ApiResponse } from '../types/common';
import type { ConversionStat, PurposeStat, AnalyticsFilter } from '../types/analytics';

export const analyticsService = {
    async getConversionStats(params: AnalyticsFilter): Promise<ConversionStat[]> {
        const res = await api.get<ApiResponse<ConversionStat[]>>('/analytics/consent/conversion', { params });
        return res.data.data;
    },

    async getPurposeStats(params: AnalyticsFilter): Promise<PurposeStat[]> {
        const res = await api.get<ApiResponse<PurposeStat[]>>('/analytics/consent/purpose', { params });
        return res.data.data;
    },
};
