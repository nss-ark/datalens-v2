import { ApiClient } from './api';
import { StateManager } from './state';
import { WidgetRenderer } from './ui/renderer';
import { WidgetOptions, WidgetConfig, ConsentDecision, PurposeRef, ConsentSession } from './types';

class DataLensConsentImpl {
    private api: ApiClient | null = null;
    private state: StateManager | null = null;
    private renderer: WidgetRenderer | null = null;
    private config: WidgetConfig | null = null;
    private initialized = false;

    async init(options: WidgetOptions) {
        if (this.initialized) return;

        console.log('Initializing DataLens Consent Widget', options);

        const apiBase = options.apiBase || 'https://api.datalens.io';
        this.api = new ApiClient(apiBase, options.widgetId);
        this.state = new StateManager();

        try {
            this.config = await this.api.fetchConfig();
            // Shim purposes for demo if not returned by API yet (depends on backend readiness)
            if (!this.config.purposes) {
                this.config.purposes = this.shimPurposes(this.config.purpose_ids);
            }

            this.renderer = new WidgetRenderer(
                this.config,
                this.handleConsent.bind(this),
                this.handleCustomize.bind(this)
            );

            // Check if we need to show
            const shouldShow = this.shouldShowWidget();
            if (shouldShow) {
                this.renderer.renderBanner();
            }

            this.initialized = true;

            // Script blocking hook
            this.applyScriptBlocking();

        } catch (e) {
            console.error('DataLens Widget Init Failed:', e);
        }
    }

    show() {
        if (this.renderer) {
            this.renderer.renderBanner(); // Defaults to banner
        }
    }

    hide() {
        if (this.renderer) {
            this.renderer.close();
        }
    }

    reset() {
        // Clear cookie (helper needed in state) but for now just show
        this.renderer?.renderBanner();
    }

    // --- Internals ---

    private shouldShowWidget(): boolean {
        // If no consent cookie exists, show
        const current = this.state?.getAllConsent();
        if (!current || Object.keys(current).length === 0) return true;

        // If new purposes added? (Enhancement)
        return false;
    }

    private handleCustomize() {
        const current = this.state?.getAllConsent() || {};
        this.renderer?.renderPreferenceCenter(current);
    }

    private handleConsent(decisions: Record<string, boolean>) {
        if (!this.config || !this.state || !this.api) return;

        // 1. Convert to array
        const decisionArray: ConsentDecision[] = Object.keys(decisions).map(k => ({
            purpose_id: k,
            granted: decisions[k]
        }));

        // 2. Save local state
        this.state.setConsent(decisionArray, this.config.consent_expiry_days);

        // 3. UI Feedback
        this.renderer?.close();

        // 4. API Record
        const session: ConsentSession = {
            widget_id: this.api.id,
            decisions: decisionArray,
            url: window.location.href,
            user_agent: navigator.userAgent
        };

        this.api.recordConsent(session);

        // 5. Unblock scripts
        this.applyScriptBlocking();
    }

    private shimPurposes(ids: string[]): PurposeRef[] {
        // Fallback if API doesn't return full objects yet
        return ids.map(id => ({
            id,
            name: id.charAt(0).toUpperCase() + id.slice(1).replace(/_/g, ' '),
            description: `Allow processing for ${id}`,
            is_essential: id === 'essential'
        }));
    }

    private applyScriptBlocking() {
        // TODO: Implement observer to unblock scripts based on consent
        // For standard "Block Until Consent" logic
    }
}

// Expose global
const instance = new DataLensConsentImpl();
// @ts-ignore
window.DataLensConsent = instance;

export default instance;
