import { api } from './api';
import type { ApiResponse } from '../types/common';
import type { IdentitySettings } from '../types/identity';

export const identityService = {
    async getSettings(): Promise<IdentitySettings> {
        // Mock response until backend is ready if needed, but assuming endpoint exists
        // If backend doesn't have settings endpoint yet, we might mock it here
        try {
            const res = await api.get<ApiResponse<IdentitySettings>>('/identity/settings');
            return res.data.data;
        } catch (error) {
            console.warn('Failed to fetch identity settings, using defaults', error);
            // Fallback defaults
            return {
                enable_digilocker: false,
                require_govt_id_for_dsr: false,
                fallback_to_email_otp: true,
                allowed_providers: [],
                min_assurance_level: 'NONE'
            };
        }
    },

    async updateSettings(settings: Partial<IdentitySettings>): Promise<IdentitySettings> {
        const res = await api.put<ApiResponse<IdentitySettings>>('/identity/settings', settings);
        return res.data.data;
    }
};
