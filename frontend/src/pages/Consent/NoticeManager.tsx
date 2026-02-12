import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Edit2, UploadCloud, Archive, Link as LinkIcon } from 'lucide-react';
import { DataTable } from '../../components/DataTable/DataTable';
import { Button } from '../../components/common/Button';
import { Modal } from '../../components/common/Modal';
import { StatusBadge } from '../../components/common/StatusBadge';
import { NoticeForm } from '../../components/Consent/NoticeForm';
import { consentService } from '../../services/consent';
import type { ConsentNotice } from '../../types/consent';
import { toast } from 'react-toastify';
import { format } from 'date-fns';

export default function NoticeManager() {
    const [isCreateOpen, setCreateOpen] = useState(false);
    const [editingNotice, setEditingNotice] = useState<ConsentNotice | null>(null);
    const [bindingNotice, setBindingNotice] = useState<ConsentNotice | null>(null);
    const queryClient = useQueryClient();

    const { data: notices = [], isLoading } = useQuery({
        queryKey: ['consent-notices'],
        queryFn: consentService.listNotices,
    });

    const publishMutation = useMutation({
        mutationFn: consentService.publishNotice,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['consent-notices'] });
            toast.success('Notice published successfully');
        },
        onError: () => toast.error('Failed to publish notice'),
    });

    const archiveMutation = useMutation({
        mutationFn: consentService.archiveNotice,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['consent-notices'] });
            toast.success('Notice archived');
        },
    });

    const columns = [
        {
            key: 'title',
            header: 'Title',
            sortable: true,
            render: (row: ConsentNotice) => (
                <div>
                    <div className="font-medium text-gray-900">{row.title}</div>
                    <div className="text-xs text-gray-500">v{row.version}</div>
                </div>
            ),
        },
        {
            key: 'regulation',
            header: 'Regulation',
            sortable: true,
        },
        {
            key: 'status',
            header: 'Status',
            render: (row: ConsentNotice) => <StatusBadge label={row.status} />,
        },
        {
            key: 'updated_at',
            header: 'Last Updated',
            render: (row: ConsentNotice) => (
                <span className="text-sm text-gray-600">
                    {format(new Date(row.updated_at || new Date()), 'MMM d, yyyy')}
                </span>
            ),
        },
        {
            key: 'actions',
            header: 'Actions',
            render: (row: ConsentNotice) => (
                <div className="flex space-x-2">
                    {row.status === 'DRAFT' && (
                        <>
                            <Button
                                size="sm"
                                variant="secondary"
                                icon={<Edit2 size={14} />}
                                onClick={() => setEditingNotice(row)}
                            >
                                Edit
                            </Button>
                            <Button
                                size="sm"
                                variant="primary"
                                icon={<UploadCloud size={14} />}
                                onClick={() => {
                                    if (confirm('Are you sure you want to publish this notice? This will increment the version.')) {
                                        publishMutation.mutate(row.id);
                                    }
                                }}
                            >
                                Publish
                            </Button>
                        </>
                    )}
                    <Button
                        size="sm"
                        variant="secondary"
                        icon={<LinkIcon size={14} />}
                        onClick={() => setBindingNotice(row)}
                        title="Bind to Widgets"
                    />
                    {row.status !== 'ARCHIVED' && (
                        <Button
                            size="sm"
                            variant="danger"
                            icon={<Archive size={14} />}
                            onClick={() => {
                                if (confirm('Archive this notice?')) {
                                    archiveMutation.mutate(row.id);
                                }
                            }}
                            title="Archive"
                        />
                    )}
                </div>
            ),
        },
    ];

    return (
        <div className="p-6">
            <div className="flex justify-between items-center mb-6">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900">Privacy Notices</h1>
                    <p className="text-gray-500">Manage privacy notices and consent policies.</p>
                </div>
                <Button icon={<Plus size={16} />} onClick={() => setCreateOpen(true)}>
                    Create Notice
                </Button>
            </div>

            <DataTable
                columns={columns}
                data={notices}
                isLoading={isLoading}
                keyExtractor={(row) => row.id}
                emptyTitle="No notices found"
                emptyDescription="Create a privacy notice to start collecting consent."
            />

            {/* Create/Edit Modal */}
            <Modal
                open={isCreateOpen || !!editingNotice}
                onClose={() => {
                    setCreateOpen(false);
                    setEditingNotice(null);
                }}
                title={editingNotice ? 'Edit Notice' : 'Create Notice'}
            >
                <NoticeForm
                    initialData={editingNotice || undefined}
                    onSuccess={() => {
                        setCreateOpen(false);
                        setEditingNotice(null);
                    }}
                    onCancel={() => {
                        setCreateOpen(false);
                        setEditingNotice(null);
                    }}
                />
            </Modal>

            {/* Bind Modal (Placeholder for now, or simple implementation) */}
            <Modal
                open={!!bindingNotice}
                onClose={() => setBindingNotice(null)}
                title="Bind to Widgets"
            >
                <BindWidgetsForm
                    notice={bindingNotice}
                    onClose={() => setBindingNotice(null)}
                />
            </Modal>
        </div>
    );
}

// Simple Bind Form Component
function BindWidgetsForm({ notice, onClose }: { notice: ConsentNotice | null; onClose: () => void }) {
    const queryClient = useQueryClient();
    const { data: widgets = [] } = useQuery({
        queryKey: ['consent-widgets'],
        queryFn: () => consentService.listWidgets().then(res => res.items),
    });

    const [selectedWidgets, setSelectedWidgets] = useState<string[]>(notice?.widget_ids || []);

    const bindMutation = useMutation({
        mutationFn: (ids: string[]) => consentService.bindNotice(notice!.id, ids),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['consent-notices'] });
            toast.success('Notice bound to widgets');
            onClose();
        },
        onError: () => toast.error('Failed to bind notice'),
    });

    if (!notice) return null;

    return (
        <div className="space-y-4">
            <p className="text-sm text-gray-600">
                Select widgets to display <strong>{notice.title}</strong> (v{notice.version}).
            </p>
            <div className="max-h-60 overflow-y-auto border rounded p-2">
                {widgets.map(w => (
                    <label key={w.id} className="flex items-center space-x-2 py-1">
                        <input
                            type="checkbox"
                            checked={selectedWidgets.includes(w.id)}
                            onChange={(e) => {
                                if (e.target.checked) setSelectedWidgets([...selectedWidgets, w.id]);
                                else setSelectedWidgets(selectedWidgets.filter(id => id !== w.id));
                            }}
                            className="rounded border-gray-300"
                        />
                        <span>{w.name} ({w.domain})</span>
                    </label>
                ))}
            </div>
            <div className="flex justify-end space-x-2">
                <Button variant="secondary" onClick={onClose}>Cancel</Button>
                <Button onClick={() => bindMutation.mutate(selectedWidgets)}>Save</Button>
            </div>
        </div>
    );
}
