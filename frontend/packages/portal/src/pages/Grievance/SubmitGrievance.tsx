import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import { portalService } from '@/services/portalService';
import { toast } from 'react-toastify';
import { ArrowLeft, Send } from 'lucide-react';

export default function SubmitGrievance() {
    const navigate = useNavigate();
    const [subject, setSubject] = useState('');
    const [category, setCategory] = useState('CONSENT_WITHDRAWAL');
    const [description, setDescription] = useState('');

    const mutation = useMutation({
        mutationFn: portalService.submitGrievance,
        onSuccess: () => {
            toast.success('Grievance submitted successfully');
            navigate('/history');
        },
        onError: () => toast.error('Failed to submit grievance')
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        mutation.mutate({ subject, category, description });
    };

    return (
        <div className="max-w-2xl mx-auto animate-fade-in">
            <button
                onClick={() => navigate(-1)}
                className="flex items-center gap-2 text-sm text-slate-500 hover:text-slate-700 font-medium transition-colors mb-6"
            >
                <ArrowLeft size={16} />
                Back
            </button>

            <div className="page-header">
                <h1>Submit a Grievance</h1>
                <p>If you have concerns about how your data is being processed, please submit a formal grievance below.</p>
            </div>

            <form onSubmit={handleSubmit} className="portal-card p-8 space-y-6">
                <div>
                    <label className="form-label">Category</label>
                    <select
                        className="form-select"
                        value={category}
                        onChange={e => setCategory(e.target.value)}
                    >
                        <option value="CONSENT_WITHDRAWAL">Consent Withdrawal Issue</option>
                        <option value="DSR_DELAY">DSR Request Delay</option>
                        <option value="DATA_BREACH">Data Breach Concern</option>
                        <option value="INCORRECT_DATA">Incorrect Data</option>
                        <option value="OTHER">Other</option>
                    </select>
                </div>

                <div>
                    <label className="form-label">Subject</label>
                    <input
                        type="text"
                        required
                        className="form-input"
                        placeholder="Brief summary of your issue"
                        value={subject}
                        onChange={e => setSubject(e.target.value)}
                    />
                </div>

                <div>
                    <label className="form-label">Description</label>
                    <textarea
                        required
                        rows={6}
                        className="form-textarea"
                        placeholder="Please provide detailed information about your grievance..."
                        value={description}
                        onChange={e => setDescription(e.target.value)}
                    />
                </div>

                <div className="flex justify-end pt-2">
                    <button
                        type="submit"
                        disabled={mutation.isPending || !subject.trim() || !description.trim()}
                        className="flex items-center gap-2 px-6 py-3 bg-blue-600 text-white rounded-xl hover:bg-blue-700 transition-all duration-200 text-sm font-semibold shadow-sm hover:shadow-md active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {mutation.isPending ? (
                            <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                        ) : (
                            <Send size={15} />
                        )}
                        Submit Grievance
                    </button>
                </div>
            </form>
        </div>
    );
}
