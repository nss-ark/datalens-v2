import React, { useState } from 'react';
import { Modal, Button, toast } from '@datalens/shared';
import { Lock, CloudUpload, User, Users, Mail, CheckCircle2 } from 'lucide-react';

interface NominationModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSubmit: (data: any) => void;
}

export const NominationModal: React.FC<NominationModalProps> = ({ isOpen, onClose, onSubmit }) => {
    const [nomineeName, setNomineeName] = useState('');
    const [relationship, setRelationship] = useState('');
    const [contact, setContact] = useState('');
    const [file, setFile] = useState<File | null>(null);

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (!nomineeName || !relationship || !contact) {
            toast.error('Please fill in all required fields');
            return;
        }
        onSubmit({ nomineeName, relationship, contact, file });
        // Reset state after successful submission
        setNomineeName('');
        setRelationship('');
        setContact('');
        setFile(null);
        onClose();
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
            title="Nomination Request"
        >
            <div style={{ padding: '28px', display: 'flex', flexDirection: 'column', gap: '28px' }}>
                <p className="text-sm text-slate-600 mb-4" style={{ marginTop: '-12px' }}>
                    Under DPDPA Section 14, you have the right to nominate a person who can exercise your data rights on your behalf in case of your death or incapacity.
                </p>
                {/* Secure Banner */}
                <div
                    className="bg-blue-50 border border-blue-100"
                    style={{ display: 'flex', alignItems: 'flex-start', gap: '16px', padding: '18px', borderRadius: '12px' }}
                >
                    <div style={{ marginTop: '2px' }}>
                        <Lock size={20} className="text-blue-600" />
                    </div>
                    <div>
                        <p style={{ fontSize: '15px', fontWeight: 600, color: '#1d4ed8', marginBottom: '4px' }}>Secure Nomination Process</p>
                        <p style={{ fontSize: '13px', color: '#64748b', lineHeight: 1.6 }}>
                            Your nominee details are encrypted. Proof of relationship is required for validation.
                        </p>
                    </div>
                </div>

                <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
                    {/* Nominee Name */}
                    <div>
                        <label style={{ display: 'block', fontSize: '14px', fontWeight: 600, color: '#1e293b', marginBottom: '10px' }}>
                            Nominee Name
                        </label>
                        <div style={{ position: 'relative' }}>
                            <div style={{ position: 'absolute', inset: '0', left: 0, paddingLeft: '14px', display: 'flex', alignItems: 'center', pointerEvents: 'none', color: '#94a3b8' }}>
                                <User size={20} />
                            </div>
                            <input
                                type="text"
                                style={{ display: 'block', width: '100%', paddingLeft: '44px', paddingTop: '14px', paddingBottom: '14px', paddingRight: '14px', borderRadius: '12px', border: '1px solid #cbd5e1', fontSize: '14px', boxShadow: '0 1px 2px 0 rgb(0 0 0 / 0.05)', outline: 'none' }}
                                placeholder="Full legal name"
                                value={nomineeName}
                                onChange={(e) => setNomineeName(e.target.value)}
                            />
                        </div>
                    </div>

                    {/* Row: Relationship & Contact */}
                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '20px' }}>
                        <div>
                            <label style={{ display: 'block', fontSize: '14px', fontWeight: 600, color: '#1e293b', marginBottom: '10px' }}>
                                Relationship
                            </label>
                            <div style={{ position: 'relative' }}>
                                <div style={{ position: 'absolute', inset: '0', left: 0, paddingLeft: '14px', display: 'flex', alignItems: 'center', pointerEvents: 'none', color: '#94a3b8' }}>
                                    <Users size={20} />
                                </div>
                                <select
                                    style={{ display: 'block', width: '100%', paddingLeft: '44px', paddingTop: '14px', paddingBottom: '14px', paddingRight: '14px', borderRadius: '12px', border: '1px solid #cbd5e1', fontSize: '14px', appearance: 'none', backgroundColor: 'white', boxShadow: '0 1px 2px 0 rgb(0 0 0 / 0.05)', outline: 'none' }}
                                    value={relationship}
                                    onChange={(e) => setRelationship(e.target.value)}
                                >
                                    <option value="" disabled>Select...</option>
                                    <option value="SPOUSE">Spouse</option>
                                    <option value="PARENT">Parent</option>
                                    <option value="CHILD">Child</option>
                                    <option value="SIBLING">Sibling</option>
                                    <option value="LEGAL_GUARDIAN">Legal Guardian</option>
                                    <option value="OTHER">Other</option>
                                </select>
                            </div>
                        </div>
                        <div>
                            <label style={{ display: 'block', fontSize: '14px', fontWeight: 600, color: '#1e293b', marginBottom: '10px' }}>
                                Contact (Email/Phone)
                            </label>
                            <div style={{ position: 'relative' }}>
                                <div style={{ position: 'absolute', inset: '0', left: 0, paddingLeft: '14px', display: 'flex', alignItems: 'center', pointerEvents: 'none', color: '#94a3b8' }}>
                                    <Mail size={20} />
                                </div>
                                <input
                                    type="text"
                                    style={{ display: 'block', width: '100%', paddingLeft: '44px', paddingTop: '14px', paddingBottom: '14px', paddingRight: '14px', borderRadius: '12px', border: '1px solid #cbd5e1', fontSize: '14px', boxShadow: '0 1px 2px 0 rgb(0 0 0 / 0.05)', outline: 'none' }}
                                    placeholder="Email or Phone"
                                    value={contact}
                                    onChange={(e) => setContact(e.target.value)}
                                />
                            </div>
                        </div>
                    </div>

                    {/* File Upload */}
                    <div>
                        <label style={{ display: 'block', fontSize: '14px', fontWeight: 600, color: '#1e293b', marginBottom: '10px' }}>
                            Upload Proof of Relationship
                        </label>
                        <div
                            style={{ display: 'flex', justifyContent: 'center', borderRadius: '12px', border: '2px dashed #cbd5e1', padding: '32px', cursor: 'pointer', backgroundColor: 'white', transition: 'all 0.2s' }}
                            onClick={() => document.getElementById('nominee-file-upload')?.click()}
                        >
                            <input
                                id="nominee-file-upload"
                                type="file"
                                className="hidden"
                                onChange={handleFileChange}
                                accept=".pdf,.jpg,.jpeg,.png"
                            />
                            <div style={{ textAlign: 'center' }}>
                                {file ? (
                                    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                                        <CheckCircle2 size={36} className="text-emerald-500" style={{ marginBottom: '12px' }} />
                                        <span style={{ fontSize: '15px', fontWeight: 500, color: '#0f172a', maxWidth: '240px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{file.name}</span>
                                        <p style={{ fontSize: '13px', color: '#64748b', marginTop: '8px' }}>Click to replace</p>
                                    </div>
                                ) : (
                                    <>
                                        <div style={{ width: '48px', height: '48px', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 12px', backgroundColor: '#f8fafc', borderRadius: '50%' }}>
                                            <CloudUpload size={28} className="text-slate-400" />
                                        </div>
                                        <div style={{ fontSize: '15px', color: '#334155' }}>
                                            <span style={{ fontWeight: 600, color: '#2563eb' }}>Click to upload</span> or drag and drop
                                        </div>
                                        <p style={{ fontSize: '13px', color: '#64748b', marginTop: '8px' }}>PDF, JPG or PNG (MAX. 5MB)</p>
                                    </>
                                )}
                            </div>
                        </div>
                    </div>

                    {/* Footer buttons */}
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
                            style={{ padding: '10px 20px', fontSize: '14px', fontWeight: 600, display: 'flex', alignItems: 'center', gap: '8px' }}
                            className="text-white bg-blue-600 hover:bg-blue-700 rounded-lg shadow-sm"
                        >
                            <CheckCircle2 size={18} />
                            Submit Nomination
                        </Button>
                    </div>
                </form>
            </div>
        </Modal>
    );
};
