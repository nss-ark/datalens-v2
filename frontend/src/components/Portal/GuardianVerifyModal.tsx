import React, { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { portalService } from '../../services/portalService';
import { Modal } from '../common/Modal';
import { Button } from '../common/Button';
import { OTPInput } from '../common/OTPInput';
import { toast } from '../../stores/toastStore';
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
            <div className="flex flex-col items-center mb-6">
                <div className="bg-blue-100 p-3 rounded-full mb-3">
                    <ShieldCheck className="w-8 h-8 text-blue-600" />
                </div>
                <p className="text-center text-gray-600 text-sm">
                    {step === 'INIT'
                        ? "Since you are under 18, we need to verify your guardian's identity before proceeding."
                        : `Please enter the 6-digit code sent to ${email}`}
                </p>
            </div>

            {step === 'INIT' ? (
                <form onSubmit={handleInitiate} className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Guardian Email Address</label>
                        <div className="relative">
                            <Mail className="absolute left-3 top-2.5 h-5 w-5 text-gray-400" />
                            <input
                                type="email"
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                                className="pl-10 w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none"
                                placeholder="guardian@example.com"
                                required
                            />
                        </div>
                    </div>
                    <div className="flex justify-end gap-3 pt-2">
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
                <form onSubmit={handleVerify} className="space-y-6">
                    <div className="flex justify-center">
                        <OTPInput
                            length={6}
                            value={otp}
                            onChange={setOtp}
                        />
                    </div>

                    <div className="flex justify-between items-center text-sm">
                        <button
                            type="button"
                            onClick={() => setStep('INIT')}
                            className="text-gray-500 hover:text-gray-700 underline"
                        >
                            Change Email
                        </button>
                        <button
                            type="button"
                            onClick={() => initiateMutation.mutate(email)}
                            className="text-blue-600 hover:text-blue-800 font-medium"
                            disabled={initiateMutation.isPending}
                        >
                            Resend Code
                        </button>
                    </div>

                    <div className="flex justify-end gap-3 pt-2">
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
        </Modal>
    );
};
