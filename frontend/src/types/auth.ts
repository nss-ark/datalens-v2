import type { ID, TenantEntity } from './common';

export type UserStatus = 'ACTIVE' | 'INVITED' | 'SUSPENDED';

export interface User extends TenantEntity {
    email: string;
    name: string;
    status: UserStatus;
    role_ids: ID[];
    mfa_enabled: boolean;
    last_login_at?: string;
}

export interface Role extends TenantEntity {
    name: string;
    description: string;
    permissions: Permission[];
    is_system: boolean;
}

export interface Permission {
    resource: string;
    actions: string[];
}

export interface LoginResponse {
    access_token: string;
    refresh_token: string;
    expires_in: number;
}

export interface AuthState {
    user: User | null;
    isAuthenticated: boolean;
    isLoading: boolean;
    token: string | null;
}
