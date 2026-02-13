// =============================================================================
// Preference Center — MODAL + SIDEBAR layouts with per-purpose toggles
// =============================================================================

import { WidgetSettings, ConsentDecision, PurposeConfig } from '../types';
import { t } from '../i18n';

export interface PrefCenterCallbacks {
    onSave: (decisions: ConsentDecision[]) => void;
    onClose: () => void;
}

/**
 * Create the preference center DOM element with per-purpose toggles.
 */
export function createPreferenceCenter(
    config: WidgetSettings,
    purposes: PurposeConfig[],
    currentDecisions: ConsentDecision[] | null,
    callbacks: PrefCenterCallbacks
) {
    const layout = config.layout === 'SIDEBAR' ? 'SIDEBAR' : 'MODAL';
    const container = document.createElement('div');
    container.className = `dl-pref dl-pref-${layout.toLowerCase()}`;

    // Build purpose toggles HTML
    const purposesHTML = purposes
        .map((purpose, i) => {
            const isEssential = purpose.is_essential;
            const isChecked = isEssential
                ? true
                : currentDecisions
                    ? currentDecisions.find(d => d.purpose_id === purpose.id)?.granted ?? false
                    : config.default_state === 'OPT_IN';

            return `
      <div class="dl-purpose">
        <div class="dl-purpose-info">
          <p class="dl-purpose-name">${escapeHTML(purpose.name)}</p>
          <p class="dl-purpose-desc">${escapeHTML(purpose.description)}</p>
          ${isEssential ? `<p class="dl-purpose-essential">${escapeHTML(t('essential_label'))}</p>` : ''}
        </div>
        <label class="dl-toggle">
          <input type="checkbox"
                 data-purpose-id="${escapeAttr(purpose.id)}"
                 ${isChecked ? 'checked' : ''}
                 ${isEssential ? 'disabled' : ''} />
          <span class="dl-toggle-track"></span>
        </label>
      </div>
    `;
        })
        .join('');

    container.innerHTML = `
    <div class="dl-pref-header" style="position:relative;">
      <p class="dl-pref-title">${escapeHTML(t('preferences_title'))}</p>
      <p class="dl-pref-desc">${escapeHTML(t('preferences_description'))}</p>
      <button class="dl-pref-close" aria-label="${escapeAttr(t('close'))}">&times;</button>
    </div>
    <div class="dl-pref-body">
      ${purposesHTML}
    </div>
    <div class="dl-pref-footer">
      <button class="dl-btn dl-btn-secondary dl-btn-reject-all">${escapeHTML(t('reject_all'))}</button>
      <button class="dl-btn dl-btn-primary dl-btn-accept-all">${escapeHTML(t('accept_all'))}</button>
      <button class="dl-btn dl-btn-primary dl-btn-save">${escapeHTML(t('save_preferences'))}</button>
    </div>
  `;

    // ---- Event handlers ----

    // Close
    container.querySelector('.dl-pref-close')!.addEventListener('click', callbacks.onClose);

    // Accept All (set all non-essential to checked)
    container.querySelector('.dl-btn-accept-all')!.addEventListener('click', () => {
        const checkboxes = container.querySelectorAll<HTMLInputElement>('input[data-purpose-id]');
        checkboxes.forEach(cb => {
            if (!cb.disabled) cb.checked = true;
        });
    });

    // Reject All (set all non-essential to unchecked)
    container.querySelector('.dl-btn-reject-all')!.addEventListener('click', () => {
        const checkboxes = container.querySelectorAll<HTMLInputElement>('input[data-purpose-id]');
        checkboxes.forEach(cb => {
            if (!cb.disabled) cb.checked = false;
        });
    });

    // Save
    container.querySelector('.dl-btn-save')!.addEventListener('click', () => {
        const checkboxes = container.querySelectorAll<HTMLInputElement>('input[data-purpose-id]');
        const decisions: ConsentDecision[] = [];
        checkboxes.forEach(cb => {
            decisions.push({
                purpose_id: cb.dataset.purposeId!,
                granted: cb.checked,
            });
        });
        callbacks.onSave(decisions);
    });

    return {
        element: container,
        show: () => {
            container.offsetHeight; // reflow
            container.classList.add('dl-visible');
        },
        hide: () => {
            container.classList.remove('dl-visible');
        },
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
