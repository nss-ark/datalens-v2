import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Plus } from 'lucide-react';
import { adminService } from '../../../services/adminService';
import { DataTable, type Column } from '../../../components/DataTable/DataTable';
import { Pagination } from '../../../components/DataTable/Pagination';
import { Button } from '../../../components/common/Button';
import { StatusBadge } from '../../../components/common/StatusBadge';
import { TenantForm } from './TenantForm';
import type { Tenant } from '../../../types/admin';

export default function TenantList() {
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(10);
    const [isCreateOpen, setIsCreateOpen] = useState(false);

    const { data, isLoading, refetch } = useQuery({
        queryKey: ['admin-tenants', page, pageSize],
        queryFn: () => adminService.getTenants({ page, limit: pageSize }),
    });

    const columns: Column<Tenant>[] = [
        {
            key: 'name',
            header: 'Organization',
            sortable: true,
            render: (row) => (
                <div>
                    <div className="font-medium text-gray-900">{row.name}</div>
                    <div className="text-xs text-gray-500">{row.domain}.datalens.com</div>
                </div>
            )
        },
        {
            key: 'plan',
            header: 'Plan',
            render: (row) => (
                <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800">
                    {row.plan}
                </span>
            )
        },
        {
            key: 'status',
            header: 'Status',
            render: (row) => (
                <StatusBadge
                    label={row.status}
                />
            )
        },
        {
            key: 'created_at',
            header: 'Joined',
            render: (row) => new Date(row.created_at).toLocaleDateString()
        }
    ];

    return (
        <div className="p-6">
            <div className="flex justify-between items-center mb-6">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900">Tenants</h1>
                    <p className="text-gray-500 mt-1">Manage organizations and subscriptions</p>
                </div>
                <Button icon={<Plus size={16} />} onClick={() => setIsCreateOpen(true)}>
                    New Tenant
                </Button>
            </div>

            <div className="bg-white rounded-lg shadow border border-gray-200">
                <DataTable
                    columns={columns}
                    data={data?.items || []}
                    isLoading={isLoading}
                    keyExtractor={(row) => row.id}
                    emptyTitle="No tenants found"
                    emptyDescription="Get started by onboarding a new tenant."
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

            <TenantForm
                isOpen={isCreateOpen}
                onClose={() => setIsCreateOpen(false)}
                onSuccess={() => refetch()}
            />
        </div>
    );
}
