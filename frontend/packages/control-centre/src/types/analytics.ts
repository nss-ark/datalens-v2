import type { ID } from '@datalens/shared';

export interface ConversionStat {
    date: string;
    total_sessions: number;
    opt_in_count: number;
    opt_out_count: number;
    conversion_rate: number;
}

export interface PurposeStat {
    purpose_id: ID;
    purpose_name: string;
    granted_count: number;
    denied_count: number;
    acceptance_rate: number;
}

export interface AnalyticsFilter {
    from: string; // YYYY-MM-DD
    to: string; // YYYY-MM-DD
    interval?: 'day' | 'week' | 'month';
}
