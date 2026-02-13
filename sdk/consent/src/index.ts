// =============================================================================
// DataLens Consent SDK — Entry Point
// =============================================================================
//
// Usage:
//   <script src="https://{host}/api/public/consent/sdk/consent.min.js"
//           data-widget-id="wdg_abc123"
//           data-api-key="pk_live_xxx"></script>
//
// The SDK self-initializes on load. It:
//   1. Reads data-widget-id and data-api-key from its own <script> tag
//   2. Fetches widget config from the API
//   3. Checks for existing consent in localStorage
//   4. If no consent → shows banner/preference center
//   5. If consent exists → shows revisit button (if configured)
//   6. Optionally blocks scripts until consent is granted

import { WidgetConfig, WidgetSettings, ConsentDecision, PurposeConfig, SDKOptions } from './types';
import { fetchConfig, submitConsent } from './api';
import { saveConsent, loadConsent, clearConsent } from './storage';
import { initI18n, t } from './i18n';
import { generateStyles } from './ui/styles';
import { createBanner, createBackdrop } from './ui/banner';
import { createPreferenceCenter } from './ui/preference-center';
import { createRevisitButton } from './ui/revisit-button';
import { initScriptBlocker, releaseScripts, destroyScriptBlocker } from './script-blocker';

// ---- SDK State ----

let widgetConfig: WidgetConfig | null = null;
let sdkOptions: SDKOptions | null = null;
let shadowRoot: ShadowRoot | null = null;
let hostElement: HTMLDivElement | null = null;

// UI handles
let bannerUI: ReturnType<typeof createBanner> | null = null;
let backdropUI: ReturnType<typeof createBackdrop> | null = null;
let prefUI: ReturnType<typeof createPreferenceCenter> | null = null;
let revisitUI: ReturnType<typeof createRevisitButton> | null = null;

/**
 * Initialize the SDK. Called automatically from the self-executing block below.
 */
async function init(options: SDKOptions): Promise<void> {
    sdkOptions = options;

    try {
        // 1. Fetch widget configuration
        widgetConfig = await fetchConfig(options.apiUrl, options.apiKey);
        const config = widgetConfig.config;
        const purposes = config.purposes || [];

        // 2. Initialize i18n
        initI18n(config);

        // 3. Initialize script blocker (if enabled)
        if (config.block_until_consent) {
            initScriptBlocker(config.blocked_script_patterns || [], purposes);
        }

        // 4. Create Shadow DOM host
        createHost(config);

        // 5. Check for existing consent
        const stored = loadConsent(widgetConfig.id);

        if (stored && stored.widget_version === widgetConfig.version) {
            // Consent exists and is current version — apply it
            applyConsentState(stored.decisions, config, purposes);
            showRevisitButton(config);
        } else {
            // No consent or outdated version — show banner
            if (stored) clearConsent(widgetConfig.id); // Clear outdated
            showBanner(config, purposes);
        }
    } catch (err) {
        console.error('[DataLens] SDK initialization failed:', err);
    }
}

// ---- UI Orchestration ----

function createHost(config: WidgetSettings): void {
    hostElement = document.createElement('div');
    hostElement.id = 'datalens-consent';
    hostElement.style.cssText = 'position:fixed;z-index:2147483645;pointer-events:none;';
    document.body.appendChild(hostElement);

    shadowRoot = hostElement.attachShadow({ mode: 'open' });

    // Inject styles
    const style = document.createElement('style');
    style.textContent = generateStyles(config.theme, config.custom_css || undefined);
    shadowRoot.appendChild(style);
}

function showBanner(config: WidgetSettings, purposes: PurposeConfig[]): void {
    if (!shadowRoot) return;

    const isModal = config.layout === 'MODAL';

    // Create backdrop for modal
    if (isModal) {
        backdropUI = createBackdrop();
        shadowRoot.appendChild(backdropUI.element);
    }

    // Create banner
    bannerUI = createBanner(config, {
        onAcceptAll: () => handleAcceptAll(config, purposes),
        onRejectAll: () => handleRejectAll(config, purposes),
        onCustomize: () => openPreferenceCenter(config, purposes),
    });
    shadowRoot.appendChild(bannerUI.element);

    // Show with animation (next frame)
    requestAnimationFrame(() => {
        bannerUI?.show();
        if (isModal) backdropUI?.show();
    });
}

function openPreferenceCenter(config: WidgetSettings, purposes: PurposeConfig[]): void {
    if (!shadowRoot) return;

    // Hide banner
    bannerUI?.hide();

    // Show backdrop
    if (!backdropUI) {
        backdropUI = createBackdrop(() => closePreferenceCenter());
        shadowRoot.appendChild(backdropUI.element);
    }
    backdropUI.show();

    // Load current decisions from storage (if any)
    const stored = widgetConfig ? loadConsent(widgetConfig.id) : null;

    // Create preference center
    prefUI = createPreferenceCenter(config, purposes, stored?.decisions || null, {
        onSave: (decisions) => handleSavePreferences(decisions, config, purposes),
        onClose: () => closePreferenceCenter(),
    });
    shadowRoot.appendChild(prefUI.element);

    requestAnimationFrame(() => prefUI?.show());
}

function closePreferenceCenter(): void {
    prefUI?.hide();
    backdropUI?.hide();

    // Remove after animation
    setTimeout(() => {
        if (prefUI?.element.parentNode) {
            prefUI.element.parentNode.removeChild(prefUI.element);
        }
        prefUI = null;

        // If no consent saved yet, re-show banner
        const stored = widgetConfig ? loadConsent(widgetConfig.id) : null;
        if (!stored) {
            bannerUI?.show();
        }
    }, 350);
}

function showRevisitButton(config: WidgetSettings): void {
    if (!shadowRoot) return;

    revisitUI = createRevisitButton('bottom-left', () => {
        revisitUI?.hide();
        openPreferenceCenter(config, config.purposes || []);
    });
    shadowRoot.appendChild(revisitUI.element);

    requestAnimationFrame(() => revisitUI?.show());
}

// ---- Consent Handlers ----

function handleAcceptAll(config: WidgetSettings, purposes: PurposeConfig[]): void {
    const decisions: ConsentDecision[] = purposes.map(p => ({
        purpose_id: p.id,
        granted: true,
    }));
    finalizeConsent(decisions, config, purposes);
}

function handleRejectAll(config: WidgetSettings, purposes: PurposeConfig[]): void {
    const decisions: ConsentDecision[] = purposes.map(p => ({
        purpose_id: p.id,
        granted: p.is_essential, // Only essential purposes remain granted
    }));
    finalizeConsent(decisions, config, purposes);
}

function handleSavePreferences(
    decisions: ConsentDecision[],
    config: WidgetSettings,
    purposes: PurposeConfig[]
): void {
    closePreferenceCenter();
    finalizeConsent(decisions, config, purposes);
}

function finalizeConsent(
    decisions: ConsentDecision[],
    config: WidgetSettings,
    purposes: PurposeConfig[]
): void {
    if (!widgetConfig || !sdkOptions) return;

    // 1. Save to storage
    saveConsent(
        widgetConfig.id,
        widgetConfig.version,
        decisions,
        config.consent_expiry_days || 365
    );

    // 2. Submit to API (async, fire-and-forget)
    submitConsent(sdkOptions.apiUrl, sdkOptions.apiKey, {
        widget_id: widgetConfig.id,
        widget_version: widgetConfig.version,
        decisions,
        user_agent: navigator.userAgent,
        page_url: window.location.href,
    }).catch(err => console.error('[DataLens] Submit failed:', err));

    // 3. Apply script blocking decisions
    applyConsentState(decisions, config, purposes);

    // 4. Hide banner, show revisit
    bannerUI?.hide();
    backdropUI?.hide();
    showRevisitButton(config);

    // 5. Dispatch custom event for site integration
    window.dispatchEvent(new CustomEvent('datalens:consent', {
        detail: { decisions },
    }));
}

function applyConsentState(
    decisions: ConsentDecision[],
    config: WidgetSettings,
    purposes: PurposeConfig[]
): void {
    if (config.block_until_consent) {
        releaseScripts(decisions);
    }
}

// ---- Public API (window.DataLensConsent) ----

export function getConsent(): ConsentDecision[] | null {
    if (!widgetConfig) return null;
    const stored = loadConsent(widgetConfig.id);
    return stored?.decisions || null;
}

export function isPurposeGranted(purposeId: string): boolean {
    const decisions = getConsent();
    if (!decisions) return false;
    return decisions.find(d => d.purpose_id === purposeId)?.granted ?? false;
}

export function showPreferences(): void {
    if (!widgetConfig) return;
    revisitUI?.hide();
    openPreferenceCenter(widgetConfig.config, widgetConfig.config.purposes || []);
}

export function revokeConsent(): void {
    if (!widgetConfig) return;
    clearConsent(widgetConfig.id);
    destroyScriptBlocker();
    // Reload to re-block scripts
    window.location.reload();
}

// ---- Self-executing initialization ----

(function bootstrap() {
    // Find our own script tag
    const scripts = document.querySelectorAll('script[data-widget-id]');
    const scriptTag = scripts[scripts.length - 1]; // Last matching = current

    if (!scriptTag) {
        console.error('[DataLens] Missing data-widget-id attribute on script tag');
        return;
    }

    const widgetId = scriptTag.getAttribute('data-widget-id')!;
    const apiKey = scriptTag.getAttribute('data-api-key') || '';

    // Derive API URL from script src or use configured value
    let apiUrl = scriptTag.getAttribute('data-api-url') || '';
    if (!apiUrl) {
        const src = scriptTag.getAttribute('src') || '';
        if (src) {
            try {
                const url = new URL(src);
                apiUrl = url.origin;
            } catch {
                apiUrl = window.location.origin;
            }
        } else {
            apiUrl = window.location.origin;
        }
    }

    // Wait for DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => init({ widgetId, apiKey, apiUrl }));
    } else {
        init({ widgetId, apiKey, apiUrl });
    }
})();
