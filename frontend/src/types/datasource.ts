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
