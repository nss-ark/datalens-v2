import { api } from './api';
import type { DataSource, CreateDataSourceInput, UpdateDataSourceInput } from '../types/datasource';
import type { ID } from '../types/common';

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

    async testConnection(id: ID): Promise<void> {
        await api.post(`/data-sources/${id}/test`);
    },
};
