import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { portalService } from '../../../services/portalService';
import { Button } from '@datalens/shared';
import { StatusBadge } from '@datalens/shared';
import { Modal } from '@datalens/shared';
import { toast } from 'react-toastify';
import { format } from 'date-fns';
import { MessageSquare } from 'lucide-react';
import type { Grievance } from '../../../types/grievance';

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
        <div className="p-6">
            <h1 className="text-2xl font-bold text-gray-900 mb-6">My Grievances</h1>

            {isLoading ? (
                <div>Loading...</div>
            ) : data?.items.length === 0 ? (
                <div className="text-center py-10 text-gray-500">
                    You haven't submitted any grievances.
                </div>
            ) : (
                <div className="space-y-4">
                    {data?.items.map((g: Grievance) => (
                        <div key={g.id} className="bg-white border rounded-lg p-4 shadow-sm">
                            <div className="flex justify-between items-start mb-2">
                                <div>
                                    <h3 className="font-semibold text-gray-900">{g.subject}</h3>
                                    <span className="text-xs text-gray-500">
                                        Submitted on {format(new Date(g.submitted_at), 'MMM d, yyyy')} • {g.category}
                                    </span>
                                </div>
                                <StatusBadge label={g.status} />
                            </div>

                            <p className="text-gray-700 text-sm mb-4 line-clamp-2">{g.description}</p>

                            {g.resolution && (
                                <div className="bg-green-50 p-3 rounded text-sm text-green-800 border border-green-200 mb-3">
                                    <strong>Resolution:</strong> {g.resolution}
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
                                <div className="text-xs text-blue-600 font-medium">
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
                <div className="space-y-4">
                    <p className="text-sm text-gray-600">How would you rate the resolution of your grievance?</p>
                    <div className="flex space-x-2">
                        {[1, 2, 3, 4, 5].map(r => (
                            <button
                                key={r}
                                onClick={() => setRating(r)}
                                className={`text-2xl ${rating >= r ? 'text-yellow-500' : 'text-gray-300'}`}
                            >
                                ★
                            </button>
                        ))}
                    </div>
                    <textarea
                        className="w-full border rounded p-2 text-sm"
                        rows={3}
                        placeholder="Optional comments..."
                        value={comment}
                        onChange={e => setComment(e.target.value)}
                    />
                    <div className="flex justify-end space-x-2">
                        <Button variant="secondary" onClick={() => setFeedbackModal(null)}>Cancel</Button>
                        <Button onClick={() => feedbackMutation.mutate()}>Submit Feedback</Button>
                    </div>
                </div>
            </Modal>
        </div>
    );
}
