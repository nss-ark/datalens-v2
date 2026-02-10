import { api } from './api';
import type { User } from '../types/auth';

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

export const authService = {
    async login(data: LoginRequest): Promise<TokenPair> {
        const res = await api.post<TokenPair>('/auth/login', data);
        return res.data;
    },

    async register(data: RegisterRequest): Promise<unknown> {
        const res = await api.post('/auth/register', data);
        return res.data;
    },

    async refreshToken(refreshToken: string): Promise<TokenPair> {
        const res = await api.post<TokenPair>('/auth/refresh', { refresh_token: refreshToken });
        return res.data;
    },

    async getMe(): Promise<User> {
        const res = await api.get<User>('/users/me');
        return res.data;
    },
};
