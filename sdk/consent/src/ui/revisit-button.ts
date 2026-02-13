// =============================================================================
// Revisit Button — Floating FAB to re-open preference center
// =============================================================================

/**
 * Create the revisit floating action button.
 * Appears after consent has been given, allowing users to change preferences.
 */
export function createRevisitButton(
    position: 'bottom-left' | 'bottom-right',
    onClick: () => void
) {
    const btn = document.createElement('button');
    btn.className = `dl-revisit dl-${position}`;
    btn.setAttribute('aria-label', 'Manage consent preferences');
    btn.innerHTML = SVG_SHIELD;
    btn.addEventListener('click', onClick);

    return {
        element: btn,
        show: () => {
            btn.offsetHeight;
            btn.classList.add('dl-visible');
        },
        hide: () => {
            btn.classList.remove('dl-visible');
        },
    };
}

/** Shield/cookie icon SVG */
const SVG_SHIELD = `<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/><circle cx="12" cy="12" r="1.5" fill="currentColor"/><circle cx="9" cy="9" r="1" fill="currentColor"/><circle cx="15" cy="10" r="1" fill="currentColor"/></svg>`;
