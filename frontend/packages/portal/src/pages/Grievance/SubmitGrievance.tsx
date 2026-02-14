import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import { portalService } from '@/services/portalService';
import { Button } from '@datalens/shared';
import { toast } from 'react-toastify';
import { ArrowLeft } from 'lucide-react';

export default function SubmitGrievance() {
    const navigate = useNavigate();
    const [subject, setSubject] = useState('');
    const [category, setCategory] = useState('CONSENT_WITHDRAWAL');
    const [description, setDescription] = useState('');

    const mutation = useMutation({
        mutationFn: portalService.submitGrievance,
        onSuccess: () => {
            toast.success('Grievance submitted successfully');
            navigate('/portal/history'); // Redirect to history/my-grievances
        },
        onError: () => toast.error('Failed to submit grievance')
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        mutation.mutate({ subject, category, description });
    };

    return (
        <div className="max-w-2xl mx-auto p-6">
            <Button variant="secondary" icon={<ArrowLeft size={16} />} onClick={() => navigate(-1)} className="mb-6">
                Back
            </Button>

            <h1 className="text-2xl font-bold text-gray-900 mb-2">Submit a Grievance</h1>
            <p className="text-gray-600 mb-8">
                If you have concerns about how your data is being processed, please submit a formal grievance below.
            </p>

            <form onSubmit={handleSubmit} className="space-y-6 bg-white p-6 rounded-lg shadow-sm border">
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Category</label>
                    <select
                        className="w-full border rounded-md p-2"
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
                    <label className="block text-sm font-medium text-gray-700 mb-1">Subject</label>
                    <input
                        type="text"
                        required
                        className="w-full border rounded-md p-2"
                        placeholder="Brief summary of your issue"
                        value={subject}
                        onChange={e => setSubject(e.target.value)}
                    />
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                    <textarea
                        required
                        rows={6}
                        className="w-full border rounded-md p-2"
                        placeholder="Please provide detailed information about your grievance..."
                        value={description}
                        onChange={e => setDescription(e.target.value)}
                    />
                </div>

                <div className="flex justify-end">
                    <Button type="submit" isLoading={mutation.isPending}>
                        Submit Grievance
                    </Button>
                </div>
            </form>
        </div>
    );
}
