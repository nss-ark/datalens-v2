import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Plus, Building2, Globe, Clock, LayoutGrid, List } from 'lucide-react';
import { adminService } from '@/services/adminService';
import { DataTable, type Column } from '@datalens/shared';
import { Pagination } from '@datalens/shared';
import { Button } from '@datalens/shared';
import { StatusBadge } from '@datalens/shared';
import { MotionList } from '@datalens/shared';
import { cn } from '@datalens/shared';
import { TenantForm } from './TenantForm';
import type { Tenant } from '@/types/admin';

type ViewMode = 'grid' | 'table';

function TenantCard({ tenant }: { tenant: Tenant }) {
    const planColors: Record<string, string> = {
        ENTERPRISE: 'bg-purple-100 text-purple-700 dark:bg-purple-500/20 dark:text-purple-300',
        PROFESSIONAL: 'bg-blue-100 text-blue-700 dark:bg-blue-500/20 dark:text-blue-300',
        STARTER: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300',
        FREE: 'bg-zinc-100 text-zinc-600 dark:bg-zinc-700 dark:text-zinc-300',
    };

    return (
        <div className="group relative bg-white dark:bg-zinc-900 rounded-xl border border-zinc-200 dark:border-zinc-800 p-5 hover:shadow-md hover:border-zinc-300 dark:hover:border-zinc-700 transition-all duration-200">
            {/* Status dot */}
            <div className={cn(
                "absolute top-4 right-4 w-2.5 h-2.5 rounded-full",
                tenant.status === 'ACTIVE' ? 'bg-emerald-500' : tenant.status === 'SUSPENDED' ? 'bg-red-500' : 'bg-zinc-400'
            )} />

            {/* Org avatar + name */}
            <div className="flex items-start gap-4 mb-4">
                <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 dark:from-blue-600 dark:to-blue-700 flex items-center justify-center text-white font-bold text-lg shadow-sm shrink-0">
                    {tenant.name.charAt(0)}
                </div>
                <div className="min-w-0 flex-1">
                    <h3 className="font-semibold text-zinc-900 dark:text-zinc-50 text-sm truncate">
                        {tenant.name}
                    </h3>
                    <p className="text-xs text-zinc-500 dark:text-zinc-400 flex items-center gap-1 mt-0.5">
                        <Globe className="h-3 w-3" />
                        {tenant.domain}.datalens.com
                    </p>
                </div>
            </div>

            {/* Info row */}
            <div className="flex items-center justify-between">
                <span className={cn(
                    "inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium",
                    planColors[tenant.plan] || planColors.FREE
                )}>
                    {tenant.plan}
                </span>
                <span className="text-xs text-zinc-400 dark:text-zinc-500 flex items-center gap-1">
                    <Clock className="h-3.5 w-3.5" />
                    {new Date(tenant.created_at).toLocaleDateString()}
                </span>
            </div>
        </div>
    );
}

export default function TenantList() {
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(10);
    const [isCreateOpen, setIsCreateOpen] = useState(false);
    const [viewMode, setViewMode] = useState<ViewMode>('grid');

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
                    <div className="font-medium text-zinc-900 dark:text-zinc-100">{row.name}</div>
                    <div className="text-xs text-zinc-500 dark:text-zinc-400">{row.domain}.datalens.com</div>
                </div>
            )
        },
        {
            key: 'plan',
            header: 'Plan',
            render: (row) => (
                <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-500/20 dark:text-blue-300">
                    {row.plan}
                </span>
            )
        },
        {
            key: 'status',
            header: 'Status',
            render: (row) => (
                <StatusBadge label={row.status} />
            )
        },
        {
            key: 'created_at',
            header: 'Joined',
            render: (row) => new Date(row.created_at).toLocaleDateString()
        }
    ];

    return (
        <div className="p-8 space-y-8">
            {/* Header */}
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold text-zinc-900 dark:text-zinc-50 tracking-tight">Tenants</h1>
                    <p className="text-zinc-500 dark:text-zinc-400 mt-2 text-lg">Manage organizations and subscriptions</p>
                </div>
                <div className="flex items-center gap-3">
                    {/* View Toggle */}
                    <div className="flex items-center bg-zinc-100 dark:bg-zinc-800 rounded-lg p-1">
                        <button
                            onClick={() => setViewMode('grid')}
                            className={cn(
                                "p-2 rounded-md transition-colors",
                                viewMode === 'grid'
                                    ? "bg-white dark:bg-zinc-700 text-zinc-900 dark:text-zinc-50 shadow-sm"
                                    : "text-zinc-500 dark:text-zinc-400 hover:text-zinc-700 dark:hover:text-zinc-300"
                            )}
                            title="Grid view"
                        >
                            <LayoutGrid className="h-4 w-4" />
                        </button>
                        <button
                            onClick={() => setViewMode('table')}
                            className={cn(
                                "p-2 rounded-md transition-colors",
                                viewMode === 'table'
                                    ? "bg-white dark:bg-zinc-700 text-zinc-900 dark:text-zinc-50 shadow-sm"
                                    : "text-zinc-500 dark:text-zinc-400 hover:text-zinc-700 dark:hover:text-zinc-300"
                            )}
                            title="Table view"
                        >
                            <List className="h-4 w-4" />
                        </button>
                    </div>
                    <Button icon={<Plus size={16} />} onClick={() => setIsCreateOpen(true)}>
                        New Tenant
                    </Button>
                </div>
            </div>

            {/* Grid View */}
            {viewMode === 'grid' && (
                <>
                    {isLoading ? (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            {[1, 2, 3, 4, 5, 6].map((i) => (
                                <div key={i} className="bg-white dark:bg-zinc-900 h-36 rounded-xl border border-zinc-200 dark:border-zinc-800 animate-pulse" />
                            ))}
                        </div>
                    ) : data?.items && data.items.length > 0 ? (
                        <MotionList
                            items={data.items}
                            className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
                            staggerDelay={0.06}
                            renderItem={(tenant) => <TenantCard tenant={tenant} />}
                        />
                    ) : (
                        <div className="text-center py-16 bg-white dark:bg-zinc-900 rounded-xl border border-zinc-200 dark:border-zinc-800">
                            <Building2 className="h-12 w-12 mx-auto text-zinc-300 dark:text-zinc-600 mb-4" />
                            <h3 className="text-lg font-medium text-zinc-900 dark:text-zinc-50 mb-1">No tenants found</h3>
                            <p className="text-zinc-500 dark:text-zinc-400 mb-4">Get started by onboarding a new tenant.</p>
                            <Button icon={<Plus size={16} />} onClick={() => setIsCreateOpen(true)}>
                                Create First Tenant
                            </Button>
                        </div>
                    )}
                </>
            )}

            {/* Table View */}
            {viewMode === 'table' && (
                <div className="bg-white dark:bg-zinc-900 rounded-xl shadow-sm border border-zinc-200 dark:border-zinc-800 overflow-hidden">
                    <DataTable
                        columns={columns}
                        data={data?.items || []}
                        isLoading={isLoading}
                        keyExtractor={(row) => row.id}
                        emptyTitle="No tenants found"
                        emptyDescription="Get started by onboarding a new tenant."
                    />
                </div>
            )}

            {/* Pagination */}
            {data && (
                <div className="flex justify-center">
                    <Pagination
                        page={page}
                        pageSize={data.page_size}
                        total={data.total}
                        onPageChange={setPage}
                        onPageSizeChange={setPageSize}
                    />
                </div>
            )}

            <TenantForm
                isOpen={isCreateOpen}
                onClose={() => setIsCreateOpen(false)}
                onSuccess={() => refetch()}
            />
        </div>
    );
}
