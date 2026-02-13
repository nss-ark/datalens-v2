// =============================================================================
// i18n — Translation helper with browser language detection
// =============================================================================

import { WidgetSettings } from './types';

/** Default English strings used as fallback */
const DEFAULT_STRINGS: Record<string, string> = {
    title: 'We value your privacy',
    description: 'We use cookies and similar technologies to improve your experience.',
    accept_all: 'Accept All',
    reject_all: 'Reject All',
    customize: 'Customize',
    save_preferences: 'Save Preferences',
    essential_label: 'Essential (Always Active)',
    powered_by: 'Powered by DataLens',
    close: 'Close',
    preferences_title: 'Privacy Preferences',
    preferences_description: 'Choose which categories of data processing you consent to.',
};

let currentLang = 'en';
let translations: Record<string, Record<string, string>> = {};

/**
 * Initialize i18n with widget config translations.
 * Auto-detects browser language, falls back to default_language.
 */
export function initI18n(config: WidgetSettings): void {
    translations = config.translations || {};

    // Detect browser language
    const browserLang = detectBrowserLanguage();
    const available = config.languages || ['en'];

    if (available.includes(browserLang)) {
        currentLang = browserLang;
    } else {
        // Try matching just the language code (e.g., "hi" from "hi-IN")
        const shortCode = browserLang.split('-')[0];
        if (available.includes(shortCode)) {
            currentLang = shortCode;
        } else {
            currentLang = config.default_language || 'en';
        }
    }
}

/**
 * Get a translated string by key.
 * Falls back: current language → default language → hardcoded English.
 */
export function t(key: string): string {
    // Try current language
    if (translations[currentLang]?.[key]) {
        return translations[currentLang][key];
    }
    // Fall back to English translations from config
    if (translations['en']?.[key]) {
        return translations['en'][key];
    }
    // Fall back to hardcoded defaults
    return DEFAULT_STRINGS[key] || key;
}

/**
 * Get the current active language code.
 */
export function getCurrentLanguage(): string {
    return currentLang;
}

/**
 * Check if the current language is RTL.
 */
export function isRTL(): boolean {
    const rtlLangs = ['ar', 'ur', 'ks', 'sd', 'he', 'fa'];
    return rtlLangs.includes(currentLang);
}

/**
 * Detect the browser's preferred language.
 */
function detectBrowserLanguage(): string {
    if (typeof navigator !== 'undefined') {
        return navigator.language || (navigator as any).userLanguage || 'en';
    }
    return 'en';
}
