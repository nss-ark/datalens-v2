import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import { portalService } from '@/services/portalService';
import { toast } from '@datalens/shared';
import { ArrowLeft, Send, AlertTriangle, ChevronDown, FileText, Info } from 'lucide-react';

const categories = [
    { value: 'CONSENT_WITHDRAWAL', label: 'Consent Withdrawal Issue', icon: '🔒' },
    { value: 'DSR_DELAY', label: 'DSR Request Delay', icon: '⏱️' },
    { value: 'DATA_BREACH', label: 'Data Breach Concern', icon: '🛡️' },
    { value: 'INCORRECT_DATA', label: 'Incorrect Data', icon: '📝' },
    { value: 'OTHER', label: 'Other', icon: '📋' },
];

export default function SubmitGrievance() {
    const navigate = useNavigate();
    const [subject, setSubject] = useState('');
    const [category, setCategory] = useState('CONSENT_WITHDRAWAL');
    const [description, setDescription] = useState('');

    const mutation = useMutation({
        mutationFn: portalService.submitGrievance,
        onSuccess: () => {
            toast.success('Grievance submitted successfully');
            navigate('/requests');
        },
        onError: () => toast.error('Failed to submit grievance')
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        mutation.mutate({ subject, category, description });
    };

    const selectedCat = categories.find(c => c.value === category);

    return (
        <div style={{ maxWidth: '720px', margin: '0 auto', padding: '40px 16px' }}>
            {/* Back Button */}
            <button
                onClick={() => navigate(-1)}
                style={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    gap: '8px',
                    fontSize: '14px',
                    color: '#64748b',
                    fontWeight: 500,
                    background: 'none',
                    border: 'none',
                    cursor: 'pointer',
                    marginBottom: '24px',
                    padding: '6px 12px 6px 8px',
                    borderRadius: '8px',
                    transition: 'all 0.15s',
                }}
                onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = '#f1f5f9'; e.currentTarget.style.color = '#0f172a'; }}
                onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent'; e.currentTarget.style.color = '#64748b'; }}
            >
                <ArrowLeft size={16} />
                Back
            </button>

            {/* Header Section */}
            <div style={{ marginBottom: '32px' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '14px', marginBottom: '12px' }}>
                    <div style={{
                        width: '48px',
                        height: '48px',
                        borderRadius: '14px',
                        background: 'linear-gradient(135deg, #fef3c7, #fde68a)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        border: '1px solid #fcd34d',
                    }}>
                        <AlertTriangle size={24} style={{ color: '#d97706' }} />
                    </div>
                    <div>
                        <h1 style={{ fontSize: '26px', fontWeight: 800, color: '#0f172a', margin: 0, letterSpacing: '-0.02em' }}>
                            Submit a Grievance
                        </h1>
                    </div>
                </div>
                <p style={{ fontSize: '15px', color: '#64748b', lineHeight: 1.7, margin: 0, maxWidth: '560px' }}>
                    If you have concerns about how your data is being processed, please submit a formal grievance below. All submissions are tracked and resolved under DPDPA regulations.
                </p>
            </div>

            {/* Form Card */}
            <form
                onSubmit={handleSubmit}
                style={{
                    background: 'white',
                    borderRadius: '20px',
                    border: '1px solid #e2e8f0',
                    boxShadow: '0 1px 3px rgba(0,0,0,0.04), 0 4px 12px rgba(0,0,0,0.02)',
                    overflow: 'hidden',
                }}
            >
                {/* Form Body */}
                <div style={{ padding: '32px', display: 'flex', flexDirection: 'column', gap: '28px' }}>
                    {/* Category */}
                    <div>
                        <label style={{ display: 'block', fontSize: '13px', fontWeight: 700, color: '#475569', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '10px' }}>
                            Category
                        </label>
                        <div style={{ position: 'relative' }}>
                            <select
                                value={category}
                                onChange={e => setCategory(e.target.value)}
                                style={{
                                    display: 'block',
                                    width: '100%',
                                    padding: '14px 44px 14px 48px',
                                    border: '1.5px solid #e2e8f0',
                                    borderRadius: '12px',
                                    fontSize: '15px',
                                    fontWeight: 500,
                                    color: '#0f172a',
                                    appearance: 'none',
                                    backgroundColor: '#f8fafc',
                                    cursor: 'pointer',
                                    outline: 'none',
                                    transition: 'border-color 0.2s',
                                }}
                                onFocus={(e) => e.currentTarget.style.borderColor = '#3b82f6'}
                                onBlur={(e) => e.currentTarget.style.borderColor = '#e2e8f0'}
                            >
                                {categories.map(cat => (
                                    <option key={cat.value} value={cat.value}>{cat.label}</option>
                                ))}
                            </select>
                            <div style={{ position: 'absolute', left: '16px', top: '50%', transform: 'translateY(-50%)', fontSize: '20px', pointerEvents: 'none' }}>
                                {selectedCat?.icon}
                            </div>
                            <div style={{ position: 'absolute', right: '16px', top: '50%', transform: 'translateY(-50%)', pointerEvents: 'none', color: '#94a3b8' }}>
                                <ChevronDown size={18} />
                            </div>
                        </div>
                    </div>

                    {/* Subject */}
                    <div>
                        <label style={{ display: 'block', fontSize: '13px', fontWeight: 700, color: '#475569', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '10px' }}>
                            Subject
                        </label>
                        <div style={{ position: 'relative' }}>
                            <div style={{ position: 'absolute', left: '16px', top: '50%', transform: 'translateY(-50%)', color: '#94a3b8', pointerEvents: 'none' }}>
                                <FileText size={18} />
                            </div>
                            <input
                                type="text"
                                required
                                placeholder="Brief summary of your issue"
                                value={subject}
                                onChange={e => setSubject(e.target.value)}
                                style={{
                                    display: 'block',
                                    width: '100%',
                                    padding: '14px 16px 14px 48px',
                                    border: '1.5px solid #e2e8f0',
                                    borderRadius: '12px',
                                    fontSize: '15px',
                                    color: '#0f172a',
                                    backgroundColor: 'white',
                                    outline: 'none',
                                    transition: 'border-color 0.2s',
                                }}
                                onFocus={(e) => e.currentTarget.style.borderColor = '#3b82f6'}
                                onBlur={(e) => e.currentTarget.style.borderColor = '#e2e8f0'}
                            />
                        </div>
                    </div>

                    {/* Description */}
                    <div>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '10px' }}>
                            <label style={{ fontSize: '13px', fontWeight: 700, color: '#475569', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                                Description
                            </label>
                            <span style={{ fontSize: '12px', color: '#94a3b8', fontWeight: 500 }}>
                                {description.length} / 2000
                            </span>
                        </div>
                        <textarea
                            required
                            rows={7}
                            maxLength={2000}
                            placeholder="Please provide detailed information about your grievance. Include dates, specific incidents, and any relevant context..."
                            value={description}
                            onChange={e => setDescription(e.target.value)}
                            style={{
                                display: 'block',
                                width: '100%',
                                padding: '16px',
                                border: '1.5px solid #e2e8f0',
                                borderRadius: '12px',
                                fontSize: '15px',
                                color: '#0f172a',
                                backgroundColor: 'white',
                                resize: 'none',
                                lineHeight: 1.7,
                                outline: 'none',
                                transition: 'border-color 0.2s',
                            }}
                            onFocus={(e) => e.currentTarget.style.borderColor = '#3b82f6'}
                            onBlur={(e) => e.currentTarget.style.borderColor = '#e2e8f0'}
                        />
                    </div>

                    {/* Info Box */}
                    <div style={{
                        display: 'flex',
                        gap: '14px',
                        padding: '16px 18px',
                        backgroundColor: '#eff6ff',
                        borderRadius: '12px',
                        border: '1px solid #bfdbfe',
                        alignItems: 'flex-start',
                    }}>
                        <Info size={18} style={{ color: '#3b82f6', flexShrink: 0, marginTop: '2px' }} />
                        <p style={{ fontSize: '13px', color: '#1e40af', lineHeight: 1.65, fontWeight: 500, margin: 0 }}>
                            Your grievance will be acknowledged within 48 hours and resolved within 30 days as per DPDPA Section 13 requirements. You'll receive email updates at each stage.
                        </p>
                    </div>
                </div>

                {/* Footer */}
                <div style={{
                    padding: '20px 32px',
                    borderTop: '1px solid #f1f5f9',
                    backgroundColor: '#fafbfc',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                }}>
                    <button
                        type="button"
                        onClick={() => navigate(-1)}
                        style={{
                            padding: '12px 24px',
                            fontSize: '14px',
                            fontWeight: 600,
                            color: '#64748b',
                            background: 'white',
                            border: '1.5px solid #e2e8f0',
                            borderRadius: '10px',
                            cursor: 'pointer',
                            transition: 'all 0.15s',
                        }}
                        onMouseEnter={(e) => { e.currentTarget.style.borderColor = '#cbd5e1'; e.currentTarget.style.color = '#334155'; }}
                        onMouseLeave={(e) => { e.currentTarget.style.borderColor = '#e2e8f0'; e.currentTarget.style.color = '#64748b'; }}
                    >
                        Cancel
                    </button>
                    <button
                        type="submit"
                        disabled={mutation.isPending || !subject.trim() || !description.trim()}
                        style={{
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: '10px',
                            padding: '12px 28px',
                            fontSize: '14px',
                            fontWeight: 700,
                            color: 'white',
                            background: (mutation.isPending || !subject.trim() || !description.trim())
                                ? '#94a3b8'
                                : 'linear-gradient(135deg, #3b82f6, #2563eb)',
                            backgroundColor: (mutation.isPending || !subject.trim() || !description.trim())
                                ? '#94a3b8'
                                : '#3b82f6',
                            border: 'none',
                            borderRadius: '10px',
                            cursor: (mutation.isPending || !subject.trim() || !description.trim()) ? 'not-allowed' : 'pointer',
                            boxShadow: (mutation.isPending || !subject.trim() || !description.trim())
                                ? 'none'
                                : '0 2px 8px rgba(59, 130, 246, 0.35)',
                            transition: 'all 0.2s',
                            letterSpacing: '0.01em',
                        }}
                    >
                        {mutation.isPending ? (
                            <div style={{
                                width: '18px', height: '18px',
                                border: '2px solid rgba(255,255,255,0.3)',
                                borderTopColor: 'white',
                                borderRadius: '50%',
                                animation: 'spin 0.6s linear infinite',
                            }} />
                        ) : (
                            <Send size={16} />
                        )}
                        Submit Grievance
                    </button>
                </div>
            </form>
        </div>
    );
}
