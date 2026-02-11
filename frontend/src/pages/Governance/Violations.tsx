import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { AlertTriangle, CheckCircle } from 'lucide-react';
import { governanceService } from '../../services/governance';
import { DataTable } from '../../components/DataTable/DataTable';
import { toast } from '../../stores/toastStore';
import type { PolicyViolation } from '../../types/governance';

const Violations = () => {
    const queryClient = useQueryClient();
    const [filter, setFilter] = useState<'all' | 'unresolved'>('all');

    const { data: violations = [], isLoading } = useQuery({
        queryKey: ['violations'],
        queryFn: governanceService.getViolations,
    });

    const resolveMutation = useMutation({
        mutationFn: governanceService.resolveViolation,
        onSuccess: () => {
            toast.success('Violation Resolved', 'The violation has been marked as resolved.');
            queryClient.invalidateQueries({ queryKey: ['violations'] });
        },
        onError: () => {
            toast.error('Error', 'Failed to resolve violation.');
        }
    });

    const filteredViolations = violations.filter((v: PolicyViolation) => {
        if (filter === 'unresolved') return v.status !== 'resolved';
        return true;
    });

    const getSeverityBadge = (severity: string) => {
        const colors: Record<string, string> = {
            critical: 'bg-red-100 text-red-800',
            high: 'bg-orange-100 text-orange-800',
            medium: 'bg-yellow-100 text-yellow-800',
            low: 'bg-blue-100 text-blue-800',
        };
        return (
            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium uppercase ${colors[severity] || 'bg-gray-100 text-gray-800'}`}>
                {severity}
            </span>
        );
    };

    const columns = [
        {
            key: 'policyName',
            header: 'Violated Policy',
            sortable: true,
            render: (row: PolicyViolation) => (
                <span className="font-medium text-gray-900">{row.policyName}</span>
            )
        },
        {
            key: 'location',
            header: 'Data Location',
            sortable: true,
            render: (row: PolicyViolation) => (
                <div className="text-sm">
                    <div className="text-gray-900">{row.dataSource}</div>
                    <div className="text-gray-500">{row.dataElement}</div>
                </div>
            )
        },
        {
            key: 'severity',
            header: 'Severity',
            sortable: true,
            render: (row: PolicyViolation) => getSeverityBadge(row.severity)
        },
        {
            key: 'status',
            header: 'Status',
            sortable: true,
            render: (row: PolicyViolation) => (
                <span className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium ${row.status === 'resolved' ? 'bg-green-100 text-green-800' : 'bg-red-50 text-red-800'
                    }`}>
                    {row.status === 'resolved' && <CheckCircle size={12} />}
                    <span className="capitalize">{row.status}</span>
                </span>
            )
        },
        {
            key: 'detectedAt',
            header: 'Detected',
            sortable: true,
            render: (row: PolicyViolation) => new Date(row.detectedAt).toLocaleDateString()
        },
        {
            key: 'actions',
            header: '',
            render: (row: PolicyViolation) => (
                <div className="flex justify-end">
                    {row.status !== 'resolved' && (
                        <button
                            onClick={(e) => {
                                e.stopPropagation();
                                resolveMutation.mutate(row.id);
                            }}
                            className="text-blue-600 hover:text-blue-800 text-sm font-medium transition-colors"
                        >
                            Resolve
                        </button>
                    )}
                </div>
            )
        }
    ];

    return (
        <div className="p-6 max-w-7xl mx-auto">
            <div className="flex justify-between items-center mb-8">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
                        <AlertTriangle className="text-red-500" />
                        Compliance Issues
                    </h1>
                    <p className="text-gray-500 mt-1">
                        Monitor and resolve policy violations across your data estate.
                    </p>
                </div>
                <div>
                    <select
                        className="border-gray-300 rounded-md shadow-sm text-sm p-2 border"
                        value={filter}
                        onChange={(e) => setFilter(e.target.value as 'all' | 'unresolved')}
                    >
                        <option value="all">All Violations</option>
                        <option value="unresolved">Unresolved Only</option>
                    </select>
                </div>
            </div>

            <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
                <DataTable
                    columns={columns}
                    data={filteredViolations}
                    isLoading={isLoading}
                    keyExtractor={(row) => row.id}
                    emptyTitle="No issues detected"
                    emptyDescription="Your data landscape is compliant with all active policies."
                />
            </div>
        </div>
    );
};

export default Violations;
