import type { ID, TenantEntity } from '@datalens/shared';

export interface PortalProfile extends TenantEntity {
    email: string;
    phone?: string;
    verification_status: 'PENDING' | 'VERIFIED';
    verified_at?: string;
    subject_id?: ID;
    last_access_at?: string;
    preferred_lang: string;
    is_minor: boolean;
    dob?: string;
    guardian_email?: string;
    guardian_verified: boolean;
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
    appeal_of?: ID;
    appeal_reason?: string;
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
    code: string;
}

export interface AuthResponse {
    token: {
        access_token: string;
        expires_in: number;
        token_type: string;
    };
    profile: PortalProfile;
}

export interface ConsentSummary {
    purpose_id: ID;
    purpose_name: string;
    status: 'GRANTED' | 'DENIED' | 'WITHDRAWN';
    last_updated: string;
}

export interface BreachNotification extends TenantEntity {
    incident_id: ID;
    data_principal_id: ID;
    title: string;
    severity: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
    occurred_at: string;
    description: string;
    affected_data: string[];
    what_we_are_doing: string;
    contact_email: string;
    is_read: boolean;
    created_at: string;
}

/* ── Activity Feed ── */
export type ActivityItemType = 'login' | 'request_update' | 'consent_update' | 'security_digest' | 'breach';
export type ActivityCategory = 'ALL' | 'REQUESTS' | 'PRIVACY';

export interface ActivityFeedItem {
    id: string;
    type: ActivityItemType;
    title: string;
    description: string;
    timestamp: string;               // ISO date
    category: ActivityCategory;
    is_read: boolean;
    /** Optional category label shown as a small tag */
    category_label?: string;
    /** Primary action, e.g. "Review Changes" */
    primary_action?: { label: string; href?: string };
    /** Secondary action pair, e.g. "This was me" / "Not me" */
    secondary_actions?: { label: string; variant: 'default' | 'danger'; href?: string }[];
    /** Reference to original breach notification, if type === 'breach' */
    breach_ref?: BreachNotification;
}
