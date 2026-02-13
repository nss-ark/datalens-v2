import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { grievanceService } from '../../services/grievanceService';
import { Button } from '../../components/common/Button';
import { StatusBadge } from '../../components/common/StatusBadge';
import { Modal } from '../../components/common/Modal';
import { toast } from 'react-toastify';
import { format } from 'date-fns';
import { ArrowLeft, CheckCircle, AlertTriangle } from 'lucide-react';

export default function GrievanceDetail() {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const queryClient = useQueryClient();
    const [resolveModalOpen, setResolveModalOpen] = useState(false);
    const [escalateModalOpen, setEscalateModalOpen] = useState(false);
    const [resolutionText, setResolutionText] = useState('');
    const [authorityText, setAuthorityText] = useState('');

    const { data: grievance, isLoading } = useQuery({
        queryKey: ['grievances', id],
        queryFn: () => grievanceService.getGrievance(id!)
    });

    const resolveMutation = useMutation({
        mutationFn: () => grievanceService.resolveGrievance(id!, resolutionText),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['grievances', id] });
            toast.success('Grievance resolved');
            setResolveModalOpen(false);
        }
    });

    const escalateMutation = useMutation({
        mutationFn: () => grievanceService.escalateGrievance(id!, authorityText),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['grievances', id] });
            toast.success('Grievance escalated');
            setEscalateModalOpen(false);
        }
    });

    if (isLoading) return <div className="p-6">Loading...</div>;
    if (!grievance) return <div className="p-6">Grievance not found</div>;

    return (
        <div className="p-6 max-w-4xl mx-auto">
            <Button variant="secondary" icon={<ArrowLeft size={16} />} onClick={() => navigate(-1)} className="mb-6">
                Back to List
            </Button>

            <div className="bg-white rounded-lg shadow-sm border p-6 mb-6">
                <div className="flex justify-between items-start mb-4">
                    <div>
                        <div className="flex items-center space-x-3 mb-2">
                            <h1 className="text-2xl font-bold text-gray-900">{grievance.subject}</h1>
                            <StatusBadge label={grievance.status} />
                            <span className="text-xs bg-gray-100 text-gray-600 px-2 py-1 rounded">
                                {grievance.priority} Priority
                            </span>
                        </div>
                        <p className="text-gray-500 text-sm">
                            Submitted by <strong>{grievance.data_subject_id}</strong> on {format(new Date(grievance.submitted_at), 'PPpp')}
                        </p>
                    </div>

                    <div className="flex space-x-2">
                        {grievance.status !== 'RESOLVED' && grievance.status !== 'CLOSED' && (
                            <>
                                <Button
                                    variant="danger"
                                    icon={<AlertTriangle size={16} />}
                                    onClick={() => setEscalateModalOpen(true)}
                                >
                                    Escalate
                                </Button>
                                <Button
                                    variant="primary"
                                    icon={<CheckCircle size={16} />}
                                    onClick={() => setResolveModalOpen(true)}
                                >
                                    Resolve
                                </Button>
                            </>
                        )}
                    </div>
                </div>

                <div className="prose max-w-none text-gray-800 bg-gray-50 p-4 rounded-md border">
                    {grievance.description}
                </div>
            </div>

            {grievance.resolution && (
                <div className="bg-green-50 rounded-lg border border-green-200 p-6 mb-6">
                    <h3 className="font-semibold text-green-800 mb-2">Resolution</h3>
                    <p className="text-green-900">{grievance.resolution}</p>
                    <div className="mt-2 text-xs text-green-700">
                        Resolved on {grievance.resolved_at ? format(new Date(grievance.resolved_at), 'PPpp') : '-'}
                    </div>
                </div>
            )}

            {grievance.feedback_rating && (
                <div className="bg-blue-50 rounded-lg border border-blue-200 p-6">
                    <h3 className="font-semibold text-blue-800 mb-2">User Feedback</h3>
                    <div className="flex items-center space-x-2 mb-2">
                        <span className="text-yellow-500 font-bold text-lg">{'★'.repeat(grievance.feedback_rating)}</span>
                        <span className="text-gray-400">{'★'.repeat(5 - grievance.feedback_rating)}</span>
                    </div>
                    {grievance.feedback_comment && (
                        <p className="text-blue-900 italic">"{grievance.feedback_comment}"</p>
                    )}
                </div>
            )}

            {/* Resolve Modal */}
            <Modal
                open={resolveModalOpen}
                onClose={() => setResolveModalOpen(false)}
                title="Resolve Grievance"
            >
                <div className="space-y-4">
                    <p className="text-sm text-gray-600">Provide a resolution message for the data principal.</p>
                    <textarea
                        className="w-full border rounded p-2 text-sm"
                        rows={4}
                        placeholder="Resolution details..."
                        value={resolutionText}
                        onChange={e => setResolutionText(e.target.value)}
                    />
                    <div className="flex justify-end space-x-2">
                        <Button variant="secondary" onClick={() => setResolveModalOpen(false)}>Cancel</Button>
                        <Button onClick={() => resolveMutation.mutate()} disabled={!resolutionText}>Submit Resolution</Button>
                    </div>
                </div>
            </Modal>

            {/* Escalate Modal */}
            <Modal
                open={escalateModalOpen}
                onClose={() => setEscalateModalOpen(false)}
                title="Escalate Grievance"
            >
                <div className="space-y-4">
                    <p className="text-sm text-gray-600">Escalate this grievance to an external authority or higher internal tier.</p>
                    <input
                        className="w-full border rounded p-2 text-sm"
                        placeholder="Authority Name (e.g. DPO, Legal, DPA)"
                        value={authorityText}
                        onChange={e => setAuthorityText(e.target.value)}
                    />
                    <div className="flex justify-end space-x-2">
                        <Button variant="secondary" onClick={() => setEscalateModalOpen(false)}>Cancel</Button>
                        <Button variant="danger" onClick={() => escalateMutation.mutate()} disabled={!authorityText}>Escalate</Button>
                    </div>
                </div>
            </Modal>
        </div>
    );
}
