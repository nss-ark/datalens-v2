import { api } from '@datalens/shared';
import type { ApiResponse } from '@datalens/shared';

// --- Types ---

export interface PurposeAssignment {
    id: string;
    tenant_id: string;
    purpose_id: string;
    scope_type: 'SERVER' | 'DATABASE' | 'SCHEMA' | 'TABLE' | 'COLUMN';
    scope_id: string;
    scope_name?: string;
    inherited: boolean;
    overridden_by?: string;
    assigned_by?: string;
    assigned_at: string;
}

export interface AssignPurposeInput {
    purpose_id: string;
    scope_type: string;
    scope_id: string;
    scope_name?: string;
}

// --- Service ---

export const purposeAssignmentService = {
    async assign(data: AssignPurposeInput): Promise<PurposeAssignment> {
        const res = await api.post<ApiResponse<PurposeAssignment>>('/purpose-assignments', data);
        return res.data.data;
    },

    async remove(id: string): Promise<void> {
        await api.delete(`/purpose-assignments/${id}`);
    },

    async getByScope(scopeType: string, scopeId: string): Promise<PurposeAssignment[]> {
        const res = await api.get<ApiResponse<PurposeAssignment[]>>('/purpose-assignments', {
            params: { scope_type: scopeType, scope_id: scopeId },
        });
        return res.data.data;
    },

    async getEffective(scopeType: string, scopeId: string): Promise<PurposeAssignment[]> {
        const res = await api.get<ApiResponse<PurposeAssignment[]>>('/purpose-assignments/effective', {
            params: { scope_type: scopeType, scope_id: scopeId },
        });
        return res.data.data;
    },

    async getAll(): Promise<PurposeAssignment[]> {
        const res = await api.get<ApiResponse<PurposeAssignment[]>>('/purpose-assignments/all');
        return res.data.data;
    },
};
