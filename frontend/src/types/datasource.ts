import type { ID, BaseEntity } from './common';

export type DataSourceType = 'postgresql' | 'mysql' | 'mongodb' | 'mssql' | 'oracle' | 'sqlite' | 's3' | 'gcs' | 'azure_blob';

export type ConnectionStatus = 'CONNECTED' | 'DISCONNECTED' | 'ERROR' | 'TESTING';

export interface DataSource extends BaseEntity {
    tenant_id: ID;
    name: string;
    type: DataSourceType;
    description: string;
    host?: string;
    port?: number;
    database?: string;
    status: ConnectionStatus;
    last_sync_at: string | null;
    error_message?: string | null;
}

export type ScanStatus = 'QUEUED' | 'RUNNING' | 'COMPLETED' | 'FAILED';

export interface ScanProgress {
    status: ScanStatus;
    progress_percentage: number;
    tables_processed: number;
    total_tables: number;
    pii_found: number;
    current_table?: string;
}

export interface ScanHistoryItem {
    id: string;
    data_source_id: string;
    status: ScanStatus;
    tables_scanned: number;
    pii_found: number;
    started_at: string;
    completed_at: string | null;
    error_message?: string;
}

export interface CreateDataSourceInput {
    name: string;
    type: DataSourceType;
    description: string;
    host: string;
    port: number;
    database: string;
    credentials: string;
}

export interface UpdateDataSourceInput {
    name?: string;
    description?: string;
    host?: string;
    port?: number;
    database?: string;
    credentials?: string;
}
