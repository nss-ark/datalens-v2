import type { BaseEntity } from './common';

export type VerificationStatus = 'NONE' | 'BASIC' | 'SUBSTANTIAL' | 'HIGH';

export interface IdentityProfile extends BaseEntity {
    user_id: string;
    assurance_level: VerificationStatus;
    full_name?: string;
    date_of_birth?: string; // YYYY-MM-DD
    gender?: string;
    masked_id?: string; // e.g., XXXXXX1234
    provider_name?: string; // e.g., DigiLocker
    provider_id?: string;
    verified_at?: string; // ISO timestamp
    valid_until?: string; // ISO timestamp
    claims?: Record<string, string>;
}

export interface IdentitySettings {
    enable_digilocker: boolean;
    require_govt_id_for_dsr: boolean;
    fallback_to_email_otp: boolean;
    allowed_providers: string[];
    min_assurance_level: VerificationStatus;
}

export interface IdentityStatusResponse {
    assurance_level: VerificationStatus;
    profile?: IdentityProfile;
}

export interface LinkIdentityRequest {
    provider_name: string;
    auth_code: string;
    redirect_uri?: string;
}
