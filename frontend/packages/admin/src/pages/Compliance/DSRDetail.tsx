import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { ArrowLeft, CheckCircle, XCircle, FileText } from 'lucide-react';
import { adminService } from '../../../services/adminService';
import { dsrService } from '../../../services/dsr'; // Using generic DSR service for status updates
import { Button } from '@datalens/shared';
import { StatusBadge } from '@datalens/shared';
import { toast } from 'react-toastify';

export default function AdminDSRDetail() {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const queryClient = useQueryClient();
    const [rejectModalOpen, setRejectModalOpen] = useState(false);
    const [rejectReason, setRejectReason] = useState('');

    const { data: dsr, isLoading, error } = useQuery({
        queryKey: ['admin-dsr', id],
        queryFn: () => adminService.getDSRById(id!),
        enabled: !!id,
    });

    const approveMutation = useMutation({
        mutationFn: () => dsrService.approve(id!),
        onSuccess: () => {
            toast.success('DSR Approved');
            queryClient.invalidateQueries({ queryKey: ['admin-dsr', id] });
        },
        onError: (err) => toast.error('Failed to approve: ' + (err as Error).message),
    });

    const rejectMutation = useMutation({
        mutationFn: (reason: string) => dsrService.reject(id!, reason),
        onSuccess: () => {
            toast.success('DSR Rejected');
            setRejectModalOpen(false);
            setRejectReason('');
            queryClient.invalidateQueries({ queryKey: ['admin-dsr', id] });
        },
        onError: (err) => toast.error('Failed to reject: ' + (err as Error).message),
    });

    if (isLoading) return <div className="p-8 text-center text-gray-500">Loading request details...</div>;
    if (error) return (
        <div className="p-8">
            <div className="bg-red-50 text-red-700 p-4 rounded mb-4">
                Error loading DSR: {(error as Error).message}
            </div>
            <Button variant="outline" onClick={() => navigate('/admin/compliance/dsr')}>
                Back to List
            </Button>
        </div>
    );

    if (!dsr) return <div className="p-8">DSR not found</div>;

    const isPending = dsr.status === 'PENDING';

    return (
        <div className="p-6 max-w-5xl mx-auto">
            <Button
                variant="ghost"
                className="mb-4 pl-0 hover:bg-transparent hover:text-blue-600"
                icon={<ArrowLeft size={16} />}
                onClick={() => navigate('/admin/compliance/dsr')}
            >
                Back to DSR Requests
            </Button>

            <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
                {/* Header */}
                <div className="p-6 border-b border-gray-100 flex justify-between items-start">
                    <div>
                        <div className="flex items-center gap-3 mb-2">
                            <h1 className="text-2xl font-bold text-gray-900">DSR-{dsr.id.substring(0, 8)}</h1>
                            <StatusBadge label={dsr.status} />
                            <span className="bg-gray-100 text-gray-600 px-2 py-0.5 rounded text-xs font-mono">
                                {dsr.request_type}
                            </span>
                        </div>
                        <p className="text-sm text-gray-500">
                            Created on {new Date(dsr.created_at).toLocaleString()} • Deadline: {new Date(dsr.sla_deadline).toLocaleDateString()}
                        </p>
                    </div>
                    {isPending && (
                        <div className="flex gap-2">
                            <button
                                className="inline-flex items-center px-3 py-2 border border-red-200 text-sm font-medium rounded-md text-red-700 bg-white hover:bg-red-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
                                onClick={() => setRejectModalOpen(true)}
                            >
                                <XCircle className="h-4 w-4 mr-2" />
                                Reject
                            </button>
                            <Button
                                className="bg-green-600 hover:bg-green-700 text-white border-transparent"
                                onClick={() => approveMutation.mutate()}
                                isLoading={approveMutation.isPending}
                                icon={<CheckCircle size={16} />}
                            >
                                Approve Request
                            </Button>
                        </div>
                    )}
                </div>

                <div className="grid grid-cols-1 md:grid-cols-3 divide-y md:divide-y-0 md:divide-x divide-gray-100">
                    {/* Subject Info */}
                    <div className="p-6">
                        <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-4">Data Subject</h3>
                        <div className="space-y-4">
                            <div>
                                <label className="text-xs text-gray-500 block">Name</label>
                                <div className="text-sm font-medium text-gray-900">{dsr.subject_name}</div>
                            </div>
                            <div>
                                <label className="text-xs text-gray-500 block">Email</label>
                                <div className="text-sm font-medium text-gray-900">{dsr.subject_email}</div>
                            </div>
                            <div>
                                <label className="text-xs text-gray-500 block">Identifiers</label>
                                <pre className="text-xs bg-gray-50 p-2 rounded mt-1 overflow-x-auto border border-gray-100 text-gray-800">
                                    {JSON.stringify(dsr.subject_identifiers, null, 2)}
                                </pre>
                            </div>
                        </div>
                    </div>

                    {/* Request Details */}
                    <div className="p-6">
                        <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-4">Context</h3>
                        <div className="space-y-4">
                            <div>
                                <label className="text-xs text-gray-500 block">Tenant</label>
                                <div className="text-sm font-medium text-gray-900">{dsr.tenant_name || 'Unknown'}</div>
                                <span className="text-gray-400 text-xs font-mono">{dsr.tenant_id}</span>
                            </div>
                            <div>
                                <label className="text-xs text-gray-500 block">Priority</label>
                                <div className="text-sm font-medium text-gray-900">{dsr.priority}</div>
                            </div>
                            <div>
                                <label className="text-xs text-gray-500 block">Notes</label>
                                <div className="text-sm text-gray-600 italic">{dsr.reason || 'No notes provided.'}</div>
                            </div>
                        </div>
                    </div>

                    {/* Admin Actions / Status */}
                    <div className="p-6 bg-gray-50">
                        <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-4">Admin Controls</h3>
                        <div className="space-y-4">
                            {!isPending && (
                                <div className="p-3 bg-white rounded border border-gray-200 text-sm shadow-sm">
                                    <div className="font-medium text-gray-900 mb-1">Workflow Status</div>
                                    <p className="text-gray-600">
                                        This request is currently <strong className="text-gray-900">{dsr.status}</strong>.
                                        {dsr.status === 'IN_PROGRESS' && " Automated tasks are running."}
                                    </p>
                                </div>
                            )}

                            <div className="pt-4 border-t border-gray-200 mt-4">
                                <h4 className="text-xs font-semibold text-gray-700 mb-2">Response File</h4>
                                <div className="flex flex-col items-center justify-center p-6 border-2 border-dashed border-gray-300 rounded-lg hover:border-gray-400 cursor-pointer bg-white transition-colors">
                                    <FileText className="h-8 w-8 text-gray-400 mb-2" />
                                    <span className="text-sm font-medium text-gray-900">
                                        Upload Response
                                    </span>
                                    <span className="text-xs text-gray-500 mt-1">PDF or ZIP up to 10MB</span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* Reject Modal */}
            {rejectModalOpen && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
                    <div className="bg-white rounded-lg p-6 max-w-md w-full shadow-xl">
                        <h3 className="text-lg font-bold mb-2 text-gray-900">Reject Request</h3>
                        <p className="text-sm text-gray-500 mb-4">
                            Please provide a reason for rejecting this request. This will be sent to the data subject.
                        </p>
                        <textarea
                            className="w-full border border-gray-300 rounded-md p-2 mb-4 h-24 focus:ring-red-500 focus:border-red-500 text-sm"
                            placeholder="Reason for rejection..."
                            value={rejectReason}
                            onChange={(e) => setRejectReason(e.target.value)}
                        />
                        <div className="flex justify-end gap-3">
                            <Button variant="ghost" onClick={() => setRejectModalOpen(false)}>Cancel</Button>
                            <Button
                                variant="primary"
                                className="bg-red-600 hover:bg-red-700 border-transparent text-white"
                                onClick={() => rejectMutation.mutate(rejectReason)}
                                disabled={!rejectReason.trim()}
                                isLoading={rejectMutation.isPending}
                            >
                                Confirm Rejection
                            </Button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
