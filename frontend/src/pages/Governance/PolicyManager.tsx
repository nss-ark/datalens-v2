import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Building2, Trash2 } from 'lucide-react';
import { governanceService } from '../../services/governance';
import { DataTable } from '../../components/DataTable/DataTable';
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
} from '../../components/ui/dialog';
import { Button } from '../../components/ui/button';
import { PolicyForm } from '../../components/Governance/PolicyForm';
import { toast } from '../../stores/toastStore';
import type { GovernancePolicy } from '../../types/governance';

const PolicyManager = () => {
    const queryClient = useQueryClient();
    const [isModalOpen, setIsModalOpen] = useState(false);

    const { data: policies = [], isLoading } = useQuery({
        queryKey: ['policies'],
        queryFn: governanceService.getPolicies,
    });

    const createMutation = useMutation({
        mutationFn: governanceService.createPolicy,
        onSuccess: () => {
            toast.success('Policy Created', 'The new compliance policy has been added.');
            setIsModalOpen(false);
            queryClient.invalidateQueries({ queryKey: ['policies'] });
        },
        onError: () => {
            toast.error('Error', 'Failed to create policy.');
        }
    });

    const deleteMutation = useMutation({
        mutationFn: governanceService.deletePolicy,
        onSuccess: () => {
            toast.success('Policy Deleted', 'The policy has been removed.');
            queryClient.invalidateQueries({ queryKey: ['policies'] });
        },
        onError: () => {
            toast.error('Error', 'Failed to delete policy.');
        }
    });

    const columns = [
        {
            key: 'name',
            header: 'Policy Name',
            sortable: true,
            render: (row: GovernancePolicy) => (
                <div>
                    <div className="font-medium text-gray-900">{row.name}</div>
                    <div className="text-xs text-gray-500">{row.description}</div>
                </div>
            )
        },
        {
            key: 'type',
            header: 'Type',
            sortable: true,
            render: (row: GovernancePolicy) => (
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 capitalize">
                    {row.type}
                </span>
            )
        },
        {
            key: 'isActive',
            header: 'Status',
            sortable: true,
            render: (row: GovernancePolicy) => (
                <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${row.isActive ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                    }`}>
                    {row.isActive ? 'Active' : 'Inactive'}
                </span>
            )
        },
        {
            key: 'createdAt',
            header: 'Created On',
            sortable: true,
            render: (row: GovernancePolicy) => new Date(row.createdAt).toLocaleDateString()
        },
        {
            key: 'actions',
            header: '',
            width: '100px',
            render: (row: GovernancePolicy) => (
                <div className="flex justify-end">
                    <Button
                        variant="ghost"
                        size="icon"
                        onClick={(e) => {
                            e.stopPropagation();
                            if (window.confirm('Are you sure you want to delete this policy?')) {
                                deleteMutation.mutate(row.id);
                            }
                        }}
                        className="text-gray-400 hover:text-red-600 hover:bg-red-50"
                        title="Delete Policy"
                    >
                        <Trash2 size={16} />
                    </Button>
                </div>
            )
        }
    ];

    return (
        <div className="p-6 max-w-7xl mx-auto">
            <div className="flex justify-between items-center mb-8">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
                        <Building2 className="text-blue-600" />
                        Policy Manager
                    </h1>
                    <p className="text-gray-500 mt-1">
                        Define and manage data governance policies.
                    </p>
                </div>
                <Button onClick={() => setIsModalOpen(true)}>
                    <Plus size={16} className="mr-2" />
                    Create Policy
                </Button>
            </div>

            <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
                <DataTable
                    columns={columns}
                    data={policies}
                    isLoading={isLoading}
                    keyExtractor={(row) => row.id}
                    emptyTitle="No policies defined"
                    emptyDescription="Create a new policy to start governing your data."
                />
            </div>

            <Dialog open={isModalOpen} onOpenChange={setIsModalOpen}>
                <DialogContent className="sm:max-w-[500px]">
                    <DialogHeader>
                        <DialogTitle>Create New Policy</DialogTitle>
                    </DialogHeader>
                    <PolicyForm
                        onSubmit={(data) => createMutation.mutate(data)}
                        onCancel={() => setIsModalOpen(false)}
                        isLoading={createMutation.isPending}
                    />
                </DialogContent>
            </Dialog>
        </div>
    );
};

export default PolicyManager;
