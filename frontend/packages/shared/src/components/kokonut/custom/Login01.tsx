import React from 'react';
import { Mail, Smartphone, ArrowRight, Loader2, ShieldCheck, KeyRound } from 'lucide-react';


interface Login01Props {
    step: 'IDENTIFIER' | 'OTP';
    identifier: string;
    setIdentifier: (val: string) => void;
    otp: string;
    setOtp: (val: string) => void;
    isLoading: boolean;
    onSendOtp: (e: React.FormEvent) => void;
    onVerifyOtp: (e: React.FormEvent) => void;
    onBack?: () => void;
}

export const Login01 = ({
    step,
    identifier,
    setIdentifier,
    otp,
    setOtp,
    isLoading,
    onSendOtp,
    onVerifyOtp,
    onBack
}: Login01Props) => {
    const isEmail = (input: string) => /\S+@\S+\.\S+/.test(input);

    return (
        <div className="w-full space-y-10">
            {/* Heading */}
            <div className="text-center lg:text-left">
                <h2 className="text-3xl font-bold tracking-tight text-slate-900">
                    {step === 'IDENTIFIER' ? 'Welcome back' : 'Verify identity'}
                </h2>
                <p className="mt-2 text-slate-500 leading-relaxed">
                    {step === 'IDENTIFIER'
                        ? 'Enter your email or phone to access your privacy portal.'
                        : `Enter the 6-digit code sent to ${identifier}`
                    }
                </p>
            </div>

            {/* Identifier Step */}
            {step === 'IDENTIFIER' ? (
                <form onSubmit={onSendOtp} className="space-y-7">
                    <div className="space-y-2">
                        <label htmlFor="identifier" className="block text-sm font-semibold text-slate-700">
                            Email or Phone
                        </label>
                        <div className="relative group">
                            <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-400 group-focus-within:text-slate-600 transition-colors">
                                {isEmail(identifier) || !identifier ? <Mail className="h-[18px] w-[18px]" /> : <Smartphone className="h-[18px] w-[18px]" />}
                            </div>
                            <input
                                id="identifier"
                                type="text"
                                value={identifier}
                                onChange={(e) => setIdentifier(e.target.value)}
                                disabled={isLoading}
                                className="block w-full pl-11 pr-4 py-3.5 border border-slate-300 rounded-xl text-slate-900 text-[15px] placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-900/10 focus:border-slate-500 transition-all bg-white hover:border-slate-400"
                                placeholder="john.doe@example.com"
                                required
                            />
                        </div>
                    </div>

                    <button
                        type="submit"
                        disabled={isLoading || !identifier.trim()}
                        className="w-full flex justify-center items-center gap-2 py-3.5 px-4 rounded-xl text-sm font-semibold text-white bg-slate-900 hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 shadow-lg shadow-slate-900/10 active:scale-[0.98]"
                    >
                        {isLoading ? (
                            <Loader2 className="w-5 h-5 animate-spin" />
                        ) : (
                            <>
                                Continue <ArrowRight className="w-4 h-4" />
                            </>
                        )}
                    </button>
                </form>
            ) : (
                /* OTP Step */
                <form onSubmit={onVerifyOtp} className="space-y-7">
                    <div className="space-y-2">
                        <label htmlFor="otp" className="block text-sm font-semibold text-slate-700">
                            Verification Code
                        </label>
                        <div className="relative group">
                            <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-400 group-focus-within:text-slate-600 transition-colors">
                                <KeyRound className="h-[18px] w-[18px]" />
                            </div>
                            <input
                                id="otp"
                                type="text"
                                value={otp}
                                onChange={(e) => setOtp(e.target.value)}
                                disabled={isLoading}
                                className="block w-full pl-11 pr-4 py-3.5 border border-slate-300 rounded-xl text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-900/10 focus:border-slate-500 transition-all font-mono tracking-widest text-center text-lg bg-white hover:border-slate-400"
                                placeholder="000000"
                                maxLength={6}
                                required
                            />
                        </div>
                    </div>

                    <button
                        type="submit"
                        disabled={isLoading || otp.length < 6}
                        className="w-full flex justify-center items-center gap-2 py-3.5 px-4 rounded-xl text-sm font-semibold text-white bg-slate-900 hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 shadow-lg shadow-slate-900/10 active:scale-[0.98]"
                    >
                        {isLoading ? (
                            <Loader2 className="w-5 h-5 animate-spin" />
                        ) : (
                            <>
                                Verify & Login <ShieldCheck className="w-4 h-4" />
                            </>
                        )}
                    </button>

                    <button
                        type="button"
                        onClick={onBack}
                        className="w-full text-sm text-slate-500 hover:text-slate-900 transition-colors py-1"
                    >
                        ← Change {isEmail(identifier) ? 'email' : 'phone'}
                    </button>
                </form>
            )}

            {/* Terms */}
            <p className="text-center text-xs text-slate-400 leading-relaxed pt-2">
                By continuing, you agree to our{' '}
                <a href="#" className="text-slate-600 hover:text-slate-900 underline underline-offset-2 transition-colors">Terms of Service</a>
                {' '}and{' '}
                <a href="#" className="text-slate-600 hover:text-slate-900 underline underline-offset-2 transition-colors">Privacy Policy</a>.
            </p>
        </div>
    );
};
