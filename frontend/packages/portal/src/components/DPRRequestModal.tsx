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
            <div style={{ padding: '28px' }}>
                <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
                    <div>
                        <label style={{ display: 'block', fontSize: '14px', fontWeight: 500, color: '#374151', marginBottom: '8px' }}>Request Type</label>
                        <select
                            value={type}
                            onChange={(e) => setType(e.target.value as CreateDPRInput['type'])}
                            style={{ width: '100%', padding: '8px 16px', border: '1px solid #d1d5db', borderRadius: '8px', outline: 'none', fontSize: '14px' }}
                        >
                            <option value="ACCESS">Access My Data (Download)</option>
                            <option value="CORRECTION">Correct My Data</option>
                            <option value="ERASURE">Erase My Data (Right to be Forgotten)</option>
                            <option value="NOMINATION">Nominate a Representative</option>
                            <option value="GRIEVANCE">File a Grievance</option>
                        </select>
                    </div>

                    <div>
                        <label style={{ display: 'block', fontSize: '14px', fontWeight: 500, color: '#374151', marginBottom: '8px' }}>Description / Details</label>
                        <textarea
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                            rows={4}
                            style={{ width: '100%', padding: '8px 16px', border: '1px solid #d1d5db', borderRadius: '8px', outline: 'none', fontSize: '14px', resize: 'none' }}
                            placeholder="Please provide details about your request..."
                            required
                        />
                    </div>

                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                        <input
                            type="checkbox"
                            id="minor"
                            checked={isMinor}
                            onChange={(e) => setIsMinor(e.target.checked)}
                            className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                        />
                        <label htmlFor="minor" style={{ fontSize: '14px', color: '#374151', userSelect: 'none' }}>
                            I am submitting this for a minor (under 18)
                        </label>
                    </div>

                    {isMinor && (
                        <div className="bg-blue-50 border border-blue-100" style={{ padding: '16px', borderRadius: '8px', display: 'flex', flexDirection: 'column', gap: '16px' }}>
                            <p style={{ fontSize: '14px', color: '#1d4ed8' }}>
                                Guardian verification is required for minors. We will send a verification code to the guardian.
                            </p>
                            <div>
                                <label style={{ display: 'block', fontSize: '14px', fontWeight: 500, color: '#374151', marginBottom: '4px' }}>Guardian Name</label>
                                <input
                                    type="text"
                                    value={guardianName}
                                    onChange={(e) => setGuardianName(e.target.value)}
                                    style={{ width: '100%', padding: '8px 12px', border: '1px solid #d1d5db', borderRadius: '6px', outline: 'none', fontSize: '14px' }}
                                    required={isMinor}
                                />
                            </div>
                            <div>
                                <label style={{ display: 'block', fontSize: '14px', fontWeight: 500, color: '#374151', marginBottom: '4px' }}>Guardian Email</label>
                                <input
                                    type="email"
                                    value={guardianEmail}
                                    onChange={(e) => setGuardianEmail(e.target.value)}
                                    style={{ width: '100%', padding: '8px 12px', border: '1px solid #d1d5db', borderRadius: '6px', outline: 'none', fontSize: '14px' }}
                                    required={isMinor}
                                />
                            </div>
                        </div>
                    )}

                    <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '12px', paddingTop: '16px' }}>
                        <Button variant="outline" onClick={onClose} type="button">Cancel</Button>
                        <Button variant="primary" type="submit" isLoading={isPending}>Submit Request</Button>
                    </div>
                </form>
            </div>
        </Modal>
    );
};
