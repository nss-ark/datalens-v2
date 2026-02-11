import { api } from './api';
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
        const response = await api.get('/governance/suggestions');
        return response.data;
    },

    acceptSuggestion: async (id: string): Promise<void> => {
        await api.post(`/governance/suggestions/${id}/accept`);
    },

    rejectSuggestion: async (id: string): Promise<void> => {
        await api.post(`/governance/suggestions/${id}/reject`);
    },

    // Policies
    getPolicies: async (): Promise<GovernancePolicy[]> => {
        const response = await api.get('/governance/policies');
        return response.data;
    },

    createPolicy: async (data: CreatePolicyRequest): Promise<GovernancePolicy> => {
        const response = await api.post('/governance/policies', data);
        return response.data;
    },

    deletePolicy: async (id: string): Promise<void> => {
        await api.delete(`/governance/policies/${id}`);
    },

    togglePolicyStatus: async (id: string, isActive: boolean): Promise<GovernancePolicy> => {
        const response = await api.patch(`/governance/policies/${id}/status`, { isActive });
        return response.data;
    },

    // Violations
    getViolations: async (): Promise<PolicyViolation[]> => {
        const response = await api.get('/governance/violations');
        return response.data;
    },

    resolveViolation: async (id: string): Promise<void> => {
        await api.post(`/governance/violations/${id}/resolve`);
    },

    // Data Lineage
    getLineage: async (): Promise<LineageGraph> => {
        const response = await api.get('/governance/lineage');
        return response.data;
    }
};
