import React, { useState } from 'react';
import { usePortalAuthStore } from '@/stores/portalAuthStore';
import { useNavigate } from 'react-router-dom';
import { portalService } from '@/services/portalService';
import { toast } from '@datalens/shared';
import { Mail, Smartphone, ArrowRight, Loader2, ShieldCheck, KeyRound } from 'lucide-react';

const PortalLogin = () => {
    const [step, setStep] = useState<'IDENTIFIER' | 'OTP'>('IDENTIFIER');
    const [identifier, setIdentifier] = useState('');
    const [otp, setOtp] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const setAuth = usePortalAuthStore((s) => s.setAuth);
    const navigate = useNavigate();

    const isEmail = (input: string) => /\S+@\S+\.\S+/.test(input);

    const handleSendOtp = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsLoading(true);
        try {
            const payload = isEmail(identifier)
                ? { email: identifier }
                : { phone: identifier };
            await portalService.requestOTP(payload);
            setStep('OTP');
            toast.success('Verification code sent!');
        } catch {
            toast.error('Failed to send OTP. Please try again.');
        } finally {
            setIsLoading(false);
        }
    };

    const handleVerifyOtp = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsLoading(true);
        try {
            const payload = isEmail(identifier)
                ? { email: identifier, code: otp }
                : { phone: identifier, code: otp };
            const res = await portalService.verifyOTP(payload);
            setAuth(res.token.access_token, res.profile);
            toast.success('Login successful!');
            navigate('/');
        } catch {
            toast.error('Invalid code. Please try again.');
        } finally {
            setIsLoading(false);
        }
    };

    /* ── Styles using the portal's CSS design tokens ── */
    const styles = {
        wrapper: {
            width: '100%',
            display: 'flex',
            flexDirection: 'column' as const,
            gap: '40px',
        },
        heading: {
            fontSize: '30px',
            fontWeight: 700,
            letterSpacing: '-0.02em',
            color: 'var(--slate-900)',
            lineHeight: 1.2,
        },
        subtitle: {
            marginTop: '8px',
            fontSize: '15px',
            color: 'var(--slate-500)',
            lineHeight: 1.6,
        },
        label: {
            display: 'block',
            fontSize: '13px',
            fontWeight: 600,
            color: 'var(--slate-700)',
            marginBottom: '8px',
        },
        inputWrapper: {
            position: 'relative' as const,
        },
        inputIcon: {
            position: 'absolute' as const,
            left: '14px',
            top: '50%',
            transform: 'translateY(-50%)',
            color: 'var(--slate-400)',
            pointerEvents: 'none' as const,
            display: 'flex',
            alignItems: 'center',
        },
        input: {
            display: 'block',
            width: '100%',
            paddingLeft: '44px',
            paddingRight: '16px',
            paddingTop: '14px',
            paddingBottom: '14px',
            fontSize: '15px',
            color: 'var(--slate-900)',
            backgroundColor: '#ffffff',
            border: '1px solid var(--slate-300)',
            borderRadius: '12px',
            outline: 'none',
            transition: 'border-color 0.2s, box-shadow 0.2s',
            fontFamily: 'inherit',
        },
        otpInput: {
            display: 'block',
            width: '100%',
            paddingLeft: '44px',
            paddingRight: '16px',
            paddingTop: '14px',
            paddingBottom: '14px',
            fontSize: '22px',
            color: 'var(--slate-900)',
            backgroundColor: '#ffffff',
            border: '1px solid var(--slate-300)',
            borderRadius: '12px',
            outline: 'none',
            transition: 'border-color 0.2s, box-shadow 0.2s',
            fontFamily: "'Courier New', monospace",
            letterSpacing: '0.3em',
            textAlign: 'center' as const,
        },
        button: {
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            gap: '8px',
            width: '100%',
            padding: '14px 16px',
            fontSize: '14px',
            fontWeight: 600,
            color: '#ffffff',
            backgroundColor: 'var(--slate-900)',
            border: 'none',
            borderRadius: '12px',
            cursor: 'pointer',
            transition: 'background-color 0.2s, transform 0.1s, opacity 0.2s',
            boxShadow: '0 4px 12px rgba(15, 23, 42, 0.15)',
            fontFamily: 'inherit',
        },
        buttonDisabled: {
            opacity: 0.5,
            cursor: 'not-allowed',
        },
        backButton: {
            display: 'block',
            width: '100%',
            padding: '8px',
            fontSize: '13px',
            fontWeight: 500,
            color: 'var(--slate-500)',
            backgroundColor: 'transparent',
            border: 'none',
            cursor: 'pointer',
            textAlign: 'center' as const,
            fontFamily: 'inherit',
            transition: 'color 0.2s',
        },
        terms: {
            textAlign: 'center' as const,
            fontSize: '12px',
            color: 'var(--slate-400)',
            lineHeight: 1.6,
        },
        termsLink: {
            color: 'var(--slate-600)',
            textDecoration: 'underline',
            textUnderlineOffset: '2px',
        },
        form: {
            display: 'flex',
            flexDirection: 'column' as const,
            gap: '24px',
        },
    };

    const isDisabled = step === 'IDENTIFIER'
        ? isLoading || !identifier.trim()
        : isLoading || otp.length < 6;

    return (
        <div style={styles.wrapper}>
            {/* ── Heading ── */}
            <div>
                <h2 style={styles.heading}>
                    {step === 'IDENTIFIER' ? 'Welcome back' : 'Verify identity'}
                </h2>
                <p style={styles.subtitle}>
                    {step === 'IDENTIFIER'
                        ? 'Enter your email or phone to access your privacy portal.'
                        : `Enter the 6-digit code sent to ${identifier}`
                    }
                </p>
            </div>

            {/* ── Forms ── */}
            {step === 'IDENTIFIER' ? (
                <form onSubmit={handleSendOtp} style={styles.form}>
                    <div>
                        <label htmlFor="identifier" style={styles.label}>
                            Email or Phone
                        </label>
                        <div style={styles.inputWrapper}>
                            <div style={styles.inputIcon}>
                                {isEmail(identifier) || !identifier
                                    ? <Mail size={18} />
                                    : <Smartphone size={18} />
                                }
                            </div>
                            <input
                                id="identifier"
                                type="text"
                                value={identifier}
                                onChange={(e) => setIdentifier(e.target.value)}
                                disabled={isLoading}
                                style={styles.input}
                                placeholder="john.doe@example.com"
                                required
                                onFocus={(e) => {
                                    e.target.style.borderColor = 'var(--primary-500)';
                                    e.target.style.boxShadow = '0 0 0 3px rgba(59,130,246,0.12)';
                                }}
                                onBlur={(e) => {
                                    e.target.style.borderColor = 'var(--slate-300)';
                                    e.target.style.boxShadow = 'none';
                                }}
                            />
                        </div>
                    </div>

                    <button
                        type="submit"
                        disabled={isDisabled}
                        style={{
                            ...styles.button,
                            ...(isDisabled ? styles.buttonDisabled : {}),
                        }}
                        onMouseEnter={(e) => {
                            if (!isDisabled) e.currentTarget.style.backgroundColor = 'var(--slate-800)';
                        }}
                        onMouseLeave={(e) => {
                            e.currentTarget.style.backgroundColor = 'var(--slate-900)';
                        }}
                    >
                        {isLoading ? (
                            <Loader2 size={20} className="animate-spin" />
                        ) : (
                            <>Continue <ArrowRight size={16} /></>
                        )}
                    </button>
                </form>
            ) : (
                <form onSubmit={handleVerifyOtp} style={styles.form}>
                    <div>
                        <label htmlFor="otp" style={styles.label}>
                            Verification Code
                        </label>
                        <div style={styles.inputWrapper}>
                            <div style={styles.inputIcon}>
                                <KeyRound size={18} />
                            </div>
                            <input
                                id="otp"
                                type="text"
                                value={otp}
                                onChange={(e) => setOtp(e.target.value)}
                                disabled={isLoading}
                                style={styles.otpInput}
                                placeholder="000000"
                                maxLength={6}
                                required
                                onFocus={(e) => {
                                    e.target.style.borderColor = 'var(--primary-500)';
                                    e.target.style.boxShadow = '0 0 0 3px rgba(59,130,246,0.12)';
                                }}
                                onBlur={(e) => {
                                    e.target.style.borderColor = 'var(--slate-300)';
                                    e.target.style.boxShadow = 'none';
                                }}
                            />
                        </div>
                    </div>

                    <button
                        type="submit"
                        disabled={isDisabled}
                        style={{
                            ...styles.button,
                            ...(isDisabled ? styles.buttonDisabled : {}),
                        }}
                        onMouseEnter={(e) => {
                            if (!isDisabled) e.currentTarget.style.backgroundColor = 'var(--slate-800)';
                        }}
                        onMouseLeave={(e) => {
                            e.currentTarget.style.backgroundColor = 'var(--slate-900)';
                        }}
                    >
                        {isLoading ? (
                            <Loader2 size={20} className="animate-spin" />
                        ) : (
                            <>Verify & Login <ShieldCheck size={16} /></>
                        )}
                    </button>

                    <button
                        type="button"
                        onClick={() => { setStep('IDENTIFIER'); setOtp(''); }}
                        style={styles.backButton}
                        onMouseEnter={(e) => { e.currentTarget.style.color = 'var(--slate-900)'; }}
                        onMouseLeave={(e) => { e.currentTarget.style.color = 'var(--slate-500)'; }}
                    >
                        ← Change {isEmail(identifier) ? 'email' : 'phone'}
                    </button>
                </form>
            )}

            {/* ── Terms ── */}
            <p style={styles.terms}>
                By continuing, you agree to our{' '}
                <a href="#" style={styles.termsLink}>Terms of Service</a>
                {' '}and{' '}
                <a href="#" style={styles.termsLink}>Privacy Policy</a>.
            </p>
        </div>
    );
};

export default PortalLogin;
