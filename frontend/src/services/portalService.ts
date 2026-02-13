import { portalApi } from './portalApi';
import type { ApiResponse, PaginatedResponse } from '../types/common';
import type {
    PortalProfile,
    AuthResponse,
    VerifyOTPInput,
    ConsentSummary,
    ConsentHistoryEntry,
    DPRRequest,
    CreateDPRInput
} from '../types/portal';
import type { IdentityStatusResponse } from '../types/identity';
import type { CreateGrievanceRequest, Grievance } from '../types/grievance';

export const portalService = {
    // --- Auth ---
    async requestOTP(identifier: { email?: string; phone?: string }): Promise<void> {
        await portalApi.post('/public/portal/auth/otp', identifier);
    },

    async verifyOTP(data: VerifyOTPInput): Promise<AuthResponse> {
        const res = await portalApi.post<ApiResponse<AuthResponse>>('/public/portal/auth/verify', data);
        return res.data.data;
    },

    async getProfile(): Promise<PortalProfile> {
        const res = await portalApi.get<ApiResponse<PortalProfile>>('/public/portal/profile');
        return res.data.data;
    },

    // --- Dashboard & Consents ---
    async getConsentSummary(): Promise<ConsentSummary[]> {
        const res = await portalApi.get<ApiResponse<ConsentSummary[]>>('/public/portal/consents');
        return res.data.data;
    },

    async withdrawConsent(purpose_id: string): Promise<void> {
        await portalApi.post('/public/portal/consent/withdraw', { purpose_id });
    },

    async grantConsent(purpose_id: string): Promise<void> { // Re-granting/Opt-in
        await portalApi.post('/public/portal/consent/grant', { purpose_id });
    },

    async getHistory(params?: { page?: number; limit?: number }): Promise<PaginatedResponse<ConsentHistoryEntry>> {
        const res = await portalApi.get<ApiResponse<PaginatedResponse<ConsentHistoryEntry>>>('/public/portal/history', { params });
        return res.data.data;
    },

    // --- DPR Requests ---
    async listRequests(params?: { page?: number; limit?: number }): Promise<PaginatedResponse<DPRRequest>> {
        const res = await portalApi.get<ApiResponse<PaginatedResponse<DPRRequest>>>('/public/portal/dpr', { params });
        return res.data.data;
    },

    async createRequest(data: CreateDPRInput): Promise<DPRRequest> {
        const res = await portalApi.post<ApiResponse<DPRRequest>>('/public/portal/dpr', data);
        return res.data.data;
    },

    async getRequest(id: string): Promise<DPRRequest> {
        const res = await portalApi.get<ApiResponse<DPRRequest>>(`/public/portal/dpr/${id}`);
        return res.data.data;
    },

    // --- Identity Verification ---
    async getIdentityStatus(): Promise<IdentityStatusResponse> {
        // Backend contract: GET /public/portal/identity/status
        const res = await portalApi.get<ApiResponse<IdentityStatusResponse>>('/public/portal/identity/status');
        return res.data.data;
    },

    async linkIdentity(provider: string, authCode: string, redirectUri?: string): Promise<IdentityStatusResponse> {
        // Backend contract: POST /public/portal/identity/link
        const res = await portalApi.post<ApiResponse<IdentityStatusResponse>>('/public/portal/identity/link', {
            provider_name: provider,
            auth_code: authCode,
            redirect_uri: redirectUri
        });
        return res.data.data;
    },

    // --- Grievance Redressal ---
    async submitGrievance(data: CreateGrievanceRequest): Promise<Grievance> {
        const res = await portalApi.post<ApiResponse<Grievance>>('/public/portal/grievance', data);
        return res.data.data;
    },

    async getGrievances(params?: { page?: number; limit?: number }): Promise<PaginatedResponse<Grievance>> {
        const res = await portalApi.get<ApiResponse<PaginatedResponse<Grievance>>>('/public/portal/grievance', { params });
        return res.data.data;
    },

    async getGrievance(id: string): Promise<Grievance> {
        const res = await portalApi.get<ApiResponse<Grievance>>(`/public/portal/grievance/${id}`);
        return res.data.data;
    },

    async submitGrievanceFeedback(id: string, rating: number, comment?: string): Promise<void> {
        await portalApi.post(`/public/portal/grievance/${id}/feedback`, { rating, comment });
    }
};
