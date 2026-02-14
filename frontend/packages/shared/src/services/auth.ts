import { api } from './api';
import type { User } from '../types/auth';
import type { ApiError } from '../types/common';

export interface LoginRequest {
    domain: string;
    email: string;
    password: string;
}

export interface RegisterRequest {
    tenant_name: string;
    domain: string;
    industry: string;
    country: string;
    email: string;
    name: string;
    password: string;
}

export interface TokenPair {
    access_token: string;
    refresh_token: string;
    expires_at: string;
}

export interface ApiResponse<T> {
    success: boolean;
    data: T;
    error?: ApiError;
    meta?: Record<string, unknown>;
}

export const authService = {
    async login(data: LoginRequest): Promise<TokenPair> {
        const res = await api.post<ApiResponse<TokenPair>>('/auth/login', data);
        return res.data.data;
    },

    async register(data: RegisterRequest): Promise<unknown> {
        const res = await api.post<ApiResponse<unknown>>('/auth/register', data);
        return res.data.data;
    },

    async refreshToken(refreshToken: string): Promise<TokenPair> {
        const res = await api.post<ApiResponse<TokenPair>>('/auth/refresh', { refresh_token: refreshToken });
        return res.data.data;
    },

    async getMe(): Promise<User> {
        const res = await api.get<ApiResponse<User>>('/users/me');
        return res.data.data;
    },

    async logout(): Promise<void> {
        // Optimistically clear local storage first
        localStorage.removeItem('token');
        localStorage.removeItem('refresh_token');

        try {
            // Attempt to notify backend, but don't block on it
            await api.post('/auth/logout');
        } catch (error) {
            // Ignore errors during logout (e.g. 401 if already expired)
            console.warn('Backend logout failed', error);
        }
    }
};
