// =============================================================================
// API Client — Lightweight fetch wrapper for DataLens public API
// =============================================================================

import { WidgetConfig, ConsentSessionPayload } from './types';

/**
 * Fetch widget configuration from the DataLens API.
 * Uses the API key for authentication via X-API-Key header.
 */
export async function fetchConfig(apiUrl: string, apiKey: string): Promise<WidgetConfig> {
    const res = await fetch(`${apiUrl}/api/public/consent/widget/config`, {
        method: 'GET',
        headers: {
            'X-API-Key': apiKey,
            'Accept': 'application/json',
        },
    });

    if (!res.ok) {
        throw new Error(`[DataLens] Failed to load widget config: ${res.status}`);
    }

    return res.json();
}

/**
 * Submit consent decisions to the DataLens API.
 * Creates an immutable consent session record.
 */
export async function submitConsent(
    apiUrl: string,
    apiKey: string,
    payload: ConsentSessionPayload
): Promise<void> {
    const res = await fetch(`${apiUrl}/api/public/consent/sessions`, {
        method: 'POST',
        headers: {
            'X-API-Key': apiKey,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
    });

    if (!res.ok) {
        console.error('[DataLens] Failed to submit consent:', res.status);
    }
}
