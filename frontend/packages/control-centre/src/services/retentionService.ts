import { api } from '@datalens/shared';
import type { ApiResponse, ID } from '@datalens/shared';
import type { RetentionPolicy, CreateRetentionPolicyInput, UpdateRetentionPolicyInput } from '../types/retention';

export const retentionService = {
    async listPolicies(): Promise<RetentionPolicy[]> {
        const res = await api.get<ApiResponse<RetentionPolicy[]>>('/v2/retention');
        return res.data.data;
    },

    async createPolicy(data: CreateRetentionPolicyInput): Promise<RetentionPolicy> {
        const res = await api.post<ApiResponse<RetentionPolicy>>('/v2/retention', data);
        return res.data.data;
    },

    async updatePolicy(id: ID, data: UpdateRetentionPolicyInput): Promise<RetentionPolicy> {
        const res = await api.put<ApiResponse<RetentionPolicy>>(`/v2/retention/${id}`, data);
        return res.data.data;
    },

    async deletePolicy(id: ID): Promise<void> {
        await api.delete(`/v2/retention/${id}`);
    }
};
