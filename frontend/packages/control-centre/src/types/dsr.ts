import type { ID } from '@datalens/shared';

// DSR Status — mirrors backend compliance.DSRStatus
export type DSRStatus =
    | 'PENDING'
    | 'IDENTITY_VERIFICATION'
    | 'APPROVED'
    | 'IN_PROGRESS'
    | 'COMPLETED'
    | 'REJECTED'
    | 'FAILED';

// DSR Request Type — mirrors backend compliance.DSRRequestType
export type DSRRequestType = 'ACCESS' | 'ERASURE' | 'CORRECTION' | 'PORTABILITY';

// DSR Task Status
export type DSRTaskStatus = 'PENDING' | 'RUNNING' | 'COMPLETED' | 'VERIFIED' | 'FAILED';

// DSR Priority
export type DSRPriority = 'HIGH' | 'MEDIUM' | 'LOW';

// DSR entity — mirrors backend compliance.DSR
export interface DSR {
    id: ID;
    tenant_id: ID;
    request_type: DSRRequestType;
    status: DSRStatus;
    subject_name: string;
    subject_email: string;
    subject_identifiers: Record<string, string>;
    priority: DSRPriority;
    sla_deadline: string;
    assigned_to?: ID;
    reason?: string;
    created_at: string;
    updated_at: string;
    completed_at?: string;
}

// DSR Task — mirrors backend compliance.DSRTask
export interface DSRTask {
    id: ID;
    dsr_id: ID;
    data_source_id: ID;
    tenant_id: ID;
    task_type: DSRRequestType;
    status: DSRTaskStatus;
    result?: unknown;
    error?: string;
    created_at: string;
    updated_at: string;
    completed_at?: string;
}

// DSR with embedded tasks — returned by GET /dsr/{id}
export interface DSRWithTasks extends DSR {
    tasks: DSRTask[];
}

// Input for creating a new DSR
export interface CreateDSRInput {
    request_type: DSRRequestType;
    subject_name: string;
    subject_email: string;
    subject_identifiers: Record<string, string>;
    priority: DSRPriority;
}

// Paginated DSR list response
export interface DSRListResponse {
    items: DSR[];
    total: number;
    page: number;
    page_size: number;
    total_pages: number;
}
