import React, { useState } from 'react';
import { usePortalAuthStore } from '../../stores/portalAuthStore';
import { useNavigate } from 'react-router-dom';
import { portalService } from '../../services/portalService';
import { toast } from '@datalens/shared';
import { Loader2 } from 'lucide-react';

const PortalLogin = () => {
    const [step, setStep] = useState<'IDENTIFIER' | 'OTP'>('IDENTIFIER');
    const [identifier, setIdentifier] = useState('');
    const [otp, setOtp] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const navigate = useNavigate();
    const setAuth = usePortalAuthStore(state => state.setAuth);

    const isEmail = (input: string) => /\S+@\S+\.\S+/.test(input);

    const handleSendOtp = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!identifier.trim()) return;

        setIsLoading(true);
        try {
            const payload = isEmail(identifier) ? { email: identifier } : { phone: identifier };
            await portalService.requestOTP(payload);
            toast.success("Code Sent", `Verification code sent to ${identifier}`);
            setStep('OTP');
        } catch (error) {
            console.error('OTP Request failed:', error);
            toast.error("Error", "Failed to send OTP. Please try again.");
        } finally {
            setIsLoading(false);
        }
    };

    const handleVerifyOtp = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!otp.trim()) return;

        setIsLoading(true);
        try {
            const payload = {
                otp,
                ...(isEmail(identifier) ? { email: identifier } : { phone: identifier })
            };
            const response = await portalService.verifyOTP(payload);

            setAuth(response.token, response.profile);
            toast.success("Welcome", `Logged in as ${response.profile.email || response.profile.phone}`);
            navigate('/portal/dashboard');
        } catch (error) {
            console.error('Verification failed:', error);
            toast.error("Invalid Code", "Please check the verification code and try again.");
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="max-w-md mx-auto bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
            <div className="p-8">
                <h2 className="text-2xl font-bold text-gray-900 mb-2">
                    {step === 'IDENTIFIER' ? 'Welcome Back' : 'Verify Identity'}
                </h2>
                <p className="text-gray-500 mb-8">
                    {step === 'IDENTIFIER'
                        ? 'Enter your email or phone to access your privacy portal.'
                        : `Enter the code sent to ${identifier}`}
                </p>

                {step === 'IDENTIFIER' ? (
                    <form onSubmit={handleSendOtp} className="space-y-6">
                        <div>
                            <label htmlFor="identifier" className="block text-sm font-medium text-gray-700 mb-2">
                                Email Address or Phone
                            </label>
                            <input
                                type="text"
                                id="identifier"
                                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-all"
                                placeholder="name@example.com"
                                value={identifier}
                                onChange={(e) => setIdentifier(e.target.value)}
                                required
                                disabled={isLoading}
                            />
                        </div>
                        <button
                            type="submit"
                            disabled={isLoading}
                            className="w-full bg-blue-600 text-white py-2.5 rounded-lg font-medium hover:bg-blue-700 transition-colors disabled:opacity-70 disabled:cursor-not-allowed flex justify-center items-center gap-2"
                        >
                            {isLoading && <Loader2 className="w-4 h-4 animate-spin" />}
                            {isLoading ? 'Sending...' : 'Send Verification Code'}
                        </button>
                    </form>
                ) : (
                    <form onSubmit={handleVerifyOtp} className="space-y-6">
                        <div>
                            <label htmlFor="otp" className="block text-sm font-medium text-gray-700 mb-2">
                                Verification Code
                            </label>
                            <input
                                type="text"
                                id="otp"
                                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-all tracking-widest text-center text-xl"
                                placeholder="0 0 0 0 0 0"
                                value={otp}
                                onChange={(e) => setOtp(e.target.value)}
                                required
                                disabled={isLoading}
                            />
                        </div>
                        <button
                            type="submit"
                            disabled={isLoading}
                            className="w-full bg-blue-600 text-white py-2.5 rounded-lg font-medium hover:bg-blue-700 transition-colors disabled:opacity-70 disabled:cursor-not-allowed flex justify-center items-center gap-2"
                        >
                            {isLoading && <Loader2 className="w-4 h-4 animate-spin" />}
                            {isLoading ? 'Verifying...' : 'Verify & Login'}
                        </button>
                        <button
                            type="button"
                            onClick={() => { setStep('IDENTIFIER'); setOtp(''); }}
                            className="w-full text-gray-500 text-sm hover:text-gray-700"
                            disabled={isLoading}
                        >
                            Back to Login
                        </button>
                    </form>
                )}
            </div>
        </div>
    );
};

export default PortalLogin;
