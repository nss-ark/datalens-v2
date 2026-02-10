import { useState } from 'react';
import { Modal } from '../common/Modal';
import { Button } from '../common/Button';
import { useCreateDSR } from '../../hooks/useDSR';
import { useToastStore } from '../../stores/toastStore';
import { Plus, Trash2 } from 'lucide-react';
import type { DSRRequestType, DSRPriority, CreateDSRInput } from '../../types/dsr';

interface CreateDSRModalProps {
    open: boolean;
    onClose: () => void;
}

const REQUEST_TYPES: { value: DSRRequestType; label: string }[] = [
    { value: 'ACCESS', label: 'Access — Right to access personal data' },
    { value: 'ERASURE', label: 'Erasure — Right to delete personal data' },
    { value: 'CORRECTION', label: 'Correction — Right to correct data' },
    { value: 'PORTABILITY', label: 'Portability — Right to export data' },
];

const PRIORITIES: { value: DSRPriority; label: string }[] = [
    { value: 'HIGH', label: 'High' },
    { value: 'MEDIUM', label: 'Medium' },
    { value: 'LOW', label: 'Low' },
];

export function CreateDSRModal({ open, onClose }: CreateDSRModalProps) {
    const { mutate, isPending } = useCreateDSR();
    const addToast = useToastStore((s) => s.addToast);

    const [requestType, setRequestType] = useState<DSRRequestType>('ACCESS');
    const [subjectName, setSubjectName] = useState('');
    const [subjectEmail, setSubjectEmail] = useState('');
    const [priority, setPriority] = useState<DSRPriority>('MEDIUM');
    const [identifiers, setIdentifiers] = useState<{ key: string; value: string }[]>([]);

    const resetForm = () => {
        setRequestType('ACCESS');
        setSubjectName('');
        setSubjectEmail('');
        setPriority('MEDIUM');
        setIdentifiers([]);
    };

    const addIdentifier = () => {
        setIdentifiers(prev => [...prev, { key: '', value: '' }]);
    };

    const removeIdentifier = (index: number) => {
        setIdentifiers(prev => prev.filter((_, i) => i !== index));
    };

    const updateIdentifier = (index: number, field: 'key' | 'value', val: string) => {
        setIdentifiers(prev => prev.map((item, i) => i === index ? { ...item, [field]: val } : item));
    };

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();

        const subjectIdentifiers: Record<string, string> = {};
        identifiers.forEach(({ key, value }) => {
            if (key.trim()) subjectIdentifiers[key.trim()] = value.trim();
        });

        const input: CreateDSRInput = {
            request_type: requestType,
            subject_name: subjectName.trim(),
            subject_email: subjectEmail.trim(),
            subject_identifiers: subjectIdentifiers,
            priority,
        };

        mutate(input, {
            onSuccess: () => {
                addToast({ title: 'DSR created successfully', variant: 'success' });
                resetForm();
                onClose();
            },
            onError: () => {
                addToast({ title: 'Failed to create DSR', variant: 'error' });
            },
        });
    };

    const inputStyle: React.CSSProperties = {
        width: '100%',
        padding: '8px 12px',
        border: '1px solid var(--border-primary)',
        borderRadius: '6px',
        fontSize: '0.875rem',
        backgroundColor: 'var(--bg-secondary)',
        color: 'var(--text-primary)',
    };

    const labelStyle: React.CSSProperties = {
        display: 'block',
        fontSize: '0.8125rem',
        fontWeight: 500,
        color: 'var(--text-secondary)',
        marginBottom: '4px',
    };

    return (
        <Modal
            open={open}
            onClose={onClose}
            title="New Data Subject Request"
            footer={
                <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
                    <Button variant="ghost" onClick={onClose}>Cancel</Button>
                    <Button onClick={handleSubmit} isLoading={isPending}>Create DSR</Button>
                </div>
            }
        >
            <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                {/* Request Type */}
                <div>
                    <label style={labelStyle}>Request Type *</label>
                    <select
                        value={requestType}
                        onChange={e => setRequestType(e.target.value as DSRRequestType)}
                        style={inputStyle}
                    >
                        {REQUEST_TYPES.map(t => (
                            <option key={t.value} value={t.value}>{t.label}</option>
                        ))}
                    </select>
                </div>

                {/* Subject Name */}
                <div>
                    <label style={labelStyle}>Subject Name *</label>
                    <input
                        type="text"
                        value={subjectName}
                        onChange={e => setSubjectName(e.target.value)}
                        placeholder="e.g. John Doe"
                        required
                        style={inputStyle}
                    />
                </div>

                {/* Subject Email */}
                <div>
                    <label style={labelStyle}>Subject Email *</label>
                    <input
                        type="email"
                        value={subjectEmail}
                        onChange={e => setSubjectEmail(e.target.value)}
                        placeholder="e.g. john@example.com"
                        required
                        style={inputStyle}
                    />
                </div>

                {/* Priority */}
                <div>
                    <label style={labelStyle}>Priority</label>
                    <select
                        value={priority}
                        onChange={e => setPriority(e.target.value as DSRPriority)}
                        style={inputStyle}
                    >
                        {PRIORITIES.map(p => (
                            <option key={p.value} value={p.value}>{p.label}</option>
                        ))}
                    </select>
                </div>

                {/* Subject Identifiers */}
                <div>
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '8px' }}>
                        <label style={{ ...labelStyle, marginBottom: 0 }}>Subject Identifiers</label>
                        <button
                            type="button"
                            onClick={addIdentifier}
                            style={{
                                display: 'flex', alignItems: 'center', gap: '4px',
                                background: 'none', border: 'none', color: 'var(--accent-blue)',
                                fontSize: '0.8125rem', cursor: 'pointer', fontWeight: 500,
                            }}
                        >
                            <Plus size={14} /> Add
                        </button>
                    </div>
                    {identifiers.map((id, idx) => (
                        <div key={idx} style={{ display: 'flex', gap: '8px', marginBottom: '8px', alignItems: 'center' }}>
                            <input
                                type="text"
                                value={id.key}
                                onChange={e => updateIdentifier(idx, 'key', e.target.value)}
                                placeholder="Key (e.g. phone)"
                                style={{ ...inputStyle, flex: 1 }}
                            />
                            <input
                                type="text"
                                value={id.value}
                                onChange={e => updateIdentifier(idx, 'value', e.target.value)}
                                placeholder="Value (e.g. +1234567890)"
                                style={{ ...inputStyle, flex: 1 }}
                            />
                            <button
                                type="button"
                                onClick={() => removeIdentifier(idx)}
                                style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--status-danger)', padding: '4px' }}
                            >
                                <Trash2 size={16} />
                            </button>
                        </div>
                    ))}
                    {identifiers.length === 0 && (
                        <p style={{ fontSize: '0.8125rem', color: 'var(--text-tertiary)', fontStyle: 'italic' }}>
                            No identifiers added. Click &quot;Add&quot; to include phone, user ID, etc.
                        </p>
                    )}
                </div>
            </form>
        </Modal>
    );
}
