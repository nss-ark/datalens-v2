import { ShieldCheck, Sparkles, Wifi } from 'lucide-react';

/**
 * Right-column Security Shield card — blue gradient with animated checkmark,
 * status indicators for Encryption and VPN.
 */
export const SecurityShield = () => (
    <div
        style={{
            background: 'linear-gradient(135deg, #4F46E5 0%, #2563EB 50%, #3B82F6 100%)',
            borderRadius: '16px',
            padding: '28px 24px',
            color: '#ffffff',
            position: 'relative',
            overflow: 'hidden',
        }}
    >
        {/* Subtle background glow */}
        <div
            style={{
                position: 'absolute',
                top: '-40%',
                right: '-30%',
                width: '200px',
                height: '200px',
                borderRadius: '50%',
                background: 'rgba(255,255,255,0.08)',
                pointerEvents: 'none',
            }}
        />

        {/* Header row */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
            <span style={{ fontSize: '11px', fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', opacity: 0.85 }}>
                SECURITY SHIELD
            </span>
            <div
                style={{
                    width: 26,
                    height: 26,
                    borderRadius: '50%',
                    backgroundColor: 'rgba(255,255,255,0.2)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                }}
            >
                <ShieldCheck size={14} />
            </div>
        </div>

        {/* Animated shield icon */}
        <div style={{ display: 'flex', justifyContent: 'center', marginBottom: '16px' }}>
            <div
                className="shield-pulse"
                style={{
                    width: 72,
                    height: 72,
                    borderRadius: '50%',
                    backgroundColor: 'rgba(255,255,255,0.15)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                }}
            >
                <ShieldCheck size={36} strokeWidth={1.8} />
            </div>
        </div>

        {/* Status text */}
        <div style={{ textAlign: 'center', marginBottom: '20px' }}>
            <h3 style={{ fontSize: '20px', fontWeight: 700, margin: '0 0 4px', letterSpacing: '-0.01em' }}>
                System Secure
            </h3>
            <p style={{ fontSize: '12.5px', opacity: 0.75, margin: 0 }}>
                Last checked: 5 mins ago
            </p>
        </div>

        {/* Status indicators */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
            <div
                style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '10px',
                    backgroundColor: 'rgba(255,255,255,0.12)',
                    borderRadius: '10px',
                    padding: '10px 14px',
                    fontSize: '13px',
                    fontWeight: 500,
                }}
            >
                <Sparkles size={16} style={{ opacity: 0.9 }} />
                Encryption: Active
            </div>
            <div
                style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '10px',
                    backgroundColor: 'rgba(255,255,255,0.12)',
                    borderRadius: '10px',
                    padding: '10px 14px',
                    fontSize: '13px',
                    fontWeight: 500,
                }}
            >
                <Wifi size={16} style={{ opacity: 0.9 }} />
                VPN Status: Healthy
            </div>
        </div>
    </div>
);
