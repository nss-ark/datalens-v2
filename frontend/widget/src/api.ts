import { WidgetConfig, ConsentSession } from './types.ts';

export class ApiClient {
    private baseUrl: string;
    private widgetId: string;

    constructor(baseUrl: string, widgetId: string) {
        this.baseUrl = baseUrl.replace(/\/$/, ''); // Remove trailing slash
        this.widgetId = widgetId;
    }

    get id() {
        return this.widgetId;
    }

    async fetchConfig(): Promise<WidgetConfig> {
        const response = await fetch(`${this.baseUrl}/api/public/consent/widget/config`, {
            headers: {
                'X-Widget-Key': this.widgetId
            }
        });

        if (!response.ok) {
            throw new Error(`Failed to fetch config: ${response.statusText}`);
        }

        const json = await response.json();
        if (!json.success || !json.data) {
            throw new Error('Invalid config response');
        }
        return json.data;
    }

    async recordConsent(session: ConsentSession): Promise<void> {
        const response = await fetch(`${this.baseUrl}/api/public/consent/sessions`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Widget-Key': this.widgetId
            },
            body: JSON.stringify(session)
        });

        if (!response.ok) {
            console.error('Failed to record consent:', response.statusText);
            // We don't throw here to avoid disrupting the user experience, just log it.
        }
    }
}
