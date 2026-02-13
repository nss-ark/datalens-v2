// =============================================================================
// Styles — Theme-driven CSS injected into Shadow DOM
// =============================================================================

import { ThemeConfig } from '../types';
import { isRTL } from '../i18n';

/**
 * Generate the full CSS stylesheet for the consent widget.
 * All selectors are scoped within a Shadow DOM — no leakage.
 */
export function generateStyles(theme: ThemeConfig, customCSS?: string): string {
    const dir = isRTL() ? 'rtl' : 'ltr';
    const radius = theme.border_radius || '8px';

    return `
:host {
  all: initial;
  font-family: ${theme.font_family || "'Inter', 'Segoe UI', system-ui, -apple-system, sans-serif"};
  color: ${theme.text_color || '#1a1a2e'};
  direction: ${dir};
  --dl-primary: ${theme.primary_color || '#6C5CE7'};
  --dl-primary-hover: ${adjustBrightness(theme.primary_color || '#6C5CE7', -15)};
  --dl-bg: ${theme.background_color || '#ffffff'};
  --dl-text: ${theme.text_color || '#1a1a2e'};
  --dl-text-secondary: ${adjustAlpha(theme.text_color || '#1a1a2e', 0.65)};
  --dl-border: ${adjustAlpha(theme.text_color || '#1a1a2e', 0.1)};
  --dl-radius: ${radius};
  --dl-shadow: 0 8px 32px rgba(0,0,0,0.12), 0 2px 8px rgba(0,0,0,0.08);
  --dl-transition: 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

/* ---- Backdrop ---- */
.dl-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(2px);
  z-index: 2147483646;
  opacity: 0;
  transition: opacity var(--dl-transition);
  pointer-events: none;
}
.dl-backdrop.dl-visible {
  opacity: 1;
  pointer-events: auto;
}

/* ---- Banner Container ---- */
.dl-banner {
  position: fixed;
  left: 0; right: 0;
  z-index: 2147483647;
  background: var(--dl-bg);
  box-shadow: var(--dl-shadow);
  padding: 20px 28px;
  display: flex;
  align-items: center;
  gap: 20px;
  transform: translateY(100%);
  transition: transform var(--dl-transition);
  border-top: 3px solid var(--dl-primary);
}
.dl-banner.dl-bottom { bottom: 0; }
.dl-banner.dl-top {
  top: 0; bottom: auto;
  transform: translateY(-100%);
  border-top: none;
  border-bottom: 3px solid var(--dl-primary);
}
.dl-banner.dl-visible {
  transform: translateY(0);
}

/* Banner as Modal */
.dl-banner.dl-modal {
  position: fixed;
  top: 50%; left: 50%;
  right: auto; bottom: auto;
  transform: translate(-50%, -50%) scale(0.9);
  opacity: 0;
  max-width: 520px;
  width: 90vw;
  border-radius: var(--dl-radius);
  border-top: 3px solid var(--dl-primary);
  flex-direction: column;
  padding: 28px;
}
.dl-banner.dl-modal.dl-visible {
  transform: translate(-50%, -50%) scale(1);
  opacity: 1;
}

/* ---- Banner Content ---- */
.dl-banner-body {
  flex: 1;
  min-width: 0;
}
.dl-banner-title {
  font-size: 16px;
  font-weight: 700;
  margin: 0 0 6px;
  color: var(--dl-text);
}
.dl-banner-desc {
  font-size: 13px;
  line-height: 1.5;
  margin: 0;
  color: var(--dl-text-secondary);
}

/* ---- Logo ---- */
.dl-logo {
  height: 28px;
  width: auto;
  flex-shrink: 0;
}

/* ---- Button Group ---- */
.dl-btn-group {
  display: flex;
  gap: 10px;
  flex-shrink: 0;
  flex-wrap: wrap;
}
.dl-modal .dl-btn-group {
  width: 100%;
  justify-content: flex-end;
  margin-top: 8px;
}

/* ---- Buttons ---- */
.dl-btn {
  padding: 10px 22px;
  border-radius: var(--dl-radius);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  border: none;
  transition: all 0.2s ease;
  white-space: nowrap;
  line-height: 1;
  letter-spacing: 0.01em;
}
.dl-btn:focus-visible {
  outline: 2px solid var(--dl-primary);
  outline-offset: 2px;
}
.dl-btn-primary {
  background: var(--dl-primary);
  color: #fff;
}
.dl-btn-primary:hover {
  background: var(--dl-primary-hover);
  box-shadow: 0 2px 8px rgba(0,0,0,0.15);
}
.dl-btn-secondary {
  background: transparent;
  color: var(--dl-text);
  border: 1.5px solid var(--dl-border);
}
.dl-btn-secondary:hover {
  border-color: var(--dl-primary);
  color: var(--dl-primary);
}
.dl-btn-text {
  background: transparent;
  color: var(--dl-primary);
  padding: 10px 14px;
}
.dl-btn-text:hover {
  text-decoration: underline;
}

/* ---- Preference Center (Modal / Sidebar) ---- */
.dl-pref {
  position: fixed;
  z-index: 2147483647;
  background: var(--dl-bg);
  box-shadow: var(--dl-shadow);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: all var(--dl-transition);
}

/* Modal layout */
.dl-pref.dl-pref-modal {
  top: 50%; left: 50%;
  transform: translate(-50%, -50%) scale(0.9);
  opacity: 0;
  pointer-events: none;
  max-width: 560px;
  width: 92vw;
  max-height: 85vh;
  border-radius: var(--dl-radius);
  border-top: 3px solid var(--dl-primary);
}
.dl-pref.dl-pref-modal.dl-visible {
  transform: translate(-50%, -50%) scale(1);
  opacity: 1;
  pointer-events: auto;
}

/* Sidebar layout */
.dl-pref.dl-pref-sidebar {
  top: 0; bottom: 0;
  right: ${dir === 'rtl' ? 'auto' : '0'};
  left: ${dir === 'rtl' ? '0' : 'auto'};
  width: 420px;
  max-width: 100vw;
  transform: translateX(${dir === 'rtl' ? '-100%' : '100%'});
  border-radius: 0;
}
.dl-pref.dl-pref-sidebar.dl-visible {
  transform: translateX(0);
}

/* Preference Header */
.dl-pref-header {
  padding: 24px 24px 16px;
  border-bottom: 1px solid var(--dl-border);
}
.dl-pref-title {
  font-size: 18px;
  font-weight: 700;
  margin: 0 0 6px;
  color: var(--dl-text);
}
.dl-pref-desc {
  font-size: 13px;
  color: var(--dl-text-secondary);
  margin: 0;
  line-height: 1.5;
}
.dl-pref-close {
  position: absolute;
  top: 16px;
  ${dir === 'rtl' ? 'left' : 'right'}: 16px;
  width: 32px; height: 32px;
  border: none;
  background: transparent;
  cursor: pointer;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--dl-text-secondary);
  font-size: 18px;
  transition: background 0.2s;
}
.dl-pref-close:hover {
  background: var(--dl-border);
}

/* Purpose List */
.dl-pref-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px 24px;
}
.dl-purpose {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 14px 0;
  border-bottom: 1px solid var(--dl-border);
  gap: 16px;
}
.dl-purpose:last-child { border-bottom: none; }
.dl-purpose-info { flex: 1; min-width: 0; }
.dl-purpose-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--dl-text);
  margin: 0 0 3px;
}
.dl-purpose-desc {
  font-size: 12px;
  color: var(--dl-text-secondary);
  margin: 0;
  line-height: 1.4;
}
.dl-purpose-essential {
  font-size: 11px;
  color: var(--dl-primary);
  font-weight: 600;
  margin-top: 2px;
}

/* Toggle Switch */
.dl-toggle {
  position: relative;
  width: 44px; height: 24px;
  flex-shrink: 0;
  margin-top: 2px;
}
.dl-toggle input {
  opacity: 0;
  width: 0; height: 0;
  position: absolute;
}
.dl-toggle-track {
  position: absolute;
  inset: 0;
  background: var(--dl-border);
  border-radius: 12px;
  cursor: pointer;
  transition: background 0.25s;
}
.dl-toggle-track::after {
  content: '';
  position: absolute;
  top: 2px;
  ${dir === 'rtl' ? 'right' : 'left'}: 2px;
  width: 20px; height: 20px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 1px 3px rgba(0,0,0,0.2);
  transition: transform 0.25s;
}
.dl-toggle input:checked + .dl-toggle-track {
  background: var(--dl-primary);
}
.dl-toggle input:checked + .dl-toggle-track::after {
  transform: translateX(${dir === 'rtl' ? '-20px' : '20px'});
}
.dl-toggle input:disabled + .dl-toggle-track {
  opacity: 0.5;
  cursor: default;
}

/* Preference Footer */
.dl-pref-footer {
  padding: 16px 24px;
  border-top: 1px solid var(--dl-border);
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}

/* ---- Revisit Button (FAB) ---- */
.dl-revisit {
  position: fixed;
  z-index: 2147483645;
  width: 44px; height: 44px;
  border-radius: 50%;
  background: var(--dl-primary);
  color: #fff;
  border: none;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(0,0,0,0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.25s;
  opacity: 0;
  transform: scale(0.7);
  pointer-events: none;
}
.dl-revisit.dl-visible {
  opacity: 1;
  transform: scale(1);
  pointer-events: auto;
}
.dl-revisit:hover {
  transform: scale(1.08);
  box-shadow: 0 6px 20px rgba(0,0,0,0.25);
}
.dl-revisit.dl-bottom-left { bottom: 20px; ${dir === 'rtl' ? 'right' : 'left'}: 20px; }
.dl-revisit.dl-bottom-right { bottom: 20px; ${dir === 'rtl' ? 'left' : 'right'}: 20px; }

/* ---- Responsive ---- */
@media (max-width: 640px) {
  .dl-banner {
    flex-direction: column;
    padding: 16px 20px;
    gap: 14px;
  }
  .dl-btn-group {
    width: 100%;
    justify-content: stretch;
  }
  .dl-btn-group .dl-btn {
    flex: 1;
    text-align: center;
  }
  .dl-pref.dl-pref-sidebar {
    width: 100vw;
  }
  .dl-pref.dl-pref-modal {
    width: 95vw;
    max-height: 90vh;
  }
}

/* ---- Animations ---- */
@keyframes dl-fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

${customCSS || ''}
`.trim();
}

// ---- Color utilities (no dependencies) ----

function adjustBrightness(hex: string, percent: number): string {
    const num = parseInt(hex.replace('#', ''), 16);
    const r = Math.min(255, Math.max(0, ((num >> 16) & 0xff) + Math.round(2.55 * percent)));
    const g = Math.min(255, Math.max(0, ((num >> 8) & 0xff) + Math.round(2.55 * percent)));
    const b = Math.min(255, Math.max(0, (num & 0xff) + Math.round(2.55 * percent)));
    return `#${((r << 16) | (g << 8) | b).toString(16).padStart(6, '0')}`;
}

function adjustAlpha(hex: string, alpha: number): string {
    const num = parseInt(hex.replace('#', ''), 16);
    const r = (num >> 16) & 0xff;
    const g = (num >> 8) & 0xff;
    const b = num & 0xff;
    return `rgba(${r},${g},${b},${alpha})`;
}
