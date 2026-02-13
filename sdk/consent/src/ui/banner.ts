// =============================================================================
// Banner Widget — BOTTOM_BAR, TOP_BAR, MODAL layouts
// =============================================================================

import { WidgetSettings, ConsentDecision, PurposeConfig } from '../types';
import { t } from '../i18n';

export interface BannerCallbacks {
    onAcceptAll: () => void;
    onRejectAll: () => void;
    onCustomize: () => void;
}

/**
 * Create the banner DOM element.
 * Returns an object with the element and show/hide controls.
 */
export function createBanner(
    config: WidgetSettings,
    callbacks: BannerCallbacks
) {
    const layout = config.layout || 'BOTTOM_BAR';

    // Build container
    const banner = document.createElement('div');
    banner.className = 'dl-banner';

    if (layout === 'MODAL') {
        banner.classList.add('dl-modal');
    } else if (layout === 'TOP_BAR') {
        banner.classList.add('dl-top');
    } else {
        banner.classList.add('dl-bottom');
    }

    // Logo
    let logoHTML = '';
    if (config.theme.logo_url) {
        logoHTML = `<img class="dl-logo" src="${escapeAttr(config.theme.logo_url)}" alt="Logo" />`;
    }

    // Content
    banner.innerHTML = `
    ${logoHTML}
    <div class="dl-banner-body">
      <p class="dl-banner-title">${escapeHTML(t('title'))}</p>
      <p class="dl-banner-desc">${escapeHTML(t('description'))}</p>
    </div>
    <div class="dl-btn-group">
      <button class="dl-btn dl-btn-secondary dl-btn-reject">${escapeHTML(t('reject_all'))}</button>
      <button class="dl-btn dl-btn-text dl-btn-customize">${escapeHTML(t('customize'))}</button>
      <button class="dl-btn dl-btn-primary dl-btn-accept">${escapeHTML(t('accept_all'))}</button>
    </div>
  `;

    // Bind events
    banner.querySelector('.dl-btn-accept')!.addEventListener('click', callbacks.onAcceptAll);
    banner.querySelector('.dl-btn-reject')!.addEventListener('click', callbacks.onRejectAll);
    banner.querySelector('.dl-btn-customize')!.addEventListener('click', callbacks.onCustomize);

    return {
        element: banner,
        show: () => {
            // Force reflow before adding class (for animation)
            banner.offsetHeight;
            banner.classList.add('dl-visible');
        },
        hide: () => {
            banner.classList.remove('dl-visible');
        },
    };
}

/**
 * Create the backdrop overlay (used for MODAL layout).
 */
export function createBackdrop(onClick?: () => void) {
    const backdrop = document.createElement('div');
    backdrop.className = 'dl-backdrop';
    if (onClick) {
        backdrop.addEventListener('click', onClick);
    }
    return {
        element: backdrop,
        show: () => backdrop.classList.add('dl-visible'),
        hide: () => backdrop.classList.remove('dl-visible'),
    };
}

// ---- Helpers ----

function escapeHTML(str: string): string {
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}

function escapeAttr(str: string): string {
    return str.replace(/"/g, '&quot;').replace(/'/g, '&#39;');
}
