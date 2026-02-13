import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { DataTable } from '../../components/DataTable/DataTable';
import { StatusBadge } from '../../components/common/StatusBadge';
import { Button } from '../../components/common/Button';
import { grievanceService } from '../../services/grievanceService';
import { useNavigate } from 'react-router-dom';
import { Eye } from 'lucide-react';
import { format } from 'date-fns';
import type { Grievance, GrievanceStatus } from '../../types/grievance';

export default function GrievanceList() {
    const navigate = useNavigate();
    const [page] = useState(1); // TODO: Add pagination controls if needed
    const [statusFilter, setStatusFilter] = useState<GrievanceStatus | ''>('');

    const { data, isLoading } = useQuery({
        queryKey: ['grievances', page, statusFilter],
        queryFn: () => grievanceService.listGrievances({
            page,
            page_size: 10,
            status: statusFilter || undefined
        })
    });

    const columns = [
        {
            key: 'subject',
            header: 'Subject',
            render: (row: Grievance) => (
                <div className="max-w-xs truncate font-medium text-gray-900" title={row.subject}>
                    {row.subject}
                </div>
            )
        },
        {
            key: 'category',
            header: 'Category',
            render: (row: Grievance) => <span className="text-xs px-2 py-1 bg-gray-100 rounded-full text-gray-600">{row.category}</span>
        },
        {
            key: 'status',
            header: 'Status',
            render: (row: Grievance) => <StatusBadge label={row.status} />
        },
        {
            key: 'priority',
            header: 'Priority',
            render: (row: Grievance) => (
                <span className={`text-xs font-bold ${row.priority === 'CRITICAL' ? 'text-red-600' :
                    row.priority === 'HIGH' ? 'text-orange-500' :
                        'text-gray-500'
                    }`}>
                    {row.priority}
                </span>
            )
        },
        {
            key: 'submitted_at',
            header: 'Submitted',
            render: (row: Grievance) => (
                <span className="text-sm text-gray-500">
                    {format(new Date(row.submitted_at), 'MMM d, yyyy')}
                </span>
            )
        },
        {
            key: 'actions',
            header: 'Actions',
            render: (row: Grievance) => (
                <Button
                    size="sm"
                    variant="secondary"
                    icon={<Eye size={14} />}
                    onClick={() => navigate(`/compliance/grievances/${row.id}`)}
                >
                    View
                </Button>
            )
        }
    ];

    return (
        <div className="p-6">
            <div className="flex justify-between items-center mb-6">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900">Grievance Redressal</h1>
                    <p className="text-gray-500">Manage and resolve data principal complaints.</p>
                </div>
            </div>

            <div className="flex space-x-2 mb-4">
                {['', 'OPEN', 'IN_PROGRESS', 'RESOLVED', 'ESCALATED'].map(status => (
                    <button
                        key={status}
                        onClick={() => setStatusFilter(status as GrievanceStatus)}
                        className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${statusFilter === status
                            ? 'bg-blue-100 text-blue-700'
                            : 'bg-white text-gray-600 hover:bg-gray-50 border'
                            }`}
                    >
                        {status === '' ? 'All' : status.replace('_', ' ')}
                    </button>
                ))}
            </div>

            <DataTable
                columns={columns}
                data={data?.items || []}
                isLoading={isLoading}
                keyExtractor={(row) => row.id}
                emptyTitle="No grievances found"
                emptyDescription="No grievances match your filters."
            />
        </div>
    );
}
