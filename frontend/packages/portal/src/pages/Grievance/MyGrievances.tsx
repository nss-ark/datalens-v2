import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { portalService } from '@/services/portalService';
import { Button } from '@datalens/shared';
import { StatusBadge } from '@datalens/shared';
import { Modal } from '@datalens/shared';
import { toast } from 'react-toastify';
import { format } from 'date-fns';
import { MessageSquare, FileText } from 'lucide-react';
import type { Grievance } from '@/types/grievance';

export default function MyGrievances() {
    const queryClient = useQueryClient();
    const [feedbackModal, setFeedbackModal] = useState<string | null>(null);
    const [rating, setRating] = useState(5);
    const [comment, setComment] = useState('');

    const { data, isLoading } = useQuery({
        queryKey: ['my-grievances'],
        queryFn: () => portalService.getGrievances()
    });

    const feedbackMutation = useMutation({
        mutationFn: () => portalService.submitGrievanceFeedback(feedbackModal!, rating, comment),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['my-grievances'] });
            toast.success('Feedback submitted');
            setFeedbackModal(null);
            setComment('');
            setRating(5);
        }
    });

    return (
        <div className="animate-fade-in">
            <div className="page-header">
                <h1>My Grievances</h1>
                <p>Track the status and resolution of your submitted grievances.</p>
            </div>

            {isLoading ? (
                <div className="space-y-4">
                    {[1, 2, 3].map(i => (
                        <div key={i} className="portal-card p-6">
                            <div className="flex justify-between items-start mb-3">
                                <div className="space-y-2 flex-1">
                                    <div className="skeleton h-5 w-56" />
                                    <div className="skeleton h-3 w-40" />
                                </div>
                                <div className="skeleton h-6 w-24 rounded-full" />
                            </div>
                            <div className="skeleton h-4 w-full" />
                        </div>
                    ))}
                </div>
            ) : data?.items.length === 0 ? (
                <div className="portal-card p-12 text-center">
                    <div className="w-14 h-14 bg-slate-100 rounded-2xl flex items-center justify-center mx-auto mb-5">
                        <FileText className="w-7 h-7 text-slate-400" />
                    </div>
                    <h3 className="text-lg font-semibold text-slate-900 mb-1.5">No grievances submitted</h3>
                    <p className="text-sm text-slate-500 max-w-sm mx-auto leading-relaxed">
                        You haven't submitted any grievances yet. If you have concerns about data processing, feel free to submit one.
                    </p>
                </div>
            ) : (
                <div className="space-y-4 stagger-children">
                    {data?.items.map((g: Grievance) => (
                        <div key={g.id} className="portal-card p-6">
                            <div className="flex justify-between items-start mb-3">
                                <div>
                                    <h3 className="font-semibold text-slate-900">{g.subject}</h3>
                                    <span className="text-xs text-slate-400 mt-1 flex items-center gap-1.5">
                                        Submitted on {format(new Date(g.submitted_at), 'MMM d, yyyy')}
                                        <span className="text-slate-200">•</span>
                                        <span className="bg-slate-50 px-2 py-0.5 rounded text-slate-500 font-medium border border-slate-100">{g.category}</span>
                                    </span>
                                </div>
                                <StatusBadge label={g.status} />
                            </div>

                            <p className="text-slate-500 text-sm mb-4 line-clamp-2 leading-relaxed">{g.description}</p>

                            {g.resolution && (
                                <div className="bg-emerald-50 p-4 rounded-xl text-sm text-emerald-800 border border-emerald-100 mb-4 flex gap-2.5">
                                    <div className="text-emerald-600 mt-0.5 flex-shrink-0">✓</div>
                                    <div>
                                        <span className="font-semibold">Resolution:</span> {g.resolution}
                                    </div>
                                </div>
                            )}

                            {g.status === 'RESOLVED' && !g.feedback_rating && (
                                <Button
                                    size="sm"
                                    variant="secondary"
                                    icon={<MessageSquare size={14} />}
                                    onClick={() => setFeedbackModal(g.id)}
                                >
                                    Provide Feedback
                                </Button>
                            )}

                            {g.feedback_rating && (
                                <div className="text-xs text-blue-600 font-medium bg-blue-50 px-3 py-1.5 rounded-full border border-blue-100 inline-flex items-center gap-1.5">
                                    ✓ Feedback submitted ({g.feedback_rating} stars)
                                </div>
                            )}
                        </div>
                    ))}
                </div>
            )}

            {/* Feedback Modal */}
            <Modal
                open={!!feedbackModal}
                onClose={() => setFeedbackModal(null)}
                title="Rate Resolution"
            >
                <div className="space-y-5">
                    <p className="text-sm text-slate-500">How would you rate the resolution of your grievance?</p>
                    <div className="flex gap-2">
                        {[1, 2, 3, 4, 5].map(r => (
                            <button
                                key={r}
                                onClick={() => setRating(r)}
                                className={`text-3xl transition-all duration-150 hover:scale-110 ${rating >= r ? 'text-amber-400' : 'text-slate-200 hover:text-amber-200'}`}
                            >
                                ★
                            </button>
                        ))}
                    </div>
                    <textarea
                        className="form-textarea"
                        rows={3}
                        placeholder="Optional comments..."
                        value={comment}
                        onChange={e => setComment(e.target.value)}
                    />
                    <div className="flex justify-end gap-3">
                        <Button variant="secondary" onClick={() => setFeedbackModal(null)}>Cancel</Button>
                        <Button onClick={() => feedbackMutation.mutate()}>Submit Feedback</Button>
                    </div>
                </div>
            </Modal>
        </div>
    );
}
