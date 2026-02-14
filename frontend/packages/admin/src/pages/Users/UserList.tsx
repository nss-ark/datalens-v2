import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Shield, Ban, CheckCircle, Search } from 'lucide-react';
import { adminService } from '../../../services/adminService';
import { DataTable, type Column } from '@datalens/shared';
import { Pagination } from '@datalens/shared';
import { Button } from '@datalens/shared';
import { StatusBadge } from '@datalens/shared';
import { RoleAssignModal } from './RoleAssignModal';
import type { AdminUser, Tenant } from '../../../types/admin';
import { toast } from '@datalens/shared';

export default function UserList() {
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(10);
    const [search, setSearch] = useState('');
    const [selectedTenant, setSelectedTenant] = useState<string>('');
    const [statusFilter, setStatusFilter] = useState<string>('');
    const [selectedUser, setSelectedUser] = useState<AdminUser | null>(null);
    const [isRoleModalOpen, setIsRoleModalOpen] = useState(false);

    // Tenants for filter and lookup
    const { data: tenantsData } = useQuery({
        queryKey: ['admin-tenants-lookup'],
        queryFn: () => adminService.getTenants({ limit: 100 }),
    });

    const tenantMap = new Map(tenantsData?.items.map((t: Tenant) => [t.id, t.name]));

    const { data, isLoading, refetch } = useQuery({
        queryKey: ['admin-users', page, pageSize, search, selectedTenant, statusFilter],
        queryFn: () => adminService.getUsers({
            page,
            limit: pageSize,
            search,
            tenant_id: selectedTenant || undefined,
            status: statusFilter || undefined
        }),
    });

    const handleStatusChange = async (user: AdminUser) => {
        const newStatus = user.status === 'SUSPENDED' ? 'ACTIVE' : 'SUSPENDED';
        const action = newStatus === 'ACTIVE' ? 'activate' : 'suspend';

        if (!window.confirm(`Are you sure you want to ${action} ${user.name}?`)) return;

        try {
            await adminService.updateUserStatus(user.id, newStatus);
            toast.success('Success', `User ${user.name} has been ${newStatus.toLowerCase()}.`);
            refetch();
        } catch (error) {
            console.error(`Failed to ${action} user`, error);
        }
    };

    const openRoleModal = (user: AdminUser) => {
        setSelectedUser(user);
        setIsRoleModalOpen(true);
    };

    const columns: Column<AdminUser>[] = [
        {
            key: 'name',
            header: 'User',
            sortable: true,
            render: (row) => (
                <div>
                    <div className="font-medium text-gray-900">{row.name}</div>
                    <div className="text-xs text-gray-500">{row.email}</div>
                </div>
            )
        },
        {
            key: 'tenant_id',
            header: 'Organization',
            render: (row) => (
                <span className="text-sm text-gray-700">
                    {tenantMap.get(row.tenant_id) || row.tenant_id}
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
            key: 'role_ids',
            header: 'Roles',
            render: (row) => (
                <div className="flex flex-wrap gap-1">
                    {row.role_ids?.length > 0 ? (
                        row.role_ids.map(roleId => (
                            <span key={roleId} className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800">
                                {roleId}
                            </span>
                        ))
                    ) : (
                        <span className="text-xs text-gray-400">No roles</span>
                    )}
                </div>
            )
        },
        {
            key: 'last_login_at',
            header: 'Last Login',
            render: (row) => row.last_login_at ? new Date(row.last_login_at).toLocaleDateString() : 'Never'
        },
        {
            key: 'id',
            header: 'Actions',
            render: (row) => (
                <div className="flex items-center space-x-2">
                    <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => openRoleModal(row)}
                        title="Manage Roles"
                    >
                        <Shield size={14} className="text-blue-600" />
                    </Button>
                    <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleStatusChange(row)}
                        title={row.status === 'SUSPENDED' ? 'Activate User' : 'Suspend User'}
                    >
                        {row.status === 'SUSPENDED' ? (
                            <CheckCircle size={14} className="text-green-600" />
                        ) : (
                            <Ban size={14} className="text-red-600" />
                        )}
                    </Button>
                </div>
            )
        }
    ];

    return (
        <div className="p-6">
            <div className="mb-6">
                <h1 className="text-2xl font-bold text-gray-900">Platform Users</h1>
                <p className="text-gray-500 mt-1">Manage users across all organizations</p>
            </div>

            <div className="flex flex-col sm:flex-row space-y-3 sm:space-y-0 sm:space-x-4 mb-6">
                <div className="relative flex-1 max-w-md">
                    <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                        <Search size={16} className="text-gray-400" />
                    </div>
                    <input
                        type="text"
                        placeholder="Search by name or email..."
                        className="block w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:ring-1 focus:ring-blue-500 sm:text-sm"
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                    />
                </div>

                <select
                    className="block w-48 pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm rounded-md"
                    value={selectedTenant}
                    onChange={(e) => setSelectedTenant(e.target.value)}
                >
                    <option value="">All Organizations</option>
                    {tenantsData?.items?.map((tenant: Tenant) => (
                        <option key={tenant.id} value={tenant.id}>
                            {tenant.name}
                        </option>
                    ))}
                </select>

                <select
                    className="block w-40 pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm rounded-md"
                    value={statusFilter}
                    onChange={(e) => setStatusFilter(e.target.value)}
                >
                    <option value="">All Statuses</option>
                    <option value="ACTIVE">Active</option>
                    <option value="INVITED">Invited</option>
                    <option value="SUSPENDED">Suspended</option>
                </select>
            </div>

            <div className="bg-white rounded-lg shadow border border-gray-200">
                <DataTable
                    columns={columns}
                    data={data?.items || []}
                    isLoading={isLoading}
                    keyExtractor={(row) => row.id}
                    emptyTitle="No users found"
                    emptyDescription="Try adjusting your search or filters."
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

            {isRoleModalOpen && selectedUser && (
                <RoleAssignModal
                    isOpen={isRoleModalOpen}
                    onClose={() => setIsRoleModalOpen(false)}
                    onSuccess={() => refetch()}
                    user={selectedUser}
                />
            )}
        </div>
    );
}
