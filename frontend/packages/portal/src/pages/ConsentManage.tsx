import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Shield, AlertTriangle } from 'lucide-react';
import { portalService } from '@/services/portalService';
import { Button } from '@datalens/shared';
import { Modal } from '@datalens/shared';
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

    const grantMutation = useMutation({
        mutationFn: portalService.grantConsent,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['portal-consents'] });
            toast.success('Consent granted successfully');
        },
        onError: () => toast.error('Failed to grant consent'),
    });

    return (
        <div className="animate-fade-in">
            <div className="page-header">
                <h1>Manage Your Consent</h1>
                <p>Review and control how your data is used. You can withdraw your consent at any time.</p>
            </div>

            {isLoading ? (
                <div className="space-y-4">
                    {[1, 2, 3].map(i => (
                        <div key={i} className="portal-card p-6 flex justify-between items-center">
                            <div className="space-y-2 flex-1">
                                <div className="skeleton h-5 w-48" />
                                <div className="skeleton h-4 w-32" />
                            </div>
                            <div className="skeleton h-9 w-36 rounded-lg" />
                        </div>
                    ))}
                </div>
            ) : consents.length === 0 ? (
                <div className="portal-card p-12 text-center">
                    <div className="w-14 h-14 bg-slate-100 rounded-2xl flex items-center justify-center mx-auto mb-5">
                        <Shield className="w-7 h-7 text-slate-400" />
                    </div>
                    <h3 className="text-lg font-semibold text-slate-900 mb-1.5">No active consents</h3>
                    <p className="text-sm text-slate-500 max-w-sm mx-auto leading-relaxed">You haven't granted any consents yet.</p>
                </div>
            ) : (
                <div className="space-y-4 stagger-children">
                    {consents.map((consent) => (
                        <div
                            key={consent.purpose_id}
                            className="portal-card p-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4"
                        >
                            <div>
                                <div className="flex items-center gap-2.5">
                                    <h3 className="text-base font-semibold text-slate-900">
                                        {consent.purpose_name}
                                    </h3>
                                    {consent.status === 'GRANTED' ? (
                                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-50 text-emerald-700 ring-1 ring-emerald-200">
                                            Active
                                        </span>
                                    ) : (
                                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-slate-100 text-slate-600 ring-1 ring-slate-200">
                                            Withdrawn
                                        </span>
                                    )}
                                </div>
                                <p className="text-sm text-slate-500 mt-1.5">
                                    Last updated: {format(new Date(consent.last_updated), 'MMM d, yyyy')}
                                </p>
                            </div>

                            <div className="flex-shrink-0">
                                {consent.status === 'GRANTED' ? (
                                    <Button
                                        variant="danger"
                                        size="sm"
                                        onClick={() => setWithdrawPurpose({ id: consent.purpose_id, name: consent.purpose_name })}
                                    >
                                        Withdraw Consent
                                    </Button>
                                ) : (
                                    <Button
                                        variant="primary"
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
                    <div className="flex items-start gap-3 p-4 bg-amber-50 rounded-xl border border-amber-100">
                        <AlertTriangle className="h-5 w-5 text-amber-600 mt-0.5 flex-shrink-0" />
                        <div className="text-sm text-amber-800">
                            Withdrawing consent for <strong>{withdrawPurpose?.name}</strong> may limit your experience.
                        </div>
                    </div>
                    <p className="text-sm text-slate-600 leading-relaxed">
                        By withdrawing consent, we will stop processing your data for this purpose immediately.
                        Are you sure you want to proceed?
                    </p>
                </div>
            </Modal>
        </div>
    );
}
