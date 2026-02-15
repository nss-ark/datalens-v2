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
            <form onSubmit={handleSubmit} className="space-y-6">
                <div className="bg-blue-50 p-4 rounded-lg border border-blue-100 text-sm text-blue-800">
                    You are exercising your right to appeal under DPDPA Section 18.
                    Please provide detailed reasons why you believe the original decision should be reconsidered.
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">Reason for Appeal</label>
                    <textarea
                        value={reason}
                        onChange={(e) => setReason(e.target.value)}
                        rows={6}
                        className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none"
                        placeholder="Explain why the decision was incorrect or provide additional context..."
                        required
                    />
                    <p className="text-xs text-gray-500 mt-1 text-right">{reason.length} / 20 characters minimum</p>
                </div>

                <div className="flex justify-end gap-3 pt-4">
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
        </Modal>
    );
};
