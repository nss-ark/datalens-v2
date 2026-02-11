import { useState } from 'react';
import { Plus, Play, Pause, Trash2, Code, Loader2 } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { Button } from '../components/common/Button';
import { DataTable } from '../components/DataTable/DataTable';
import { StatusBadge } from '../components/common/StatusBadge';
import { Modal } from '../components/common/Modal';
import { useWidgets, useDeleteWidget, useActivateWidget, usePauseWidget } from '../hooks/useConsent';
import type { ConsentWidget } from '../types/consent';
import WidgetBuilder from '../components/Consent/WidgetBuilder';

export default function ConsentWidgets() {
    const navigate = useNavigate();
    const [page, setPage] = useState(1);
    const [isBuilderOpen, setIsBuilderOpen] = useState(false);

    // Data Fetching
    const { data, isLoading, isError } = useWidgets({ page, page_size: 10 });
    const deleteMutation = useDeleteWidget();
    const activateMutation = useActivateWidget();
    const pauseMutation = usePauseWidget();

    const handleDelete = async (id: string) => {
        if (confirm('Are you sure you want to delete this widget? This action cannot be undone.')) {
            await deleteMutation.mutateAsync(id);
        }
    };

    const handleToggleStatus = async (widget: ConsentWidget) => {
        if (widget.status === 'ACTIVE') {
            await pauseMutation.mutateAsync(widget.id);
        } else {
            await activateMutation.mutateAsync(widget.id);
        }
    };

    const columns = [
        {
            key: 'name',
            header: 'Name',
            render: (row: ConsentWidget) => (
                <div className="flex flex-col">
                    <span className="font-medium text-gray-900 dark:text-gray-100">{row.name}</span>
                    <span className="text-xs text-gray-500">{row.domain}</span>
                </div>
            ),
        },
        {
            key: 'type',
            header: 'Type',
            render: (row: ConsentWidget) => <span className="text-sm text-gray-600 dark:text-gray-400">{row.type.replace('_', ' ')}</span>,
        },
        {
            key: 'status',
            header: 'Status',
            render: (row: ConsentWidget) => {
                let variant: 'success' | 'warning' | 'neutral' | 'error' = 'neutral';
                if (row.status === 'ACTIVE') variant = 'success';
                if (row.status === 'PAUSED') variant = 'warning';
                if (row.status === 'DRAFT') variant = 'neutral';
                return <StatusBadge label={row.status} variant={variant} />;
            },
        },
        {
            key: 'created_at',
            header: 'Created',
            render: (row: ConsentWidget) => <span className="text-sm text-gray-500">{row.created_at ? new Date(row.created_at).toLocaleDateString() : '-'}</span>,
        },
        {
            key: 'actions',
            header: 'Actions',
            render: (row: ConsentWidget) => (
                <div className="flex gap-2">
                    <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleToggleStatus(row)}
                        title={row.status === 'ACTIVE' ? "Pause Widget" : "Activate Widget"}
                    >
                        {row.status === 'ACTIVE' ? <Pause size={16} /> : <Play size={16} />}
                    </Button>
                    <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => navigate(`/consent/widgets/${row.id}`)}
                        title="View Code & Analytics"
                    >
                        <Code size={16} />
                    </Button>
                    <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleDelete(row.id)}
                        title="Delete Widget"
                        className="text-red-600 hover:text-red-700 hover:bg-red-50"
                    >
                        <Trash2 size={16} />
                    </Button>
                </div>
            ),
        },
    ];

    if (isError) return <div className="p-8 text-center text-red-500">Failed to load widgets</div>;

    return (
        <div className="p-6 space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Consent Widgets</h1>
                    <p className="text-gray-500 dark:text-gray-400">Manage your consent collection widgets and notices.</p>
                </div>
                <Button onClick={() => setIsBuilderOpen(true)} icon={<Plus size={16} />}>
                    New Widget
                </Button>
            </div>

            {isLoading ? (
                <div className="flex justify-center p-12">
                    <Loader2 className="animate-spin text-blue-600" size={32} />
                </div>
            ) : (
                <>
                    <DataTable<ConsentWidget>
                        columns={columns}
                        data={data?.items || []}
                        keyExtractor={(row) => row.id}
                    />
                    {/* Simple Pagination */}
                    <div className="flex justify-between items-center mt-4">
                        <span className="text-sm text-gray-500">
                            Page {page} of {data?.total_pages || 1}
                        </span>
                        <div className="flex gap-2">
                            <Button
                                variant="secondary"
                                size="sm"
                                disabled={page === 1}
                                onClick={() => setPage(p => Math.max(1, p - 1))}
                            >
                                Previous
                            </Button>
                            <Button
                                variant="secondary"
                                size="sm"
                                disabled={page === (data?.total_pages || 1)}
                                onClick={() => setPage(p => p + 1)}
                            >
                                Next
                            </Button>
                        </div>
                    </div>
                </>
            )}

            {/* Widget Builder Modal */}
            <Modal
                open={isBuilderOpen}
                onClose={() => setIsBuilderOpen(false)}
                title="Create New Widget"
            >
                <WidgetBuilder onClose={() => setIsBuilderOpen(false)} />
            </Modal>
        </div>
    );
}
