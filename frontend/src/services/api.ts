import axios, { type AxiosError, type InternalAxiosRequestConfig } from 'axios';
import { useAuthStore } from '../stores/authStore';
import { toast } from '../stores/toastStore';

// Create axios instance with base URL
export const api = axios.create({
    baseURL: import.meta.env.VITE_API_URL || '/api/v2',
    headers: {
        'Content-Type': 'application/json',
    },
});

// Request interceptor to add auth token
api.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
        const token = useAuthStore.getState().token;
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error: AxiosError) => {
        return Promise.reject(error);
    }
);

// Response interceptor to handle errors
api.interceptors.response.use(
    (response) => response,
    (error: AxiosError<{ message?: string; error?: { message?: string } }>) => {
        const status = error.response?.status;

        // Handle 401 Unauthorized globally
        if (status === 401) {
            useAuthStore.getState().logout();
            window.location.href = '/login';
            return Promise.reject(error);
        }

        // Show toast for server errors (skip if component handles it)
        if (status && status >= 500) {
            const msg = error.response?.data?.message
                || error.response?.data?.error?.message
                || 'An unexpected server error occurred.';
            toast.error('Server Error', msg);
        }

        return Promise.reject(error);
    }
);

