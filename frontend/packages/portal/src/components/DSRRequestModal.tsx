import React, { useState } from 'react';
import { Modal, Button, toast } from '@datalens/shared';
import { Lock, FileText, CheckCircle2, UploadCloud } from 'lucide-react';

interface DSRRequestModalProps {
    isOpen: boolean;
    onClose: () => void;
    type: 'ACCESS' | 'CORRECTION' | 'ERASURE';
    onSubmit: (data: { description: string; file?: File | null }) => void;
    isLoading?: boolean;
}

export const DSRRequestModal: React.FC<DSRRequestModalProps> = ({
    isOpen, onClose, type, onSubmit, isLoading
}) => {
    const [description, setDescription] = useState('');
    const [file, setFile] = useState<File | null>(null);
    const [isDragging, setIsDragging] = useState(false);

    // Dynamic content based on request type
    const config = {
        ACCESS: {
            title: 'Right to Access',
            subtitle: 'Request a copy of your personal data.',
            placeholder: 'Please specify what data you would like to access (e.g., account history, profile details)...',
            showUpload: false
        },
        CORRECTION: {
            title: 'Right to Correction',
            subtitle: 'Update or correct inaccurate personal data.',
            placeholder: 'Describe the incorrect data and provide the correct information...',
            showUpload: true,
            uploadLabel: 'Upload Supporting Documents (Optional)'
        },
        ERASURE: {
            title: 'Right to Erasure',
            subtitle: 'Request deletion of your personal data.',
            placeholder: 'Please provide the reason for erasure (e.g., no longer necessary, withdrawal of consent)...',
            showUpload: false
        }
    }[type];

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (!description.trim()) {
            toast.error('Description is required');
            return;
        }
        onSubmit({ description, file });
        // Reset form on success handled by parent or here if preferred
        // We'll reset here for now assuming optimistic update
        setDescription('');
        setFile(null);
    };

    const handleDragOver = (e: React.DragEvent) => {
        e.preventDefault();
        setIsDragging(true);
    };

    const handleDragLeave = () => {
        setIsDragging(false);
    };

    const handleDrop = (e: React.DragEvent) => {
        e.preventDefault();
        setIsDragging(false);
        if (e.dataTransfer.files && e.dataTransfer.files[0]) {
            setFile(e.dataTransfer.files[0]);
        }
    };

    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.target.files && e.target.files[0]) {
            setFile(e.target.files[0]);
        }
    };

    return (
        <Modal
            open={isOpen}
            onClose={onClose}
            title={config.title}
        >
            <div style={{ padding: '28px', display: 'flex', flexDirection: 'column', gap: '28px' }}>
                {/* Badge Grid */}
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '24px' }}>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                        <label style={{ fontSize: '11px', fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                            Request Type
                        </label>
                        <div
                            className="bg-slate-100 border border-slate-200"
                            style={{ display: 'inline-flex', alignItems: 'center', gap: '10px', padding: '10px 16px', borderRadius: '8px', fontSize: '14px', fontWeight: 600, color: '#0f172a', cursor: 'not-allowed', width: 'fit-content' }}
                        >
                            {type === 'ACCESS' && <div className="text-indigo-500"><FileText size={20} /></div>}
                            {type === 'CORRECTION' && <div className="text-emerald-500"><CheckCircle2 size={20} /></div>}
                            {type === 'ERASURE' && <div className="text-rose-500"><Lock size={20} /></div>}
                            {config.title}
                        </div>
                    </div>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                        <label style={{ fontSize: '11px', fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                            Identity Verification
                        </label>
                        <div
                            className="bg-emerald-50 border border-emerald-100"
                            style={{ display: 'flex', alignItems: 'center', gap: '10px', padding: '10px 16px', color: '#15803d', borderRadius: '8px', width: 'fit-content' }}
                        >
                            <CheckCircle2 size={20} />
                            <span style={{ fontSize: '14px', fontWeight: 600 }}>Verified</span>
                        </div>
                    </div>
                </div>

                <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
                    {/* Description Field */}
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
                        <label style={{ fontSize: '14px', fontWeight: 600, color: '#0f172a' }}>
                            Additional Details <span style={{ color: '#94a3b8', fontWeight: 400 }}>(Optional)</span>
                        </label>
                        <p style={{ fontSize: '13px', color: '#64748b' }}>
                            {config.subtitle}
                        </p>
                        <textarea
                            rows={5}
                            style={{ display: 'block', width: '100%', borderRadius: '12px', border: '1px solid #cbd5e1', backgroundColor: 'white', color: '#0f172a', fontSize: '14px', resize: 'none', padding: '14px 16px', lineHeight: 1.6, boxShadow: '0 1px 2px 0 rgb(0 0 0 / 0.05)', outline: 'none' }}
                            placeholder={config.placeholder}
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                        />
                    </div>

                    {/* Optional File Upload for Correction */}
                    {config.showUpload && (
                        <div>
                            <label style={{ display: 'block', fontSize: '14px', fontWeight: 600, color: '#1e293b', marginBottom: '10px' }}>
                                {config.uploadLabel}
                            </label>
                            <div
                                style={{
                                    border: `2px dashed ${isDragging ? '#3b82f6' : '#cbd5e1'}`,
                                    borderRadius: '12px',
                                    padding: '32px',
                                    textAlign: 'center',
                                    cursor: 'pointer',
                                    backgroundColor: isDragging ? 'rgba(239, 246, 255, 0.5)' : 'white',
                                    transition: 'all 0.2s'
                                }}
                                onDragOver={handleDragOver}
                                onDragLeave={handleDragLeave}
                                onDrop={handleDrop}
                                onClick={() => document.getElementById('dsr-file-upload')?.click()}
                            >
                                <input
                                    id="dsr-file-upload"
                                    type="file"
                                    className="hidden"
                                    onChange={handleFileChange}
                                    accept=".pdf,.jpg,.png,.doc,.docx"
                                />
                                <div style={{ width: '56px', height: '56px', backgroundColor: 'white', borderRadius: '50%', boxShadow: '0 1px 3px 0 rgb(0 0 0 / 0.06)', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 20px', border: '1px solid #f1f5f9' }}>
                                    <UploadCloud size={28} className="text-blue-500" />
                                </div>
                                {file ? (
                                    <div className="bg-blue-50 border border-blue-100" style={{ display: 'inline-flex', alignItems: 'center', gap: '12px', padding: '8px 16px', borderRadius: '9999px', maxWidth: '100%' }}>
                                        <FileText size={18} className="text-blue-500" />
                                        <span style={{ fontSize: '14px', fontWeight: 500, color: '#1d4ed8', maxWidth: '200px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{file.name}</span>
                                        <button
                                            type="button"
                                            onClick={(e) => { e.stopPropagation(); setFile(null); }}
                                            style={{ color: '#93c5fd', cursor: 'pointer', background: 'none', border: 'none', fontSize: '18px', marginLeft: '8px' }}
                                        >
                                            ×
                                        </button>
                                    </div>
                                ) : (
                                    <>
                                        <p style={{ fontSize: '15px', fontWeight: 700, color: '#2563eb', marginBottom: '8px' }}>
                                            Click to upload <span style={{ color: '#64748b', fontWeight: 400 }}>or drop file</span>
                                        </p>
                                        <p style={{ fontSize: '13px', color: '#94a3b8', fontWeight: 500 }}>
                                            PDF, JPG, PNG (Max 5MB)
                                        </p>
                                    </>
                                )}
                            </div>
                        </div>
                    )}

                    {/* Blue Info Box */}
                    <div
                        className="bg-blue-50 border border-blue-100"
                        style={{ display: 'flex', gap: '14px', padding: '16px 18px', borderRadius: '12px', alignItems: 'flex-start' }}
                    >
                        <div className="text-blue-600" style={{ flexShrink: 0, marginTop: '2px' }}>
                            <FileText size={20} />
                        </div>
                        <p style={{ fontSize: '13px', color: '#1e40af', lineHeight: 1.6, fontWeight: 500 }}>
                            Processing time is typically 30 days. We will notify you via email once your data package is ready for secure download.
                        </p>
                    </div>

                    {/* Footer */}
                    <div style={{ paddingTop: '24px', borderTop: '1px solid #e2e8f0', display: 'flex', justifyContent: 'flex-end', gap: '12px' }}>
                        <Button
                            type="button"
                            variant="outline"
                            onClick={onClose}
                            style={{ padding: '10px 20px', fontSize: '14px', fontWeight: 600 }}
                            className="text-slate-700 bg-white border border-slate-300 hover:bg-slate-50"
                        >
                            Cancel
                        </Button>
                        <Button
                            type="submit"
                            variant="primary"
                            style={{ padding: '10px 20px', fontSize: '14px', fontWeight: 600 }}
                            className="text-white bg-blue-600 hover:bg-blue-700 shadow-sm"
                            isLoading={isLoading}
                        >
                            Submit Request
                        </Button>
                    </div>
                </form>
            </div>
        </Modal>
    );
};
