import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
    Shield, Plus, Trash2, Edit2, Archive, X
} from 'lucide-react';
import { adminService } from '@/services/adminService';
import { toast, Button, DataTable, StatusBadge, type Column } from '@datalens/shared';
import type { RetentionPolicy } from '@/types/admin';

export default function RetentionPolicies() {
    const queryClient = useQueryClient();
    const [selectedTenant, setSelectedTenant] = useState<string>('');
    const [isCreateOpen, setIsCreateOpen] = useState(false);
    const [editingPolicy, setEditingPolicy] = useState<RetentionPolicy | null>(null);

    // Fetch Tenants for dropdown
    const { data: tenants } = useQuery({
        queryKey: ['admin-tenants-list'],
        queryFn: async () => {
            const res = await adminService.getTenants({ page: 1, limit: 100 });
            return res.items; // Simplified for now, should handle all tenants
        },
    });

    // Fetch Policies
    const { data: policies, isLoading } = useQuery({
        queryKey: ['retention-policies', selectedTenant],
        queryFn: () => adminService.getRetentionPolicies(selectedTenant),
        enabled: !!selectedTenant,
    });

    // Mutations
    const createMut = useMutation({
        mutationFn: adminService.createRetentionPolicy,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['retention-policies', selectedTenant] });
            toast.success('Policy created');
            setIsCreateOpen(false);
        },
        onError: () => toast.error('Failed to create policy'),
    });

    const updateMut = useMutation({
        mutationFn: (data: { id: string; policy: Partial<RetentionPolicy> }) =>
            adminService.updateRetentionPolicy(data.id, data.policy),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['retention-policies', selectedTenant] });
            toast.success('Policy updated');
            setEditingPolicy(null);
        },
        onError: () => toast.error('Failed to update policy'),
    });

    const deleteMut = useMutation({
        mutationFn: adminService.deleteRetentionPolicy,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['retention-policies', selectedTenant] });
            toast.success('Policy deleted');
        },
        onError: () => toast.error('Failed to delete policy'),
    });

    const columns: Column<RetentionPolicy>[] = [
        {
            key: 'purpose_id',
            header: 'Purpose ID',
            render: row => <span className="font-mono text-xs">{row.purpose_id}</span>
        },
        {
            key: 'max_retention_days',
            header: 'Retention (Days)',
            render: row => <span className="font-medium">{row.max_retention_days} days</span>
        },
        {
            key: 'data_categories',
            header: 'Categories',
            render: row => (
                <div className="flex gap-1 flex-wrap max-w-xs">
                    {row.data_categories.map(c => (
                        <span key={c} className="px-1.5 py-0.5 bg-gray-100 text-gray-700 text-xs rounded border border-gray-200">
                            {c}
                        </span>
                    ))}
                </div>
            )
        },
        {
            key: 'status',
            header: 'Status',
            render: row => <StatusBadge label={row.status} />
        },
        {
            key: 'auto_erase',
            header: 'Auto-Erase',
            render: row => (
                <span className={`text-xs font-semibold ${row.auto_erase ? 'text-red-600' : 'text-gray-500'}`}>
                    {row.auto_erase ? 'ENABLED' : 'MANUAL'}
                </span>
            )
        },
        {
            key: 'actions',
            header: '',
            render: row => (
                <div className="flex items-center gap-2 justify-end">
                    <button
                        onClick={() => setEditingPolicy(row)}
                        className="p-1 text-gray-500 hover:text-blue-600 transition-colors"
                        title="Edit Policy"
                    >
                        <Edit2 size={16} />
                    </button>
                    <button
                        onClick={() => {
                            if (confirm('Are you sure you want to delete this policy?')) {
                                deleteMut.mutate(row.id);
                            }
                        }}
                        className="p-1 text-gray-500 hover:text-red-600 transition-colors"
                        title="Delete Policy"
                    >
                        <Trash2 size={16} />
                    </button>
                </div>
            )
        }
    ];

    if (!selectedTenant && tenants && tenants.length > 0) {
        // Auto-select first tenant for convenience if none selected
        // In a real app we might want a "Select Tenant" empty state
    }

    return (
        <div className="p-8 max-w-7xl mx-auto">
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-8">
                <div>
                    <h1 className="text-2xl font-bold text-[var(--text-primary)] flex items-center gap-2">
                        <Archive className="text-[var(--accent-primary)]" />
                        Retention Policies
                    </h1>
                    <p className="text-[var(--text-secondary)] mt-1">
                        Configure data retention rules per tenant and purpose (DPDP Rule R8).
                    </p>
                </div>
                <div className="flex items-center gap-3">
                    <select
                        className="bg-[var(--bg-secondary)] border border-[var(--border-primary)] rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--accent-primary)]"
                        value={selectedTenant}
                        onChange={e => setSelectedTenant(e.target.value)}
                    >
                        <option value="" disabled>Select Tenant to Manage</option>
                        {tenants?.map(t => (
                            <option key={t.id} value={t.id}>{t.name}</option>
                        ))}
                    </select>
                    <Button
                        variant="primary"
                        disabled={!selectedTenant}
                        onClick={() => setIsCreateOpen(true)}
                        icon={<Plus size={18} />}
                    >
                        Create Policy
                    </Button>
                </div>
            </div>

            {!selectedTenant ? (
                <div className="text-center py-20 bg-[var(--bg-secondary)] rounded-xl border border-[var(--border-primary)] border-dashed">
                    <Shield className="mx-auto h-12 w-12 text-[var(--text-secondary)] opacity-50 mb-4" />
                    <h3 className="text-lg font-medium text-[var(--text-primary)]">Select a Tenant</h3>
                    <p className="text-[var(--text-secondary)]">Please select a tenant from the dropdown above to manage retention policies.</p>
                </div>
            ) : (
                <div className="bg-[var(--bg-secondary)] rounded-xl border border-[var(--border-primary)] shadow-sm overflow-hidden">
                    <DataTable
                        columns={columns}
                        data={policies ?? []}
                        isLoading={isLoading}
                        keyExtractor={row => row.id}
                        emptyTitle="No retention policies"
                        emptyDescription="Create a policy to define how long data should be kept."
                    />
                </div>
            )}

            {/* Create/Edit Modal (Simplified as inline conditional rendering for now) */}
            {(isCreateOpen || editingPolicy) && (
                <PolicyModal
                    policy={editingPolicy}
                    onClose={() => { setIsCreateOpen(false); setEditingPolicy(null); }}
                    onSave={(data) => {
                        if (editingPolicy) {
                            updateMut.mutate({ id: editingPolicy.id, policy: data });
                        } else {
                            createMut.mutate({ ...data, tenant_id: selectedTenant });
                        }
                    }}
                    saving={createMut.isPending || updateMut.isPending}
                />
            )}
        </div>
    );
}

function PolicyModal({
    policy, onClose, onSave, saving
}: {
    policy: RetentionPolicy | null;
    onClose: () => void;
    onSave: (data: any) => void;
    saving: boolean;
}) {
    // Form state
    const [purposeId, setPurposeId] = useState(policy?.purpose_id ?? '');
    const [days, setDays] = useState(policy?.max_retention_days ?? 365);
    const [categories, setCategories] = useState(policy?.data_categories?.join(', ') ?? '');
    const [status, setStatus] = useState<RetentionPolicy['status']>(policy?.status ?? 'ACTIVE');
    const [autoErase, setAutoErase] = useState(policy?.auto_erase ?? false);
    const [desc, setDesc] = useState(policy?.description ?? '');

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSave({
            purpose_id: purposeId,
            max_retention_days: Number(days),
            data_categories: categories.split(',').map(s => s.trim()).filter(Boolean),
            status,
            auto_erase: autoErase,
            description: desc,
        });
    };

    return (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50 backdrop-blur-sm">
            <div className="bg-[var(--bg-primary)] rounded-xl shadow-xl w-full max-w-md overflow-hidden border border-[var(--border-primary)]">
                <div className="p-4 border-b border-[var(--border-primary)] flex justify-between items-center bg-[var(--bg-secondary)]">
                    <h2 className="font-semibold text-[var(--text-primary)]">
                        {policy ? 'Edit Retention Policy' : 'New Retention Policy'}
                    </h2>
                    <button onClick={onClose} className="text-[var(--text-secondary)] hover:text-[var(--text-primary)]">
                        <X size={20} />
                    </button>
                </div>
                <form onSubmit={handleSubmit} className="p-6 space-y-4">
                    <div>
                        <label className="block text-xs font-semibold text-[var(--text-secondary)] uppercase mb-1">Purpose ID</label>
                        <input
                            required
                            className="w-full px-3 py-2 bg-[var(--bg-secondary)] border border-[var(--border-primary)] rounded-lg text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-primary)]"
                            placeholder="e.g., pur_marketing_01"
                            value={purposeId}
                            onChange={e => setPurposeId(e.target.value)}
                        />
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <label className="block text-xs font-semibold text-[var(--text-secondary)] uppercase mb-1">Retention Days</label>
                            <input
                                type="number"
                                required
                                min={1}
                                className="w-full px-3 py-2 bg-[var(--bg-secondary)] border border-[var(--border-primary)] rounded-lg text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-primary)]"
                                value={days}
                                onChange={e => setDays(Number(e.target.value))}
                            />
                        </div>
                        <div>
                            <label className="block text-xs font-semibold text-[var(--text-secondary)] uppercase mb-1">Status</label>
                            <select
                                className="w-full px-3 py-2 bg-[var(--bg-secondary)] border border-[var(--border-primary)] rounded-lg text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-primary)]"
                                value={status}
                                onChange={e => setStatus(e.target.value as any)}
                            >
                                <option value="ACTIVE">Active</option>
                                <option value="PAUSED">Paused</option>
                            </select>
                        </div>
                    </div>
                    <div>
                        <label className="block text-xs font-semibold text-[var(--text-secondary)] uppercase mb-1">Data Categories</label>
                        <input
                            className="w-full px-3 py-2 bg-[var(--bg-secondary)] border border-[var(--border-primary)] rounded-lg text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-primary)]"
                            placeholder="contact, financial (comma separated)"
                            value={categories}
                            onChange={e => setCategories(e.target.value)}
                        />
                    </div>
                    <div>
                        <label className="block text-xs font-semibold text-[var(--text-secondary)] uppercase mb-1">Description</label>
                        <input
                            className="w-full px-3 py-2 bg-[var(--bg-secondary)] border border-[var(--border-primary)] rounded-lg text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-primary)]"
                            placeholder="Optional description"
                            value={desc}
                            onChange={e => setDesc(e.target.value)}
                        />
                    </div>
                    <div className="flex items-center gap-3 p-3 bg-[var(--bg-secondary)] rounded-lg border border-[var(--border-primary)]">
                        <input
                            type="checkbox"
                            checked={autoErase}
                            onChange={e => setAutoErase(e.target.checked)}
                            className="h-4 w-4 rounded border-gray-300 text-[var(--accent-primary)] focus:ring-[var(--accent-primary)]"
                            id="autoErase"
                        />
                        <label htmlFor="autoErase" className="text-sm text-[var(--text-primary)] cursor-pointer select-none">
                            <span className="font-medium">Enable Auto-Erase</span>
                            <span className="block text-xs text-[var(--text-secondary)]">Automatically trigger erasure DSR on expiry</span>
                        </label>
                    </div>

                    <div className="flex justify-end gap-3 pt-2">
                        <Button variant="ghost" type="button" onClick={onClose}>Cancel</Button>
                        <Button variant="primary" type="submit" disabled={saving}>
                            {saving ? 'Saving...' : 'Save Policy'}
                        </Button>
                    </div>
                </form>
            </div>
        </div>
    );
}
