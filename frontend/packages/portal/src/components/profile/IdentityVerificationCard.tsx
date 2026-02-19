import { ShieldCheck, ArrowRight } from 'lucide-react';
import type { PortalProfile } from '@/types/portal';
import { toast } from '@datalens/shared';

interface IdentityVerificationCardProps {
    profile: PortalProfile;
    onVerifyClick?: () => void;
}

export const IdentityVerificationCard = ({ profile, onVerifyClick }: IdentityVerificationCardProps) => {
    const isVerified = profile.verification_status === 'VERIFIED';

    const handleVerify = () => {
        if (onVerifyClick) {
            onVerifyClick();
        } else {
            toast.info('Redirecting to DigiLocker verification...');
        }
    };

    return (
        <div className="bg-white dark:bg-slate-900 rounded-2xl p-10 border border-gray-200/60 dark:border-gray-800 shadow-sm hover:shadow-md transition-all duration-300 relative overflow-hidden flex flex-col justify-between h-full">
            {/* Background decoration */}
            <div className="absolute top-0 right-0 w-80 h-80 bg-slate-50 dark:bg-slate-800/50 rounded-bl-[120px] -z-0 pointer-events-none"></div>

            <div className="relative z-10 w-full">
                <div className="flex justify-between items-start mb-8">
                    <div className="flex items-center gap-4">
                        <ShieldCheck className={isVerified ? "text-emerald-500 w-9 h-9" : "text-gray-400 w-9 h-9"} />
                        <h2 className="text-2xl font-bold text-gray-900 dark:text-white">Identity Verification</h2>
                    </div>
                    {isVerified ? (
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-100 text-emerald-600 dark:bg-emerald-900 dark:text-emerald-300 tracking-widest uppercase">
                            Verified
                        </span>
                    ) : (
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400 tracking-widest uppercase">
                            Unverified
                        </span>
                    )}
                </div>

                <p className="text-gray-500 dark:text-gray-400 mb-8 max-w-lg">
                    {isVerified
                        ? "Your identity has been fully verified. You have access to all sensitive data requests and advanced privacy controls."
                        : "Verify your identity to access sensitive data requests and advanced privacy controls. Your current level restricts some actions."
                    }
                </p>

                <div className="bg-slate-50 dark:bg-slate-800/40 rounded-xl p-10 mb-12 border border-slate-100 dark:border-slate-700/50">
                    <div className="flex items-center justify-between mb-4">
                        <span className="text-xs font-bold text-slate-400 uppercase tracking-wider">Current Level</span>
                    </div>
                    <div className="text-3xl font-bold text-gray-900 dark:text-white">
                        {isVerified ? 'Verified Account' : 'Basic Account'}
                    </div>
                    <div className="mt-6 w-full bg-gray-200 dark:bg-gray-700 h-3 rounded-full overflow-hidden">
                        <div
                            className={`h-full rounded-full ${isVerified ? 'bg-emerald-500 w-full' : 'bg-[#1e3a8a] w-[30%]'}`}
                        ></div>
                    </div>
                    <div className="mt-4 text-sm text-gray-400 flex justify-between">
                        <span>Email Confirmed</span>
                        <span className={isVerified ? "text-emerald-600 font-medium" : "text-gray-300 dark:text-gray-600"}>
                            {isVerified ? 'ID Verified' : 'ID Pending'}
                        </span>
                    </div>
                </div>
            </div>

            <div className="mt-auto relative z-10 sm:max-w-md">
                {!isVerified ? (
                    <>
                        <button
                            onClick={handleVerify}
                            className="w-full bg-[#1e3a8a] hover:bg-[#172554] text-white py-3.5 px-4 rounded-lg font-medium shadow-sm transition-all duration-200 flex items-center justify-center gap-3 group"
                        >
                            <span className="font-bold tracking-tight">Verify with DigiLocker</span>
                            <ArrowRight className="text-white/70 group-hover:translate-x-1 transition-transform w-5 h-5" />
                        </button>
                        <div className="mt-4 text-center">
                            <button className="text-sm text-gray-500 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200 transition-colors bg-transparent border-none cursor-pointer">
                                Continue with Email (Restricted)
                            </button>
                        </div>
                    </>
                ) : (
                    <button className="w-full bg-emerald-50 text-emerald-700 py-3.5 px-4 rounded-lg font-medium border border-emerald-100 flex items-center justify-center gap-2 cursor-default">
                        <ShieldCheck size={18} />
                        Identity Secured
                    </button>
                )}
            </div>
        </div>
    );
};
