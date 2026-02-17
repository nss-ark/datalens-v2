import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { Shield, CheckCircle, ArrowRight } from 'lucide-react';
import { portalService } from '@/services/portalService';
import { toast } from '@datalens/shared';
import { cn } from '@datalens/shared';

export const IdentityCard = () => {
    const navigate = useNavigate();
    const { data: identityStatus, isLoading } = useQuery({
        queryKey: ['portal-identity'],
        queryFn: portalService.getIdentityStatus
    });

    const handleVerify = async () => {
        try {
            toast.info('Redirecting to DigiLocker...');
            // In production: window.location.href = `${API_BASE}/auth/digilocker/login?redirect_uri=...`
        } catch (error) {
            console.error('Failed to initiate DigiLocker verification:', error);
            toast.error('Failed to initiate verification');
        }
    };

    if (isLoading) {
        return (
            <div className="portal-card p-8 h-full">
                <div className="flex justify-between items-start mb-6">
                    <div className="space-y-2 flex-1">
                        <div className="skeleton h-6 w-48" />
                        <div className="skeleton h-4 w-64" />
                    </div>
                    <div className="skeleton h-7 w-24 rounded-full" />
                </div>
                <div className="skeleton h-28 w-full rounded-xl mb-6" />
                <div className="skeleton h-12 w-full rounded-xl" />
            </div>
        );
    }

    const level = identityStatus?.assurance_level || 'NONE';
    const profile = identityStatus?.profile;
    const isVerified = level === 'SUBSTANTIAL' || level === 'HIGH';

    return (
        <div className="portal-card p-6 h-full flex flex-col justify-between relative overflow-hidden group hover:border-slate-300/80 transition-all duration-300">
            {/* Subtle accent gradient */}
            {isVerified && (
                <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-emerald-400 to-emerald-500" />
            )}

            <div>
                <div className="flex justify-between items-start mb-4">
                    <h3 className="text-lg font-bold text-slate-900 flex items-center gap-2.5">
                        <Shield className={cn("w-5 h-5", isVerified ? "text-emerald-600" : "text-slate-400")} />
                        Identity Verification
                    </h3>
                    {isVerified ? (
                        <span className="bg-emerald-50 text-emerald-700 px-2.5 py-1 rounded-full text-[11px] font-bold tracking-wide flex items-center gap-1 ring-1 ring-emerald-200/50">
                            <CheckCircle size={10} strokeWidth={2.5} />
                            VERIFIED
                        </span>
                    ) : (
                        <span className="bg-slate-100 text-slate-500 px-2.5 py-1 rounded-full text-[11px] font-bold tracking-wide">
                            UNVERIFIED
                        </span>
                    )}
                </div>

                <p className="text-sm text-slate-500 leading-relaxed max-w-sm">
                    {isVerified
                        ? "Your identity has been verified via DigiLocker."
                        : "Verify your identity to access sensitive data requests."}
                </p>
            </div>

            <div className="mt-6">
                <div className="bg-slate-50 rounded-xl p-4 mb-4 border border-slate-100/80">
                    <div className="flex justify-between items-end">
                        <div>
                            <div className="text-[10px] uppercase tracking-wider text-slate-400 font-bold mb-1">
                                Current Level
                            </div>
                            <div className={cn("text-xl font-bold", isVerified ? "text-emerald-700" : "text-slate-800")}>
                                {level === 'NONE' ? 'Basic Account' : level}
                            </div>
                        </div>
                        {profile?.provider_name && (
                            <div className="text-right">
                                <div className="text-[10px] text-slate-400 mb-1 uppercase tracking-wider font-bold">Source</div>
                                <div className="font-semibold text-slate-700 text-xs bg-white px-2 py-1 rounded border border-slate-200">{profile.provider_name}</div>
                            </div>
                        )}
                    </div>
                </div>

                {!isVerified ? (
                    <div className="space-y-3">
                        <button
                            onClick={handleVerify}
                            className="w-full bg-[#002f56] hover:bg-[#003d70] text-white py-3 rounded-xl font-semibold text-sm flex items-center justify-center gap-2.5 transition-all duration-200 shadow-sm hover:shadow-md active:scale-[0.98]"
                        >
                            <img
                                src="https://upload.wikimedia.org/wikipedia/commons/thumb/f/f4/DigiLocker.svg/1200px-DigiLocker.svg.png"
                                alt="DigiLocker"
                                className="h-4 brightness-0 invert"
                            />
                            Verify with DigiLocker
                        </button>
                        <button className="w-full bg-white border border-slate-200 text-slate-600 hover:bg-slate-50 hover:border-slate-300 py-2.5 rounded-xl font-semibold text-xs transition-all duration-200">
                            Continue with Email (Restricted)
                        </button>
                    </div>
                ) : (
                    <button
                        onClick={() => navigate('/profile')}
                        className="w-full flex items-center justify-center gap-2 bg-white border border-slate-200 text-slate-700 hover:bg-slate-50 hover:border-slate-300 py-3 rounded-xl text-sm font-semibold transition-all duration-200 group"
                    >
                        View Details
                        <ArrowRight className="w-4 h-4 transition-transform group-hover:translate-x-1" />
                    </button>
                )}
            </div>
        </div>
    );
};
