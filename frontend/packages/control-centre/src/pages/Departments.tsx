import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Building2, Plus, Trash2, Edit2, Mail, Check, X } from 'lucide-react';
import { departmentService, type Department } from '../services/departmentService';
import {
    DataTable,
    toast,
    Button,
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogFooter,
    Input,
    Label,
} from '@datalens/shared';

type DepartmentForm = Partial<Department>;

const INITIAL_FORM_STATE: DepartmentForm = {
    name: '',
    description: '',
    owner_name: '',
    owner_email: '',
    responsibilities: [],
    notification_enabled: false,
};

export default function Departments() {
    const queryClient = useQueryClient();
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [isNotifyOpen, setIsNotifyOpen] = useState(false);
    const [editingId, setEditingId] = useState<string | null>(null);
    const [notifyId, setNotifyId] = useState<string | null>(null);

    const [formData, setFormData] = useState<DepartmentForm>(INITIAL_FORM_STATE);
    const [notifyData, setNotifyData] = useState({ subject: '', body: '' });
    const [responsibilitiesInput, setResponsibilitiesInput] = useState('');

    const { data: departments = [], isLoading } = useQuery({
        queryKey: ['departments'],
        queryFn: departmentService.list,
    });

    const createMutation = useMutation({
        mutationFn: departmentService.create,
        onSuccess: () => {
            toast.success('Department Created', 'The department was successfully added.');
            handleCloseModal();
            queryClient.invalidateQueries({ queryKey: ['departments'] });
        },
        onError: () => toast.error('Error', 'Failed to create department.'),
    });

    const updateMutation = useMutation({
        mutationFn: ({ id, data }: { id: string; data: Partial<Department> }) =>
            departmentService.update(id, data),
        onSuccess: () => {
            toast.success('Department Updated', 'The department was successfully updated.');
            handleCloseModal();
            queryClient.invalidateQueries({ queryKey: ['departments'] });
        },
        onError: () => toast.error('Error', 'Failed to update department.'),
    });

    const deleteMutation = useMutation({
        mutationFn: departmentService.remove,
        onSuccess: () => {
            toast.success('Department Deleted', 'The department was removed.');
            queryClient.invalidateQueries({ queryKey: ['departments'] });
        },
        onError: () => toast.error('Error', 'Failed to delete department.'),
    });

    const notifyMutation = useMutation({
        mutationFn: ({ id, subject, body }: { id: string; subject: string; body: string }) =>
            departmentService.notify(id, subject, body),
        onSuccess: () => {
            toast.success('Notification Sent', 'Email has been dispatched to the owner.');
            handleCloseNotify();
        },
        onError: () => toast.error('Error', 'Failed to send notification.'),
    });

    const handleOpenCreate = () => {
        setFormData(INITIAL_FORM_STATE);
        setResponsibilitiesInput('');
        setEditingId(null);
        setIsModalOpen(true);
    };

    const handleOpenEdit = (dept: Department) => {
        setFormData({
            name: dept.name,
            description: dept.description,
            owner_name: dept.owner_name,
            owner_email: dept.owner_email,
            notification_enabled: dept.notification_enabled,
        });
        setResponsibilitiesInput((dept.responsibilities || []).join(', '));
        setEditingId(dept.id);
        setIsModalOpen(true);
    };

    const handleOpenNotify = (dept: Department) => {
        setNotifyData({ subject: '', body: '' });
        setNotifyId(dept.id);
        setIsNotifyOpen(true);
    };

    const handleCloseModal = () => {
        setIsModalOpen(false);
        setEditingId(null);
    };

    const handleCloseNotify = () => {
        setIsNotifyOpen(false);
        setNotifyId(null);
    };

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        const finalData = {
            ...formData,
            responsibilities: responsibilitiesInput.split(',').map((s) => s.trim()).filter(Boolean),
        };

        if (editingId) {
            updateMutation.mutate({ id: editingId, data: finalData });
        } else {
            createMutation.mutate(finalData);
        }
    };

    const handleNotifySubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (!notifyId) return;
        notifyMutation.mutate({ id: notifyId, ...notifyData });
    };

    const columns = [
        {
            key: 'name',
            header: 'Name',
            render: (row: Department) => (
                <div className="font-medium text-gray-900">{row.name}</div>
            ),
        },
        {
            key: 'owner_name',
            header: 'Owner',
            render: (row: Department) => (
                <div className="text-sm text-gray-700">{row.owner_name || '—'}</div>
            ),
        },
        {
            key: 'owner_email',
            header: 'Email',
            render: (row: Department) => (
                <div className="text-sm text-gray-500">{row.owner_email || '—'}</div>
            ),
        },
        {
            key: 'responsibilities',
            header: 'Responsibilities',
            render: (row: Department) => (
                <div className="flex flex-wrap gap-1 max-w-[250px]">
                    {row.responsibilities?.map((item, idx) => (
                        <span key={idx} className="bg-blue-50 text-blue-700 px-2 py-0.5 rounded-sm text-xs border border-blue-100 whitespace-nowrap">
                            {item}
                        </span>
                    )) || <span className="text-gray-400 text-xs italic">None</span>}
                </div>
            ),
        },
        {
            key: 'notification_enabled',
            header: 'Notif',
            render: (row: Department) => (
                <div className="flex justify-center">
                    {row.notification_enabled ? (
                        <Check className="text-green-600" size={16} />
                    ) : (
                        <X className="text-gray-400" size={16} />
                    )}
                </div>
            ),
        },
        {
            key: 'actions',
            header: 'Actions',
            render: (row: Department) => (
                <div className="flex justify-end gap-2 text-gray-500">
                    <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleOpenEdit(row)}
                        title="Edit Department"
                    >
                        <Edit2 size={16} />
                    </Button>
                    <Button
                        variant="ghost"
                        size="sm"
                        className="text-red-600 hover:text-red-700 hover:bg-red-50"
                        onClick={() => {
                            if (window.confirm('Are you sure you want to delete this department?')) {
                                deleteMutation.mutate(row.id);
                            }
                        }}
                        title="Delete Department"
                    >
                        <Trash2 size={16} />
                    </Button>
                    {row.notification_enabled && row.owner_email && (
                        <Button
                            variant="ghost"
                            size="sm"
                            className="text-blue-600 hover:text-blue-700 hover:bg-blue-50"
                            onClick={() => handleOpenNotify(row)}
                            title="Send Notification"
                        >
                            <Mail size={16} />
                        </Button>
                    )}
                </div>
            ),
        },
    ];

    return (
        <div className="p-6 max-w-7xl mx-auto py-8">
            <div className="flex justify-between items-center mb-8">
                <div>
                    <h1 className="text-3xl font-bold text-gray-900 flex items-center gap-3">
                        <Building2 className="text-blue-600 h-8 w-8" />
                        Departments
                    </h1>
                    <p className="text-gray-500 mt-2 text-lg">Manage organizational departments and assign compliance owners.</p>
                </div>
                <Button onClick={handleOpenCreate} icon={<Plus size={16} />}>
                    New Department
                </Button>
            </div>

            <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                <DataTable
                    columns={columns}
                    data={departments}
                    isLoading={isLoading}
                    keyExtractor={(row) => row.id}
                    emptyTitle="Create your first department"
                    emptyDescription="Set up departments to assign ownership and map data flows."
                />
            </div>

            {/* Create/Edit Dialog */}
            <Dialog open={isModalOpen} onOpenChange={setIsModalOpen}>
                <DialogContent className="sm:max-w-[600px]">
                    <DialogHeader>
                        <DialogTitle>{editingId ? 'Edit Department' : 'Create New Department'}</DialogTitle>
                    </DialogHeader>

                    <form onSubmit={handleSubmit} className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label htmlFor="name">Name <span className="text-red-500">*</span></Label>
                            <Input
                                id="name"
                                required
                                value={formData.name || ''}
                                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                placeholder="e.g. Engineering"
                            />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="description">Description</Label>
                            <textarea
                                id="description"
                                value={formData.description || ''}
                                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                                placeholder="Core responsibilities and functions"
                                className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                            />
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="owner_name">Owner Name</Label>
                                <Input
                                    id="owner_name"
                                    value={formData.owner_name || ''}
                                    onChange={(e) => setFormData({ ...formData, owner_name: e.target.value })}
                                    placeholder="e.g. Jane Doe"
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="owner_email">Owner Email</Label>
                                <Input
                                    id="owner_email"
                                    type="email"
                                    value={formData.owner_email || ''}
                                    onChange={(e) => setFormData({ ...formData, owner_email: e.target.value })}
                                    placeholder="e.g. jane@acme.com"
                                />
                            </div>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="responsibilitiesInput">Responsibilities (comma-separated)</Label>
                            <Input
                                id="responsibilitiesInput"
                                value={responsibilitiesInput}
                                onChange={(e) => setResponsibilitiesInput(e.target.value)}
                                placeholder="e.g. Data Privacy, GDPR Compliance, Audits"
                            />
                            {responsibilitiesInput && (
                                <div className="flex flex-wrap gap-1 mt-2">
                                    {responsibilitiesInput.split(',').map((s) => s.trim()).filter(Boolean).map((cat, idx) => (
                                        <span key={idx} className="bg-blue-50 text-blue-700 px-2 py-0.5 rounded-sm text-xs border border-blue-100">
                                            {cat}
                                        </span>
                                    ))}
                                </div>
                            )}
                        </div>

                        <div className="flex items-center space-x-2 pt-2">
                            <input
                                type="checkbox"
                                id="notification_enabled"
                                checked={formData.notification_enabled}
                                onChange={(e) => setFormData({ ...formData, notification_enabled: e.target.checked })}
                                className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-600"
                            />
                            <Label htmlFor="notification_enabled" className="cursor-pointer">Enable Email Notifications</Label>
                        </div>

                        <DialogFooter className="pt-4">
                            <Button type="button" variant="secondary" onClick={handleCloseModal}>
                                Cancel
                            </Button>
                            <Button type="submit" disabled={createMutation.isPending || updateMutation.isPending}>
                                {editingId ? 'Save Changes' : 'Create Department'}
                            </Button>
                        </DialogFooter>
                    </form>
                </DialogContent>
            </Dialog>

            {/* Notify Dialog */}
            <Dialog open={isNotifyOpen} onOpenChange={setIsNotifyOpen}>
                <DialogContent className="sm:max-w-[500px]">
                    <DialogHeader>
                        <DialogTitle>Send Notification to Owner</DialogTitle>
                    </DialogHeader>

                    <form onSubmit={handleNotifySubmit} className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label htmlFor="notify_subject">Subject <span className="text-red-500">*</span></Label>
                            <Input
                                id="notify_subject"
                                required
                                value={notifyData.subject}
                                onChange={(e) => setNotifyData({ ...notifyData, subject: e.target.value })}
                                placeholder="Action Required: Compliance Review"
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="notify_body">Message Body <span className="text-red-500">*</span></Label>
                            <textarea
                                id="notify_body"
                                required
                                value={notifyData.body}
                                onChange={(e) => setNotifyData({ ...notifyData, body: e.target.value })}
                                placeholder="Please review the latest compliance data..."
                                className="flex min-h-[150px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                            />
                        </div>
                        <DialogFooter className="pt-4">
                            <Button type="button" variant="secondary" onClick={handleCloseNotify}>
                                Cancel
                            </Button>
                            <Button type="submit" disabled={notifyMutation.isPending}>
                                Send Email
                            </Button>
                        </DialogFooter>
                    </form>
                </DialogContent>
            </Dialog>
        </div>
    );
}
