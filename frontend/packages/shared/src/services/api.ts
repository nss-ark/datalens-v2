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
        console.log('[API Interceptor] Token:', token);
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
    async (error: AxiosError<{ message?: string; error?: { message?: string } }>) => {
        const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };
        const status = error.response?.status;

        // Handle 401 Unauthorized globally with Token Refresh
        if (status === 401 && !originalRequest._retry) {
            originalRequest._retry = true;
            const state = useAuthStore.getState();
            const refreshToken = state.refreshToken;

            if (refreshToken) {
                try {
                    // Call refresh endpoint with a clean axios instance to avoid interceptor loops
                    const response = await axios.post<{ data: { access_token: string; refresh_token: string } }>(
                        `${api.defaults.baseURL}/auth/refresh`,
                        { refresh_token: refreshToken }
                    );

                    const { access_token, refresh_token } = response.data.data;

                    // Update store
                    useAuthStore.setState({
                        token: access_token,
                        refreshToken: refresh_token
                    });

                    // Update header and retry original request
                    originalRequest.headers.Authorization = `Bearer ${access_token}`;
                    return api(originalRequest);

                } catch (refreshError) {
                    console.error('Available refresh token failed to rotate:', refreshError);
                    state.logout();
                    window.location.href = '/login';
                    return Promise.reject(refreshError);
                }
            }

            // No refresh token available
            state.logout();
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
