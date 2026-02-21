import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2, Edit2, ShieldAlert } from 'lucide-react';
import { retentionService } from '../services/retentionService';
import { DataTable, toast, StatusBadge, Button, Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, Input, Label, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@datalens/shared';
import type { RetentionPolicy, CreateRetentionPolicyInput, RetentionStatus } from '../types/retention';

const INITIAL_FORM_STATE: CreateRetentionPolicyInput = {
    purpose_id: '',
    max_retention_days: 30,
    data_categories: [],
    status: 'ACTIVE',
    auto_erase: false,
    description: '',
};

export default function RetentionPolicies() {
    const queryClient = useQueryClient();
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [editingId, setEditingId] = useState<string | null>(null);
    const [formData, setFormData] = useState<CreateRetentionPolicyInput>(INITIAL_FORM_STATE);
    const [categoriesInput, setCategoriesInput] = useState('');

    const { data: policies = [] as RetentionPolicy[], isLoading } = useQuery({
        queryKey: ['retention-policies'],
        queryFn: retentionService.listPolicies,
    });

    const createMutation = useMutation({
        mutationFn: retentionService.createPolicy,
        onSuccess: () => {
            toast.success('Policy Created', 'Retention policy has been successfully added.');
            handleCloseModal();
            queryClient.invalidateQueries({ queryKey: ['retention-policies'] });
        },
        onError: () => toast.error('Error', 'Failed to create policy.'),
    });

    const updateMutation = useMutation({
        mutationFn: ({ id, data }: { id: string; data: Partial<CreateRetentionPolicyInput> }) =>
            retentionService.updatePolicy(id, data),
        onSuccess: () => {
            toast.success('Policy Updated', 'Retention policy has been successfully updated.');
            handleCloseModal();
            queryClient.invalidateQueries({ queryKey: ['retention-policies'] });
        },
        onError: () => toast.error('Error', 'Failed to update policy.'),
    });

    const deleteMutation = useMutation({
        mutationFn: retentionService.deletePolicy,
        onSuccess: () => {
            toast.success('Policy Deleted', 'The policy has been removed.');
            queryClient.invalidateQueries({ queryKey: ['retention-policies'] });
        },
        onError: () => toast.error('Error', 'Failed to delete policy.'),
    });

    const handleOpenCreate = () => {
        setFormData(INITIAL_FORM_STATE);
        setCategoriesInput('');
        setEditingId(null);
        setIsModalOpen(true);
    };

    const handleOpenEdit = (policy: RetentionPolicy) => {
        setFormData({
            description: policy.description,
            purpose_id: policy.purpose_id,
            max_retention_days: policy.max_retention_days,
            status: policy.status,
            auto_erase: policy.auto_erase,
            data_categories: policy.data_categories,
        });
        setCategoriesInput(policy.data_categories.join(', '));
        setEditingId(policy.id);
        setIsModalOpen(true);
    };

    const handleCloseModal = () => {
        setIsModalOpen(false);
        setEditingId(null);
    };

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        const finalData = {
            ...formData,
            data_categories: categoriesInput.split(',').map((s) => s.trim()).filter(Boolean),
        };

        if (editingId) {
            updateMutation.mutate({ id: editingId, data: finalData });
        } else {
            createMutation.mutate(finalData);
        }
    };

    const columns = [
        {
            key: 'description',
            header: 'Description',
            render: (row: RetentionPolicy) => (
                <div className="font-medium text-gray-900 max-w-[200px] truncate" title={row.description}>
                    {row.description}
                </div>
            ),
        },
        {
            key: 'purpose_id',
            header: 'Purpose ID',
            render: (row: RetentionPolicy) => (
                <div className="text-gray-500 font-mono text-xs max-w-[120px] truncate" title={row.purpose_id}>
                    {row.purpose_id}
                </div>
            ),
        },
        {
            key: 'retention_period',
            header: 'Retention Period',
            render: (row: RetentionPolicy) => (
                <span className="text-sm text-gray-700">{row.max_retention_days} days</span>
            ),
        },
        {
            key: 'data_categories',
            header: 'Data Categories',
            render: (row: RetentionPolicy) => (
                <div className="flex flex-wrap gap-1">
                    {row.data_categories.map((cat: string, idx: number) => (
                        <span key={idx} className="bg-gray-100 text-gray-600 px-2 py-0.5 rounded-full text-xs">
                            {cat}
                        </span>
                    ))}
                    {row.data_categories.length === 0 && <span className="text-gray-400 text-xs italic">None</span>}
                </div>
            ),
        },
        {
            key: 'status',
            header: 'Status',
            render: (row: RetentionPolicy) => (
                <div className="flex items-center">
                    <StatusBadge label={row.status} variant={row.status === 'ACTIVE' ? 'success' : 'warning'} />
                </div>
            ),
        },
        {
            key: 'auto_erase',
            header: 'Auto-Erase',
            render: (row: RetentionPolicy) => (
                <div className="flex items-center justify-center">
                    <span className={`px-2 py-1 rounded text-xs font-semibold ${row.auto_erase ? 'bg-red-100 text-red-700' : 'bg-gray-100 text-gray-500'}`}>
                        {row.auto_erase ? 'Enabled' : 'Disabled'}
                    </span>
                </div>
            ),
        },
        {
            key: 'actions',
            header: 'Actions',
            render: (row: RetentionPolicy) => (
                <div className="flex justify-end gap-2">
                    <Button
                        variant="secondary"
                        size="sm"
                        onClick={() => handleOpenEdit(row)}
                        icon={<Edit2 size={14} />}
                    >
                        Edit
                    </Button>
                    <Button
                        variant="secondary"
                        size="sm"
                        className="text-red-600 hover:text-red-700 hover:bg-red-50"
                        onClick={() => {
                            if (window.confirm('Are you sure you want to delete this retention policy?')) {
                                deleteMutation.mutate(row.id);
                            }
                        }}
                        icon={<Trash2 size={14} />}
                    >
                        Delete
                    </Button>
                </div>
            ),
        },
    ];

    return (
        <div className="p-6 max-w-7xl mx-auto">
            <div className="flex justify-between items-center mb-8">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
                        <ShieldAlert className="text-blue-600" />
                        Retention Policies
                    </h1>
                    <p className="text-gray-500 mt-1">Manage data retention rules and auto-erasure schedules.</p>
                </div>
                <Button onClick={handleOpenCreate} icon={<Plus size={16} />}>
                    Add Policy
                </Button>
            </div>

            <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
                <DataTable
                    columns={columns}
                    data={policies}
                    isLoading={isLoading}
                    keyExtractor={(row) => row.id}
                    emptyTitle="No retention policies"
                    emptyDescription="Create a policy to start enforcing data retention rules."
                />
            </div>

            <Dialog open={isModalOpen} onOpenChange={setIsModalOpen}>
                <DialogContent className="sm:max-w-[500px]">
                    <DialogHeader>
                        <DialogTitle>{editingId ? 'Edit Policy' : 'Create New Policy'}</DialogTitle>
                    </DialogHeader>

                    <form onSubmit={handleSubmit} className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label htmlFor="description">Description</Label>
                            <Input
                                id="description"
                                required
                                value={formData.description}
                                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                                placeholder="e.g. Inactive User Data Retention"
                            />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="purpose_id">Purpose ID</Label>
                            <Input
                                id="purpose_id"
                                required
                                value={formData.purpose_id}
                                onChange={(e) => setFormData({ ...formData, purpose_id: e.target.value })}
                                placeholder="Enter mapped purpose ID"
                            />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="max_retention_days">Max Retention (Days)</Label>
                            <Input
                                id="max_retention_days"
                                type="number"
                                required
                                min={1}
                                value={formData.max_retention_days}
                                onChange={(e) => setFormData({ ...formData, max_retention_days: parseInt(e.target.value, 10) || 0 })}
                            />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="categoriesInput">Data Categories (comma-separated)</Label>
                            <Input
                                id="categoriesInput"
                                value={categoriesInput}
                                onChange={(e) => setCategoriesInput(e.target.value)}
                                placeholder="e.g. EMAIL, PHONE, BIOMETRIC"
                            />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="status">Status</Label>
                            <Select
                                value={formData.status}
                                onValueChange={(val: RetentionStatus) => setFormData({ ...formData, status: val })}
                            >
                                <SelectTrigger>
                                    <SelectValue placeholder="Select status" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="ACTIVE">Active</SelectItem>
                                    <SelectItem value="PAUSED">Paused</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>

                        <div className="flex items-center space-x-2 pt-2">
                            <input
                                type="checkbox"
                                id="auto_erase"
                                checked={formData.auto_erase}
                                onChange={(e) => setFormData({ ...formData, auto_erase: e.target.checked })}
                                className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-600"
                            />
                            <Label htmlFor="auto_erase" className="cursor-pointer">Enable Auto-Erase on Expiry</Label>
                        </div>

                        <DialogFooter className="pt-4">
                            <Button type="button" variant="secondary" onClick={handleCloseModal}>
                                Cancel
                            </Button>
                            <Button type="submit" disabled={createMutation.isPending || updateMutation.isPending}>
                                {editingId ? 'Save Changes' : 'Create Policy'}
                            </Button>
                        </DialogFooter>
                    </form>
                </DialogContent>
            </Dialog>
        </div>
    );
}
