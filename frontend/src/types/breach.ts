import type { TenantEntity } from './common';

export type IncidentStatus = 'OPEN' | 'INVESTIGATING' | 'CONTAINED' | 'RESOLVED' | 'REPORTED' | 'CLOSED';
export type IncidentSeverity = 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';

export interface BreachIncident extends TenantEntity {
    title: string;
    description: string;
    type: string;
    severity: IncidentSeverity;
    status: IncidentStatus;

    // Timestamps
    detected_at: string;
    occurred_at: string;
    reported_to_cert_in_at?: string;
    reported_to_dpb_at?: string;
    closed_at?: string;

    // Impact
    affected_systems: string[];
    affected_data_subject_count: number;
    pii_categories: string[];

    // Response
    is_reportable_cert_in: boolean;
    is_reportable_dpb: boolean;

    // PoC
    poc_name: string;
    poc_role: string;
    poc_email: string;
}

export interface CreateIncidentInput {
    title: string;
    description: string;
    type: string;
    severity: IncidentSeverity;
    detected_at: string;
    occurred_at: string;
    affected_systems: string[];
    affected_data_subject_count: number;
    pii_categories: string[];
    poc_name: string;
    poc_role: string;
    poc_email: string;
}

export interface UpdateIncidentInput {
    title?: string;
    description?: string;
    type?: string;
    severity?: IncidentSeverity;
    status?: IncidentStatus;
    detected_at?: string; // ISO string
    occurred_at?: string; // ISO string
    reported_to_cert_in_at?: string;
    reported_to_dpb_at?: string;
    closed_at?: string;
    affected_systems?: string[];
    affected_data_subject_count?: number;
    pii_categories?: string[];
    poc_name?: string;
    poc_role?: string;
    poc_email?: string;
}

export interface BreachFilter {
    status?: IncidentStatus;
    severity?: IncidentSeverity;
}

export interface SLAData {
    time_remaining_cert_in: string; // duration string
    time_remaining_dpb: string;     // duration string
    cert_in_deadline: string;       // ISO timestamp
    dpb_deadline: string;           // ISO timestamp
    overdue_cert_in: boolean;
    overdue_dpb: boolean;
}

export interface IncidentDetailResponse {
    incident: BreachIncident;
    sla: SLAData;
}
