// =============================================================================
// Consent Storage — localStorage with cookie fallback
// =============================================================================

import { StoredConsent, ConsentDecision } from './types';

const STORAGE_PREFIX = 'dl_consent_';

/**
 * Get the storage key for a widget.
 */
function storageKey(widgetId: string): string {
    return `${STORAGE_PREFIX}${widgetId}`;
}

/**
 * Save consent decisions to localStorage (with cookie fallback).
 */
export function saveConsent(
    widgetId: string,
    widgetVersion: number,
    decisions: ConsentDecision[],
    expiryDays: number
): void {
    const expiresAt = new Date();
    expiresAt.setDate(expiresAt.getDate() + expiryDays);

    const stored: StoredConsent = {
        widget_id: widgetId,
        widget_version: widgetVersion,
        decisions,
        timestamp: new Date().toISOString(),
        expires_at: expiresAt.toISOString(),
    };

    const json = JSON.stringify(stored);
    const key = storageKey(widgetId);

    try {
        localStorage.setItem(key, json);
    } catch {
        // Fallback to cookie (for iframe / restricted contexts)
        setCookie(key, json, expiryDays);
    }
}

/**
 * Load stored consent. Returns null if not found or expired.
 */
export function loadConsent(widgetId: string): StoredConsent | null {
    const key = storageKey(widgetId);

    let json: string | null = null;
    try {
        json = localStorage.getItem(key);
    } catch {
        json = getCookie(key);
    }

    if (!json) return null;

    try {
        const stored: StoredConsent = JSON.parse(json);

        // Check expiry
        if (new Date(stored.expires_at) < new Date()) {
            clearConsent(widgetId);
            return null;
        }

        return stored;
    } catch {
        return null;
    }
}

/**
 * Clear stored consent for a widget.
 */
export function clearConsent(widgetId: string): void {
    const key = storageKey(widgetId);
    try {
        localStorage.removeItem(key);
    } catch {
        // noop
    }
    deleteCookie(key);
}

/**
 * Check if a specific purpose has been granted consent.
 */
export function isPurposeGranted(widgetId: string, purposeId: string): boolean {
    const stored = loadConsent(widgetId);
    if (!stored) return false;
    const decision = stored.decisions.find(d => d.purpose_id === purposeId);
    return decision?.granted ?? false;
}

// ---- Cookie helpers ----

function setCookie(name: string, value: string, days: number): void {
    const expires = new Date();
    expires.setDate(expires.getDate() + days);
    document.cookie = `${encodeURIComponent(name)}=${encodeURIComponent(value)};expires=${expires.toUTCString()};path=/;SameSite=Lax`;
}

function getCookie(name: string): string | null {
    const encoded = encodeURIComponent(name) + '=';
    const parts = document.cookie.split(';');
    for (const part of parts) {
        const trimmed = part.trim();
        if (trimmed.startsWith(encoded)) {
            return decodeURIComponent(trimmed.substring(encoded.length));
        }
    }
    return null;
}

function deleteCookie(name: string): void {
    document.cookie = `${encodeURIComponent(name)}=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/;SameSite=Lax`;
}
