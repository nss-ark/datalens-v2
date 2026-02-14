import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Activity, ArrowLeft, FileText, User } from 'lucide-react';
import { useBreach, useUpdateIncident } from '../../hooks/useBreach';
import { Button } from '@datalens/shared';
import { BreachStatusBadge } from '../../components/Breach/BreachStatusBadge';
import { SLATimer } from '../../components/Breach/SLATimer';
import { BreachForm } from '../../components/Breach/BreachForm';
import { CertInReportModal } from '../../components/Breach/CertInReportModal';
import { breachService } from '../../services/breach';
import { toast } from 'react-toastify';
import type { UpdateIncidentInput } from '../../types/breach';

const IncidentDetail = () => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const { data, isLoading, isError } = useBreach(id!);
    const updateMutation = useUpdateIncident();

    const [isEditing, setIsEditing] = useState(false);
    const [reportModalOpen, setReportModalOpen] = useState(false);
    const [reportData, setReportData] = useState<Record<string, unknown> | null>(null);

    if (isLoading) return <div className="p-10 text-center">Loading incident details...</div>;
    if (isError || !data) return <div className="p-10 text-center text-red-500">Incident not found</div>;

    const { incident, sla } = data;

    const handleUpdate = (updatedData: UpdateIncidentInput) => { // keeping any here as it comes from form, but could be specific
        updateMutation.mutate(
            { id: id!, data: updatedData },
            {
                onSuccess: () => {
                    toast.success('Incident updated successfully');
                    setIsEditing(false);
                },
                onError: () => toast.error('Failed to update incident')
            }
        );
    };

    const handleGenerateReport = async () => {
        try {
            const report = await breachService.generateCertInReport(id!);
            setReportData(report);
            setReportModalOpen(true);
        } catch {
            toast.error('Failed to generate report');
        }
    };

    if (isEditing) {
        return (
            <div className="p-6 max-w-4xl mx-auto">
                <Button variant="ghost" className="mb-4" onClick={() => setIsEditing(false)}>
                    <ArrowLeft size={16} className="mr-2" /> Back to Details
                </Button>
                <h1 className="text-2xl font-bold mb-6">Edit Incident</h1>
                <BreachForm
                    initialData={incident}
                    onSubmit={handleUpdate}
                    isLoading={updateMutation.isPending}
                    isEdit
                />
            </div>
        );
    }

    return (
        <div className="p-6 max-w-[1600px] mx-auto">
            {/* Header */}
            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-8">
                <div>
                    <Button variant="ghost" size="sm" onClick={() => navigate('/breach')} className="mb-2 pl-0">
                        <ArrowLeft size={16} className="mr-1" /> Back to Incidents
                    </Button>
                    <div className="flex items-center gap-3">
                        <h1 className="text-2xl font-bold text-gray-900">{incident.title}</h1>
                        <BreachStatusBadge status={incident.status} />
                    </div>
                </div>
                <div className="flex gap-3">
                    <Button variant="outline" onClick={() => setIsEditing(true)}>Edit Incident</Button>
                    {(incident.is_reportable_cert_in || incident.severity === 'HIGH' || incident.severity === 'CRITICAL') && (
                        <Button icon={<FileText />} onClick={handleGenerateReport}>
                            Generate CERT-In Report
                        </Button>
                    )}
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                {/* Main Content */}
                <div className="lg:col-span-2 space-y-6">
                    {/* SLA Timers */}
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                        <SLATimer
                            label="CERT-In Deadline (6h)"
                            deadline={sla.cert_in_deadline}
                            isOverdue={sla.overdue_cert_in}
                        />
                        <SLATimer
                            label="DPB Deadline (72h)"
                            deadline={sla.dpb_deadline}
                            isOverdue={sla.overdue_dpb}
                        />
                    </div>

                    {/* Details Card */}
                    <div className="bg-white rounded-lg border shadow-sm p-6">
                        <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
                            <Activity size={20} /> Incident Overview
                        </h3>
                        <div className="space-y-4">
                            <div>
                                <h4 className="text-sm font-medium text-gray-500">Description</h4>
                                <p className="mt-1 text-gray-900 leading-relaxed">{incident.description}</p>
                            </div>
                            <div className="grid grid-cols-2 gap-6">
                                <div>
                                    <h4 className="text-sm font-medium text-gray-500">Type</h4>
                                    <p className="mt-1">{incident.type}</p>
                                </div>
                                <div>
                                    <h4 className="text-sm font-medium text-gray-500">Severity</h4>
                                    <p className="mt-1 font-medium">{incident.severity}</p>
                                </div>
                                <div>
                                    <h4 className="text-sm font-medium text-gray-500">Detected At</h4>
                                    <p className="mt-1">{new Date(incident.detected_at).toLocaleString()}</p>
                                </div>
                                <div>
                                    <h4 className="text-sm font-medium text-gray-500">Occurred At</h4>
                                    <p className="mt-1">{new Date(incident.occurred_at).toLocaleString()}</p>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Impact Card */}
                    <div className="bg-white rounded-lg border shadow-sm p-6">
                        <h3 className="text-lg font-semibold mb-4">Impact Assessment</h3>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <div>
                                <h4 className="text-sm font-medium text-gray-500">Affected Systems</h4>
                                <div className="mt-2 flex flex-wrap gap-2">
                                    {incident.affected_systems?.map((sys, idx) => (
                                        <span key={idx} className="px-2 py-1 bg-gray-100 rounded text-xs font-medium border">
                                            {sys}
                                        </span>
                                    ))}
                                    {(!incident.affected_systems || incident.affected_systems.length === 0) && (
                                        <span className="text-gray-400 italic">No systems listed</span>
                                    )}
                                </div>
                            </div>
                            <div>
                                <h4 className="text-sm font-medium text-gray-500">PII Categories</h4>
                                <div className="mt-2 flex flex-wrap gap-2">
                                    {incident.pii_categories?.map((cat, idx) => (
                                        <span key={idx} className="px-2 py-1 bg-blue-50 text-blue-700 rounded text-xs font-medium border border-blue-100">
                                            {cat}
                                        </span>
                                    ))}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Sidebar Info */}
                <div className="space-y-6">
                    <div className="bg-white rounded-lg border shadow-sm p-6">
                        <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
                            <User size={20} /> Point of Contact
                        </h3>
                        <div className="space-y-3">
                            <div>
                                <span className="text-xs text-gray-500 block">Name</span>
                                <span className="font-medium">{incident.poc_name || 'N/A'}</span>
                            </div>
                            <div>
                                <span className="text-xs text-gray-500 block">Role</span>
                                <span className="font-medium">{incident.poc_role || 'N/A'}</span>
                            </div>
                            <div>
                                <span className="text-xs text-gray-500 block">Email</span>
                                <a href={`mailto:${incident.poc_email} `} className="text-blue-600 hover:underline text-sm">
                                    {incident.poc_email || 'N/A'}
                                </a>
                            </div>
                        </div>
                    </div>

                    <div className="bg-gray-50 rounded-lg border p-4">
                        <h4 className="font-semibold text-sm mb-2">Compliance Check</h4>
                        <ul className="space-y-2 text-sm">
                            <li className="flex justify-between">
                                <span>Reportable to CERT-In?</span>
                                <span className={incident.is_reportable_cert_in ? "text-red-600 font-bold" : "text-gray-500"}>
                                    {incident.is_reportable_cert_in ? "YES" : "NO"}
                                </span>
                            </li>
                            <li className="flex justify-between">
                                <span>Reportable to DPB?</span>
                                <span className={incident.is_reportable_dpb ? "text-red-600 font-bold" : "text-gray-500"}>
                                    {incident.is_reportable_dpb ? "YES" : "NO"}
                                </span>
                            </li>
                        </ul>
                    </div>
                </div>
            </div>

            <CertInReportModal
                isOpen={reportModalOpen}
                onClose={() => setReportModalOpen(false)}
                reportData={reportData}
            />
        </div>
    );
};

export default IncidentDetail;
