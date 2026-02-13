import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { Eye } from 'lucide-react';
import { adminService } from '../../../services/adminService';
import { DataTable, type Column } from '../../../components/DataTable/DataTable';
import { Pagination } from '../../../components/DataTable/Pagination';
import { StatusBadge } from '../../../components/common/StatusBadge';
import { Button } from '../../../components/common/Button';
import type { AdminDSR } from '../../../types/admin';

export default function AdminDSRList() {
    const navigate = useNavigate();
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(10);
    const [statusFilter, setStatusFilter] = useState<string>('');
    const [typeFilter, setTypeFilter] = useState<string>('');

    const { data, isLoading, error } = useQuery({
        queryKey: ['admin-dsrs', page, pageSize, statusFilter, typeFilter],
        queryFn: () => adminService.getDSRs({
            page,
            limit: pageSize,
            status: statusFilter || undefined,
            type: typeFilter || undefined
        }),
    });

    const columns: Column<AdminDSR>[] = [
        {
            key: 'id',
            header: 'ID',
            render: (row) => <span className="font-mono text-xs">{row.id.substring(0, 8)}...</span>
        },
        {
            key: 'tenant_name',
            header: 'Tenant',
            render: (row) => (
                <div>
                    {/* Fallback if tenant_name is missing due to API gap */}
                    <div className="font-medium text-gray-900">{row.tenant_name || 'Unknown Tenant'}</div>
                    <div className="text-xs text-gray-500">{row.tenant_id}</div>
                </div>
            )
        },
        {
            key: 'subject_name',
            header: 'Subject',
            render: (row) => (
                <div>
                    <div className="font-medium text-gray-900">{row.subject_name}</div>
                    <div className="text-xs text-gray-500">{row.subject_email}</div>
                </div>
            )
        },
        {
            key: 'request_type',
            header: 'Type',
            render: (row) => (
                <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800">
                    {row.request_type}
                </span>
            )
        },
        {
            key: 'status',
            header: 'Status',
            render: (row) => <StatusBadge label={row.status} />
        },
        {
            key: 'commences_at', // Using generic render for date
            header: 'Deadline',
            render: (row) => {
                const deadline = new Date(row.sla_deadline);
                const isOverdue = deadline < new Date() && row.status !== 'COMPLETED' && row.status !== 'REJECTED';
                return (
                    <span className={isOverdue ? 'text-red-600 font-bold' : ''}>
                        {deadline.toLocaleDateString()}
                    </span>
                );
            }
        },
        {
            key: 'actions',
            header: '',
            render: (row) => (
                <Button
                    variant="ghost"
                    size="sm"
                    icon={<Eye size={16} />}
                    onClick={() => navigate(`/admin/compliance/dsr/${row.id}`)}
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
                    <h1 className="text-2xl font-bold text-gray-900">DSR Requests</h1>
                    <p className="text-gray-500 mt-1">Monitor compliance requests across all tenants</p>
                </div>
            </div>

            <div className="bg-white p-4 rounded-lg shadow border border-gray-200 mb-6 flex gap-4">
                <div className="w-48">
                    <label className="block text-sm font-medium text-gray-700 mb-1">Status</label>
                    <select
                        className="w-full border-gray-300 rounded-md shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                        value={statusFilter}
                        onChange={(e) => setStatusFilter(e.target.value)}
                    >
                        <option value="">All Statuses</option>
                        <option value="PENDING">Pending</option>
                        <option value="IN_PROGRESS">In Progress</option>
                        <option value="COMPLETED">Completed</option>
                        <option value="REJECTED">Rejected</option>
                    </select>
                </div>
                <div className="w-48">
                    <label className="block text-sm font-medium text-gray-700 mb-1">Type</label>
                    <select
                        className="w-full border-gray-300 rounded-md shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                        value={typeFilter}
                        onChange={(e) => setTypeFilter(e.target.value)}
                    >
                        <option value="">All Types</option>
                        <option value="ACCESS">Access</option>
                        <option value="ERASURE">Erasure</option>
                        <option value="CORRECTION">Correction</option>
                        <option value="PORTABILITY">Portability</option>
                    </select>
                </div>
            </div>

            {error ? (
                <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded relative" role="alert">
                    <strong className="font-bold">Error loading DSRs: </strong>
                    <span className="block sm:inline">{(error as Error).message}</span>
                    <p className="mt-2 text-sm">Note: Cross-tenant DSR listing endpoint may not be implemented yet.</p>
                </div>
            ) : (
                <div className="bg-white rounded-lg shadow border border-gray-200">
                    <DataTable
                        columns={columns}
                        data={data?.items || []}
                        isLoading={isLoading}
                        keyExtractor={(row) => row.id}
                        emptyTitle="No DSRs found"
                        emptyDescription="Adjust filters or check back later."
                    />
                    {data && (
                        <Pagination
                            page={page}
                            pageSize={data.page_size}
                            total={data.total}
                            onPageChange={setPage}
                            onPageSizeChange={setPageSize}
                        />
                    )}
                </div>
            )}
        </div>
    );
}
