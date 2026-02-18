import React from 'react';
import {
    Eye, Pencil, Trash2, UserPlus,
    CheckCircle2
} from 'lucide-react';

/* ────────────────────────────────────────
   Inline Styles for RequestSidebar
   ──────────────────────────────────────── */
const styles = {
    sidebar: {
        display: 'flex',
        flexDirection: 'column' as const,
        gap: '32px',
    },
    // Stepper Section
    sectionTitle: {
        fontSize: '12px',
        fontWeight: 700,
        textTransform: 'uppercase' as const,
        letterSpacing: '0.05em',
        color: '#6b7280', // slate-500
        marginBottom: '16px',
        display: 'flex',
        alignItems: 'center',
        gap: '8px',
    },
    stepper: {
        display: 'flex',
        flexDirection: 'column' as const,
        gap: '0',
        position: 'relative' as const,
    },
    stepItem: {
        display: 'flex',
        gap: '16px',
        paddingBottom: '24px',
        position: 'relative' as const,
    },
    stepLine: {
        position: 'absolute' as const,
        top: '24px',
        left: '12px',
        bottom: '0',
        width: '1px',
        backgroundColor: '#e5e7eb', // slate-200
        zIndex: 0,
    },
    stepCircle: {
        width: '24px',
        height: '24px',
        borderRadius: '50%',
        backgroundColor: '#eff6ff', // blue-50
        border: '1px solid #dbeafe', // blue-100
        color: '#3b82f6', // blue-500
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        fontSize: '11px',
        fontWeight: 600,
        position: 'relative' as const,
        zIndex: 1,
        flexShrink: 0,
    },
    stepContent: {
        paddingTop: '2px',
    },
    stepTitle: {
        fontSize: '13px',
        fontWeight: 600,
        color: '#111827', // slate-900
        marginBottom: '2px',
    },
    stepDesc: {
        fontSize: '12px',
        color: '#6b7280', // slate-500
        lineHeight: 1.4,
    },

    // Grid Section
    grid: {
        display: 'grid',
        gridTemplateColumns: 'repeat(2, 1fr)',
        gap: '12px',
    },
    card: {
        backgroundColor: '#ffffff',
        border: '1px solid #e2e8f0', // slate-200
        borderRadius: '12px',
        padding: '16px',
        display: 'flex',
        flexDirection: 'column' as const,
        alignItems: 'center',
        justifyContent: 'center',
        gap: '8px',
        cursor: 'pointer',
        transition: 'all 0.2s ease',
        textAlign: 'center' as const,
    },
    cardHover: {
        borderColor: '#93c5fd', // blue-300
        backgroundColor: '#eff6ff', // blue-50
        transform: 'translateY(-1px)',
    },
    iconWrapper: {
        width: '32px',
        height: '32px',
        borderRadius: '8px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        marginBottom: '4px',
    },
    cardLabel: {
        fontSize: '11px',
        fontWeight: 600,
        color: '#475569', // slate-600
    },

    // Colors for icons
    colorBlue: { backgroundColor: '#eff6ff', color: '#3b82f6' },
    colorGreen: { backgroundColor: '#f0fdf4', color: '#10b981' },
    colorRed: { backgroundColor: '#fef2f2', color: '#ef4444' },
    colorOrange: { backgroundColor: '#fff7ed', color: '#f97316' },
};

interface RequestSidebarProps {
    onRequestTypeSelect: (type: 'ACCESS' | 'CORRECTION' | 'ERASURE' | 'NOMINATION') => void;
}

export const RequestSidebar: React.FC<RequestSidebarProps> = ({ onRequestTypeSelect }) => {
    // Helper to handle hover state manually since we use inline styles
    const handleMouseEnter = (e: React.MouseEvent<HTMLDivElement>) => {
        Object.assign(e.currentTarget.style, styles.cardHover);
    };
    const handleMouseLeave = (e: React.MouseEvent<HTMLDivElement>) => {
        e.currentTarget.style.borderColor = '#e2e8f0';
        e.currentTarget.style.backgroundColor = '#ffffff';
        e.currentTarget.style.transform = 'none';
    };

    return (
        <div style={styles.sidebar}>
            {/* ── How It Works ── */}
            <div>
                <div style={styles.sectionTitle}>
                    <CheckCircle2 size={14} />
                    How It Works
                </div>
                <div style={styles.stepper}>
                    <div style={styles.stepItem}>
                        {/* Wrapper div to hide last line */}
                        <div style={styles.stepLine} />
                        <div style={styles.stepCircle}>1</div>
                        <div style={styles.stepContent}>
                            <div style={styles.stepTitle}>Select Type</div>
                            <div style={styles.stepDesc}>Choose the right you want to exercise.</div>
                        </div>
                    </div>
                    <div style={styles.stepItem}>
                        <div style={styles.stepLine} />
                        <div style={styles.stepCircle}>2</div>
                        <div style={styles.stepContent}>
                            <div style={styles.stepTitle}>Verify ID</div>
                            <div style={styles.stepDesc}>Securely confirm your identity.</div>
                        </div>
                    </div>
                    <div style={{ ...styles.stepItem, paddingBottom: 0 }}>
                        {/* No line for last item */}
                        <div style={styles.stepCircle}>3</div>
                        <div style={styles.stepContent}>
                            <div style={styles.stepTitle}>Receive Data</div>
                            <div style={styles.stepDesc}>Get results within 30 days.</div>
                        </div>
                    </div>
                </div>
            </div>

            {/* ── Request Types Grid ── */}
            <div className="bg-slate-50/50 p-4 rounded-xl border border-slate-100">
                <div style={styles.sectionTitle}>
                    Request Types
                </div>
                <div style={styles.grid}>
                    <div
                        style={styles.card}
                        onMouseEnter={handleMouseEnter}
                        onMouseLeave={handleMouseLeave}
                        onClick={() => onRequestTypeSelect('ACCESS')}
                    >
                        <div style={{ ...styles.iconWrapper, ...styles.colorBlue }}>
                            <Eye size={16} />
                        </div>
                        <span style={styles.cardLabel}>Access</span>
                    </div>

                    <div
                        style={styles.card}
                        onMouseEnter={handleMouseEnter}
                        onMouseLeave={handleMouseLeave}
                        onClick={() => onRequestTypeSelect('CORRECTION')}
                    >
                        <div style={{ ...styles.iconWrapper, ...styles.colorGreen }}>
                            <Pencil size={16} />
                        </div>
                        <span style={styles.cardLabel}>Correct</span>
                    </div>

                    <div
                        style={styles.card}
                        onMouseEnter={handleMouseEnter}
                        onMouseLeave={handleMouseLeave}
                        onClick={() => onRequestTypeSelect('ERASURE')}
                    >
                        <div style={{ ...styles.iconWrapper, ...styles.colorRed }}>
                            <Trash2 size={16} />
                        </div>
                        <span style={styles.cardLabel}>Erase</span>
                    </div>

                    <div
                        style={styles.card}
                        onMouseEnter={handleMouseEnter}
                        onMouseLeave={handleMouseLeave}
                        onClick={() => onRequestTypeSelect('NOMINATION')}
                    >
                        <div style={{ ...styles.iconWrapper, ...styles.colorOrange }}>
                            <UserPlus size={16} />
                        </div>
                        <span style={styles.cardLabel}>Nomination</span>
                    </div>
                </div>
            </div>
        </div>
    );
};
