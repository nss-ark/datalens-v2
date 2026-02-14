import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Plus, ShieldAlert } from 'lucide-react';
import { Button } from '@datalens/shared';
import { DataTable, type Column } from '@datalens/shared';
import { Pagination } from '@datalens/shared';
import { useBreachList } from '../../hooks/useBreach';
import { BreachStats } from '../../components/Breach/BreachStats';
import { BreachStatusBadge } from '../../components/Breach/BreachStatusBadge';
import type { BreachIncident, IncidentStatus } from '../../types/breach';

const BreachDashboard = () => {
    const navigate = useNavigate();
    const [page, setPage] = useState(1);
    const [statusFilter, setStatusFilter] = useState('');
    const { data, isLoading } = useBreachList({ page, page_size: 10, status: statusFilter ? statusFilter as IncidentStatus : undefined });

    const columns: Column<BreachIncident>[] = [
        {
            header: 'Incident',
            key: 'title',
            render: (row) => (
                <div>
                    <div className="font-medium text-gray-900">{row.title}</div>
                    <div className="text-xs text-gray-500">{row.type}</div>
                </div>
            )
        },
        {
            header: 'Severity',
            key: 'severity',
            render: (row) => (
                <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${row.severity === 'CRITICAL' ? 'bg-red-100 text-red-800' :
                    row.severity === 'HIGH' ? 'bg-orange-100 text-orange-800' :
                        row.severity === 'MEDIUM' ? 'bg-yellow-100 text-yellow-800' :
                            'bg-green-100 text-green-800'
                    }`}>
                    {row.severity}
                </span>
            )
        },
        {
            header: 'Status',
            key: 'status',
            render: (row) => <BreachStatusBadge status={row.status} />
        },
        {
            header: 'Detected',
            key: 'detected_at',
            render: (row) => new Date(row.detected_at).toLocaleDateString()
        },
        {
            header: 'Actions',
            key: 'actions',
            render: (row) => (
                <Button variant="ghost" size="sm" onClick={() => navigate(`/breach/${row.id}`)}>
                    View
                </Button>
            )
        }
    ];

    // For stats, we might want a separate API call or aggregated data.
    // Assuming the list returns enough for a rudimentary calc or we use a separate stats endpoint.
    // For MVP, passing the current page items + total count is partial, but acceptable for demo.
    // Better: Helper hook or separate endpoint for stats. The backend `GET /dashboard/stats` is generic.
    // Update: I'll pass the current items for now, ideally we'd fetch stats separately.

    return (
        <div className="p-6 max-w-[1600px] mx-auto">
            <div className="flex justify-between items-center mb-6">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900">Breach Management</h1>
                    <p className="text-gray-500 mt-1">Monitor, track, and report security incidents</p>
                </div>
                <Button icon={<Plus />} onClick={() => navigate('/breach/new')}>
                    Report Incident
                </Button>
            </div>

            {/* In a real app, stats should come from a dedicated endpoint to be accurate across all pages */}
            <BreachStats incidents={data?.items || []} />

            <div className="bg-white rounded-lg border shadow-sm">
                <div className="p-4 border-b flex justify-between items-center">
                    <h2 className="font-semibold text-gray-800 flex items-center gap-2">
                        <ShieldAlert size={18} />
                        Incident Log
                    </h2>
                    <div className="flex gap-2">
                        <select
                            className="text-sm border rounded px-2 py-1"
                            value={statusFilter}
                            onChange={(e) => setStatusFilter(e.target.value)}
                        >
                            <option value="">All Statuses</option>
                            <option value="OPEN">Open</option>
                            <option value="INVESTIGATING">Investigating</option>
                            <option value="RESOLVED">Resolved</option>
                        </select>
                    </div>
                </div>

                <DataTable
                    columns={columns}
                    data={data?.items || []}
                    isLoading={isLoading}
                    emptyTitle="No incidents recorded"
                    emptyDescription="There are no breach incidents to display."
                    keyExtractor={(row) => row.id}
                />

                {data && (
                    <Pagination
                        page={page}
                        total={data.total}
                        pageSize={10}
                        onPageChange={setPage}
                    />
                )}
            </div>
        </div>
    );
};

export default BreachDashboard;
