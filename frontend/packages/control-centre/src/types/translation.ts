import type { BaseEntity, ID } from '@datalens/shared';

export type SupportedLanguage =
    | 'en' | 'hi' | 'bn' | 'te' | 'mr' | 'ta' | 'gu' | 'kn' | 'ml' | 'pa' | 'or' | 'as' | 'ur' // Top Indic
    | 'es' | 'fr' | 'de' | 'it' | 'pt' | 'ru' | 'ja' | 'zh' | 'ko'; // International

export const SUPPORTED_LANGUAGES: { code: string; name: string }[] = [
    { code: 'hi', name: 'Hindi' }, { code: 'bn', name: 'Bengali' }, { code: 'te', name: 'Telugu' },
    { code: 'mr', name: 'Marathi' }, { code: 'ta', name: 'Tamil' }, { code: 'gu', name: 'Gujarati' },
    { code: 'kn', name: 'Kannada' }, { code: 'ml', name: 'Malayalam' }, { code: 'pa', name: 'Punjabi' },
    { code: 'or', name: 'Odia' }, { code: 'as', name: 'Assamese' }, { code: 'ur', name: 'Urdu' },
    { code: 'en', name: 'English' }, { code: 'es', name: 'Spanish' }, { code: 'fr', name: 'French' },
    { code: 'de', name: 'German' }, { code: 'it', name: 'Italian' }, { code: 'pt', name: 'Portuguese' },
    { code: 'ru', name: 'Russian' }, { code: 'ja', name: 'Japanese' }, { code: 'zh', name: 'Chinese' },
    { code: 'ko', name: 'Korean' }
];

export type TranslationStatus = 'PENDING' | 'COMPLETED' | 'FAILED' | 'OUTDATED' | 'MANUAL';

export interface Translation extends BaseEntity {
    notice_id: ID;
    language: string;
    content: Record<string, string>; // Key-value pairs of translated text
    status: TranslationStatus;
    last_translated_at?: string;
    provider: 'AI' | 'MANUAL';
}

export interface TranslationResult {
    language: string;
    status: TranslationStatus;
    translation?: Translation;
    error?: string;
}
