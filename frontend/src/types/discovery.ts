import type { ID, BaseEntity } from './common';

// ─── Enums ──────────────────────────────────────────────────────────────

export type PIICategory =
    | 'IDENTITY' | 'CONTACT' | 'FINANCIAL' | 'HEALTH'
    | 'BIOMETRIC' | 'GENETIC' | 'LOCATION' | 'BEHAVIORAL'
    | 'PROFESSIONAL' | 'GOVERNMENT_ID' | 'MINOR';

export type PIIType =
    | 'NAME' | 'EMAIL' | 'PHONE' | 'ADDRESS'
    | 'AADHAAR' | 'PAN' | 'PASSPORT' | 'SSN' | 'NATIONAL_ID'
    | 'DATE_OF_BIRTH' | 'GENDER'
    | 'BANK_ACCOUNT' | 'CREDIT_CARD'
    | 'IP_ADDRESS' | 'MAC_ADDRESS' | 'DEVICE_ID'
    | 'BIOMETRIC' | 'MEDICAL_RECORD' | 'PHOTO' | 'SIGNATURE';

export type SensitivityLevel = 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';

export type DetectionMethod = 'AI' | 'REGEX' | 'HEURISTIC' | 'INDUSTRY' | 'MANUAL';

export type VerificationStatus = 'PENDING' | 'VERIFIED' | 'REJECTED';

export type FeedbackType = 'VERIFIED' | 'CORRECTED' | 'REJECTED';

// ─── Entities ───────────────────────────────────────────────────────────

export interface PIIClassification extends BaseEntity {
    field_id: ID;
    data_source_id: ID;
    entity_name: string;
    field_name: string;
    category: PIICategory;
    type: PIIType;
    sensitivity: SensitivityLevel;
    confidence: number;
    detection_method: DetectionMethod;
    status: VerificationStatus;
    verified_by?: ID | null;
    verified_at?: string | null;
    reasoning: string;
}

export interface SubmitFeedbackInput {
    classification_id: ID;
    feedback_type: FeedbackType;
    corrected_category?: PIICategory;
    corrected_type?: PIIType;
    notes: string;
}

export interface DetectionFeedback extends BaseEntity {
    classification_id: ID;
    tenant_id: ID;
    feedback_type: FeedbackType;
    original_category: PIICategory;
    original_type: PIIType;
    original_confidence: number;
    original_method: DetectionMethod;
    corrected_category?: PIICategory | null;
    corrected_type?: PIIType | null;
    corrected_by: ID;
    corrected_at: string;
    notes: string;
    column_name: string;
    table_name: string;
}

export interface AccuracyStats {
    method: DetectionMethod;
    total: number;
    verified: number;
    corrected: number;
    rejected: number;
    accuracy: number;
}

export interface FeedbackResponse {
    feedback: DetectionFeedback;
    classification: PIIClassification;
}

// ─── Constants for dropdowns ────────────────────────────────────────────

export const PII_CATEGORIES: { value: PIICategory; label: string }[] = [
    { value: 'IDENTITY', label: 'Identity' },
    { value: 'CONTACT', label: 'Contact' },
    { value: 'FINANCIAL', label: 'Financial' },
    { value: 'HEALTH', label: 'Health' },
    { value: 'BIOMETRIC', label: 'Biometric' },
    { value: 'GENETIC', label: 'Genetic' },
    { value: 'LOCATION', label: 'Location' },
    { value: 'BEHAVIORAL', label: 'Behavioral' },
    { value: 'PROFESSIONAL', label: 'Professional' },
    { value: 'GOVERNMENT_ID', label: 'Government ID' },
    { value: 'MINOR', label: 'Minor' },
];

export const PII_TYPES: { value: PIIType; label: string }[] = [
    { value: 'NAME', label: 'Name' },
    { value: 'EMAIL', label: 'Email' },
    { value: 'PHONE', label: 'Phone' },
    { value: 'ADDRESS', label: 'Address' },
    { value: 'AADHAAR', label: 'Aadhaar' },
    { value: 'PAN', label: 'PAN' },
    { value: 'PASSPORT', label: 'Passport' },
    { value: 'SSN', label: 'SSN' },
    { value: 'NATIONAL_ID', label: 'National ID' },
    { value: 'DATE_OF_BIRTH', label: 'Date of Birth' },
    { value: 'GENDER', label: 'Gender' },
    { value: 'BANK_ACCOUNT', label: 'Bank Account' },
    { value: 'CREDIT_CARD', label: 'Credit Card' },
    { value: 'IP_ADDRESS', label: 'IP Address' },
    { value: 'MAC_ADDRESS', label: 'MAC Address' },
    { value: 'DEVICE_ID', label: 'Device ID' },
    { value: 'BIOMETRIC', label: 'Biometric' },
    { value: 'MEDICAL_RECORD', label: 'Medical Record' },
    { value: 'PHOTO', label: 'Photo' },
    { value: 'SIGNATURE', label: 'Signature' },
];

export const DETECTION_METHODS: { value: DetectionMethod; label: string }[] = [
    { value: 'AI', label: 'AI' },
    { value: 'REGEX', label: 'Regex' },
    { value: 'HEURISTIC', label: 'Heuristic' },
    { value: 'INDUSTRY', label: 'Industry' },
    { value: 'MANUAL', label: 'Manual' },
];
