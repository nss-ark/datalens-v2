import { Monitor, CheckCircle, Shield, Settings, AlertTriangle } from 'lucide-react';
import type { ActivityFeedItem } from '@/types/portal';

/* ── Icon + color mapping per activity type ── */
const typeConfig: Record<string, { icon: typeof Shield; bg: string; color: string }> = {
    login: { icon: Monitor, bg: '#DBEAFE', color: '#2563EB' },
    request_update: { icon: CheckCircle, bg: '#D1FAE5', color: '#059669' },
    consent_update: { icon: Shield, bg: '#FEE2E2', color: '#DC2626' },
    security_digest: { icon: Settings, bg: '#F3F4F6', color: '#6B7280' },
    breach: { icon: AlertTriangle, bg: '#FEF3C7', color: '#D97706' },
};

/* ── Relative-time helper ── */
function relativeTime(iso: string): string {
    const diff = Date.now() - new Date(iso).getTime();
    const mins = Math.floor(diff / 60_000);
    if (mins < 1) return 'Just now';
    if (mins < 60) return `${mins} minute${mins > 1 ? 's' : ''} ago`;
    const hrs = Math.floor(mins / 60);
    if (hrs < 24) return `${hrs} hour${hrs > 1 ? 's' : ''} ago`;
    const days = Math.floor(hrs / 24);
    if (days === 1) return 'Yesterday';
    return `${days} days ago`;
}

interface Props {
    item: ActivityFeedItem;
}

export const ActivityFeedCard = ({ item }: Props) => {
    const cfg = typeConfig[item.type] || typeConfig.login;
    const Icon = cfg.icon;

    return (
        <div
            className="portal-card"
            style={{
                padding: '20px 24px',
                display: 'flex',
                gap: '16px',
                alignItems: 'flex-start',
                transition: 'box-shadow 0.2s, transform 0.15s',
            }}
        >
            {/* Icon circle */}
            <div
                style={{
                    width: 44,
                    height: 44,
                    minWidth: 44,
                    borderRadius: '12px',
                    backgroundColor: cfg.bg,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                }}
            >
                <Icon size={20} style={{ color: cfg.color }} />
            </div>

            {/* Content */}
            <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: '12px', marginBottom: '4px' }}>
                    <h3 style={{ fontSize: '15px', fontWeight: 600, color: '#111827', margin: 0, lineHeight: 1.4 }}>
                        {item.title}
                    </h3>
                    <span style={{ fontSize: '12px', color: '#9CA3AF', whiteSpace: 'nowrap', flexShrink: 0, lineHeight: 1.75 }}>
                        {relativeTime(item.timestamp)}
                    </span>
                </div>

                <p style={{ fontSize: '13.5px', color: '#6B7280', margin: 0, lineHeight: 1.6 }}>
                    {item.description}
                </p>

                {/* Category label */}
                {item.category_label && (
                    <span
                        style={{
                            display: 'inline-block',
                            marginTop: '8px',
                            fontSize: '11px',
                            fontWeight: 600,
                            textTransform: 'uppercase',
                            letterSpacing: '0.04em',
                            color: '#9CA3AF',
                        }}
                    >
                        CATEGORY: {item.category_label}
                    </span>
                )}

                {/* Action buttons */}
                {(item.secondary_actions || item.primary_action) && (
                    <div style={{ display: 'flex', gap: '12px', marginTop: '10px', flexWrap: 'wrap' }}>
                        {item.secondary_actions?.map((a, i) => (
                            <button
                                key={i}
                                style={{
                                    fontSize: '12.5px',
                                    fontWeight: 600,
                                    color: a.variant === 'danger' ? '#DC2626' : '#2563EB',
                                    background: 'none',
                                    border: 'none',
                                    padding: 0,
                                    cursor: 'pointer',
                                    textDecoration: 'underline',
                                    textUnderlineOffset: '2px',
                                }}
                            >
                                {a.label}
                            </button>
                        ))}
                        {item.primary_action && (
                            <button
                                style={{
                                    fontSize: '12.5px',
                                    fontWeight: 600,
                                    color: '#ffffff',
                                    backgroundColor: '#2563EB',
                                    border: 'none',
                                    padding: '5px 14px',
                                    borderRadius: '6px',
                                    cursor: 'pointer',
                                    transition: 'background-color 0.2s',
                                }}
                            >
                                {item.primary_action.label}
                            </button>
                        )}
                    </div>
                )}
            </div>
        </div>
    );
};
