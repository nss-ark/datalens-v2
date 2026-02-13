import { api } from './api';
import type { ApiResponse, ID } from '../types/common';
import type {
    DataSource,
    CreateDataSourceInput,
    UpdateDataSourceInput,
    ScanHistoryItem,
    ScanProgress,
    M365User,
    SharePointSite,
    M365ScopeConfig,
    GoogleScopeConfig
} from '../types/datasource';

export const dataSourceService = {
    async list(): Promise<DataSource[]> {
        const res = await api.get<ApiResponse<DataSource[]>>('/data-sources');
        return res.data.data;
    },

    async getById(id: ID): Promise<DataSource> {
        const res = await api.get<ApiResponse<DataSource>>(`/data-sources/${id}`);
        return res.data.data;
    },

    async create(data: CreateDataSourceInput): Promise<DataSource> {
        const res = await api.post<ApiResponse<DataSource>>('/data-sources', data);
        return res.data.data;
    },

    async update(data: UpdateDataSourceInput & { id: ID }): Promise<DataSource> {
        const res = await api.put<ApiResponse<DataSource>>(`/data-sources/${data.id}`, data);
        return res.data.data;
    },

    async delete(id: ID): Promise<void> {
        await api.delete(`/data-sources/${id}`);
    },

    async scan(id: ID): Promise<void> {
        await api.post(`/data-sources/${id}/scan`);
    },

    async getScanStatus(id: ID): Promise<ScanProgress> {
        const res = await api.get<ApiResponse<ScanProgress>>(`/data-sources/${id}/scan/status`);
        return res.data.data;
    },

    async getScanHistory(id: ID): Promise<ScanHistoryItem[]> {
        const res = await api.get<ApiResponse<ScanHistoryItem[]>>(`/data-sources/${id}/scan/history`);
        return res.data.data;
    },

    // --- M365 Scope Management ---

    async getM365Users(dataSourceId: ID): Promise<M365User[]> {
        const res = await api.get<ApiResponse<M365User[]>>(`/data-sources/${dataSourceId}/m365/users`);
        return res.data.data;
    },

    async getSharePointSites(dataSourceId: ID): Promise<SharePointSite[]> {
        const res = await api.get<ApiResponse<SharePointSite[]>>(`/data-sources/${dataSourceId}/m365/sites`);
        return res.data.data;
    },

    async updateScope(dataSourceId: ID, config: M365ScopeConfig | GoogleScopeConfig): Promise<DataSource> {
        // We persist this by updating the 'config' field of the DataSource
        const configJson = JSON.stringify(config);
        const res = await api.put<ApiResponse<DataSource>>(`/data-sources/${dataSourceId}`, { config: configJson });
        return res.data.data;
    },

    // --- OAuth Helpers ---

    getM365AuthUrl(): string {
        // Returns the backend endpoint that redirects to Microsoft
        // We might want to fetch this if it's dynamic, but the handler redirects directly.
        // If we want to open it in a popup, we point the popup to this URL.
        return `${import.meta.env.VITE_API_URL || '/api/v2'}/m365/connect`;
    },

    getGoogleAuthUrl(): string {
        return `${import.meta.env.VITE_API_URL || '/api/v2'}/google/connect`;
    }
};
