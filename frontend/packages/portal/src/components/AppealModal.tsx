import React, { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { portalService } from '@/services/portalService';
import { Modal, Button, toast } from '@datalens/shared';
import { AxiosError } from 'axios';

interface AppealModalProps {
    isOpen: boolean;
    onClose: () => void;
    dprId: string;
    onSuccess: () => void;
}

export const AppealModal: React.FC<AppealModalProps> = ({ isOpen, onClose, dprId, onSuccess }) => {
    const [reason, setReason] = useState('');
    const queryClient = useQueryClient();

    const { mutate, isPending } = useMutation({
        mutationFn: () => portalService.appealDPR(dprId, reason),
        onSuccess: () => {
            toast.success('Appeal Submitted', 'Your appeal has been received and will be reviewed by the DPO.');
            queryClient.invalidateQueries({ queryKey: ['portal-requests'] });
            onSuccess();
            setReason('');
            onClose();
        },
        onError: (err: unknown) => {
            const error = err as AxiosError<{ message: string }>;
            const message = error.response?.data?.message || 'Failed to submit appeal';
            toast.error('Appeal Failed', message);
        }
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (reason.trim().length < 20) {
            toast.error('Validation Error', 'Appeal reason must be at least 20 characters.');
            return;
        }
        mutate();
    };

    return (
        <Modal
            open={isOpen}
            onClose={onClose}
            title="Appeal Request Decision"
        >
            <div style={{ padding: '28px' }}>
                <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
                    <div className="bg-blue-50 border border-blue-100" style={{ padding: '20px', borderRadius: '8px', fontSize: '14px', color: '#1e40af', lineHeight: 1.6 }}>
                        You are exercising your right to appeal under DPDPA Section 18.
                        Please provide detailed reasons why you believe the original decision should be reconsidered.
                    </div>

                    <div>
                        <label style={{ display: 'block', fontSize: '14px', fontWeight: 600, color: '#374151', marginBottom: '8px' }}>Reason for Appeal</label>
                        <textarea
                            value={reason}
                            onChange={(e) => setReason(e.target.value)}
                            rows={6}
                            style={{ width: '100%', padding: '12px 16px', border: '1px solid #d1d5db', borderRadius: '8px', outline: 'none', resize: 'none', fontSize: '14px' }}
                            placeholder="Explain why the decision was incorrect or provide additional context..."
                            required
                        />
                        <p style={{ fontSize: '12px', color: '#6b7280', marginTop: '8px', textAlign: 'right' }}>{reason.length} / 20 characters minimum</p>
                    </div>

                    <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '12px', paddingTop: '24px', borderTop: '1px solid #f3f4f6' }}>
                        <Button variant="outline" onClick={onClose} type="button">Cancel</Button>
                        <Button
                            variant="primary"
                            type="submit"
                            isLoading={isPending}
                            disabled={reason.trim().length < 20}
                        >
                            Submit Appeal
                        </Button>
                    </div>
                </form>
            </div>
        </Modal>
    );
};
