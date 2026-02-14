import { api } from '@datalens/shared';
import type { DataSource, CreateDataSourceInput, UpdateDataSourceInput, ScanProgress, ScanHistoryItem } from '../types/datasource';
import type { ID } from '@datalens/shared';

export const dataSourceService = {
    async list(): Promise<DataSource[]> {
        const res = await api.get<DataSource[]>('/data-sources');
        return res.data;
    },

    async getById(id: ID): Promise<DataSource> {
        const res = await api.get<DataSource>(`/data-sources/${id}`);
        return res.data;
    },

    async create(data: CreateDataSourceInput): Promise<DataSource> {
        const res = await api.post<DataSource>('/data-sources', data);
        return res.data;
    },

    async update(id: ID, data: UpdateDataSourceInput): Promise<DataSource> {
        const res = await api.put<DataSource>(`/data-sources/${id}`, data);
        return res.data;
    },

    async remove(id: ID): Promise<void> {
        await api.delete(`/data-sources/${id}`);
    },

    async scan(id: ID): Promise<void> {
        await api.post(`/data-sources/${id}/scan`);
    },

    async getScanStatus(id: ID): Promise<ScanProgress> {
        const res = await api.get<ScanProgress>(`/data-sources/${id}/scan/status`);
        return res.data;
    },

    async getScanHistory(id: ID): Promise<ScanHistoryItem[]> {
        const res = await api.get<ScanHistoryItem[]>(`/data-sources/${id}/scan/history`);
        return res.data;
    },

    async testConnection(id: ID): Promise<void> {
        await api.post(`/data-sources/${id}/test`);
    },
};
