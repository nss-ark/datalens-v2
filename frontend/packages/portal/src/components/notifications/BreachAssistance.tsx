import { AlertTriangle, Headphones, ShieldCheck, ExternalLink } from 'lucide-react';

const items = [
    {
        icon: AlertTriangle,
        iconBg: '#FEF2F2',
        iconColor: '#DC2626',
        title: 'Recent Breach Reports',
        description: 'Check if your data was part of a known breach.',
    },
    {
        icon: Headphones,
        iconBg: '#EFF6FF',
        iconColor: '#2563EB',
        title: 'Contact Security Officer',
        description: 'Direct line to our emergency response team.',
    },
    {
        icon: ShieldCheck,
        iconBg: '#EFF6FF',
        iconColor: '#2563EB',
        title: 'Identity Protection',
        description: 'Tools to monitor your identity footprint.',
    },
];

/**
 * Breach Assistance panel — sits below SecurityShield in the right column.
 */
export const BreachAssistance = () => (
    <div
        className="portal-card"
        style={{ padding: '24px' }}
    >
        <h4
            style={{
                fontSize: '11px',
                fontWeight: 700,
                letterSpacing: '0.08em',
                textTransform: 'uppercase',
                color: '#9CA3AF',
                margin: '0 0 16px',
            }}
        >
            BREACH ASSISTANCE
        </h4>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
            {items.map((item) => (
                <div
                    key={item.title}
                    style={{
                        display: 'flex',
                        gap: '12px',
                        alignItems: 'flex-start',
                        cursor: 'pointer',
                    }}
                >
                    <div
                        style={{
                            width: 36,
                            height: 36,
                            minWidth: 36,
                            borderRadius: '10px',
                            backgroundColor: item.iconBg,
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                        }}
                    >
                        <item.icon size={16} style={{ color: item.iconColor }} />
                    </div>
                    <div>
                        <p style={{ fontSize: '13.5px', fontWeight: 600, color: '#111827', margin: '0 0 2px' }}>
                            {item.title}
                        </p>
                        <p style={{ fontSize: '12px', color: '#9CA3AF', margin: 0, lineHeight: 1.5 }}>
                            {item.description}
                        </p>
                    </div>
                </div>
            ))}
        </div>

        {/* View All Resources link */}
        <button
            style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '6px',
                width: '100%',
                marginTop: '20px',
                padding: '10px',
                fontSize: '13px',
                fontWeight: 600,
                color: '#2563EB',
                backgroundColor: 'transparent',
                border: '1px solid #E5E7EB',
                borderRadius: '10px',
                cursor: 'pointer',
                transition: 'background-color 0.2s',
            }}
        >
            View All Resources
            <ExternalLink size={13} />
        </button>
    </div>
);
