import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { portalService } from '@/services/portalService';
import { ChevronDown } from 'lucide-react';
import type { ActivityFeedItem, ActivityCategory, BreachNotification } from '@/types/portal';
import { ActivityFeedCard } from '@/components/notifications/ActivityFeedCard';
import { SecurityShield } from '@/components/notifications/SecurityShield';
import { BreachAssistance } from '@/components/notifications/BreachAssistance';

/* ────────────────────────────────────────────
 * Mock activity items (supplement real breach data)
 * ──────────────────────────────────────────── */
const now = Date.now();
const mins = (m: number) => new Date(now - m * 60_000).toISOString();
const hrs = (h: number) => new Date(now - h * 3_600_000).toISOString();
const days = (d: number) => new Date(now - d * 86_400_000).toISOString();

const MOCK_ITEMS: ActivityFeedItem[] = [
    {
        id: 'mock-login-1',
        type: 'login',
        title: 'New login detected',
        description: 'A new login was successful from Chrome on a MacOS device located in San Francisco, CA.',
        timestamp: mins(2),
        category: 'ALL',
        is_read: false,
        secondary_actions: [
            { label: 'This was me', variant: 'default' },
            { label: 'Not me – Secure account', variant: 'danger' },
        ],
    },
    {
        id: 'mock-dsr-update-1',
        type: 'request_update',
        title: 'Request #DS-2940 Status Update',
        description: 'Your data access request has been completed. You can now download your archive.',
        timestamp: hrs(3),
        category: 'REQUESTS',
        is_read: false,
        category_label: 'DATA PRIVACY',
    },
    {
        id: 'mock-consent-1',
        type: 'consent_update',
        title: 'Updated Privacy Consent',
        description: 'We\'ve updated our data processing terms for marketing services. Please review and provide consent.',
        timestamp: days(1),
        category: 'PRIVACY',
        is_read: true,
        primary_action: { label: 'Review Changes' },
    },
    {
        id: 'mock-digest-1',
        type: 'security_digest',
        title: 'Monthly Security Digest Available',
        description: 'Your summary of account security activities for the month of January is ready for viewing.',
        timestamp: days(2),
        category: 'ALL',
        is_read: true,
    },
];

/* ── Map real breach notifications into ActivityFeedItems ── */
function breachToFeedItem(b: BreachNotification): ActivityFeedItem {
    return {
        id: `breach-${b.id}`,
        type: 'breach',
        title: b.title,
        description: b.description,
        timestamp: b.created_at,
        category: 'PRIVACY',
        is_read: b.is_read,
        category_label: b.severity,
        breach_ref: b,
    };
}

/* ── Filter tabs ── */
const TABS: { label: string; value: ActivityCategory }[] = [
    { label: 'ALL', value: 'ALL' },
    { label: 'REQUESTS', value: 'REQUESTS' },
    { label: 'PRIVACY', value: 'PRIVACY' },
];

/* ════════════════════════════════════════════
 * Page Component
 * ════════════════════════════════════════════ */
const BreachNotifications = () => {
    const [activeTab, setActiveTab] = useState<ActivityCategory>('ALL');

    // Fetch real breach notifications
    const { data: breachData, isLoading } = useQuery({
        queryKey: ['breach-notifications'],
        queryFn: () => portalService.getBreachNotifications(),
    });

    // Merge breach data + mock items into a single feed, sorted by timestamp desc
    const feedItems = useMemo(() => {
        const breachItems = (breachData?.items || []).map(breachToFeedItem);
        const all = [...MOCK_ITEMS, ...breachItems].sort(
            (a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime(),
        );
        return all;
    }, [breachData]);

    // Apply tab filter
    const filteredItems = useMemo(() => {
        if (activeTab === 'ALL') return feedItems;
        return feedItems.filter((item) => item.category === activeTab);
    }, [feedItems, activeTab]);

    return (
        <div className="animate-fade-in">
            {/* ── 2-column grid ── */}
            <div className="notification-feed-grid">
                {/* ════ LEFT COLUMN — Activity Feed ════ */}
                <div>
                    {/* Header + tabs */}
                    <div
                        style={{
                            display: 'flex',
                            justifyContent: 'space-between',
                            alignItems: 'flex-start',
                            flexWrap: 'wrap',
                            gap: '16px',
                            marginBottom: '28px',
                        }}
                    >
                        <div>
                            <h1 style={{ fontSize: '28px', fontWeight: 700, color: '#111827', margin: '0 0 4px', letterSpacing: '-0.02em' }}>
                                Activity Feed
                            </h1>
                            <p style={{ fontSize: '14.5px', color: '#6B7280', margin: 0 }}>
                                Stay updated with your latest security and system events.
                            </p>
                        </div>

                        {/* Tabs */}
                        <div style={{ display: 'flex', gap: '6px' }}>
                            {TABS.map((tab) => (
                                <button
                                    key={tab.value}
                                    onClick={() => setActiveTab(tab.value)}
                                    style={{
                                        padding: '7px 16px',
                                        borderRadius: '8px',
                                        fontSize: '12px',
                                        fontWeight: 600,
                                        letterSpacing: '0.02em',
                                        cursor: 'pointer',
                                        transition: 'all 0.2s',
                                        border: activeTab === tab.value ? 'none' : '1px solid #E5E7EB',
                                        backgroundColor: activeTab === tab.value ? '#2563EB' : '#ffffff',
                                        color: activeTab === tab.value ? '#ffffff' : '#374151',
                                    }}
                                >
                                    {tab.label}
                                </button>
                            ))}
                        </div>
                    </div>

                    {/* Feed items */}
                    {isLoading ? (
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                            {[1, 2, 3, 4].map((i) => (
                                <div key={i} className="portal-card" style={{ padding: '20px 24px' }}>
                                    <div style={{ display: 'flex', gap: '16px' }}>
                                        <div className="skeleton" style={{ width: 44, height: 44, borderRadius: '12px', flexShrink: 0 }} />
                                        <div style={{ flex: 1 }}>
                                            <div className="skeleton" style={{ height: 16, width: '60%', marginBottom: 8, borderRadius: 6 }} />
                                            <div className="skeleton" style={{ height: 14, width: '90%', borderRadius: 6 }} />
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    ) : filteredItems.length === 0 ? (
                        <div className="portal-card" style={{ padding: '48px 24px', textAlign: 'center' }}>
                            <p style={{ fontSize: '14px', color: '#9CA3AF', margin: 0 }}>
                                No activity for the selected filter.
                            </p>
                        </div>
                    ) : (
                        <div className="stagger-children" style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                            {filteredItems.map((item) => (
                                <ActivityFeedCard key={item.id} item={item} />
                            ))}
                        </div>
                    )}

                    {/* Load older activity */}
                    {filteredItems.length > 0 && (
                        <button
                            style={{
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                gap: '6px',
                                width: '100%',
                                marginTop: '20px',
                                padding: '12px',
                                fontSize: '13px',
                                fontWeight: 500,
                                color: '#9CA3AF',
                                backgroundColor: 'transparent',
                                border: 'none',
                                cursor: 'pointer',
                                transition: 'color 0.2s',
                            }}
                        >
                            Load older activity
                            <ChevronDown size={14} />
                        </button>
                    )}
                </div>

                {/* ════ RIGHT COLUMN — Shield + Assistance ════ */}
                <div className="notification-sidebar">
                    <SecurityShield />
                    <BreachAssistance />
                </div>
            </div>
        </div>
    );
};

export default BreachNotifications;
