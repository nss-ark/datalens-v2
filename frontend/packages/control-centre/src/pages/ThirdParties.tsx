import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Globe, Plus, Trash2, Edit2, ShieldAlert } from 'lucide-react';
import { thirdPartyService, type ThirdParty } from '../services/thirdPartyService';
import { api } from '@datalens/shared';
import {
    DataTable,
    toast,
    StatusBadge,
    Button,
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogFooter,
    Input,
    Label,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@datalens/shared';

type ThirdPartyForm = Partial<ThirdParty>;

const INITIAL_FORM_STATE: ThirdPartyForm = {
    name: '',
    type: 'PROCESSOR',
    country: '',
    purpose_ids: [],
    dpa_status: 'NONE',
    dpa_doc_path: '',
    dpa_signed_at: '',
    dpa_expires_at: '',
    dpa_notes: '',
    contact_name: '',
    contact_email: '',
};

const dpaStatusColors: Record<string, 'success' | 'warning' | 'danger' | 'neutral'> = {
    SIGNED: 'success',
    PENDING: 'warning',
    EXPIRED: 'danger',
    NONE: 'neutral',
};

const typeColors: Record<string, string> = {
    PROCESSOR: 'bg-blue-100 text-blue-800',
    CONTROLLER: 'bg-purple-100 text-purple-800',
    VENDOR: 'bg-orange-100 text-orange-800',
};

export default function ThirdParties() {
    const queryClient = useQueryClient();
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [editingId, setEditingId] = useState<string | null>(null);
    const [viewMode, setViewMode] = useState<'SIMPLE' | 'FULL'>('SIMPLE');

    const [formData, setFormData] = useState<ThirdPartyForm>(INITIAL_FORM_STATE);

    const { data: thirdParties = [], isLoading } = useQuery({
        queryKey: ['third-parties'],
        queryFn: thirdPartyService.list,
    });

    const { data: purposes = [] } = useQuery({
        queryKey: ['purposes'],
        queryFn: async () => {
            try {
                const res = await api.get('/v2/purposes');
                return res.data?.data || [];
            } catch {
                return [];
            }
        },
    });

    const createMutation = useMutation({
        mutationFn: thirdPartyService.create,
        onSuccess: () => {
            toast.success('Third Party Added', 'The third party was successfully added.');
            handleCloseModal();
            queryClient.invalidateQueries({ queryKey: ['third-parties'] });
        },
        onError: () => toast.error('Error', 'Failed to add third party.'),
    });

    const updateMutation = useMutation({
        mutationFn: ({ id, data }: { id: string; data: Partial<ThirdParty> }) =>
            thirdPartyService.update(id, data),
        onSuccess: () => {
            toast.success('Third Party Updated', 'The third party was successfully updated.');
            handleCloseModal();
            queryClient.invalidateQueries({ queryKey: ['third-parties'] });
        },
        onError: () => toast.error('Error', 'Failed to update third party.'),
    });

    const deleteMutation = useMutation({
        mutationFn: thirdPartyService.remove,
        onSuccess: () => {
            toast.success('Third Party Deleted', 'The third party was removed.');
            queryClient.invalidateQueries({ queryKey: ['third-parties'] });
        },
        onError: () => toast.error('Error', 'Failed to delete third party.'),
    });

    const handleOpenCreate = () => {
        setFormData(INITIAL_FORM_STATE);
        setEditingId(null);
        setIsModalOpen(true);
    };

    const handleOpenEdit = (tp: ThirdParty) => {
        setFormData({
            name: tp.name,
            type: tp.type,
            country: tp.country,
            purpose_ids: tp.purpose_ids || [],
            dpa_status: tp.dpa_status,
            dpa_doc_path: tp.dpa_doc_path,
            dpa_signed_at: tp.dpa_signed_at,
            dpa_expires_at: tp.dpa_expires_at,
            dpa_notes: tp.dpa_notes,
            contact_name: tp.contact_name,
            contact_email: tp.contact_email,
        });
        setEditingId(tp.id);
        setIsModalOpen(true);
    };

    const handleCloseModal = () => {
        setIsModalOpen(false);
        setEditingId(null);
    };

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        const finalData = { ...formData };
        if (!finalData.dpa_signed_at) delete finalData.dpa_signed_at;
        if (!finalData.dpa_expires_at) delete finalData.dpa_expires_at;

        if (editingId) {
            updateMutation.mutate({ id: editingId, data: finalData });
        } else {
            createMutation.mutate(finalData);
        }
    };

    const togglePurposeId = (id: string) => {
        setFormData((prev) => {
            const ids = prev.purpose_ids || [];
            if (ids.includes(id)) {
                return { ...prev, purpose_ids: ids.filter((p) => p !== id) };
            }
            return { ...prev, purpose_ids: [...ids, id] };
        });
    };

    const baseColumns = [
        {
            key: 'name',
            header: 'Name',
            render: (row: ThirdParty) => (
                <div className="font-medium text-gray-900">{row.name}</div>
            ),
        },
        {
            key: 'type',
            header: 'Type',
            render: (row: ThirdParty) => (
                <span className={`px-2 py-1 rounded text-xs font-semibold ${typeColors[row.type] || 'bg-gray-100 text-gray-800'}`}>
                    {row.type}
                </span>
            ),
        },
        {
            key: 'country',
            header: 'Country',
            render: (row: ThirdParty) => (
                <div className="text-sm text-gray-700">{row.country}</div>
            ),
        },
    ];

    const fullDpaColumns = [
        {
            key: 'dpa_status',
            header: 'DPA Status',
            render: (row: ThirdParty) => (
                <StatusBadge label={row.dpa_status} variant={dpaStatusColors[row.dpa_status] || 'neutral'} />
            ),
        },
        {
            key: 'contact_email',
            header: 'Contact Email',
            render: (row: ThirdParty) => (
                <div className="text-sm text-gray-500">{row.contact_email || '—'}</div>
            ),
        },
    ];

    const actionColumn = {
        key: 'actions',
        header: 'Actions',
        render: (row: ThirdParty) => (
            <div className="flex justify-end gap-2 text-gray-500">
                <Button variant="ghost" size="sm" onClick={() => handleOpenEdit(row)} title="Edit">
                    <Edit2 size={16} />
                </Button>
                <Button
                    variant="ghost"
                    size="sm"
                    className="text-red-600 hover:text-red-700 hover:bg-red-50"
                    onClick={() => {
                        if (window.confirm('Are you sure you want to delete this third party?')) {
                            deleteMutation.mutate(row.id);
                        }
                    }}
                    title="Delete"
                >
                    <Trash2 size={16} />
                </Button>
            </div>
        ),
    };

    const columns = viewMode === 'SIMPLE'
        ? [...baseColumns, actionColumn]
        : [...baseColumns, ...fullDpaColumns, actionColumn];

    return (
        <div className="p-6 max-w-7xl mx-auto py-8">
            <div className="flex justify-between items-center mb-8">
                <div>
                    <h1 className="text-3xl font-bold text-gray-900 flex items-center gap-3">
                        <Globe className="text-blue-600 h-8 w-8" />
                        Third Parties
                    </h1>
                    <p className="text-gray-500 mt-2 text-lg">Manage processors, controllers, and vendor DPAs.</p>
                </div>
                <div className="flex items-center gap-4">
                    <div className="flex bg-gray-100 p-1 rounded-md">
                        <button
                            onClick={() => setViewMode('SIMPLE')}
                            className={`px-3 py-1 text-sm rounded ${viewMode === 'SIMPLE' ? 'bg-white shadow' : 'text-gray-500 hover:text-gray-700'}`}
                        >
                            Simple View
                        </button>
                        <button
                            onClick={() => setViewMode('FULL')}
                            className={`px-3 py-1 text-sm rounded ${viewMode === 'FULL' ? 'bg-white shadow' : 'text-gray-500 hover:text-gray-700'}`}
                        >
                            Full DPA
                        </button>
                    </div>
                    <Button onClick={handleOpenCreate} icon={<Plus size={16} />}>
                        Add Third Party
                    </Button>
                </div>
            </div>

            <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                <DataTable
                    columns={columns}
                    data={thirdParties}
                    isLoading={isLoading}
                    keyExtractor={(row) => row.id}
                    emptyTitle="Add your first third party"
                    emptyDescription="Track vendors and their Data Processing Agreements."
                />
            </div>

            <Dialog open={isModalOpen} onOpenChange={setIsModalOpen}>
                <DialogContent className="sm:max-w-[700px] max-h-[90vh] overflow-y-auto">
                    <DialogHeader>
                        <DialogTitle>{editingId ? 'Edit Third Party' : 'Add Third Party'}</DialogTitle>
                    </DialogHeader>

                    <form onSubmit={handleSubmit} className="space-y-6 py-4">
                        {/* Basic Info */}
                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="name">Name <span className="text-red-500">*</span></Label>
                                <Input
                                    id="name"
                                    required
                                    value={formData.name || ''}
                                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                />
                            </div>
                            <div className="space-y-2">
                                <Label>Type</Label>
                                <Select
                                    value={formData.type}
                                    onValueChange={(val: any) => setFormData({ ...formData, type: val })}
                                >
                                    <SelectTrigger>
                                        <SelectValue placeholder="Select type" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="PROCESSOR">Processor</SelectItem>
                                        <SelectItem value="CONTROLLER">Controller</SelectItem>
                                        <SelectItem value="VENDOR">Vendor</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="country">Country <span className="text-red-500">*</span></Label>
                                <Input
                                    id="country"
                                    required
                                    value={formData.country || ''}
                                    onChange={(e) => setFormData({ ...formData, country: e.target.value })}
                                    placeholder="e.g. USA, UK, India"
                                />
                            </div>
                        </div>

                        <hr className="border-gray-100" />

                        {/* Purposes */}
                        <div className="space-y-2">
                            <Label>Purposes</Label>
                            <p className="text-xs text-gray-500 mb-2">Select the purposes this third party is involved in.</p>
                            <div className="grid grid-cols-2 gap-2 max-h-32 overflow-y-auto p-2 border border-gray-200 rounded-md">
                                {purposes.length > 0 ? purposes.map((p: any) => (
                                    <label key={p.id} className="flex items-center gap-2 cursor-pointer">
                                        <input
                                            type="checkbox"
                                            checked={formData.purpose_ids?.includes(p.id) || false}
                                            onChange={() => togglePurposeId(p.id)}
                                            className="rounded border-gray-300 text-blue-600 focus:ring-blue-600"
                                        />
                                        <span className="text-sm truncate" title={p.name}>{p.name}</span>
                                    </label>
                                )) : (
                                    <span className="text-sm text-gray-400">No purposes found.</span>
                                )}
                            </div>
                        </div>

                        <hr className="border-gray-100" />

                        {/* DPA Info */}
                        <div>
                            <h3 className="text-lg font-semibold text-gray-800 mb-4 flex items-center gap-2">
                                <ShieldAlert size={18} className="text-gray-500" /> Data Processing Agreement
                            </h3>
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label>DPA Status</Label>
                                    <Select
                                        value={formData.dpa_status}
                                        onValueChange={(val: any) => setFormData({ ...formData, dpa_status: val })}
                                    >
                                        <SelectTrigger>
                                            <SelectValue placeholder="Select DPA status" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="NONE">None</SelectItem>
                                            <SelectItem value="PENDING">Pending</SelectItem>
                                            <SelectItem value="SIGNED">Signed</SelectItem>
                                            <SelectItem value="EXPIRED">Expired</SelectItem>
                                        </SelectContent>
                                    </Select>
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="dpa_doc_path">Doc Path / URL</Label>
                                    <Input
                                        id="dpa_doc_path"
                                        value={formData.dpa_doc_path || ''}
                                        onChange={(e) => setFormData({ ...formData, dpa_doc_path: e.target.value })}
                                        placeholder="URL or internal reference"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="dpa_signed_at">Signed Date</Label>
                                    <Input
                                        id="dpa_signed_at"
                                        type="date"
                                        value={formData.dpa_signed_at ? formData.dpa_signed_at.substring(0, 10) : ''}
                                        onChange={(e) => setFormData({ ...formData, dpa_signed_at: e.target.value })}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="dpa_expires_at">Expiry Date</Label>
                                    <Input
                                        id="dpa_expires_at"
                                        type="date"
                                        value={formData.dpa_expires_at ? formData.dpa_expires_at.substring(0, 10) : ''}
                                        onChange={(e) => setFormData({ ...formData, dpa_expires_at: e.target.value })}
                                    />
                                </div>
                                <div className="space-y-2 col-span-2">
                                    <Label htmlFor="dpa_notes">DPA Notes</Label>
                                    <textarea
                                        id="dpa_notes"
                                        value={formData.dpa_notes || ''}
                                        onChange={(e) => setFormData({ ...formData, dpa_notes: e.target.value })}
                                        className="flex min-h-[60px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                    />
                                </div>
                            </div>
                        </div>

                        <hr className="border-gray-100" />

                        {/* Contact */}
                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="contact_name">Contact Name</Label>
                                <Input
                                    id="contact_name"
                                    value={formData.contact_name || ''}
                                    onChange={(e) => setFormData({ ...formData, contact_name: e.target.value })}
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="contact_email">Contact Email</Label>
                                <Input
                                    id="contact_email"
                                    type="email"
                                    value={formData.contact_email || ''}
                                    onChange={(e) => setFormData({ ...formData, contact_email: e.target.value })}
                                />
                            </div>
                        </div>

                        <DialogFooter className="pt-4 mt-6 border-t border-gray-100">
                            <Button type="button" variant="secondary" onClick={handleCloseModal}>
                                Cancel
                            </Button>
                            <Button type="submit" disabled={createMutation.isPending || updateMutation.isPending}>
                                {editingId ? 'Save Changes' : 'Add Third Party'}
                            </Button>
                        </DialogFooter>
                    </form>
                </DialogContent>
            </Dialog>
        </div>
    );
}
