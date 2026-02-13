import { ConsentDecision } from './types.ts';

const COOKIE_NAME = '_dl_consent';
const DEFAULT_EXPIRY_DAYS = 365;

export class StateManager {
    private listeners: Record<string, Function[]> = {};
    private consentState: Record<string, boolean> = {}; // purposeId -> granted

    constructor() {
        this.loadFromCookie();
    }

    // --- State Management ---

    getConsent(purposeId: string): boolean {
        return this.consentState[purposeId] === true;
    }

    getAllConsent(): Record<string, boolean> {
        return { ...this.consentState };
    }

    setConsent(decisions: ConsentDecision[], expiryDays: number = DEFAULT_EXPIRY_DAYS) {
        decisions.forEach(d => {
            this.consentState[d.purpose_id] = d.granted;
        });

        this.saveToCookie(expiryDays);
        this.emit('consent', this.consentState);
    }

    // --- Cookie Persistence ---

    private loadFromCookie() {
        const cookie = document.cookie.split('; ').find(row => row.startsWith(COOKIE_NAME + '='));
        if (cookie) {
            try {
                const val = decodeURIComponent(cookie.split('=')[1]);
                this.consentState = JSON.parse(val);
                // Validate structure (simple check)
                if (typeof this.consentState !== 'object') {
                    this.consentState = {};
                }
            } catch (e) {
                console.warn('Failed to parse consent cookie', e);
                this.consentState = {};
            }
        }
    }

    private saveToCookie(expiryDays: number) {
        const d = new Date();
        d.setTime(d.getTime() + (expiryDays * 24 * 60 * 60 * 1000));
        const expires = "expires=" + d.toUTCString();
        const value = JSON.stringify(this.consentState);
        document.cookie = `${COOKIE_NAME}=${encodeURIComponent(value)};${expires};path=/;SameSite=Lax`;
    }

    // --- Event Emitter ---

    on(event: string, callback: Function) {
        if (!this.listeners[event]) {
            this.listeners[event] = [];
        }
        this.listeners[event].push(callback);
    }

    private emit(event: string, data: any) {
        if (this.listeners[event]) {
            this.listeners[event].forEach(cb => cb(data));
        }
    }
}
