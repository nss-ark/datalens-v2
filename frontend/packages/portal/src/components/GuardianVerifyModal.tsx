import React, { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { portalService } from '@/services/portalService';
import { Modal } from '@datalens/shared';
import { Button } from '@datalens/shared';
import { OTPInput } from './OTPInput';
import { toast } from '@datalens/shared';
import { Mail, ShieldCheck } from 'lucide-react';
import { AxiosError } from 'axios';

interface GuardianVerifyModalProps {
    isOpen: boolean;
    onClose: () => void;
    onVerified: () => void;
    guardianEmail?: string; // Pre-fill if known
}

export const GuardianVerifyModal: React.FC<GuardianVerifyModalProps> = ({
    isOpen,
    onClose,
    onVerified,
    guardianEmail: initialEmail = ''
}) => {
    const [step, setStep] = useState<'INIT' | 'VERIFY'>('INIT');
    const [email, setEmail] = useState(initialEmail);
    const [otp, setOtp] = useState('');

    // Step 1: Initiate Verification
    const initiateMutation = useMutation({
        mutationFn: portalService.initiateGuardianVerify,
        onSuccess: () => {
            setStep('VERIFY');
            toast.success('Code Sent', `Verification code sent to ${email}`);
        },
        onError: (err: unknown) => {
            const error = err as AxiosError<{ message: string }>;
            toast.error('Failed to send code', error.response?.data?.message || 'Unknown error');
        }
    });

    // Step 2: Verify Code
    const verifyMutation = useMutation({
        mutationFn: portalService.verifyGuardian,
        onSuccess: () => {
            toast.success('Verified', 'Guardian identity verified successfully');
            onVerified();
            handleClose();
        },
        onError: (err: unknown) => {
            const error = err as AxiosError<{ message: string }>;
            toast.error('Verification Failed', error.response?.data?.message || 'Invalid code');
        }
    });

    const handleClose = () => {
        setStep('INIT');
        setOtp('');
        // Don't clear email if it was passed in props, but we are using state so it's fine.
        onClose();
    };

    const handleInitiate = (e: React.FormEvent) => {
        e.preventDefault();
        if (!email) return;
        initiateMutation.mutate(email);
    };

    const handleVerify = (e: React.FormEvent) => {
        e.preventDefault();
        if (otp.length !== 6) return;
        verifyMutation.mutate(otp);
    };

    return (
        <Modal
            open={isOpen}
            onClose={handleClose}
            title={step === 'INIT' ? "Guardian Verification Required" : "Enter Verification Code"}
        >
            <div style={{ padding: '28px' }}>
                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', marginBottom: '24px' }}>
                    <div className="bg-blue-100" style={{ padding: '12px', borderRadius: '50%', marginBottom: '12px' }}>
                        <ShieldCheck className="text-blue-600" style={{ width: '32px', height: '32px' }} />
                    </div>
                    <p style={{ textAlign: 'center', color: '#4b5563', fontSize: '14px' }}>
                        {step === 'INIT'
                            ? "Since you are under 18, we need to verify your guardian's identity before proceeding."
                            : `Please enter the 6-digit code sent to ${email}`}
                    </p>
                </div>

                {step === 'INIT' ? (
                    <form onSubmit={handleInitiate} style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                        <div>
                            <label style={{ display: 'block', fontSize: '14px', fontWeight: 500, color: '#374151', marginBottom: '4px' }}>Guardian Email Address</label>
                            <div style={{ position: 'relative' }}>
                                <Mail style={{ position: 'absolute', left: '12px', top: '10px', width: '20px', height: '20px', color: '#9ca3af' }} />
                                <input
                                    type="email"
                                    value={email}
                                    onChange={(e) => setEmail(e.target.value)}
                                    style={{ paddingLeft: '40px', width: '100%', padding: '8px 16px 8px 40px', border: '1px solid #d1d5db', borderRadius: '8px', outline: 'none', fontSize: '14px' }}
                                    placeholder="guardian@example.com"
                                    required
                                />
                            </div>
                        </div>
                        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '12px', paddingTop: '8px' }}>
                            <Button variant="outline" onClick={handleClose} type="button">Cancel</Button>
                            <Button
                                variant="primary"
                                type="submit"
                                isLoading={initiateMutation.isPending}
                                disabled={!email}
                            >
                                Send Verification Code
                            </Button>
                        </div>
                    </form>
                ) : (
                    <form onSubmit={handleVerify} style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
                        <div style={{ display: 'flex', justifyContent: 'center' }}>
                            <OTPInput
                                length={6}
                                value={otp}
                                onChange={setOtp}
                            />
                        </div>

                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '14px' }}>
                            <button
                                type="button"
                                onClick={() => setStep('INIT')}
                                style={{ color: '#6b7280', textDecoration: 'underline', background: 'none', border: 'none', cursor: 'pointer', fontSize: '14px' }}
                            >
                                Change Email
                            </button>
                            <button
                                type="button"
                                onClick={() => initiateMutation.mutate(email)}
                                style={{ color: '#2563eb', fontWeight: 500, background: 'none', border: 'none', cursor: 'pointer', fontSize: '14px' }}
                                disabled={initiateMutation.isPending}
                            >
                                Resend Code
                            </button>
                        </div>

                        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '12px', paddingTop: '8px' }}>
                            <Button variant="outline" onClick={handleClose} type="button">Cancel</Button>
                            <Button
                                variant="primary"
                                type="submit"
                                isLoading={verifyMutation.isPending}
                                disabled={otp.length !== 6}
                            >
                                Verify & Proceed
                            </Button>
                        </div>
                    </form>
                )}
            </div>
        </Modal>
    );
};
