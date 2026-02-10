import { api } from './api';
import type { DSR, DSRWithTasks, DSRListResponse, CreateDSRInput } from '../types/dsr';
import type { ID } from '../types/common';

export const dsrService = {
    async list(params?: { page?: number; page_size?: number; status?: string }): Promise<DSRListResponse> {
        const res = await api.get<DSRListResponse>('/dsr', { params });
        return res.data;
    },

    async getById(id: ID): Promise<DSRWithTasks> {
        const res = await api.get<DSRWithTasks>(`/dsr/${id}`);
        return res.data;
    },

    async create(data: CreateDSRInput): Promise<DSR> {
        const res = await api.post<DSR>('/dsr', data);
        return res.data;
    },

    async approve(id: ID): Promise<DSR> {
        const res = await api.put<DSR>(`/dsr/${id}/approve`);
        return res.data;
    },

    async reject(id: ID, reason: string): Promise<DSR> {
        const res = await api.put<DSR>(`/dsr/${id}/reject`, { reason });
        return res.data;
    },

    async getResult(id: ID): Promise<unknown> {
        const res = await api.get(`/dsr/${id}/result`);
        return res.data;
    },

    async execute(id: ID): Promise<void> {
        await api.post(`/dsr/${id}/execute`);
    },
};
