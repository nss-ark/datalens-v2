import type { ID, BaseEntity } from './common';

export type DataSourceType = 'postgresql' | 'mysql' | 'mongodb' | 'mssql' | 'oracle' | 'sqlite' | 's3' | 'gcs' | 'azure_blob' | 'google_workspace' | 'onedrive' | 'sharepoint' | 'outlook' | 'm365';

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
    config?: string; // JSON string for scope configuration
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
    config?: string; // JSON string for scope configuration
}

// --- M365 Scope Configuration Types ---

export interface M365User {
    id: string; // User Principal Name or ID
    displayName: string;
    email: string;
    scanOneDrive: boolean;
    scanOutlook: boolean;
}

export interface SharePointSite {
    id: string; // Site ID or URL
    name: string;
    url: string;
    scanDocuments: boolean;
}

export interface M365ScopeConfig {
    users: M365User[];
    sites: SharePointSite[];
    excludedExtensions?: string[];
    scanAllUsers?: boolean; // If true, automatically scan new users
    scanAllSites?: boolean; // If true, automatically scan new sites
}

export interface GoogleScopeConfig {
    scanMyDrive: boolean;
    scanSharedDrives: boolean;
    scanGmail: boolean;
    excludedExtensions?: string[];
}
