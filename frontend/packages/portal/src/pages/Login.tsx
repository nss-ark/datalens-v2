import React, { useState } from 'react';
import { usePortalAuthStore } from '@/stores/portalAuthStore';
import { useNavigate } from 'react-router-dom';
import { portalService } from '@/services/portalService';
import { toast } from '@datalens/shared';
import { Loader2, Mail, ArrowRight, Smartphone, ShieldCheck, KeyRound } from 'lucide-react';

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
                code: otp,
                ...(isEmail(identifier) ? { email: identifier } : { phone: identifier })
            };
            const response = await portalService.verifyOTP(payload);

            setAuth(response.token.access_token, response.profile);
            toast.success("Welcome", `Logged in as ${response.profile.email || response.profile.phone}`);
            navigate('/dashboard');
        } catch (error) {
            console.error('Verification failed:', error);
            toast.error("Invalid Code", "Please check the verification code and try again.");
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="mt-8 space-y-6 animate-fade-in-up">
            {step === 'IDENTIFIER' ? (
                <form onSubmit={handleSendOtp} className="space-y-5">
                    <div className="space-y-1.5">
                        <label htmlFor="identifier" className="form-label">
                            Email address or Phone number
                        </label>
                        <div className="relative group">
                            <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 group-focus-within:text-blue-500 transition-colors">
                                {isEmail(identifier) || !identifier ? <Mail className="h-5 w-5" /> : <Smartphone className="h-5 w-5" />}
                            </div>
                            <input
                                id="identifier"
                                name="identifier"
                                type="text"
                                autoComplete="username"
                                required
                                value={identifier}
                                onChange={(e) => setIdentifier(e.target.value)}
                                disabled={isLoading}
                                className="form-input !pl-12 !py-3.5 text-base"
                                placeholder="name@example.com"
                            />
                        </div>
                    </div>

                    <button
                        type="submit"
                        disabled={isLoading || !identifier.trim()}
                        className="w-full flex justify-center items-center gap-2 py-3.5 px-4 rounded-xl text-sm font-semibold text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 shadow-sm hover:shadow-md active:scale-[0.98]"
                    >
                        {isLoading ? (
                            <Loader2 className="w-5 h-5 animate-spin" />
                        ) : (
                            <>
                                Continue <ArrowRight className="w-4 h-4" />
                            </>
                        )}
                    </button>

                    <div className="pt-2">
                        <div className="relative">
                            <div className="absolute inset-0 flex items-center">
                                <div className="w-full border-t border-slate-200" />
                            </div>
                            <div className="relative flex justify-center text-xs">
                                <span className="px-3 bg-white text-slate-400 font-medium">
                                    Secure Login via OTP
                                </span>
                            </div>
                        </div>
                    </div>
                </form>
            ) : (
                <form onSubmit={handleVerifyOtp} className="space-y-5 animate-fade-in-up">
                    {/* Sent-to indicator */}
                    <div className="bg-blue-50 border border-blue-100 rounded-xl p-4 flex items-center gap-3">
                        <div className="bg-blue-100 p-2 rounded-lg">
                            {isEmail(identifier) ? <Mail className="h-4 w-4 text-blue-600" /> : <Smartphone className="h-4 w-4 text-blue-600" />}
                        </div>
                        <div>
                            <p className="text-sm font-medium text-blue-900">Code sent to</p>
                            <p className="text-sm text-blue-700">{identifier}</p>
                        </div>
                    </div>

                    <div className="space-y-1.5">
                        <label htmlFor="otp" className="form-label">
                            Verification Code
                        </label>
                        <div className="relative group">
                            <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 group-focus-within:text-blue-500 transition-colors">
                                <KeyRound className="h-5 w-5" />
                            </div>
                            <input
                                id="otp"
                                name="otp"
                                type="text"
                                inputMode="numeric"
                                autoComplete="one-time-code"
                                required
                                value={otp}
                                onChange={(e) => setOtp(e.target.value.replace(/\D/g, '').slice(0, 6))}
                                disabled={isLoading}
                                className="form-input !pl-12 !py-3.5 tracking-[0.35em] font-mono text-center text-lg"
                                placeholder="• • • • • •"
                                maxLength={6}
                            />
                        </div>
                        <p className="text-xs text-slate-400 mt-1.5">
                            Enter the 6-digit code sent to your {isEmail(identifier) ? 'email' : 'phone'}
                        </p>
                    </div>

                    <button
                        type="submit"
                        disabled={isLoading || otp.length < 6}
                        className="w-full flex justify-center items-center gap-2 py-3.5 px-4 rounded-xl text-sm font-semibold text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 shadow-sm hover:shadow-md active:scale-[0.98]"
                    >
                        {isLoading ? (
                            <Loader2 className="w-5 h-5 animate-spin" />
                        ) : (
                            <>
                                <ShieldCheck className="w-4 h-4" />
                                Verify & Login
                            </>
                        )}
                    </button>

                    <button
                        type="button"
                        onClick={() => { setStep('IDENTIFIER'); setOtp(''); }}
                        className="w-full text-center text-sm text-slate-500 hover:text-blue-600 font-medium transition-colors py-1"
                        disabled={isLoading}
                    >
                        ← Change {isEmail(identifier) ? 'email address' : 'phone number'}
                    </button>
                </form>
            )}
        </div>
    );
};

export default PortalLogin;
