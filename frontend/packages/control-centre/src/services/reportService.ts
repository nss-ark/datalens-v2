import { api } from '@datalens/shared';
import type { ApiResponse } from '@datalens/shared';

// ── Types ────────────────────────────────────────────────────────────────

export interface CompliancePillarConsent {
    score: number;
    total_consents: number;
    active_consents: number;
    withdrawal_rate: string;
    notices_published: number;
    notices_compliant: number;
}

export interface CompliancePillarDSR {
    score: number;
    total_requests: number;
    completed_on_time: number;
    overdue: number;
    avg_resolution_days: number;
    by_type: Record<string, number>;
}

export interface CompliancePillarBreach {
    score: number;
    total_incidents: number;
    reported_to_cert_in: number;
    avg_notification_hours: number;
}

export interface CompliancePillarGovernance {
    score: number;
    departments_with_owners: number;
    departments_total: number;
    third_parties_with_dpa: number;
    third_parties_total: number;
    purposes_mapped: number;
    ropa_published: boolean;
    ropa_version: string;
}

export interface ComplianceSnapshot {
    generated_at: string;
    period: { from: string; to: string };
    overall_score: number;
    pillars: {
        consent_management: CompliancePillarConsent;
        dsr_compliance: CompliancePillarDSR;
        breach_management: CompliancePillarBreach;
        data_governance: CompliancePillarGovernance;
    };
    recommendations: string[];
}

export type ExportEntity =
    | 'dsr'
    | 'breaches'
    | 'consent-records'
    | 'audit-logs'
    | 'departments'
    | 'third-parties'
    | 'purposes';

// ── Service ──────────────────────────────────────────────────────────────

export const reportService = {
    async getComplianceSnapshot(from?: string, to?: string): Promise<ComplianceSnapshot> {
        const params: Record<string, string> = {};
        if (from) params.from = from;
        if (to) params.to = to;
        const res = await api.get<ApiResponse<ComplianceSnapshot>>('/reports/compliance-snapshot', { params });
        return res.data.data;
    },

    exportEntity(entity: ExportEntity, format: 'csv' | 'json' = 'csv') {
        const baseURL = api.defaults.baseURL || '';
        window.open(`${baseURL}/reports/export/${entity}?format=${format}`, '_blank');
    },
};
