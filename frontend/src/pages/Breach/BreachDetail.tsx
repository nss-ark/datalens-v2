import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
    Clock,
    ShieldAlert,
    FileText,
    User,
    Calendar,
    Edit,
    ArrowLeft,
    Download
} from 'lucide-react';
import { toast } from 'react-toastify';
import { breachService } from '../../services/breach';
import { Button } from '../../components/common/Button';
import { BreachStatusBadge } from '../../components/Breach/BreachStatusBadge';
import type { IncidentStatus } from '../../types/breach';

const BreachDetail = () => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const queryClient = useQueryClient();

    const { data, isLoading, isError } = useQuery({
        queryKey: ['breach', id],
        queryFn: () => breachService.getById(id!),
        enabled: !!id
    });

    const updateStatusMutation = useMutation({
        mutationFn: ({ status }: { status: IncidentStatus }) =>
            breachService.update(id!, { status }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['breach', id] });
            toast.success('Status updated successfully');
        },
        onError: () => toast.error('Failed to update status')
    });

    const generateReportMutation = useMutation({
        mutationFn: () => breachService.generateCertInReport(id!),
        onSuccess: () => {
            toast.success('CERT-In Report generated (Check console/download)');
            // In a real app, this would trigger a file download or open a preview
        },
        onError: () => toast.error('Failed to generate report')
    });

    if (isLoading) return <div className="p-8 text-center">Loading incident details...</div>;
    if (isError || !data) return <div className="p-8 text-center text-red-600">Failed to load incident</div>;

    const { incident, sla } = data;

    const handleStatusChange = (newStatus: IncidentStatus) => {
        if (confirm(`Are you sure you want to change status to ${newStatus}?`)) {
            updateStatusMutation.mutate({ status: newStatus });
        }
    };

    return (
        <div className="p-6 max-w-[1600px] mx-auto space-y-6">
            {/* Header */}
            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                <div className="flex items-center gap-2">
                    <Button variant="ghost" size="sm" onClick={() => navigate('/breach')}>
                        <ArrowLeft size={16} />
                    </Button>
                    <div>
                        <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
                            {incident.title}
                            <BreachStatusBadge status={incident.status} />
                        </h1>
                        <p className="text-gray-500 text-sm mt-1">
                            ID: {incident.id} • Type: {incident.type}
                        </p>
                    </div>
                </div>
                <div className="flex gap-2">
                    <Button
                        variant="outline"
                        icon={<Edit size={16} />}
                        onClick={() => navigate(`/breach/${id}/edit`)}
                    >
                        Edit
                    </Button>
                    <Button
                        variant="secondary"
                        icon={<Download size={16} />}
                        onClick={() => generateReportMutation.mutate()}
                        isLoading={generateReportMutation.isPending}
                    >
                        CERT-In Report
                    </Button>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Left Column - Main Info */}
                <div className="lg:col-span-2 space-y-6">
                    {/* SLA Cards */}
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div className={`p-4 rounded-lg border border-l-4 ${sla.overdue_cert_in ? 'bg-red-50 border-red-500' : 'bg-blue-50 border-blue-500'}`}>
                            <div className="flex justify-between items-start">
                                <div>
                                    <h3 className="font-semibold text-gray-900">CERT-In Deadline (6h)</h3>
                                    <p className="text-sm mt-1">Reporting Window</p>
                                </div>
                                <Clock size={20} className={sla.overdue_cert_in ? 'text-red-500' : 'text-blue-500'} />
                            </div>
                            <div className="mt-4">
                                <span className={`text-2xl font-bold ${sla.overdue_cert_in ? 'text-red-700' : 'text-blue-700'}`}>
                                    {sla.time_remaining_cert_in}
                                </span>
                                <p className="text-xs text-gray-500 mt-1">Due: {new Date(sla.cert_in_deadline).toLocaleString()}</p>
                            </div>
                        </div>

                        <div className={`p-4 rounded-lg border border-l-4 ${sla.overdue_dpb ? 'bg-red-50 border-red-500' : 'bg-purple-50 border-purple-500'}`}>
                            <div className="flex justify-between items-start">
                                <div>
                                    <h3 className="font-semibold text-gray-900">DPB Deadline (72h)</h3>
                                    <p className="text-sm mt-1">Data Protection Board</p>
                                </div>
                                <Clock size={20} className={sla.overdue_dpb ? 'text-red-500' : 'text-purple-500'} />
                            </div>
                            <div className="mt-4">
                                <span className={`text-2xl font-bold ${sla.overdue_dpb ? 'text-red-700' : 'text-purple-700'}`}>
                                    {sla.time_remaining_dpb}
                                </span>
                                <p className="text-xs text-gray-500 mt-1">Due: {new Date(sla.dpb_deadline).toLocaleString()}</p>
                            </div>
                        </div>
                    </div>

                    {/* Description */}
                    <div className="bg-white rounded-lg border shadow-sm p-6">
                        <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
                            <FileText size={20} />
                            Incident Description
                        </h2>
                        <p className="text-gray-700 whitespace-pre-wrap leading-relaxed">
                            {incident.description}
                        </p>
                    </div>

                    {/* Impact Analysis */}
                    <div className="bg-white rounded-lg border shadow-sm p-6">
                        <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
                            <ShieldAlert size={20} />
                            Impact Analysis
                        </h2>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <div>
                                <h4 className="text-sm font-medium text-gray-500 uppercase tracking-wider mb-2">Affected Systems</h4>
                                <ul className="list-disc pl-5 space-y-1">
                                    {incident.affected_systems.map((sys, idx) => (
                                        <li key={idx} className="text-gray-800">{sys}</li>
                                    ))}
                                </ul>
                            </div>
                            <div>
                                <h4 className="text-sm font-medium text-gray-500 uppercase tracking-wider mb-2">PII Categories Exposed</h4>
                                <div className="flex flex-wrap gap-2">
                                    {incident.pii_categories.map((cat, idx) => (
                                        <span key={idx} className="px-2 py-1 bg-red-100 text-red-800 rounded text-xs font-medium">
                                            {cat}
                                        </span>
                                    ))}
                                </div>
                            </div>
                        </div>
                        <div className="mt-6 pt-4 border-t">
                            <div className="flex items-center gap-2">
                                <span className="text-gray-600">Estimated Data Subjects Affected:</span>
                                <span className="font-bold text-gray-900 text-lg">{incident.affected_data_subject_count}</span>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Right Column - Status & Meta */}
                <div className="space-y-6">
                    {/* Status Workflow */}
                    <div className="bg-white rounded-lg border shadow-sm p-6">
                        <h2 className="text-lg font-semibold mb-4">Status & Action</h2>
                        <div className="space-y-2">
                            <p className="text-sm text-gray-500 mb-2">Move incident to next stage:</p>
                            {incident.status === 'OPEN' && (
                                <Button className="w-full" onClick={() => handleStatusChange('INVESTIGATING')}>Start Investigation</Button>
                            )}
                            {incident.status === 'INVESTIGATING' && (
                                <Button className="w-full" onClick={() => handleStatusChange('CONTAINED')}>Mark as Contained</Button>
                            )}
                            {incident.status === 'CONTAINED' && (
                                <Button className="w-full" onClick={() => handleStatusChange('RESOLVED')}>Mark as Resolved</Button>
                            )}
                            {incident.status === 'RESOLVED' && (
                                <Button className="w-full" onClick={() => handleStatusChange('CLOSED')}>Close Incident</Button>
                            )}

                            {incident.status === 'CLOSED' && (
                                <div className="p-3 bg-gray-50 text-gray-600 text-center rounded text-sm">
                                    Incident is closed. Re-open if necessary.
                                    <Button variant="ghost" size="sm" onClick={() => handleStatusChange('OPEN')}>Re-open</Button>
                                </div>
                            )}
                        </div>
                    </div>

                    {/* Meta Details */}
                    <div className="bg-white rounded-lg border shadow-sm p-6">
                        <h2 className="text-lg font-semibold mb-4">Metadata</h2>
                        <dl className="space-y-4">
                            <div>
                                <dt className="text-sm font-medium text-gray-500">Severity</dt>
                                <dd className={`mt-1 font-semibold ${incident.severity === 'CRITICAL' ? 'text-red-700' :
                                    incident.severity === 'HIGH' ? 'text-orange-700' : 'text-gray-900'
                                    }`}>
                                    {incident.severity}
                                </dd>
                            </div>
                            <div>
                                <dt className="text-sm font-medium text-gray-500">Detected At</dt>
                                <dd className="mt-1 text-sm text-gray-900 flex items-center gap-2">
                                    <Calendar size={14} />
                                    {new Date(incident.detected_at).toLocaleString()}
                                </dd>
                            </div>
                            <div>
                                <dt className="text-sm font-medium text-gray-500">Occurred At</dt>
                                <dd className="mt-1 text-sm text-gray-900 flex items-center gap-2">
                                    <Calendar size={14} />
                                    {new Date(incident.occurred_at).toLocaleString()}
                                </dd>
                            </div>
                        </dl>
                    </div>

                    {/* PoC Info */}
                    <div className="bg-white rounded-lg border shadow-sm p-6">
                        <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
                            <User size={20} />
                            Point of Contact
                        </h2>
                        <dl className="space-y-3">
                            <div>
                                <dt className="text-xs font-medium text-gray-500 uppercase">Name</dt>
                                <dd className="text-sm font-medium">{incident.poc_name}</dd>
                            </div>
                            <div>
                                <dt className="text-xs font-medium text-gray-500 uppercase">Role</dt>
                                <dd className="text-sm text-gray-700">{incident.poc_role}</dd>
                            </div>
                            <div>
                                <dt className="text-xs font-medium text-gray-500 uppercase">Email</dt>
                                <dd className="text-sm text-blue-600 truncate">{incident.poc_email}</dd>
                            </div>
                        </dl>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default BreachDetail;
