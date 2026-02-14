import { useForm } from 'react-hook-form';
import { Button } from '@datalens/shared';
import type { CreateIncidentInput, IncidentStatus } from '../../types/breach';
import { useNavigate } from 'react-router-dom';

interface BreachFormProps {
    initialData?: Partial<CreateIncidentInput> & { status?: IncidentStatus };
    onSubmit: (data: CreateIncidentInput) => void;
    isLoading?: boolean;
    isEdit?: boolean;
}

export const BreachForm = ({ initialData, onSubmit, isLoading, isEdit }: BreachFormProps) => {
    const navigate = useNavigate();
    const { register, handleSubmit, formState: { errors } } = useForm<CreateIncidentInput & { status?: IncidentStatus }>({
        defaultValues: initialData || {
            title: '',
            description: '',
            type: 'Data Breach',
            severity: 'LOW',
            detected_at: new Date().toISOString().slice(0, 16), // datetime-local format
            occurred_at: new Date().toISOString().slice(0, 16),
            affected_systems: [],
            pii_categories: [],
            affected_data_subject_count: 0,
            poc_name: '',
            poc_role: '',
            poc_email: ''
        }
    });

    const onFormSubmit = (data: CreateIncidentInput & { status?: IncidentStatus }) => {
        // Transform datetime-local strings back to ISO if needed, or handle in service
        const payload = {
            ...data,
            detected_at: new Date(data.detected_at).toISOString(),
            occurred_at: new Date(data.occurred_at).toISOString(),
            affected_data_subject_count: Number(data.affected_data_subject_count),
            // simple array parsing from comma-separated string for MVP
            affected_systems: typeof data.affected_systems === 'string' ? (data.affected_systems as string).split(',').map((s: string) => s.trim()) : data.affected_systems as unknown as string[],
            pii_categories: typeof data.pii_categories === 'string' ? (data.pii_categories as string).split(',').map((s: string) => s.trim()) : data.pii_categories as unknown as string[],
        };
        onSubmit(payload);
    };

    return (
        <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-6 max-w-4xl">
            <div className="bg-white p-6 rounded-lg border shadow-sm space-y-4">
                <h3 className="text-lg font-semibold mb-4">Incident Details</h3>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Title</label>
                        <input
                            {...register('title', { required: 'Title is required' })}
                            className="w-full px-3 py-2 border rounded-md focus:ring-blue-500 focus:border-blue-500"
                            placeholder="e.g. Unauthorized Access to HR DB"
                        />
                        {errors.title && <span className="text-red-500 text-sm">{String(errors.title.message)}</span>}
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Type</label>
                        <select
                            {...register('type')}
                            className="w-full px-3 py-2 border rounded-md focus:ring-blue-500 focus:border-blue-500"
                        >
                            <option value="Data Breach">Data Breach</option>
                            <option value="Malware">Malware / Ransomware</option>
                            <option value="DoS">Denial of Service</option>
                            <option value="Identity Theft">Identity Theft</option>
                            <option value="Unlawful Disclosure">Unlawful Disclosure</option>
                        </select>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Severity</label>
                        <select
                            {...register('severity')}
                            className="w-full px-3 py-2 border rounded-md focus:ring-blue-500 focus:border-blue-500"
                        >
                            <option value="LOW">Low</option>
                            <option value="MEDIUM">Medium</option>
                            <option value="HIGH">High (CERT-In Reportable)</option>
                            <option value="CRITICAL">Critical (CERT-In Reportable)</option>
                        </select>
                    </div>

                    {isEdit && (
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Status</label>
                            <select
                                {...register('status')}
                                className="w-full px-3 py-2 border rounded-md focus:ring-blue-500 focus:border-blue-500"
                            >
                                <option value="OPEN">Open</option>
                                <option value="INVESTIGATING">Investigating</option>
                                <option value="CONTAINED">Contained</option>
                                <option value="RESOLVED">Resolved</option>
                                <option value="REPORTED">Reported</option>
                                <option value="CLOSED">Closed</option>
                            </select>
                        </div>
                    )}
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                    <textarea
                        {...register('description')}
                        rows={3}
                        className="w-full px-3 py-2 border rounded-md focus:ring-blue-500 focus:border-blue-500"
                        placeholder="Detailed description of the incident..."
                    />
                </div>
            </div>

            <div className="bg-white p-6 rounded-lg border shadow-sm space-y-4">
                <h3 className="text-lg font-semibold mb-4">Timeline & Impact</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Detected At</label>
                        <input
                            type="datetime-local"
                            {...register('detected_at', { required: true })}
                            className="w-full px-3 py-2 border rounded-md focus:ring-blue-500 focus:border-blue-500"
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Occurred At (Estimated)</label>
                        <input
                            type="datetime-local"
                            {...register('occurred_at')}
                            className="w-full px-3 py-2 border rounded-md focus:ring-blue-500 focus:border-blue-500"
                        />
                    </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Affected Systems (comma separated)</label>
                        <input
                            {...register('affected_systems')}
                            className="w-full px-3 py-2 border rounded-md focus:ring-blue-500 focus:border-blue-500"
                            placeholder="Server-01, DB-Prod, 192.168.1.5"
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Affected Data Subjects (Est.)</label>
                        <input
                            type="number"
                            {...register('affected_data_subject_count')}
                            className="w-full px-3 py-2 border rounded-md focus:ring-blue-500 focus:border-blue-500"
                        />
                    </div>
                </div>
            </div>

            <div className="bg-white p-6 rounded-lg border shadow-sm space-y-4">
                <h3 className="text-lg font-semibold mb-4">Point of Contact (PoC)</h3>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
                        <input
                            {...register('poc_name')}
                            className="w-full px-3 py-2 border rounded-md focus:ring-blue-500 focus:border-blue-500"
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Role</label>
                        <input
                            {...register('poc_role')}
                            className="w-full px-3 py-2 border rounded-md focus:ring-blue-500 focus:border-blue-500"
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
                        <input
                            type="email"
                            {...register('poc_email')}
                            className="w-full px-3 py-2 border rounded-md focus:ring-blue-500 focus:border-blue-500"
                        />
                    </div>
                </div>
            </div>

            <div className="flex justify-end gap-3">
                <Button variant="outline" onClick={() => navigate('/breach')} type="button">
                    Cancel
                </Button>
                <Button type="submit" isLoading={isLoading} >
                    {isEdit ? 'Update Incident' : 'Create Incident'}
                </Button>
            </div>
        </form>
    );
};
