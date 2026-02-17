import React from 'react';
import { Shield, Lock, Fingerprint, Globe } from 'lucide-react';

interface AuthLayoutProps {
    children: React.ReactNode;
    title?: string;
    subtitle?: string;
}

export const AuthLayout: React.FC<AuthLayoutProps> = ({ children, title, subtitle }) => {
    return (
        <div className="min-h-screen flex font-sans">
            {/* Left Side — Branding & Visual */}
            <div className="hidden lg:flex lg:w-[52%] relative overflow-hidden">
                {/* Animated gradient background */}
                <div className="absolute inset-0 bg-gradient-to-br from-slate-900 via-blue-950 to-slate-900" />
                <div className="absolute inset-0 opacity-30"
                    style={{
                        backgroundImage: `radial-gradient(circle at 20% 80%, rgba(59,130,246,0.3), transparent 50%),
                                         radial-gradient(circle at 80% 20%, rgba(99,102,241,0.2), transparent 50%),
                                         radial-gradient(circle at 50% 50%, rgba(37,99,235,0.15), transparent 70%)`
                    }}
                />
                {/* Subtle grid pattern */}
                <div className="absolute inset-0 opacity-[0.04]"
                    style={{
                        backgroundImage: `linear-gradient(rgba(255,255,255,.1) 1px, transparent 1px),
                                         linear-gradient(90deg, rgba(255,255,255,.1) 1px, transparent 1px)`,
                        backgroundSize: '64px 64px'
                    }}
                />

                <div className="relative z-10 w-full flex flex-col justify-between p-12 xl:p-16 text-white">
                    {/* Logo */}
                    <div className="flex items-center gap-3">
                        <div className="bg-white/10 p-2.5 rounded-xl backdrop-blur-sm border border-white/10">
                            <Shield className="w-7 h-7 text-white" />
                        </div>
                        <span className="text-2xl font-bold tracking-tight">DataLens</span>
                    </div>

                    {/* Hero Copy */}
                    <div className="space-y-8 max-w-lg">
                        <h1 className="text-5xl font-extrabold leading-[1.1] tracking-tight">
                            Your Privacy,<br />
                            <span className="bg-clip-text text-transparent bg-gradient-to-r from-blue-400 to-cyan-300">
                                Under Your Control.
                            </span>
                        </h1>
                        <p className="text-blue-200/80 text-lg leading-relaxed">
                            Manage your consent, exercise your data rights, and track how your
                            information is used—all from one secure dashboard.
                        </p>

                        {/* Trust Badges */}
                        <div className="flex flex-wrap items-center gap-4 pt-2">
                            <div className="flex items-center gap-2 bg-white/[0.06] border border-white/10 rounded-full px-4 py-2 text-sm text-blue-200/90 backdrop-blur-sm">
                                <Lock className="w-3.5 h-3.5" />
                                <span>End-to-End Encrypted</span>
                            </div>
                            <div className="flex items-center gap-2 bg-white/[0.06] border border-white/10 rounded-full px-4 py-2 text-sm text-blue-200/90 backdrop-blur-sm">
                                <Fingerprint className="w-3.5 h-3.5" />
                                <span>DPDPA Compliant</span>
                            </div>
                            <div className="flex items-center gap-2 bg-white/[0.06] border border-white/10 rounded-full px-4 py-2 text-sm text-blue-200/90 backdrop-blur-sm">
                                <Globe className="w-3.5 h-3.5" />
                                <span>ISO 27001</span>
                            </div>
                        </div>
                    </div>

                    {/* Footer */}
                    <div className="text-sm text-blue-300/60">
                        © {new Date().getFullYear()} ComplyArk. Secure Privacy Infrastructure.
                    </div>
                </div>
            </div>

            {/* Right Side — Form */}
            <div className="w-full lg:w-[48%] flex flex-col justify-center items-center px-6 py-12 sm:px-12 lg:px-16 xl:px-20 relative bg-white">
                {/* Subtle top gradient accent */}
                <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-blue-600 via-indigo-500 to-blue-600" />

                <div className="w-full max-w-[420px] animate-fade-in">
                    {/* Mobile Branding */}
                    <div className="lg:hidden flex justify-center mb-10">
                        <div className="flex items-center gap-2.5">
                            <div className="bg-blue-600 p-2 rounded-xl shadow-sm">
                                <Shield className="w-6 h-6 text-white" />
                            </div>
                            <span className="text-xl font-bold text-slate-900">DataLens</span>
                        </div>
                    </div>

                    <div className="text-center lg:text-left mb-2">
                        <h2 className="text-3xl font-bold tracking-tight text-slate-900">
                            {title || 'Welcome back'}
                        </h2>
                        <p className="mt-3 text-slate-500 leading-relaxed">
                            {subtitle || 'Please verify your identity to continue.'}
                        </p>
                    </div>

                    {children}

                    {/* Bottom helper text */}
                    <p className="mt-10 text-center text-xs text-slate-400 leading-relaxed">
                        By continuing, you agree to our{' '}
                        <a href="#" className="text-blue-600 hover:text-blue-700 transition-colors">Terms of Service</a>
                        {' '}and{' '}
                        <a href="#" className="text-blue-600 hover:text-blue-700 transition-colors">Privacy Policy</a>.
                    </p>
                </div>
            </div>
        </div>
    );
};
