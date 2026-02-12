import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Shield, AlertTriangle } from 'lucide-react';
import { portalService } from '../../services/portalService';
import { Button } from '../../components/common/Button';
import { Modal } from '../../components/common/Modal'; // Using reusable modal
import { toast } from 'react-toastify';
import { useState } from 'react';
import { format } from 'date-fns';

export default function ConsentManage() {
    const queryClient = useQueryClient();
    const [withdrawPurpose, setWithdrawPurpose] = useState<{ id: string; name: string } | null>(null);

    const { data: consents = [], isLoading } = useQuery({
        queryKey: ['portal-consents'],
        queryFn: portalService.getConsentSummary,
    });

    const withdrawMutation = useMutation({
        mutationFn: portalService.withdrawConsent,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['portal-consents'] });
            toast.success('Consent withdrawn successfully');
            setWithdrawPurpose(null);
        },
        onError: () => toast.error('Failed to withdraw consent'),
    });

    // We can also handle re-granting if the API supports it
    const grantMutation = useMutation({
        mutationFn: portalService.grantConsent,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['portal-consents'] });
            toast.success('Consent granted successfully');
        },
        onError: () => toast.error('Failed to grant consent'),
    });

    return (
        <div className="p-6 max-w-4xl mx-auto">
            <div className="mb-8">
                <h1 className="text-2xl font-bold text-gray-900 mb-2">Manage Your Consent</h1>
                <p className="text-gray-600">
                    Review and control how your data is used. You can withdraw your consent at any time.
                </p>
            </div>

            {isLoading ? (
                <div className="space-y-4">
                    {[1, 2, 3].map(i => (
                        <div key={i} className="h-24 bg-gray-100 rounded animate-pulse" />
                    ))}
                </div>
            ) : consents.length === 0 ? (
                <div className="text-center py-12 bg-white rounded-lg shadow-sm">
                    <Shield className="mx-auto h-12 w-12 text-gray-400" />
                    <h3 className="mt-2 text-sm font-medium text-gray-900">No active consents</h3>
                    <p className="mt-1 text-sm text-gray-500">You haven't granted any consents yet.</p>
                </div>
            ) : (
                <div className="space-y-4">
                    {consents.map((consent) => (
                        <div
                            key={consent.purpose_id}
                            className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4"
                        >
                            <div>
                                <div className="flex items-center space-x-2">
                                    <h3 className="text-lg font-medium text-gray-900">
                                        {consent.purpose_name}
                                    </h3>
                                    {consent.status === 'GRANTED' ? (
                                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                                            Active
                                        </span>
                                    ) : (
                                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                                            Withdrawn
                                        </span>
                                    )}
                                </div>
                                <p className="text-sm text-gray-500 mt-1">
                                    Last updated: {format(new Date(consent.last_updated), 'MMM d, yyyy')}
                                </p>
                            </div>

                            <div className="flex-shrink-0">
                                {consent.status === 'GRANTED' ? (
                                    <Button
                                        variant="danger" // Or secondary/outline danger
                                        size="sm"
                                        onClick={() => setWithdrawPurpose({ id: consent.purpose_id, name: consent.purpose_name })}
                                    >
                                        Withdraw Consent
                                    </Button>
                                ) : (
                                    <Button
                                        variant="primary" // Or secondary
                                        size="sm"
                                        onClick={() => grantMutation.mutate(consent.purpose_id)}
                                    >
                                        Grant Again
                                    </Button>
                                )}
                            </div>
                        </div>
                    ))}
                </div>
            )}

            {/* Withdrawal Confirmation Modal */}
            <Modal
                title="Confirm Withdrawal"
                open={!!withdrawPurpose}
                onClose={() => setWithdrawPurpose(null)}
                footer={
                    <div className="flex justify-end space-x-3 w-full">
                        <Button variant="secondary" onClick={() => setWithdrawPurpose(null)}>
                            Cancel
                        </Button>
                        <Button
                            variant="danger"
                            isLoading={withdrawMutation.isPending}
                            onClick={() => {
                                if (withdrawPurpose) {
                                    withdrawMutation.mutate(withdrawPurpose.id);
                                }
                            }}
                        >
                            Confirm Withdrawal
                        </Button>
                    </div>
                }
            >
                <div className="space-y-4">
                    <div className="flex items-start space-x-3 p-4 bg-amber-50 rounded-md">
                        <AlertTriangle className="h-5 w-5 text-amber-600 mt-0.5" />
                        <div className="text-sm text-amber-800">
                            Withdrawing consent for <strong>{withdrawPurpose?.name}</strong> may limit your experience.
                        </div>
                    </div>
                    <p className="text-sm text-gray-600">
                        By withdrawing consent, we will stop processing your data for this purpose immediately.
                        Are you sure you want to proceed?
                    </p>
                </div>
            </Modal>
        </div>
    );
}
