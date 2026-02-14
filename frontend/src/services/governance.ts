import { api } from './api';
import type { ApiResponse } from '../types/common';
import type {
    PurposeSuggestion,
    GovernancePolicy,
    PolicyViolation,
    CreatePolicyRequest,
    LineageGraph
} from '../types/governance';

export const governanceService = {
    // Purpose Suggestions
    getPurposeSuggestions: async (): Promise<PurposeSuggestion[]> => {
        const response = await api.get<ApiResponse<PurposeSuggestion[]>>('/governance/suggestions');
        return response.data.data;
    },

    acceptSuggestion: async (id: string): Promise<void> => {
        await api.post(`/governance/suggestions/${id}/accept`);
    },

    rejectSuggestion: async (id: string): Promise<void> => {
        await api.post(`/governance/suggestions/${id}/reject`);
    },

    // Policies
    getPolicies: async (): Promise<GovernancePolicy[]> => {
        const response = await api.get<ApiResponse<GovernancePolicy[]>>('/governance/policies');
        return response.data.data;
    },

    createPolicy: async (data: CreatePolicyRequest): Promise<GovernancePolicy> => {
        const response = await api.post<ApiResponse<GovernancePolicy>>('/governance/policies', data);
        return response.data.data;
    },

    deletePolicy: async (id: string): Promise<void> => {
        await api.delete(`/governance/policies/${id}`);
    },

    togglePolicyStatus: async (id: string, isActive: boolean): Promise<GovernancePolicy> => {
        const response = await api.patch<ApiResponse<GovernancePolicy>>(`/governance/policies/${id}/status`, { isActive });
        return response.data.data;
    },

    // Violations
    getViolations: async (): Promise<PolicyViolation[]> => {
        const response = await api.get<ApiResponse<PolicyViolation[]>>('/governance/violations');
        return Array.isArray(response.data.data) ? response.data.data : [];
    },

    resolveViolation: async (id: string): Promise<void> => {
        await api.post(`/governance/violations/${id}/resolve`);
    },

    // Data Lineage
    getLineage: async (): Promise<LineageGraph> => {
        const response = await api.get<ApiResponse<LineageGraph>>('/governance/lineage');
        return response.data.data;
    }
};
