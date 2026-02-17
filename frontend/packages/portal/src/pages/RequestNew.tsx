import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import { portalService } from '@/services/portalService';
import { toast } from '@datalens/shared';
import { usePortalAuthStore } from '@/stores/portalAuthStore';
import { GuardianVerifyModal } from '@/components/GuardianVerifyModal';
import { ArrowLeft, ShieldAlert, Send } from 'lucide-react';

const PortalRequestNew = () => {
    const navigate = useNavigate();
    const profile = usePortalAuthStore(state => state.profile);

    const [type, setType] = useState<'ACCESS' | 'CORRECTION' | 'ERASURE' | 'NOMINATION' | 'GRIEVANCE'>('ACCESS');
    const [description, setDescription] = useState('');
    const [guardianOpen, setGuardianOpen] = useState(false);

    const mutation = useMutation({
        mutationFn: () => portalService.createRequest({ type, description }),
        onSuccess: () => {
            toast.success('Request submitted successfully');
            navigate('/requests');
        },
        onError: () => toast.error('Failed to submit request'),
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (profile?.is_minor && !profile?.guardian_verified) {
            setGuardianOpen(true);
            return;
        }
        mutation.mutate();
    };

    return (
        <div className="max-w-2xl mx-auto animate-fade-in">
            <button
                onClick={() => navigate(-1)}
                className="flex items-center gap-2 text-sm text-slate-500 hover:text-slate-700 font-medium transition-colors mb-6"
            >
                <ArrowLeft size={16} />
                Back to Requests
            </button>

            <div className="page-header">
                <h1>New Data Request</h1>
                <p>Exercise your data privacy rights by submitting a formal request.</p>
            </div>

            {/* Minor Warning Banner */}
            {profile?.is_minor && !profile?.guardian_verified && (
                <div className="bg-orange-50 border border-orange-200 rounded-xl p-5 flex gap-3 mb-6">
                    <div className="bg-orange-100 p-2 rounded-lg h-fit">
                        <ShieldAlert className="w-5 h-5 text-orange-600" />
                    </div>
                    <div>
                        <p className="text-orange-900 font-medium text-sm">Guardian Verification Required</p>
                        <p className="text-sm text-orange-700 mt-1 leading-relaxed">
                            As a minor, you need guardian approval to submit data requests.
                            You'll be prompted to verify your guardian before the request is submitted.
                        </p>
                    </div>
                </div>
            )}

            <form onSubmit={handleSubmit} className="portal-card p-8 space-y-6">
                <div>
                    <label htmlFor="requestType" className="form-label">
                        Request Type
                    </label>
                    <select
                        id="requestType"
                        className="form-select"
                        value={type}
                        onChange={e => setType(e.target.value as typeof type)}
                    >
                        <option value="ACCESS">Right to Access</option>
                        <option value="CORRECTION">Right to Correction</option>
                        <option value="ERASURE">Right to Erasure</option>
                        <option value="NOMINATION">Nominate Data Fiduciary</option>
                    </select>
                </div>

                <div>
                    <label htmlFor="requestDesc" className="form-label">
                        Description
                    </label>
                    <textarea
                        id="requestDesc"
                        className="form-textarea"
                        rows={5}
                        required
                        placeholder="Describe what data you would like to access, correct, or erase..."
                        value={description}
                        onChange={e => setDescription(e.target.value)}
                    />
                    <p className="text-xs text-slate-400 mt-2">
                        Provide as much detail as possible to help process your request efficiently.
                    </p>
                </div>

                <div className="flex justify-end pt-2">
                    <button
                        type="submit"
                        disabled={mutation.isPending || !description.trim()}
                        className="flex items-center gap-2 px-6 py-3 bg-blue-600 text-white rounded-xl hover:bg-blue-700 transition-all duration-200 text-sm font-semibold shadow-sm hover:shadow-md active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {mutation.isPending ? (
                            <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                        ) : (
                            <Send size={15} />
                        )}
                        Submit Request
                    </button>
                </div>
            </form>

            <GuardianVerifyModal
                isOpen={guardianOpen}
                onClose={() => setGuardianOpen(false)}
                onVerified={() => {
                    setGuardianOpen(false);
                    mutation.mutate();
                }}
                guardianEmail={profile?.guardian_email}
            />
        </div>
    );
};

export default PortalRequestNew;
