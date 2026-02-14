import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMutation, useQuery } from '@tanstack/react-query';
import { portalService } from '@/services/portalService';
import { usePortalAuthStore } from '@/stores/portalAuthStore';
import { Button } from '@datalens/shared';
import { toast } from '@datalens/shared';
import type { CreateDPRInput } from '@/types/portal';
import { ArrowLeft, ShieldCheck } from 'lucide-react';
import { GuardianVerifyModal } from '@/components/GuardianVerifyModal';
import { AxiosError } from 'axios';

const RequestNew = () => {
    const navigate = useNavigate();
    const profile = usePortalAuthStore(state => state.profile);
    const { refetch: refetchProfile } = useQuery({
        queryKey: ['portal-profile'],
        queryFn: portalService.getProfile
    });

    const [type, setType] = useState<CreateDPRInput['type']>('ACCESS');
    const [description, setDescription] = useState('');
    const [isGuardianModalOpen, setGuardianModalOpen] = useState(false);

    // Initial check for minor
    const isMinor = profile?.is_minor;
    const isGuardianVerified = profile?.guardian_verified;

    const { mutate, isPending } = useMutation({
        mutationFn: portalService.createRequest,
        onSuccess: () => {
            toast.success('Request Submitted', 'Your request has been successfully submitted.');
            navigate('/portal/requests');
        },
        onError: (err: unknown) => {
            const error = err as AxiosError<{ message: string }>;
            const message = error.response?.data?.message || 'Failed to submit request';
            toast.error('Submission Failed', message);
        }
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();

        if (isMinor && !isGuardianVerified) {
            setGuardianModalOpen(true);
            return;
        }

        mutate({
            type,
            description,
            is_minor: isMinor,
            // Guardian details are now implicitly handled by the backend checking the profile verification status
            // or we could pass them if the backend expects them again, but verification is the key.
            // Based on previous modal, we sent them. But if profile is verified, we might not need to re-send.
            // Let's assume backend checks profile.guardian_verified.
        });
    };

    return (
        <div className="max-w-2xl mx-auto">
            <button
                onClick={() => navigate('/portal/requests')}
                className="flex items-center gap-2 text-gray-500 hover:text-gray-900 mb-6 transition-colors"
            >
                <ArrowLeft size={20} />
                Back to Requests
            </button>

            <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
                <div className="p-6 border-b border-gray-100">
                    <h1 className="text-xl font-bold text-gray-900">Submit New Request</h1>
                    <p className="text-gray-500 text-sm mt-1">Exercise your data rights under DPDPA.</p>
                </div>

                <div className="p-6">
                    {isMinor && !isGuardianVerified && (
                        <div className="mb-6 bg-orange-50 border border-orange-200 rounded-lg p-4 flex gap-3">
                            <ShieldCheck className="w-5 h-5 text-orange-600 flex-shrink-0 mt-0.5" />
                            <div>
                                <h3 className="font-medium text-orange-900">Guardian Verification Required</h3>
                                <p className="text-sm text-orange-700 mt-1">
                                    Since you are a minor, your guardian must verify their identity before you can submit this request.
                                </p>
                                <button
                                    onClick={() => setGuardianModalOpen(true)}
                                    className="text-sm font-medium text-orange-800 underline mt-2"
                                >
                                    Verify Guardian Now
                                </button>
                            </div>
                        </div>
                    )}

                    <form onSubmit={handleSubmit} className="space-y-6">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">Request Type</label>
                            <select
                                value={type}
                                onChange={(e) => setType(e.target.value as CreateDPRInput['type'])}
                                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none bg-white"
                            >
                                <option value="ACCESS">Access My Data (Download)</option>
                                <option value="CORRECTION">Correct My Data</option>
                                <option value="ERASURE">Erase My Data (Right to be Forgotten)</option>
                                <option value="NOMINATION">Nominate a Representative</option>
                                <option value="GRIEVANCE">File a Grievance</option>
                            </select>
                            <p className="text-xs text-gray-500 mt-2">
                                {type === 'ACCESS' && "Download a copy of all your personal data."}
                                {type === 'CORRECTION' && "Request corrections to inaccurate data."}
                                {type === 'ERASURE' && "Request deletion of your data (subject to retention policies)."}
                            </p>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">Description / Details</label>
                            <textarea
                                value={description}
                                onChange={(e) => setDescription(e.target.value)}
                                rows={5}
                                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none"
                                placeholder="Please provide specific details to help us fulfill your request..."
                                required
                            />
                        </div>

                        <div className="flex justify-end gap-3 pt-4 border-t border-gray-100">
                            <Button variant="outline" onClick={() => navigate('/portal/requests')} type="button">Cancel</Button>
                            <Button
                                variant="primary"
                                type="submit"
                                isLoading={isPending}
                                disabled={isMinor && !isGuardianVerified}
                            >
                                Submit Request
                            </Button>
                        </div>
                    </form>
                </div>
            </div>

            <GuardianVerifyModal
                isOpen={isGuardianModalOpen}
                onClose={() => setGuardianModalOpen(false)}
                onVerified={() => {
                    refetchProfile();
                    setGuardianModalOpen(false);
                }}
                guardianEmail={profile?.guardian_email}
            />
        </div>
    );
};

export default RequestNew;
