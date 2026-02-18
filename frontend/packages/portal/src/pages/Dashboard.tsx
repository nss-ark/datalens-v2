
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { portalService } from '@/services/portalService';
import { toast } from '@datalens/shared';
import {
    Shield, Lock, FileText, User, Activity, ArrowRight,
    TrendingUp, CheckCircle, ChevronRight, Sparkles, Clock,
    Bell, BadgeCheck, LockKeyhole
} from 'lucide-react';

/* ────────────────────────────────────────
   Inline style objects — guarantees rendering
   regardless of Tailwind v4 compilation quirks
   ──────────────────────────────────────── */

const s = {
    /* ── Layout ── */
    wrapper: {
        display: 'flex',
        flexDirection: 'column' as const,
        gap: '48px',
        paddingBottom: '40px',
    },

    /* ── Hero ── */
    heroCenter: {
        textAlign: 'center' as const,
        marginTop: '16px',
    },
    heroBadge: {
        display: 'inline-flex',
        alignItems: 'center',
        gap: '6px',
        padding: '4px 14px',
        borderRadius: '999px',
        backgroundColor: 'rgba(59,130,246,0.08)',
        border: '1px solid rgba(59,130,246,0.15)',
        color: '#3b82f6',
        fontSize: '12px',
        fontWeight: 600,
        marginBottom: '16px',
    },
    heroTitle: {
        fontSize: 'clamp(28px, 4vw, 44px)',
        fontWeight: 700,
        color: '#111827',
        letterSpacing: '-0.02em',
        lineHeight: 1.15,
        marginBottom: '16px',
    },
    heroAccent: {
        color: '#3b82f6',
    },
    heroSub: {
        color: '#6b7280',
        fontSize: '17px',
        lineHeight: 1.7,
        maxWidth: '680px',
        margin: '0 auto',
    },

    /* ── Shared card base ── */
    card: {
        backgroundColor: '#ffffff',
        borderRadius: '16px',
        padding: '24px',
        border: '1px solid #f3f4f6',
        boxShadow: '0 4px 20px -2px rgba(0,0,0,0.05)',
        transition: 'box-shadow 0.3s, transform 0.3s',
        overflow: 'hidden' as const,
        position: 'relative' as const,
    },

    /* ── Stats cards ── */
    statsGrid: {
        display: 'grid',
        gridTemplateColumns: 'repeat(4, 1fr)',
        gap: '24px',
    },
    statsLabel: {
        fontSize: '11px',
        fontWeight: 700,
        color: '#6b7280',
        textTransform: 'uppercase' as const,
        letterSpacing: '0.05em',
    },
    statsValue: {
        fontSize: '36px',
        fontWeight: 700,
        color: '#111827',
        lineHeight: 1,
    },
    statsIconWrap: (bg: string, fg: string) => ({
        padding: '8px',
        borderRadius: '10px',
        backgroundColor: bg,
        color: fg,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        transition: 'background-color 0.3s, color 0.3s',
    }),
    trendRow: {
        display: 'flex',
        alignItems: 'center',
        gap: '4px',
        marginTop: '16px',
        fontSize: '12px',
        fontWeight: 500,
        color: '#059669',
    },
    trendMuted: {
        color: '#6b7280',
        fontWeight: 400,
        marginLeft: '4px',
    },
    progressBar: {
        marginTop: '16px',
        height: '4px',
        width: '100%',
        backgroundColor: '#f3f4f6',
        borderRadius: '999px',
        overflow: 'hidden' as const,
    },

    /* ── Feature row (3 cols: identity 2-span, consents gradient, data requests) ── */
    featureGrid: {
        display: 'grid',
        gridTemplateColumns: 'repeat(4, 1fr)',
        gap: '24px',
    },

    /* ── Identity verification card ── */
    identityCard: {
        gridColumn: 'span 2',
        backgroundColor: '#ffffff',
        borderRadius: '16px',
        padding: '24px',
        border: '1px solid #f3f4f6',
        boxShadow: '0 4px 20px -2px rgba(0,0,0,0.05)',
        display: 'flex',
        flexDirection: 'column' as const,
        justifyContent: 'space-between',
        minHeight: '280px',
    },
    identityBadge: {
        padding: '4px 10px',
        borderRadius: '4px',
        backgroundColor: '#f3f4f6',
        fontSize: '10px',
        fontWeight: 700,
        color: '#6b7280',
        textTransform: 'uppercase' as const,
    },
    digilockerBtn: {
        width: '100%',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        gap: '12px',
        padding: '12px 16px',
        borderRadius: '10px',
        backgroundColor: '#0F3460',
        color: '#ffffff',
        border: 'none',
        fontSize: '14px',
        fontWeight: 600,
        cursor: 'pointer',
        transition: 'background-color 0.2s',
        boxShadow: '0 4px 6px rgba(15,52,96,0.2)',
        fontFamily: 'inherit',
    },
    emailBtn: {
        width: '100%',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '10px 16px',
        borderRadius: '10px',
        backgroundColor: '#ffffff',
        color: '#6b7280',
        border: '1px solid #e5e7eb',
        fontSize: '13px',
        fontWeight: 500,
        cursor: 'pointer',
        transition: 'background-color 0.2s',
        fontFamily: 'inherit',
    },

    /* ── Consents gradient card ── */
    consentsCard: {
        gridColumn: 'span 1',
        background: 'linear-gradient(135deg, #3b82f6 0%, #4f46e5 100%)',
        borderRadius: '16px',
        padding: '24px',
        color: '#ffffff',
        position: 'relative' as const,
        overflow: 'hidden' as const,
        display: 'flex',
        flexDirection: 'column' as const,
        justifyContent: 'space-between',
        boxShadow: '0 10px 15px -3px rgba(59,130,246,0.25)',
        minHeight: '280px',
    },
    consentsGlow: {
        position: 'absolute' as const,
        top: '-40px',
        right: '-40px',
        width: '128px',
        height: '128px',
        backgroundColor: 'rgba(255,255,255,0.1)',
        borderRadius: '50%',
        filter: 'blur(40px)',
        pointerEvents: 'none' as const,
    },
    consentsManageBtn: {
        fontSize: '10px',
        fontWeight: 700,
        backgroundColor: 'rgba(255,255,255,0.2)',
        padding: '4px 10px',
        borderRadius: '4px',
        border: 'none',
        color: '#ffffff',
        textTransform: 'uppercase' as const,
        letterSpacing: '0.05em',
        cursor: 'pointer',
        transition: 'background-color 0.2s',
        fontFamily: 'inherit',
    },
    consentsReviewBtn: {
        width: '100%',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        gap: '8px',
        padding: '10px 16px',
        borderRadius: '10px',
        backgroundColor: 'rgba(255,255,255,0.2)',
        backdropFilter: 'blur(8px)',
        border: '1px solid rgba(255,255,255,0.1)',
        color: '#ffffff',
        fontSize: '13px',
        fontWeight: 600,
        cursor: 'pointer',
        transition: 'background-color 0.2s',
        fontFamily: 'inherit',
    },

    /* ── Data requests card ── */
    dataRequestsCard: {
        gridColumn: 'span 1',
        backgroundColor: '#ffffff',
        borderRadius: '16px',
        padding: '24px',
        border: '1px solid #f3f4f6',
        boxShadow: '0 4px 20px -2px rgba(0,0,0,0.05)',
        display: 'flex',
        flexDirection: 'column' as const,
        justifyContent: 'space-between',
        minHeight: '280px',
    },
    viewHistoryBtn: {
        width: '100%',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '10px 16px',
        borderRadius: '10px',
        backgroundColor: '#f9fafb',
        color: '#111827',
        border: '1px solid #f3f4f6',
        fontSize: '13px',
        fontWeight: 600,
        cursor: 'pointer',
        transition: 'background-color 0.2s',
        fontFamily: 'inherit',
    },

    /* ── Quick actions ── */
    quickActionsGrid: {
        display: 'grid',
        gridTemplateColumns: 'repeat(3, 1fr)',
        gap: '24px',
    },
    actionCard: {
        backgroundColor: '#ffffff',
        borderRadius: '16px',
        padding: '24px',
        border: '1px solid #f3f4f6',
        boxShadow: '0 4px 20px -2px rgba(0,0,0,0.05)',
        cursor: 'pointer',
        transition: 'box-shadow 0.3s, border-color 0.3s, transform 0.2s',
        textDecoration: 'none',
        display: 'block',
    },
    actionIcon: {
        width: '40px',
        height: '40px',
        borderRadius: '10px',
        backgroundColor: 'rgba(59,130,246,0.08)',
        color: '#3b82f6',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        marginBottom: '16px',
        transition: 'transform 0.3s',
    },

    /* ── Safety alerts ── */
    alertBanner: {
        backgroundColor: '#ffffff',
        borderRadius: '16px',
        padding: '24px',
        border: '1px solid #f3f4f6',
        boxShadow: '0 4px 20px -2px rgba(0,0,0,0.05)',
        display: 'flex',
        alignItems: 'center',
        gap: '16px',
        cursor: 'pointer',
        transition: 'box-shadow 0.3s',
    },
    alertIcon: {
        flexShrink: 0,
        padding: '12px',
        backgroundColor: 'rgba(59,130,246,0.08)',
        borderRadius: '12px',
        color: '#3b82f6',
        display: 'flex',
    },
};

const PortalDashboard = () => {
    const navigate = useNavigate();

    const { data: consentSummary } = useQuery({
        queryKey: ['consents-summary'],
        queryFn: portalService.getConsentSummary
    });

    const { data: identityStatus } = useQuery({
        queryKey: ['portal-identity'],
        queryFn: portalService.getIdentityStatus
    });

    const activeConsents = consentSummary?.filter(c => c.status === 'GRANTED').length || 0;
    const totalConsents = consentSummary?.length || 0;
    const identityLevel = identityStatus?.assurance_level || 'NONE';
    const isVerified = identityLevel === 'SUBSTANTIAL' || identityLevel === 'HIGH';

    const handleVerify = () => {
        toast.info('Redirecting to DigiLocker...');
    };

    return (
        <div style={s.wrapper}>
            {/* ══════════════════════════════════════
                HERO SECTION
               ══════════════════════════════════════ */}
            <div style={s.heroCenter}>
                <div style={s.heroBadge}>
                    <Sparkles size={14} />
                    Privacy Dashboard
                </div>
                <h1 style={s.heroTitle}>
                    Manage your data privacy<br />
                    <span style={s.heroAccent}>with confidence.</span>
                </h1>
                <p style={s.heroSub}>
                    Gain full visibility into how your data is used. Manage consents, track
                    requests, and exercise your rights from a single, secure control center.
                </p>
            </div>

            {/* ══════════════════════════════════════
                STATS ROW (4 columns)
               ══════════════════════════════════════ */}
            <div style={s.statsGrid} className="portal-stats-grid">
                {/* Active Consents */}
                <div style={s.card}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '16px' }}>
                        <span style={s.statsLabel}>Active Consents</span>
                        <div style={s.statsIconWrap('#eff6ff', '#3b82f6')}>
                            <Shield size={20} />
                        </div>
                    </div>
                    <span style={s.statsValue}>{activeConsents}</span>
                    <div style={s.trendRow}>
                        <TrendingUp size={14} />
                        1 New <span style={s.trendMuted}>vs last month</span>
                    </div>
                </div>

                {/* Open Requests */}
                <div style={s.card}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '16px' }}>
                        <span style={s.statsLabel}>Open Requests</span>
                        <div style={s.statsIconWrap('#fff7ed', '#f97316')}>
                            <Clock size={20} />
                        </div>
                    </div>
                    <span style={s.statsValue}>0</span>
                    <div style={s.progressBar}>
                        <div style={{ height: '100%', width: '0%', backgroundColor: '#f97316', borderRadius: '999px' }} />
                    </div>
                </div>

                {/* Recent Activity */}
                <div style={s.card}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '16px' }}>
                        <span style={s.statsLabel}>Recent Activity</span>
                        <div style={s.statsIconWrap('#faf5ff', '#a855f7')}>
                            <Activity size={20} />
                        </div>
                    </div>
                    <span style={s.statsValue}>2</span>
                    <div style={{ marginTop: '16px', display: 'flex' }}>
                        <div style={{
                            width: '24px', height: '24px', borderRadius: '50%', backgroundColor: '#e5e7eb',
                            display: 'flex', alignItems: 'center', justifyContent: 'center',
                            fontSize: '10px', color: '#6b7280', border: '2px solid #fff',
                        }}>S</div>
                        <div style={{
                            width: '24px', height: '24px', borderRadius: '50%', backgroundColor: '#e5e7eb',
                            display: 'flex', alignItems: 'center', justifyContent: 'center',
                            fontSize: '10px', color: '#6b7280', border: '2px solid #fff',
                            marginLeft: '-8px',
                        }}>L</div>
                    </div>
                </div>

                {/* Verification */}
                <div style={{ ...s.card, display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
                    <span style={s.statsLabel}>Verification</span>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginTop: '8px' }}>
                        <span style={{ fontSize: '24px', fontWeight: 700, color: '#111827' }}>
                            {isVerified ? 'Verified' : 'Unverified'}
                        </span>
                        {isVerified && <CheckCircle size={20} style={{ color: '#3b82f6' }} />}
                    </div>
                    <button
                        onClick={() => navigate('/profile')}
                        style={{
                            display: 'flex', alignItems: 'center', gap: '4px',
                            marginTop: '16px', fontSize: '13px', fontWeight: 600,
                            color: '#3b82f6', backgroundColor: 'transparent',
                            border: 'none', cursor: 'pointer', padding: 0,
                            fontFamily: 'inherit', transition: 'color 0.2s',
                        }}
                    >
                        View Profile <ArrowRight size={14} />
                    </button>
                </div>
            </div>

            {/* ══════════════════════════════════════
                FEATURE ROW (Identity 2-span + Consents + Data Requests)
               ══════════════════════════════════════ */}
            <div style={s.featureGrid} className="portal-feature-grid">
                {/* ── Identity Verification (2-col span) ── */}
                <div style={s.identityCard}>
                    <div>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '8px' }}>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                                <BadgeCheck size={20} style={{ color: '#6b7280' }} />
                                <h3 style={{ fontSize: '17px', fontWeight: 700, color: '#111827' }}>
                                    Identity Verification
                                </h3>
                            </div>
                            <span style={s.identityBadge}>
                                {isVerified ? 'Verified' : 'Unverified'}
                            </span>
                        </div>
                        <p style={{ fontSize: '13px', color: '#6b7280', marginBottom: '32px' }}>
                            Verify your identity to access sensitive data requests securely.
                        </p>
                    </div>

                    <div>
                        <div style={{ marginBottom: '16px' }}>
                            <div style={{ fontSize: '11px', fontWeight: 600, color: '#6b7280', textTransform: 'uppercase' as const, letterSpacing: '0.05em', marginBottom: '4px' }}>
                                Current Level
                            </div>
                            <h4 style={{ fontSize: '20px', fontWeight: 700, color: '#111827' }}>
                                {identityLevel === 'NONE' ? 'Basic Account' : identityLevel}
                            </h4>
                        </div>
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                            <button
                                onClick={handleVerify}
                                style={s.digilockerBtn}
                                onMouseEnter={e => { e.currentTarget.style.backgroundColor = '#16213e'; }}
                                onMouseLeave={e => { e.currentTarget.style.backgroundColor = '#0F3460'; }}
                            >
                                <Lock size={18} />
                                Verify with DigiLocker
                            </button>
                            <button
                                style={s.emailBtn}
                                onMouseEnter={e => { e.currentTarget.style.backgroundColor = '#f9fafb'; }}
                                onMouseLeave={e => { e.currentTarget.style.backgroundColor = '#ffffff'; }}
                            >
                                Continue with Email (Restricted)
                            </button>
                        </div>
                    </div>
                </div>

                {/* ── Total Consents (blue gradient) ── */}
                <div style={s.consentsCard}>
                    <div style={s.consentsGlow} />
                    <div style={{ position: 'relative', zIndex: 1 }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '24px' }}>
                            <LockKeyhole size={24} style={{ opacity: 0.8 }} />
                            <button
                                onClick={() => navigate('/consent')}
                                style={s.consentsManageBtn}
                                onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'rgba(255,255,255,0.3)'; }}
                                onMouseLeave={e => { e.currentTarget.style.backgroundColor = 'rgba(255,255,255,0.2)'; }}
                            >
                                Manage
                            </button>
                        </div>
                        <span style={{ fontSize: '13px', fontWeight: 500, opacity: 0.8, display: 'block', marginBottom: '4px' }}>
                            Total Consents
                        </span>
                        <span style={{ fontSize: '36px', fontWeight: 700 }}>{totalConsents}</span>
                    </div>
                    <div style={{ marginTop: '24px', position: 'relative', zIndex: 1 }}>
                        <button
                            onClick={() => navigate('/consent')}
                            style={s.consentsReviewBtn}
                            onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'rgba(255,255,255,0.3)'; }}
                            onMouseLeave={e => { e.currentTarget.style.backgroundColor = 'rgba(255,255,255,0.2)'; }}
                        >
                            Review All <ArrowRight size={14} />
                        </button>
                    </div>
                </div>

                {/* ── Data Requests ── */}
                <div style={s.dataRequestsCard}>
                    <div>
                        <FileText size={24} style={{ color: '#6b7280', marginBottom: '8px' }} />
                        <span style={{ fontSize: '13px', fontWeight: 500, color: '#6b7280', display: 'block', marginBottom: '4px' }}>
                            Data Requests
                        </span>
                        <span style={{ fontSize: '36px', fontWeight: 700, color: '#111827' }}>0</span>
                    </div>
                    <button
                        onClick={() => navigate('/requests')}
                        style={s.viewHistoryBtn}
                        onMouseEnter={e => { e.currentTarget.style.backgroundColor = '#f3f4f6'; }}
                        onMouseLeave={e => { e.currentTarget.style.backgroundColor = '#f9fafb'; }}
                    >
                        View History
                    </button>
                </div>
            </div>

            {/* ══════════════════════════════════════
                QUICK ACTIONS
               ══════════════════════════════════════ */}
            <div>
                <h2 style={{ fontSize: '24px', fontWeight: 700, color: '#111827', marginBottom: '4px' }}>
                    Quick Actions
                </h2>
                <p style={{ fontSize: '14px', color: '#6b7280', marginBottom: '24px' }}>
                    Exercise your privacy rights
                </p>
                <div style={s.quickActionsGrid} className="portal-actions-grid">
                    {[
                        {
                            title: 'Submit Request',
                            desc: 'Exercise your rights to access, correct, or erase your personal data.',
                            icon: FileText,
                            onClick: () => navigate('/requests/new'),
                        },
                        {
                            title: 'Manage Consents',
                            desc: 'Review and control which applications can access your data.',
                            icon: Lock,
                            onClick: () => navigate('/consent'),
                        },
                        {
                            title: 'My Profile',
                            desc: 'Update your personal details and verification status.',
                            icon: User,
                            onClick: () => navigate('/profile'),
                        },
                    ].map(item => (
                        <div
                            key={item.title}
                            onClick={item.onClick}
                            style={s.actionCard}
                            onMouseEnter={e => {
                                e.currentTarget.style.boxShadow = '0 10px 25px -5px rgba(0,0,0,0.08)';
                                e.currentTarget.style.borderColor = '#bfdbfe';
                            }}
                            onMouseLeave={e => {
                                e.currentTarget.style.boxShadow = '0 4px 20px -2px rgba(0,0,0,0.05)';
                                e.currentTarget.style.borderColor = '#f3f4f6';
                            }}
                        >
                            <div style={s.actionIcon}>
                                <item.icon size={20} />
                            </div>
                            <h3 style={{ fontSize: '17px', fontWeight: 700, color: '#111827', marginBottom: '8px' }}>
                                {item.title}
                            </h3>
                            <p style={{ fontSize: '13px', color: '#6b7280', lineHeight: 1.6 }}>
                                {item.desc}
                            </p>
                        </div>
                    ))}
                </div>
            </div>

            {/* ══════════════════════════════════════
                SAFETY ALERTS BANNER
               ══════════════════════════════════════ */}
            <div
                onClick={() => navigate('/notifications/breach')}
                style={s.alertBanner}
                onMouseEnter={e => { e.currentTarget.style.boxShadow = '0 10px 25px -5px rgba(0,0,0,0.08)'; }}
                onMouseLeave={e => { e.currentTarget.style.boxShadow = '0 4px 20px -2px rgba(0,0,0,0.05)'; }}
            >
                <div style={s.alertIcon}>
                    <Bell size={24} />
                </div>
                <div style={{ flex: 1 }}>
                    <h3 style={{ fontSize: '17px', fontWeight: 700, color: '#111827' }}>Safety Alerts</h3>
                    <p style={{ fontSize: '13px', color: '#6b7280' }}>
                        View breach notifications and security alerts related to your accounts.
                    </p>
                </div>
                <ChevronRight size={20} style={{ color: '#d1d5db', flexShrink: 0 }} />
            </div>
        </div>
    );
};

export default PortalDashboard;
