import type { ID, TenantEntity } from './common';

export type NotificationChannel = 'EMAIL' | 'SMS' | 'WEBHOOK' | 'IN_APP';
export type NotificationStatus = 'PENDING' | 'SENT' | 'DELIVERED' | 'FAILED';

export interface ConsentNotification extends TenantEntity {
    recipient_type: string;
    recipient_id: string; // Email, Phone, or ID
    event_type: string;
    channel: NotificationChannel;
    template_id?: ID;
    payload: Record<string, unknown>;
    status: NotificationStatus;
    sent_at?: string;
    failure_reason?: string;
}

export interface NotificationTemplate extends TenantEntity {
    name: string;
    event_type: string;
    channel: NotificationChannel;
    subject: string;
    body_template: string; // HTML or Text with Go template syntax
    is_active: boolean;
}

export interface NotificationFilter {
    recipient_id?: string;
    event_type?: string;
    channel?: NotificationChannel;
    status?: NotificationStatus;
}
