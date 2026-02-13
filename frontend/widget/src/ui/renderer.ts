import { WidgetConfig } from '../types.ts';
import { injectStyles } from './styles.ts';
import * as layout from './layouts.ts';

export class WidgetRenderer {
    private container: HTMLElement | null = null;
    private backdrop: HTMLElement | null = null;
    private config: WidgetConfig;
    private onConsent: (decisions: Record<string, boolean>) => void;
    private onCustomize: () => void; // Switch to preference center

    constructor(
        config: WidgetConfig,
        onConsent: (decisions: Record<string, boolean>) => void,
        onCustomize: () => void
    ) {
        this.config = config;
        this.onConsent = onConsent;
        this.onCustomize = onCustomize;
        injectStyles(config.theme, config.custom_css);
    }

    renderBanner() {
        this.clear();

        // 1. Container
        this.container = layout.renderMainContainer(this.config.layout);

        // 2. Content
        const lang = this.config.default_language;
        const texts = this.config.translations[lang] || {};

        this.container.appendChild(layout.renderHeader(texts.title || 'Privacy Preference'));
        this.container.appendChild(layout.renderDescription(texts.description || 'We use cookies.'));

        // 3. Actions
        const actions = layout.renderActions();

        // TODO: Map texts properly
        actions.appendChild(layout.renderButton(texts.customize || 'Customize', 'secondary', () => this.onCustomize()));
        actions.appendChild(layout.renderButton(texts.reject_all || 'Reject All', 'secondary', () => this.handleRejectAll()));
        actions.appendChild(layout.renderButton(texts.accept_all || 'Accept All', 'primary', () => this.handleAcceptAll()));

        this.container.appendChild(actions);

        // 4. Mount
        document.body.appendChild(this.container);

        // 5. Backdrop if modal
        if (this.config.layout === 'MODAL') {
            this.backdrop = layout.renderBackdrop();
            document.body.appendChild(this.backdrop);
        }
    }

    renderPreferenceCenter(currentConsent: Record<string, boolean>) {
        this.clear();

        // 1. Container
        this.container = layout.renderMainContainer('MODAL'); // Pref center is always detailed, effectively a modal for now or reuse layout

        // 2. Content
        // 2. Content
        this.container.appendChild(layout.renderHeader('Privacy Preferences'));
        this.container.appendChild(layout.renderDescription('Manage your consent preferences below.'));

        // 3. Toggles
        const decisions: Record<string, boolean> = { ...currentConsent };
        const list = document.createElement('div');
        list.style.maxHeight = '300px';
        list.style.overflowY = 'auto';
        list.style.margin = '1rem 0';

        // Ensure purposes exist in config (populated by API)
        const purposes = this.config.purposes || [];

        purposes.forEach(p => {
            const isGranted = decisions[p.id] ?? false;
            // Essential purposes cannot be disabled
            const disabled = p.is_essential;
            const checked = disabled ? true : isGranted;

            const toggle = layout.renderToggle(p.id, p.name, checked, disabled, (val) => {
                decisions[p.id] = val;
            });
            list.appendChild(toggle);
        });
        this.container.appendChild(list);

        // 4. Actions
        const actions = layout.renderActions();
        actions.appendChild(layout.renderButton('Cancel', 'text', () => {
            // Re-render banner? Or just close?
            // For now, re-render banner if we came from there contextually
            this.renderBanner();
        }));
        actions.appendChild(layout.renderButton('Save Preferences', 'primary', () => this.onConsent(decisions)));
        this.container.appendChild(actions);

        // 5. Mount
        document.body.appendChild(this.container);
        this.backdrop = layout.renderBackdrop();
        document.body.appendChild(this.backdrop);
    }

    close() {
        this.clear();
    }

    private clear() {
        if (this.container) {
            this.container.remove();
            this.container = null;
        }
        if (this.backdrop) {
            this.backdrop.remove();
            this.backdrop = null;
        }
    }

    private handleAcceptAll() {
        const decisions: Record<string, boolean> = {};
        (this.config.purposes || []).forEach(p => decisions[p.id] = true);
        this.onConsent(decisions);
    }

    private handleRejectAll() {
        const decisions: Record<string, boolean> = {};
        (this.config.purposes || []).forEach(p => decisions[p.id] = p.is_essential);
        this.onConsent(decisions);
    }
}
