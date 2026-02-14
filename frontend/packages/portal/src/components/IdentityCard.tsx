import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { Shield, CheckCircle } from 'lucide-react';
import { portalService } from '../../services/portalService';
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
            // Initiate DigiLocker flow
            // Ideally this redirects to an OAuth URL or opens a popup
            // For now, we simulate the link request which returns the auth URL (in a real scenario)
            // But based on the contract: POST /link returns status.
            // Beacuse this is an OAuth flow, typically we redirect the browser.
            // Let's assume for now we just show a toast saying "Redirecting to DigiLocker..."

            toast.info('Redirecting to DigiLocker...');

            // In a real implementation:
            // window.location.href = `${API_BASE}/auth/digilocker/login?redirect_uri=${window.location.origin}/portal/verify`;

            // Mocking the behavior for the UI demo if no backend:
            // await portalService.linkIdentity('digilocker', 'mock_code');
        } catch (error) {
            console.error('Failed to initiate DigiLocker verification:', error);
            toast.error('Failed to initiate verification');
        }
    };

    if (isLoading) {
        return (
            <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm animate-pulse h-48">
                <div className="h-6 bg-gray-200 rounded w-1/3 mb-4"></div>
                <div className="h-12 bg-gray-100 rounded mb-4"></div>
            </div>
        );
    }

    const level = identityStatus?.assurance_level || 'NONE';
    const profile = identityStatus?.profile;

    const isVerified = level === 'SUBSTANTIAL' || level === 'HIGH';

    return (
        <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm relative overflow-hidden">
            <div className="flex justify-between items-start mb-4">
                <div>
                    <h3 className="text-lg font-bold text-gray-900 flex items-center gap-2">
                        <Shield className={cn("w-5 h-5", isVerified ? "text-green-600" : "text-gray-400")} />
                        Identity Verification
                    </h3>
                    <p className="text-sm text-gray-500 mt-1">
                        {isVerified
                            ? "Your identity has been verified via DigiLocker."
                            : "Verify your identity to access sensitive data requests."}
                    </p>
                </div>
                {isVerified ? (
                    <span className="bg-green-100 text-green-700 px-3 py-1 rounded-full text-xs font-bold flex items-center gap-1">
                        <CheckCircle size={12} />
                        VERIFIED
                    </span>
                ) : (
                    <span className="bg-gray-100 text-gray-600 px-3 py-1 rounded-full text-xs font-bold">
                        UNVERIFIED
                    </span>
                )}
            </div>

            <div className="bg-gray-50 rounded-lg p-4 mb-6">
                <div className="flex justify-between items-end">
                    <div>
                        <div className="text-xs uppercase tracking-wider text-gray-500 font-semibold mb-1">
                            Current Level
                        </div>
                        <div className={cn("text-2xl font-bold", isVerified ? "text-green-700" : "text-gray-700")}>
                            {level === 'NONE' ? 'Basic Account' : level}
                        </div>
                    </div>
                    {profile?.provider_name && (
                        <div className="text-right">
                            <div className="text-xs text-gray-400 mb-1">Source</div>
                            <div className="font-medium text-gray-700">{profile.provider_name}</div>
                        </div>
                    )}
                </div>
            </div>

            {!isVerified && (
                <div className="space-y-3">
                    <button
                        onClick={handleVerify}
                        className="w-full bg-[#002f56] hover:bg-[#003d70] text-white py-2.5 rounded-lg font-medium flex items-center justify-center gap-2 transition-all"
                    >
                        <img
                            src="https://upload.wikimedia.org/wikipedia/commons/thumb/f/f4/DigiLocker.svg/1200px-DigiLocker.svg.png"
                            alt="DigiLocker"
                            className="h-5 brightness-0 invert"
                        />
                        Verify with DigiLocker
                    </button>

                    <button className="w-full bg-white border border-gray-200 text-gray-600 hover:bg-gray-50 py-2.5 rounded-lg font-medium text-sm transition-all">
                        Continue with Email (Restricted)
                    </button>
                </div>
            )}

            {isVerified && (
                <div className="flex gap-2">
                    <button
                        onClick={() => navigate('/portal/profile')}
                        className="flex-1 bg-white border border-gray-200 text-gray-700 hover:bg-gray-50 py-2 rounded-lg text-sm font-medium"
                    >
                        View Details
                    </button>
                </div>
            )}
        </div>
    );
};
