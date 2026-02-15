import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { portalService } from '@/services/portalService';
import { IdentityCard } from '@/components/IdentityCard';
import { GuardianVerifyModal } from '@/components/GuardianVerifyModal';
import { User, Mail, Phone, Calendar, ShieldAlert } from 'lucide-react';
import { format } from 'date-fns';
import { StatusBadge } from '@datalens/shared';

const PortalProfile = () => {
    const { data: profile, isLoading, refetch } = useQuery({
        queryKey: ['portal-profile'],
        queryFn: portalService.getProfile
    });

    const [isGuardianModalOpen, setGuardianModalOpen] = useState(false);

    if (isLoading) {
        return <div className="p-8 text-center text-gray-500">Loading profile...</div>;
    }

    if (!profile) return null;

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <h1 className="text-2xl font-bold text-gray-900">My Profile</h1>
                <StatusBadge label={profile.verification_status} />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* Personal Info */}
                <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm">
                    <h2 className="text-lg font-bold text-gray-900 mb-4 flex items-center gap-2">
                        <User className="w-5 h-5 text-gray-500" />
                        Personal Information
                    </h2>

                    <div className="space-y-4">
                        <div className="flex items-start gap-3">
                            <Mail className="w-5 h-5 text-gray-400 mt-0.5" />
                            <div>
                                <div className="text-sm text-gray-500">Email Address</div>
                                <div className="font-medium text-gray-900">{profile.email}</div>
                            </div>
                        </div>

                        {profile.phone && (
                            <div className="flex items-start gap-3">
                                <Phone className="w-5 h-5 text-gray-400 mt-0.5" />
                                <div>
                                    <div className="text-sm text-gray-500">Phone Number</div>
                                    <div className="font-medium text-gray-900">{profile.phone}</div>
                                </div>
                            </div>
                        )}

                        <div className="flex items-start gap-3">
                            <Calendar className="w-5 h-5 text-gray-400 mt-0.5" />
                            <div>
                                <div className="text-sm text-gray-500">Member Since</div>
                                <div className="font-medium text-gray-900">
                                    {profile.created_at ? format(new Date(profile.created_at), 'MMMM d, yyyy') : '-'}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Identity Verification */}
                <div className="space-y-6">
                    <IdentityCard />

                    {/* Guardian Verification Status (Only for Minors) */}
                    {profile.is_minor && (
                        <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm">
                            <h2 className="text-lg font-bold text-gray-900 mb-4 flex items-center gap-2">
                                <ShieldAlert className="w-5 h-5 text-orange-500" />
                                Guardian Verification
                            </h2>

                            {profile.guardian_verified ? (
                                <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                                    <p className="text-green-800 font-medium">Guardian Verified</p>
                                    <p className="text-sm text-green-600 mt-1">
                                        Your guardian ({profile.guardian_email}) has verified their identity.
                                    </p>
                                </div>
                            ) : (
                                <div className="bg-orange-50 border border-orange-200 rounded-lg p-4">
                                    <p className="text-orange-800 font-medium">Verification Required</p>
                                    <p className="text-sm text-orange-600 mt-1 mb-3">
                                        As a minor, you need guardian approval for certain actions.
                                    </p>
                                    <button
                                        onClick={() => setGuardianModalOpen(true)}
                                        className="text-sm bg-white border border-orange-200 text-orange-700 px-3 py-1.5 rounded-md hover:bg-orange-100 font-medium transition-colors"
                                    >
                                        Verify Guardian Now
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
