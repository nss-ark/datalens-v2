import React, { useState } from 'react';
import { usePortalAuthStore } from '../../stores/portalAuthStore';
import { useNavigate } from 'react-router-dom';

const PortalLogin = () => {
    const [step, setStep] = useState<'IDENTIFIER' | 'OTP'>('IDENTIFIER');
    const [identifier, setIdentifier] = useState('');
    const [otp, setOtp] = useState('');
    const navigate = useNavigate();
    const setAuth = usePortalAuthStore(state => state.setAuth);

    const handleSendOtp = (e: React.FormEvent) => {
        e.preventDefault();
        // TODO: Call API
        console.log('Sending OTP to', identifier);
        setStep('OTP');
    };

    const handleVerifyOtp = (e: React.FormEvent) => {
        e.preventDefault();
        // TODO: Call API
        console.log('Verifying OTP', otp);
        // Mock success
        setAuth('mock-portal-token', {
            id: 'p-123',
            tenant_id: 't-123',
            email: identifier,
            verification_status: 'VERIFIED',
            preferred_lang: 'en',
            is_minor: false,
            guardian_verified: false,
        });
        navigate('/portal/dashboard');
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
                            />
                        </div>
                        <button
                            type="submit"
                            className="w-full bg-blue-600 text-white py-2.5 rounded-lg font-medium hover:bg-blue-700 transition-colors"
                        >
                            Send Verification Code
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
                            />
                        </div>
                        <button
                            type="submit"
                            className="w-full bg-blue-600 text-white py-2.5 rounded-lg font-medium hover:bg-blue-700 transition-colors"
                        >
                            Verify & Login
                        </button>
                        <button
                            type="button"
                            onClick={() => setStep('IDENTIFIER')}
                            className="w-full text-gray-500 text-sm hover:text-gray-700"
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
