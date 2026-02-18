import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { portalService } from '@/services/portalService';
import { IdentityCard } from '@/components/IdentityCard';
import { GuardianVerifyModal } from '@/components/GuardianVerifyModal';
import { ShieldAlert } from 'lucide-react';
import { format } from 'date-fns';
import { StatusBadge, Card09 } from '@datalens/shared';

const PortalProfile = () => {
    const { data: profile, isLoading, refetch } = useQuery({
        queryKey: ['portal-profile'],
        queryFn: portalService.getProfile
    });

    const [isGuardianModalOpen, setGuardianModalOpen] = useState(false);

    if (isLoading) {
        return (
            <div className="space-y-8">
                <div className="page-header">
                    <div className="skeleton h-8 w-40 mb-2" />
                    <div className="skeleton h-4 w-72" />
                </div>
                <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
                    <div className="lg:col-span-4">
                        <div className="portal-card p-8 space-y-6">
                            <div className="flex flex-col items-center">
                                <div className="skeleton h-24 w-24 rounded-full mb-4" />
                                <div className="skeleton h-5 w-40 mb-2" />
                                <div className="skeleton h-4 w-24" />
                            </div>
                        </div>
                    </div>
                    <div className="lg:col-span-8">
                        <div className="portal-card p-8">
                            <div className="skeleton h-48 w-full rounded-xl" />
                        </div>
                    </div>
                </div>
            </div>
        );
    }

    if (!profile) return null;

    return (
        <div className="space-y-8 animate-fade-in">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div className="page-header !mb-0">
                    <h1>My Profile</h1>
                    <p>Manage your personal information and verification status.</p>
                </div>
                <StatusBadge label={profile.verification_status} />
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
                {/* Left — Personal Info */}
                <div className="lg:col-span-4 space-y-6">
                    <Card09
                        name={profile.email?.split('@')[0] || 'User'}
                        role="Data Principal"
                        stats={[
                            { label: "Email", value: profile.email || '-' },
                            { label: "Phone", value: profile.phone || '-' },
                            { label: "Joined", value: profile.created_at ? format(new Date(profile.created_at), 'MMM d, yyyy') : '-' }
                        ]}
                    />
                </div>

                {/* Right — Identity & Verification */}
                <div className="lg:col-span-8 space-y-6">
                    <IdentityCard />

                    {/* Guardian Verification (Minors only) */}
                    {profile.is_minor && (
                        <div className="portal-card p-8">
                            <h2 className="text-lg font-bold text-slate-900 mb-6 flex items-center gap-2.5">
                                <ShieldAlert className="w-5 h-5 text-orange-500" />
                                Guardian Verification
                            </h2>

                            {profile.guardian_verified ? (
                                <div className="bg-emerald-50 border border-emerald-200 rounded-xl p-5 flex gap-3">
                                    <div className="bg-emerald-100 p-2 rounded-lg h-fit">
                                        <ShieldAlert className="w-5 h-5 text-emerald-600" />
                                    </div>
                                    <div>
                                        <p className="text-emerald-800 font-medium">Guardian Verified</p>
                                        <p className="text-sm text-emerald-700 mt-1 leading-relaxed">
                                            Your guardian <strong>{profile.guardian_email}</strong> has verified their identity. You can now access all features.
                                        </p>
                                    </div>
                                </div>
                            ) : (
                                <div className="bg-orange-50 border border-orange-200 rounded-xl p-5">
                                    <div className="flex gap-3 mb-4">
                                        <div className="bg-orange-100 p-2 rounded-lg h-fit">
                                            <ShieldAlert className="w-5 h-5 text-orange-600" />
                                        </div>
                                        <div>
                                            <p className="text-orange-900 font-medium">Verification Required</p>
                                            <p className="text-sm text-orange-700 mt-1 leading-relaxed">
                                                As a minor, you need guardian approval to submit sensitive requests.
                                            </p>
                                        </div>
                                    </div>
                                    <button
                                        onClick={() => setGuardianModalOpen(true)}
                                        className="text-sm bg-white border border-orange-200 text-orange-700 px-5 py-2.5 rounded-xl hover:bg-orange-50 font-medium transition-all duration-200 shadow-sm"
                                    >
                                        Verify Guardian Details
                                    </button>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            </div>

            <GuardianVerifyModal
                isOpen={isGuardianModalOpen}
                onClose={() => setGuardianModalOpen(false)}
                onVerified={() => {
                    refetch();
                    setGuardianModalOpen(false);
                }}
                guardianEmail={profile.guardian_email}
            />
        </div>
    );
};

export default PortalProfile;
