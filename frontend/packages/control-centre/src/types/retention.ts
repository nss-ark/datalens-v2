import type { ID } from '@datalens/shared';

export type RetentionStatus = 'ACTIVE' | 'PAUSED';

export interface RetentionPolicy {
    id: ID;
    tenant_id: ID;
    purpose_id: ID;
    max_retention_days: number;
    data_categories: string[];
    status: RetentionStatus;
    auto_erase: boolean;
    description: string;
    created_at: string;
    updated_at: string;
}

export interface CreateRetentionPolicyInput {
    purpose_id: ID;
    max_retention_days: number;
    data_categories: string[];
    status: RetentionStatus;
    auto_erase: boolean;
    description: string;
}

export interface UpdateRetentionPolicyInput {
    purpose_id?: ID;
    max_retention_days?: number;
    data_categories?: string[];
    status?: RetentionStatus;
    auto_erase?: boolean;
    description?: string;
}
