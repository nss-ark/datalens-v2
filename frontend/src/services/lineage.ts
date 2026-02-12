import { api } from './api';
import type { ApiResponse } from '../types/common';
import type { LineageGraph, DataFlow } from '../types/lineage';

export const lineageService = {
    async getGraph(): Promise<LineageGraph> {
        const res = await api.get<ApiResponse<LineageGraph>>('/governance/lineage');
        return res.data.data;
    },

    async trackFlow(flow: Partial<DataFlow>): Promise<DataFlow> {
        const res = await api.post<ApiResponse<DataFlow>>('/governance/lineage', flow);
        return res.data.data;
    }
};
