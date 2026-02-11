import type { ID, TenantEntity } from './common';

export interface PortalProfile extends TenantEntity {
    email: string;
    phone?: string;
    verification_status: 'PENDING' | 'VERIFIED';
    verified_at?: string;
    subject_id?: ID;
    last_access_at?: string;
    preferred_lang: string;
}

export interface ConsentHistoryEntry extends TenantEntity {
    subject_id: ID;
    widget_id?: ID;
    widget_name?: string;
    purpose_id: ID;
    purpose_name: string;
    previous_status?: 'GRANTED' | 'DENIED' | 'WITHDRAWN' | 'EXPIRED' | 'PENDING';
    new_status: 'GRANTED' | 'DENIED' | 'WITHDRAWN' | 'EXPIRED' | 'PENDING';
    source: 'BANNER' | 'PORTAL' | 'API' | 'IMPORT';
    notice_version: string;
    signature: string;
}

export interface DPRRequest extends TenantEntity {
    profile_id: ID;
    dsr_id?: ID; // ID of the internal DSR if created
    type: 'ACCESS' | 'CORRECTION' | 'ERASURE' | 'NOMINATION' | 'GRIEVANCE';
    description?: string;
    status: 'SUBMITTED' | 'PENDING_VERIFICATION' | 'VERIFIED' | 'IN_PROGRESS' | 'COMPLETED' | 'REJECTED' | 'APPEALED' | 'ESCALATED';
    submitted_at: string;
    deadline?: string;
    completed_at?: string;
    response_summary?: string;
    download_url?: string;
    is_minor: boolean;
    guardian_name?: string;
    guardian_email?: string;
    guardian_verified: boolean;
}

export interface CreateDPRInput {
    type: DPRRequest['type'];
    description: string;
    is_minor?: boolean;
    guardian_name?: string;
    guardian_email?: string;
}

export interface VerifyOTPInput {
    email?: string;
    phone?: string;
    otp: string;
}

export interface AuthResponse {
    token: string; // Session token
    profile: PortalProfile;
}

export interface ConsentSummary {
    purpose_id: ID;
    purpose_name: string;
    status: 'GRANTED' | 'DENIED' | 'WITHDRAWN';
    last_updated: string;
}
