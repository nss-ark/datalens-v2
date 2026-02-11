import { api } from './api';
import type { DSR, DSRWithTasks, DSRListResponse, CreateDSRInput } from '../types/dsr';
import type { ID, ApiResponse } from '../types/common';

export const dsrService = {
    async list(params?: { page?: number; page_size?: number; status?: string }): Promise<DSRListResponse> {
        const res = await api.get<ApiResponse<DSRListResponse>>('/dsr', { params });
        return res.data.data;
    },

    async getById(id: ID): Promise<DSRWithTasks> {
        const res = await api.get<ApiResponse<DSRWithTasks>>(`/dsr/${id}`);
        return res.data.data;
    },

    async create(data: CreateDSRInput): Promise<DSR> {
        const res = await api.post<ApiResponse<DSR>>('/dsr', data);
        return res.data.data;
    },

    async approve(id: ID): Promise<DSR> {
        const res = await api.put<ApiResponse<DSR>>(`/dsr/${id}/approve`);
        return res.data.data;
    },

    async reject(id: ID, reason: string): Promise<DSR> {
        const res = await api.put<ApiResponse<DSR>>(`/dsr/${id}/reject`, { reason });
        return res.data.data;
    },

    async getResult(id: ID): Promise<unknown> {
        const res = await api.get<ApiResponse<unknown>>(`/dsr/${id}/result`);
        return res.data.data;
    },

    async execute(id: ID): Promise<void> {
        // Execute might return status message, but we just await completion
        await api.post<ApiResponse<any>>(`/dsr/${id}/execute`);
    },
};
