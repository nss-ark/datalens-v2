import React from 'react';
import { Shield, Lock, Fingerprint, Globe } from 'lucide-react';

interface AuthLayoutProps {
    children: React.ReactNode;
}

export const AuthLayout: React.FC<AuthLayoutProps> = ({ children }) => {
    return (
        <div style={{
            display: 'flex',
            minHeight: '100vh',
            fontFamily: "var(--font-sans, 'Inter', -apple-system, sans-serif)",
        }}>
            {/* ── Left Side — Branding & Visual ── */}
            <div style={{
                display: 'none',
                width: '52%',
                position: 'relative',
                overflow: 'hidden',
            }}
                className="lg:!flex"
            >
                {/* Gradient background */}
                <div style={{
                    position: 'absolute',
                    inset: 0,
                    background: 'linear-gradient(135deg, #0f172a 0%, #1e293b 40%, #0f172a 100%)',
                }} />
                {/* Accent radials */}
                <div style={{
                    position: 'absolute',
                    inset: 0,
                    opacity: 0.3,
                    backgroundImage: `radial-gradient(circle at 20% 80%, rgba(59,130,246,0.3), transparent 50%),
                                     radial-gradient(circle at 80% 20%, rgba(99,102,241,0.2), transparent 50%),
                                     radial-gradient(circle at 50% 50%, rgba(37,99,235,0.15), transparent 70%)`,
                }} />
                {/* Grid pattern */}
                <div style={{
                    position: 'absolute',
                    inset: 0,
                    opacity: 0.04,
                    backgroundImage: `linear-gradient(rgba(255,255,255,.1) 1px, transparent 1px),
                                     linear-gradient(90deg, rgba(255,255,255,.1) 1px, transparent 1px)`,
                    backgroundSize: '64px 64px',
                }} />

                <div style={{
                    position: 'relative',
                    zIndex: 10,
                    width: '100%',
                    display: 'flex',
                    flexDirection: 'column',
                    justifyContent: 'space-between',
                    padding: '48px',
                    color: '#ffffff',
                }}>
                    {/* Logo */}
                    <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                        <div style={{
                            background: 'rgba(255,255,255,0.1)',
                            padding: '10px',
                            borderRadius: '12px',
                            backdropFilter: 'blur(8px)',
                            border: '1px solid rgba(255,255,255,0.1)',
                        }}>
                            <Shield style={{ width: '28px', height: '28px', color: '#ffffff' }} />
                        </div>
                        <span style={{
                            fontSize: '24px',
                            fontWeight: 700,
                            letterSpacing: '-0.02em',
                        }}>DataLens</span>
                    </div>

                    {/* Hero Copy */}
                    <div style={{ maxWidth: '480px' }}>
                        <h1 style={{
                            fontSize: '48px',
                            fontWeight: 800,
                            lineHeight: 1.1,
                            letterSpacing: '-0.03em',
                            margin: '0 0 24px 0',
                        }}>
                            Your Privacy,<br />
                            <span style={{
                                background: 'linear-gradient(90deg, #60a5fa, #67e8f9)',
                                WebkitBackgroundClip: 'text',
                                WebkitTextFillColor: 'transparent',
                            }}>
                                Under Your Control.
                            </span>
                        </h1>
                        <p style={{
                            color: 'rgba(191, 219, 254, 0.8)',
                            fontSize: '17px',
                            lineHeight: 1.7,
                            margin: '0 0 32px 0',
                        }}>
                            Manage your consent, exercise your data rights, and track how your
                            information is used—all from one secure dashboard.
                        </p>

                        {/* Trust Badges */}
                        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '12px' }}>
                            {[
                                { icon: Lock, label: 'End-to-End Encrypted' },
                                { icon: Fingerprint, label: 'DPDPA Compliant' },
                                { icon: Globe, label: 'ISO 27001' },
                            ].map(({ icon: Icon, label }) => (
                                <div key={label} style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: '8px',
                                    background: 'rgba(255,255,255,0.06)',
                                    border: '1px solid rgba(255,255,255,0.1)',
                                    borderRadius: '999px',
                                    padding: '8px 16px',
                                    fontSize: '13px',
                                    color: 'rgba(191, 219, 254, 0.9)',
                                    backdropFilter: 'blur(8px)',
                                }}>
                                    <Icon style={{ width: '14px', height: '14px' }} />
                                    <span>{label}</span>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Footer */}
                    <div style={{
                        fontSize: '13px',
                        color: 'rgba(147, 197, 253, 0.5)',
                    }}>
                        © {new Date().getFullYear()} ComplyArk. Secure Privacy Infrastructure.
                    </div>
                </div>
            </div>

            {/* ── Right Side — Form ── */}
            <div style={{
                width: '100%',
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'center',
                alignItems: 'center',
                padding: '48px 24px',
                position: 'relative',
                backgroundColor: '#ffffff',
            }}
                className="lg:!w-[48%]"
            >
                <div className="animate-fade-in" style={{
                    width: '100%',
                    maxWidth: '400px',
                }}>
                    {/* Mobile Branding — only shows on small screens */}
                    <div className="lg:hidden" style={{
                        display: 'flex',
                        justifyContent: 'center',
                        marginBottom: '48px',
                    }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                            <div style={{
                                backgroundColor: 'var(--primary-600)',
                                padding: '8px',
                                borderRadius: '12px',
                            }}>
                                <Shield style={{ width: '24px', height: '24px', color: '#ffffff' }} />
                            </div>
                            <span style={{
                                fontSize: '20px',
                                fontWeight: 700,
                                color: 'var(--slate-900)',
                            }}>DataLens</span>
                        </div>
                    </div>

                    {children}
                </div>
            </div>
        </div>
    );
};
