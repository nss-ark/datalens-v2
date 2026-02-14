import React, { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { portalService } from '@/services/portalService';
import { Modal } from '@datalens/shared';
import { Button } from '@datalens/shared';
import { toast } from '@datalens/shared';
import type { CreateDPRInput } from '@/types/portal';
import { AxiosError } from 'axios';

interface DPRRequestModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSuccess: () => void;
}

export const DPRRequestModal: React.FC<DPRRequestModalProps> = ({ isOpen, onClose, onSuccess }) => {
    const [type, setType] = useState<CreateDPRInput['type']>('ACCESS');
    const [description, setDescription] = useState('');
    const [isMinor, setIsMinor] = useState(false);
    const [guardianName, setGuardianName] = useState('');
    const [guardianEmail, setGuardianEmail] = useState('');

    const resetForm = () => {
        setType('ACCESS');
        setDescription('');
        setIsMinor(false);
        setGuardianName('');
        setGuardianEmail('');
    };

    const { mutate, isPending } = useMutation({
        mutationFn: portalService.createRequest,
        onSuccess: () => {
            toast.success('Request Submitted', 'Your request has been successfully submitted.');
            onSuccess();
            resetForm();
        },
        onError: (err: unknown) => {
            const error = err as AxiosError<{ message: string }>;
            const message = error.response?.data?.message || 'Failed to submit request';
            toast.error('Submission Failed', message);
        }
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        mutate({
            type,
            description,
            is_minor: isMinor,
            guardian_name: isMinor ? guardianName : undefined,
            guardian_email: isMinor ? guardianEmail : undefined
        });
    };

    return (
        <Modal
            open={isOpen}
            onClose={onClose}
            title="Submit Data Request"
        >
            <form onSubmit={handleSubmit} className="space-y-6">
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">Request Type</label>
                    <select
                        value={type}
                        onChange={(e) => setType(e.target.value as CreateDPRInput['type'])}
                        className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none"
                    >
                        <option value="ACCESS">Access My Data (Download)</option>
                        <option value="CORRECTION">Correct My Data</option>
                        <option value="ERASURE">Erase My Data (Right to be Forgotten)</option>
                        <option value="NOMINATION">Nominate a Representative</option>
                        <option value="GRIEVANCE">File a Grievance</option>
                    </select>
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">Description / Details</label>
                    <textarea
                        value={description}
                        onChange={(e) => setDescription(e.target.value)}
                        rows={4}
                        className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none"
                        placeholder="Please provide details about your request..."
                        required
                    />
                </div>

                <div className="flex items-center gap-2">
                    <input
                        type="checkbox"
                        id="minor"
                        checked={isMinor}
                        onChange={(e) => setIsMinor(e.target.checked)}
                        className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                    />
                    <label htmlFor="minor" className="text-sm text-gray-700 select-none">
                        I am submitting this for a minor (under 18)
                    </label>
                </div>

                {isMinor && (
                    <div className="bg-blue-50 p-4 rounded-lg space-y-4 border border-blue-100">
                        <p className="text-sm text-blue-700">
                            Guardian verification is required for minors. We will send a verification code to the guardian.
                        </p>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Guardian Name</label>
                            <input
                                type="text"
                                value={guardianName}
                                onChange={(e) => setGuardianName(e.target.value)}
                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 outline-none"
                                required={isMinor}
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Guardian Email</label>
                            <input
                                type="email"
                                value={guardianEmail}
                                onChange={(e) => setGuardianEmail(e.target.value)}
                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 outline-none"
                                required={isMinor}
                            />
                        </div>
                    </div>
                )}

                <div className="flex justify-end gap-3 pt-4">
                    <Button variant="outline" onClick={onClose} type="button">Cancel</Button>
                    <Button variant="primary" type="submit" isLoading={isPending}>Submit Request</Button>
                </div>
            </form>
        </Modal>
    );
};
