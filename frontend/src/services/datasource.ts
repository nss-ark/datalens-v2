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
    M365ScopeConfig
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

    // --- M365 Scope Management (Mocked for now) ---

    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    async getM365Users(_dataSourceId: ID): Promise<M365User[]> {
        // TODO: Replace with actual API call when backend endpoint is ready
        // GET /data-sources/{id}/m365/users
        // return api.get<ApiResponse<M365User[]>>(\`/data-sources/\${dataSourceId}/m365/users\`).then(res => res.data.data);

        // Mock Data
        await new Promise(resolve => setTimeout(resolve, 800)); // Simulate latency
        return [
            { id: 'user1@example.com', displayName: 'Alice Johnson', email: 'alice@example.com', scanOneDrive: true, scanOutlook: true },
            { id: 'user2@example.com', displayName: 'Bob Smith', email: 'bob@example.com', scanOneDrive: false, scanOutlook: true },
            { id: 'user3@example.com', displayName: 'Carol Williams', email: 'carol@example.com', scanOneDrive: true, scanOutlook: false },
            { id: 'user4@example.com', displayName: 'David Brown', email: 'david@example.com', scanOneDrive: false, scanOutlook: false },
        ];
    },

    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    async getSharePointSites(_dataSourceId: ID): Promise<SharePointSite[]> {
        // TODO: Replace with actual API call when backend endpoint is ready
        // GET /data-sources/{id}/m365/sites

        // Mock Data
        await new Promise(resolve => setTimeout(resolve, 800));
        return [
            { id: 'site1', name: 'HR Confidential', url: 'https://example.sharepoint.com/sites/hr', scanDocuments: true },
            { id: 'site2', name: 'Engineering Team', url: 'https://example.sharepoint.com/sites/engineering', scanDocuments: true },
            { id: 'site3', name: 'Public Documents', url: 'https://example.sharepoint.com/sites/public', scanDocuments: false },
        ];
    },

    async updateScope(dataSourceId: ID, config: M365ScopeConfig): Promise<DataSource> {
        // We persist this by updating the 'config' field of the DataSource
        const configJson = JSON.stringify(config);
        const res = await api.put<ApiResponse<DataSource>>(`/data-sources/${dataSourceId}`, { config: configJson });
        return res.data.data;
    }
};
