import { api } from '@datalens/shared';
import type { ApiResponse, PaginatedResponse, ID } from '@datalens/shared';
import type { Grievance, GrievanceFilter } from '../types/grievance';

export const grievanceService = {
    async listGrievances(params?: GrievanceFilter & { page?: number; page_size?: number }): Promise<PaginatedResponse<Grievance>> {
        const res = await api.get<ApiResponse<PaginatedResponse<Grievance>>>('/compliance/grievances', { params });
        return res.data.data;
    },

    async getGrievance(id: ID): Promise<Grievance> {
        const res = await api.get<ApiResponse<Grievance>>(`/compliance/grievances/${id}`);
        return res.data.data;
    },

    async assignGrievance(id: ID, userId: ID): Promise<void> {
        await api.post(`/compliance/grievances/${id}/assign`, { user_id: userId });
    },

    async resolveGrievance(id: ID, resolution: string): Promise<void> {
        await api.post(`/compliance/grievances/${id}/resolve`, { resolution });
    },

    async escalateGrievance(id: ID, authority: string): Promise<void> {
        await api.post(`/compliance/grievances/${id}/escalate`, { authority });
    }
};
