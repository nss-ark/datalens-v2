// =============================================================================
// Script Blocker — Opt-in script interception à la Cookie Gate
// =============================================================================
//
// When block_until_consent is enabled in widget config:
//   1. On init, find all <script data-consent-purpose="..."> tags, neutralize them
//   2. Install MutationObserver to catch dynamically added scripts
//   3. After consent, re-enable scripts for consented purposes
//   4. Essential purposes are never blocked
//
// Site integration example:
//   <script data-consent-purpose="analytics"
//           type="text/plain"
//           data-src="https://www.google-analytics.com/analytics.js"></script>

import { BlockedScriptPattern, ConsentDecision, PurposeConfig } from './types';

interface BlockedScript {
    original: HTMLScriptElement;
    purposeId: string;
    src?: string;
    inline?: string;
}

let observer: MutationObserver | null = null;
const blockedScripts: BlockedScript[] = [];
let essentialPurposeIds: Set<string> = new Set();
let patterns: BlockedScriptPattern[] = [];

/**
 * Initialize script blocking.
 * Call this BEFORE the page finishes loading for maximum effectiveness.
 */
export function initScriptBlocker(
    configPatterns: BlockedScriptPattern[],
    purposes: PurposeConfig[]
): void {
    patterns = configPatterns || [];
    essentialPurposeIds = new Set(
        purposes.filter(p => p.is_essential).map(p => p.id)
    );

    // Phase 1: Block existing scripts with data-consent-purpose attribute
    blockExistingScripts();

    // Phase 2: Watch for dynamically injected scripts
    startObserver();
}

/**
 * Release scripts for consented purposes.
 */
export function releaseScripts(decisions: ConsentDecision[]): void {
    const grantedPurposes = new Set(
        decisions.filter(d => d.granted).map(d => d.purpose_id)
    );

    // Also always release essential
    essentialPurposeIds.forEach(id => grantedPurposes.add(id));

    const toRelease = blockedScripts.filter(bs => grantedPurposes.has(bs.purposeId));

    toRelease.forEach(bs => {
        injectScript(bs);
        // Remove from blocked list
        const idx = blockedScripts.indexOf(bs);
        if (idx >= 0) blockedScripts.splice(idx, 1);
    });
}

/**
 * Block all non-essential script execution.
 * Call when user rejects consent.
 */
export function blockAllScripts(): void {
    // Already blocked on init — nothing extra to do.
    // Any future scripts will be caught by the MutationObserver.
}

/**
 * Stop the observer (cleanup).
 */
export function destroyScriptBlocker(): void {
    if (observer) {
        observer.disconnect();
        observer = null;
    }
    blockedScripts.length = 0;
}

// ---- Internal ----

function blockExistingScripts(): void {
    // Find scripts with data-consent-purpose attribute
    const scripts = document.querySelectorAll<HTMLScriptElement>(
        'script[data-consent-purpose]'
    );

    scripts.forEach(script => {
        const purposeId = script.getAttribute('data-consent-purpose');
        if (!purposeId) return;

        // Don't block essential purposes
        if (essentialPurposeIds.has(purposeId)) return;

        neutralizeScript(script, purposeId);
    });

    // Also check scripts matching URL patterns from config
    if (patterns.length > 0) {
        const allScripts = document.querySelectorAll<HTMLScriptElement>('script[src]');
        allScripts.forEach(script => {
            // Skip if already handled
            if (script.hasAttribute('data-consent-purpose')) return;
            if (script.hasAttribute('data-dl-blocked')) return;

            const src = script.getAttribute('src') || '';
            const match = patterns.find(p => src.includes(p.pattern));
            if (match && !essentialPurposeIds.has(match.purpose_id)) {
                neutralizeScript(script, match.purpose_id);
            }
        });
    }
}

function neutralizeScript(script: HTMLScriptElement, purposeId: string): void {
    // Store original info
    const blocked: BlockedScript = {
        original: script,
        purposeId,
        src: script.src || script.getAttribute('data-src') || undefined,
        inline: script.src ? undefined : script.textContent || undefined,
    };
    blockedScripts.push(blocked);

    // Neutralize: change type to prevent execution
    script.type = 'text/plain';
    script.setAttribute('data-dl-blocked', 'true');

    // If it has a src, remove it to prevent loading
    if (script.src) {
        script.setAttribute('data-src', script.src);
        script.removeAttribute('src');
    }
}

function injectScript(blocked: BlockedScript): void {
    // Create a fresh script element (required to trigger execution)
    const newScript = document.createElement('script');

    if (blocked.src) {
        newScript.src = blocked.src;
    } else if (blocked.inline) {
        newScript.textContent = blocked.inline;
    }

    newScript.type = 'text/javascript';
    newScript.setAttribute('data-consent-purpose', blocked.purposeId);
    newScript.setAttribute('data-dl-released', 'true');

    // Replace the neutralized script or append
    if (blocked.original.parentNode) {
        blocked.original.parentNode.replaceChild(newScript, blocked.original);
    } else {
        document.head.appendChild(newScript);
    }
}

function startObserver(): void {
    if (typeof MutationObserver === 'undefined') return;

    observer = new MutationObserver(mutations => {
        for (const mutation of mutations) {
            for (const node of Array.from(mutation.addedNodes)) {
                if (node instanceof HTMLScriptElement) {
                    handleNewScript(node);
                }
            }
        }
    });

    observer.observe(document.documentElement, {
        childList: true,
        subtree: true,
    });
}

function handleNewScript(script: HTMLScriptElement): void {
    // Skip our own released scripts
    if (script.hasAttribute('data-dl-released')) return;
    if (script.hasAttribute('data-dl-blocked')) return;

    // Check data-consent-purpose attribute
    const purposeId = script.getAttribute('data-consent-purpose');
    if (purposeId && !essentialPurposeIds.has(purposeId)) {
        neutralizeScript(script, purposeId);
        return;
    }

    // Check pattern matching
    const src = script.src || '';
    if (src && patterns.length > 0) {
        const match = patterns.find(p => src.includes(p.pattern));
        if (match && !essentialPurposeIds.has(match.purpose_id)) {
            neutralizeScript(script, match.purpose_id);
        }
    }
}
