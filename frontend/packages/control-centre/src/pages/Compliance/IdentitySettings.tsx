import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Save, AlertTriangle, Shield } from 'lucide-react';
import { identityService } from '../../services/identity';
import { Button } from '@datalens/shared';
import { toast } from '@datalens/shared';
import { cn } from '@datalens/shared';

const IdentitySettings = () => {
    const queryClient = useQueryClient();
    const [isEditing, setIsEditing] = useState(false);

    // Fetch settings
    const { data: settings, isLoading, isError } = useQuery({
        queryKey: ['identity-settings'],
        queryFn: identityService.getSettings
    });

    // Local state for form
    const [formData, setFormData] = useState({
        enable_digilocker: false,
        require_govt_id_for_dsr: false,
        fallback_to_email_otp: true
    });

    useEffect(() => {
        if (settings) {
            // eslint-disable-next-line react-hooks/set-state-in-effect
            setFormData(prev => {
                if (
                    prev.enable_digilocker === settings.enable_digilocker &&
                    prev.require_govt_id_for_dsr === settings.require_govt_id_for_dsr &&
                    prev.fallback_to_email_otp === settings.fallback_to_email_otp
                ) {
                    return prev;
                }
                return {
                    enable_digilocker: settings.enable_digilocker,
                    require_govt_id_for_dsr: settings.require_govt_id_for_dsr,
                    fallback_to_email_otp: settings.fallback_to_email_otp
                };
            });
        }
    }, [settings]);

    // Update mutation
    const mutation = useMutation({
        mutationFn: identityService.updateSettings,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['identity-settings'] });
            toast.success('Identity settings updated successfully');
            setIsEditing(false);
        },
        onError: () => {
            toast.error('Failed to update settings');
        }
    });

    const handleSave = () => {
        mutation.mutate({
            ...settings, // keep other fields
            ...formData
        });
    };

    const handleCancel = () => {
        if (settings) {
            setFormData({
                enable_digilocker: settings.enable_digilocker,
                require_govt_id_for_dsr: settings.require_govt_id_for_dsr,
                fallback_to_email_otp: settings.fallback_to_email_otp
            });
        }
        setIsEditing(false);
    };

    if (isLoading) return <div className="p-8">Loading settings...</div>;
    if (isError) return <div className="p-8 text-red-500">Failed to load identity settings</div>;

    return (
        <div className="p-6 max-w-4xl mx-auto">
            <div className="flex items-center justify-between mb-8">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
                        <Shield className="w-8 h-8 text-blue-600" />
                        Identity Verification
                    </h1>
                    <p className="text-gray-500 mt-1">Configure how you verify Data Principals' identities.</p>
                </div>
                <div className="flex gap-3">
                    {isEditing ? (
                        <>
                            <Button variant="secondary" onClick={handleCancel}>Cancel</Button>
                            <Button
                                icon={<Save size={18} />}
                                onClick={handleSave}
                                disabled={mutation.isPending}
                            >
                                {mutation.isPending ? 'Saving...' : 'Save Changes'}
                            </Button>
                        </>
                    ) : (
                        <Button onClick={() => setIsEditing(true)}>Edit Settings</Button>
                    )}
                </div>
            </div>

            <div className="space-y-6">
                {/* Integration Status Card */}
                <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm">
                    <h2 className="text-lg font-semibold text-gray-900 mb-4">Identity Providers</h2>

                    <div className="flex items-center justify-between p-4 bg-gray-50 rounded-lg border border-gray-100">
                        <div className="flex items-center gap-4">
                            <div className="h-10 w-10 rounded-lg bg-blue-100 flex items-center justify-center">
                                <Shield className="h-5 w-5 text-blue-600" />
                            </div>
                            <div>
                                <h3 className="font-medium text-gray-900">DigiLocker (India)</h3>
                                <p className="text-sm text-gray-500">Verify identities using government-issued documents (Aadhaar, PAN).</p>
                            </div>
                        </div>
                        <div className="flex items-center gap-2">
                            <span className={cn(
                                "px-2.5 py-0.5 rounded-full text-xs font-medium",
                                formData.enable_digilocker ? "bg-green-100 text-green-700" : "bg-gray-100 text-gray-600"
                            )}>
                                {formData.enable_digilocker ? 'Enabled' : 'Disabled'}
                            </span>
                        </div>
                    </div>
                </div>

                {/* Configuration Form */}
                <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm">
                    <h2 className="text-lg font-semibold text-gray-900 mb-6">Verification Policy</h2>

                    <div className="space-y-6">
                        {/* Toggle 1: Enable DigiLocker */}
                        <div className="flex items-start justify-between">
                            <div>
                                <label className="text-base font-medium text-gray-900">Enable DigiLocker Verification</label>
                                <p className="text-sm text-gray-500 mt-1">Allow Data Principals to verify their identity using DigiLocker OAuth.</p>
                            </div>
                            <div className="flex items-center h-6">
                                <button
                                    type="button"
                                    disabled={!isEditing}
                                    onClick={() => setFormData({ ...formData, enable_digilocker: !formData.enable_digilocker })}
                                    className={cn(
                                        "relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed",
                                        formData.enable_digilocker ? "bg-blue-600" : "bg-gray-200"
                                    )}
                                >
                                    <span className={cn(
                                        "pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out",
                                        formData.enable_digilocker ? "translate-x-5" : "translate-x-0"
                                    )} />
                                </button>
                            </div>
                        </div>

                        <hr className="border-gray-100" />

                        {/* Toggle 2: Require Govt ID for DSR */}
                        <div className="flex items-start justify-between">
                            <div>
                                <label className="text-base font-medium text-gray-900">Require Government ID for DSRs</label>
                                <p className="text-sm text-gray-500 mt-1">
                                    Enforce Substantial Assurance (IAL2) for sensitive requests like Access or Erasure.
                                </p>
                            </div>
                            <div className="flex items-center h-6">
                                <button
                                    type="button"
                                    disabled={!isEditing}
                                    onClick={() => setFormData({ ...formData, require_govt_id_for_dsr: !formData.require_govt_id_for_dsr })}
                                    className={cn(
                                        "relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed",
                                        formData.require_govt_id_for_dsr ? "bg-blue-600" : "bg-gray-200"
                                    )}
                                >
                                    <span className={cn(
                                        "pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out",
                                        formData.require_govt_id_for_dsr ? "translate-x-5" : "translate-x-0"
                                    )} />
                                </button>
                            </div>
                        </div>

                        <hr className="border-gray-100" />

                        {/* Toggle 3: Fallback to Email */}
                        <div className="flex items-start justify-between">
                            <div>
                                <label className="text-base font-medium text-gray-900">Fallback to Email OTP</label>
                                <p className="text-sm text-gray-500 mt-1">
                                    Allow users to verify via Email OTP if DigiLocker is unavailable or fails.
                                </p>
                                {formData.fallback_to_email_otp && (
                                    <div className="mt-2 flex items-center gap-2 text-yellow-600 text-sm bg-yellow-50 px-3 py-1.5 rounded-md">
                                        <AlertTriangle size={14} />
                                        <span>Security Warning: Email OTP is Basic Assurance (IAL1) only.</span>
                                    </div>
                                )}
                            </div>
                            <div className="flex items-center h-6">
                                <button
                                    type="button"
                                    disabled={!isEditing}
                                    onClick={() => setFormData({ ...formData, fallback_to_email_otp: !formData.fallback_to_email_otp })}
                                    className={cn(
                                        "relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed",
                                        formData.fallback_to_email_otp ? "bg-blue-600" : "bg-gray-200"
                                    )}
                                >
                                    <span className={cn(
                                        "pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out",
                                        formData.fallback_to_email_otp ? "translate-x-5" : "translate-x-0"
                                    )} />
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default IdentitySettings;
