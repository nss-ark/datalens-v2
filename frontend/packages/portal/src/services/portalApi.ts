import axios, { AxiosError, type InternalAxiosRequestConfig } from 'axios';
import { usePortalAuthStore } from '../stores/portalAuthStore';

// Separate axios instance for Portal to keep auth distinct from Admin
export const portalApi = axios.create({
    baseURL: import.meta.env.VITE_API_URL || '/api/v2', // Same base, different endpoints
    headers: {
        'Content-Type': 'application/json',
    },
});

// Request interceptor: Add Portal Token
portalApi.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
        const token = usePortalAuthStore.getState().token;
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error: AxiosError) => {
        return Promise.reject(error);
    }
);

// Response interceptor: Handle Portal 401s
portalApi.interceptors.response.use(
    (response) => response,
    (error: AxiosError) => {
        const status = error.response?.status;
        if (status === 401) {
            usePortalAuthStore.getState().logout();
            // Optional: Redirect to portal login if we have a way to know we are in portal context
            if (window.location.pathname.startsWith('/portal') && !window.location.pathname.includes('/login')) {
                window.location.href = '/portal/login';
            }
        }
        return Promise.reject(error);
    }
);
