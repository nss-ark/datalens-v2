import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { portalService } from '@/services/portalService';
import { GuardianVerifyModal } from '@/components/GuardianVerifyModal';
import { ProfileHeader } from '@/components/profile/ProfileHeader';
import { ProfileInfoCard } from '@/components/profile/ProfileInfoCard';
import { IdentityVerificationCard } from '@/components/profile/IdentityVerificationCard';
import { SecurityScoreCard } from '@/components/profile/SecurityScoreCard';
import { LastActivityCard } from '@/components/profile/LastActivityCard';
import { DataRightsCard } from '@/components/profile/DataRightsCard';

const PortalProfile = () => {
    const { data: profile, isLoading, refetch } = useQuery({
        queryKey: ['portal-profile'],
        queryFn: portalService.getProfile
    });

    const [isGuardianModalOpen, setGuardianModalOpen] = useState(false);

    if (isLoading) {
        return (
            <div className="space-y-8">
                <div className="flex flex-col md:flex-row md:items-end justify-between mb-8">
                    <div>
                        <div className="skeleton h-10 w-48 mb-3" />
                        <div className="skeleton h-5 w-96" />
                    </div>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                    <div className="md:col-span-1 h-[400px] skeleton rounded-2xl" />
                    <div className="md:col-span-2 h-[400px] skeleton rounded-2xl" />
                </div>
            </div>
        );
    }

    if (!profile) return null;

    return (
        <div className="animate-fade-in pb-12">
            <ProfileHeader verificationStatus={profile.verification_status} />

            <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                {/* Left Column - Profile Info */}
                <div className="md:col-span-1">
                    <ProfileInfoCard profile={profile} />
                </div>

                {/* Right Column - Identity & Stats */}
                <div className="md:col-span-2 flex flex-col gap-8">
                    {/* Identity Verification - Spans full width of this column */}
                    <div className="flex-1 min-h-[380px]">
                        <IdentityVerificationCard
                            profile={profile}
                            onVerifyClick={profile.is_minor ? () => setGuardianModalOpen(true) : undefined}
                        />
                    </div>

                    {/* Bottom Row - Stats Cards */}
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                        <SecurityScoreCard />
                        <LastActivityCard />
                        <DataRightsCard />
                    </div>
                </div>
            </div>

            {/* Guardian Verification Modal for Minors */}
            {profile.is_minor && (
                <GuardianVerifyModal
                    isOpen={isGuardianModalOpen}
                    onClose={() => setGuardianModalOpen(false)}
                    onVerified={() => {
                        refetch();
                        setGuardianModalOpen(false);
                    }}
                    guardianEmail={profile.guardian_email}
                />
            )}
        </div>
    );
};

export default PortalProfile;
