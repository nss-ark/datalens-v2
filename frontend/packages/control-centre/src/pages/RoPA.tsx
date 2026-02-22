import { useState, useMemo, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
    FileText, RefreshCw, Save, Upload, ArrowUpCircle,
    ChevronDown, ChevronRight, Clock, AlertCircle, Sparkles,
    Database, Shield, Tag, Building2
} from 'lucide-react';
import {
    Button, Badge, toast,
    Card, CardHeader, CardContent,
    Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription,
    Input, Textarea,
    Table, TableHeader, TableHead, TableBody, TableRow, TableCell,
    StatusBadge,
} from '@datalens/shared';
import { ropaService } from '../services/ropaService';
import type { RoPAVersion, RoPAContent } from '../services/ropaService';

// ─── Collapsible Section ────────────────────────────────────────────────────

function CollapsibleSection({
    title,
    icon: Icon,
    count,
    defaultOpen = false,
    children,
}: {
    title: string;
    icon: React.ElementType;
    count: number;
    defaultOpen?: boolean;
    children: React.ReactNode;
}) {
    const [open, setOpen] = useState(defaultOpen);

    return (
        <Card className="mb-4">
            <button
                onClick={() => setOpen(!open)}
                className="w-full flex items-center justify-between p-4 hover:bg-gray-50 rounded-t-lg transition-colors cursor-pointer"
            >
                <div className="flex items-center gap-3">
                    {open ? <ChevronDown size={18} /> : <ChevronRight size={18} />}
                    <Icon size={18} className="text-blue-600" />
                    <span className="font-medium text-gray-900">{title}</span>
                    <Badge variant="secondary" className="ml-2">{count}</Badge>
                </div>
            </button>
            {open && <CardContent className="pt-0 pb-4 px-4">{children}</CardContent>}
        </Card>
    );
}

// ─── Status Badge Mapping ───────────────────────────────────────────────────

const STATUS_VARIANT: Record<string, 'success' | 'warning' | 'info' | 'neutral'> = {
    DRAFT: 'warning',
    PUBLISHED: 'success',
    ARCHIVED: 'neutral',
};

// ─── Main Page Component ────────────────────────────────────────────────────

const RoPA = () => {
    const queryClient = useQueryClient();

    // --- State ---
    const [editedOrgName, setEditedOrgName] = useState<string | null>(null);
    const [showSaveDialog, setShowSaveDialog] = useState(false);
    const [changeSummary, setChangeSummary] = useState('');
    const [showVersionHistory, setShowVersionHistory] = useState(false);
    const [viewingVersion, setViewingVersion] = useState<RoPAVersion | null>(null);
    const [versionPage, setVersionPage] = useState(1);

    // --- Queries ---
    const { data: latestVersion, isLoading, isError, refetch } = useQuery({
        queryKey: ['ropa', 'latest'],
        queryFn: ropaService.getLatest,
    });

    const { data: versionsData } = useQuery({
        queryKey: ['ropa', 'versions', versionPage],
        queryFn: () => ropaService.listVersions(versionPage, 10),
        enabled: showVersionHistory,
    });

    // --- Mutations ---
    const generateMutation = useMutation({
        mutationFn: ropaService.generate,
        onSuccess: () => {
            toast.success('RoPA Generated', 'A new RoPA version has been auto-generated from live data.');
            queryClient.invalidateQueries({ queryKey: ['ropa'] });
            setEditedOrgName(null);
        },
        onError: () => toast.error('Error', 'Failed to generate RoPA.'),
    });

    const saveMutation = useMutation({
        mutationFn: ({ content, summary }: { content: RoPAContent; summary: string }) =>
            ropaService.saveEdit(content, summary),
        onSuccess: () => {
            toast.success('Changes Saved', 'Your edits have been saved as a new version.');
            queryClient.invalidateQueries({ queryKey: ['ropa'] });
            setShowSaveDialog(false);
            setChangeSummary('');
            setEditedOrgName(null);
        },
        onError: () => toast.error('Error', 'Failed to save changes.'),
    });

    const publishMutation = useMutation({
        mutationFn: ropaService.publish,
        onSuccess: () => {
            toast.success('Published', 'This RoPA version has been published.');
            queryClient.invalidateQueries({ queryKey: ['ropa'] });
        },
        onError: () => toast.error('Error', 'Failed to publish RoPA.'),
    });

    const promoteMutation = useMutation({
        mutationFn: ropaService.promote,
        onSuccess: () => {
            toast.success('Major Version', 'A new major version has been created.');
            queryClient.invalidateQueries({ queryKey: ['ropa'] });
        },
        onError: () => toast.error('Error', 'Failed to create major version.'),
    });

    // --- Derived State ---
    const currentVersion = viewingVersion || latestVersion;
    const isViewingOldVersion = viewingVersion !== null;
    const isDraft = currentVersion?.status === 'DRAFT';
    const isPublished = currentVersion?.status === 'PUBLISHED';

    const isDirty = useMemo(() => {
        if (!currentVersion || isViewingOldVersion) return false;
        return editedOrgName !== null && editedOrgName !== currentVersion.content.organization_name;
    }, [editedOrgName, currentVersion, isViewingOldVersion]);

    const currentOrgName = editedOrgName ?? currentVersion?.content?.organization_name ?? '';

    // --- Handlers ---
    const handleSaveChanges = useCallback(() => {
        if (!currentVersion) return;
        const updatedContent: RoPAContent = {
            ...currentVersion.content,
            organization_name: editedOrgName ?? currentVersion.content.organization_name,
        };
        saveMutation.mutate({ content: updatedContent, summary: changeSummary });
    }, [currentVersion, editedOrgName, changeSummary, saveMutation]);

    const handlePublish = useCallback(() => {
        if (!currentVersion) return;
        if (window.confirm('Publish this version? Previous published versions will be archived.')) {
            publishMutation.mutate(currentVersion.id);
        }
    }, [currentVersion, publishMutation]);

    const handlePromote = useCallback(() => {
        if (window.confirm('Create a new major version? This copies the current content to a new major version.')) {
            promoteMutation.mutate();
        }
    }, [promoteMutation]);

    const handleViewVersion = useCallback(async (versionStr: string) => {
        try {
            const v = await ropaService.getVersion(versionStr);
            setViewingVersion(v);
            setEditedOrgName(null);
        } catch {
            toast.error('Error', 'Failed to load version.');
        }
    }, []);

    // ─── Loading State ──────────────────────────────────────────────────────
    if (isLoading) {
        return (
            <div className="p-6 max-w-7xl mx-auto flex justify-center py-24">
                <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-600" />
            </div>
        );
    }

    // ─── Error State ────────────────────────────────────────────────────────
    if (isError) {
        return (
            <div className="p-6 max-w-7xl mx-auto">
                <Card className="p-8 text-center">
                    <AlertCircle className="mx-auto h-12 w-12 text-red-400 mb-4" />
                    <h2 className="text-lg font-medium text-gray-900 mb-2">Failed to Load RoPA</h2>
                    <p className="text-gray-500 mb-4">Something went wrong while fetching the RoPA data.</p>
                    <Button onClick={() => refetch()} variant="primary">
                        <RefreshCw size={16} className="mr-2" /> Retry
                    </Button>
                </Card>
            </div>
        );
    }

    // ─── Empty State — No Versions ──────────────────────────────────────────
    if (!latestVersion && !viewingVersion) {
        return (
            <div className="p-6 max-w-7xl mx-auto">
                <Card className="p-12 text-center border-2 border-dashed border-blue-200 bg-blue-50/30">
                    <Sparkles className="mx-auto h-16 w-16 text-blue-500 mb-6" />
                    <h2 className="text-2xl font-bold text-gray-900 mb-3">
                        Generate Your First RoPA
                    </h2>
                    <p className="text-gray-500 mb-6 max-w-md mx-auto">
                        Auto-generate a Record of Processing Activities from your existing
                        purposes, data sources, and retention policies.
                    </p>
                    <Button
                        onClick={() => generateMutation.mutate()}
                        variant="primary"
                        className="bg-blue-600 hover:bg-blue-700 text-white px-8 py-3"
                        disabled={generateMutation.isPending}
                    >
                        {generateMutation.isPending ? (
                            <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2" />
                        ) : (
                            <FileText size={18} className="mr-2" />
                        )}
                        Generate RoPA
                    </Button>
                </Card>
            </div>
        );
    }

    // ─── Main Content ───────────────────────────────────────────────────────
    const content = currentVersion?.content;

    return (
        <div className="p-6 max-w-7xl mx-auto py-8">
            {/* Header */}
            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-8">
                <div className="flex items-center gap-3">
                    <FileText className="text-blue-600" size={28} />
                    <div>
                        <h1 className="text-2xl font-bold text-gray-900">
                            Record of Processing Activities
                        </h1>
                        <div className="flex items-center gap-2 mt-1">
                            <Badge variant="outline" className="font-mono">
                                v{currentVersion?.version}
                            </Badge>
                            <StatusBadge
                                label={currentVersion?.status || 'DRAFT'}
                                variant={STATUS_VARIANT[currentVersion?.status || 'DRAFT']}
                            />
                            {isViewingOldVersion && (
                                <Badge variant="secondary" className="bg-amber-100 text-amber-800">
                                    Viewing Historical Version
                                </Badge>
                            )}
                            {isDirty && (
                                <Badge variant="secondary" className="bg-yellow-100 text-yellow-700">
                                    Unsaved Changes
                                </Badge>
                            )}
                        </div>
                    </div>
                </div>

                <div className="flex flex-wrap items-center gap-2">
                    {isViewingOldVersion && (
                        <Button
                            onClick={() => { setViewingVersion(null); setEditedOrgName(null); }}
                            variant="outline"
                        >
                            ← Back to Latest
                        </Button>
                    )}

                    {!isViewingOldVersion && (
                        <>
                            <Button
                                onClick={() => generateMutation.mutate()}
                                variant="outline"
                                disabled={generateMutation.isPending}
                            >
                                <RefreshCw size={16} className="mr-1" />
                                Regenerate
                            </Button>

                            {isDraft && isDirty && (
                                <Button
                                    onClick={() => setShowSaveDialog(true)}
                                    variant="primary"
                                    className="bg-blue-600 hover:bg-blue-700 text-white"
                                >
                                    <Save size={16} className="mr-1" />
                                    Save Changes
                                </Button>
                            )}

                            {isDraft && (
                                <Button
                                    onClick={handlePublish}
                                    variant="primary"
                                    className="bg-green-600 hover:bg-green-700 text-white"
                                    disabled={publishMutation.isPending}
                                >
                                    <Upload size={16} className="mr-1" />
                                    Publish
                                </Button>
                            )}

                            {isPublished && (
                                <Button
                                    onClick={handlePromote}
                                    variant="outline"
                                    disabled={promoteMutation.isPending}
                                >
                                    <ArrowUpCircle size={16} className="mr-1" />
                                    New Major Version
                                </Button>
                            )}
                        </>
                    )}
                </div>
            </div>

            {/* Organization Name (Editable) */}
            {content && (
                <Card className="mb-6">
                    <CardContent className="p-4">
                        <label className="block text-sm font-medium text-gray-600 mb-1">Organization</label>
                        {isDraft && !isViewingOldVersion ? (
                            <Input
                                value={currentOrgName}
                                onChange={(e) => setEditedOrgName(e.target.value)}
                                className="max-w-md"
                                placeholder="Enter organization name"
                            />
                        ) : (
                            <p className="text-lg font-medium text-gray-900">{content.organization_name || '—'}</p>
                        )}
                    </CardContent>
                </Card>
            )}

            {/* Collapsible Sections */}
            {content && (
                <>
                    {/* Purposes */}
                    <CollapsibleSection
                        title="Purposes"
                        icon={Sparkles}
                        count={content.purposes?.length ?? 0}
                        defaultOpen
                    >
                        {content.purposes?.length > 0 ? (
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>Name</TableHead>
                                        <TableHead>Code</TableHead>
                                        <TableHead>Legal Basis</TableHead>
                                        <TableHead>Description</TableHead>
                                        <TableHead>Active</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {content.purposes.map((p) => (
                                        <TableRow key={p.id}>
                                            <TableCell className="font-medium">{p.name}</TableCell>
                                            <TableCell><code className="text-xs bg-gray-100 px-1 py-0.5 rounded">{p.code}</code></TableCell>
                                            <TableCell>{p.legal_basis}</TableCell>
                                            <TableCell className="max-w-xs truncate">{p.description}</TableCell>
                                            <TableCell>
                                                <Badge variant={p.is_active ? 'default' : 'secondary'}>
                                                    {p.is_active ? 'Active' : 'Inactive'}
                                                </Badge>
                                            </TableCell>
                                        </TableRow>
                                    ))}
                                </TableBody>
                            </Table>
                        ) : (
                            <p className="text-gray-400 text-sm py-2">No purposes defined.</p>
                        )}
                    </CollapsibleSection>

                    {/* Data Sources */}
                    <CollapsibleSection title="Data Sources" icon={Database} count={content.data_sources?.length ?? 0}>
                        {content.data_sources?.length > 0 ? (
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>Name</TableHead>
                                        <TableHead>Type</TableHead>
                                        <TableHead>Active</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {content.data_sources.map((ds) => (
                                        <TableRow key={ds.id}>
                                            <TableCell className="font-medium">{ds.name}</TableCell>
                                            <TableCell>{ds.type}</TableCell>
                                            <TableCell>
                                                <Badge variant={ds.is_active ? 'default' : 'secondary'}>
                                                    {ds.is_active ? 'Active' : 'Inactive'}
                                                </Badge>
                                            </TableCell>
                                        </TableRow>
                                    ))}
                                </TableBody>
                            </Table>
                        ) : (
                            <p className="text-gray-400 text-sm py-2">No data sources found.</p>
                        )}
                    </CollapsibleSection>

                    {/* Retention Policies */}
                    <CollapsibleSection title="Retention Policies" icon={Shield} count={content.retention_policies?.length ?? 0}>
                        {content.retention_policies?.length > 0 ? (
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>Purpose</TableHead>
                                        <TableHead>Max Retention (days)</TableHead>
                                        <TableHead>Categories</TableHead>
                                        <TableHead>Auto-Erase</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {content.retention_policies.map((r) => (
                                        <TableRow key={r.id}>
                                            <TableCell className="font-medium">{r.purpose_name}</TableCell>
                                            <TableCell>{r.max_retention_days}</TableCell>
                                            <TableCell>
                                                <div className="flex flex-wrap gap-1">
                                                    {r.data_categories?.map((c) => (
                                                        <Badge key={c} variant="outline" className="text-xs">{c}</Badge>
                                                    ))}
                                                </div>
                                            </TableCell>
                                            <TableCell>
                                                <Badge variant={r.auto_erase ? 'destructive' : 'secondary'}>
                                                    {r.auto_erase ? 'Yes' : 'No'}
                                                </Badge>
                                            </TableCell>
                                        </TableRow>
                                    ))}
                                </TableBody>
                            </Table>
                        ) : (
                            <p className="text-gray-400 text-sm py-2">No retention policies defined.</p>
                        )}
                    </CollapsibleSection>

                    {/* Third Parties */}
                    <CollapsibleSection title="Third Parties" icon={Building2} count={content.third_parties?.length ?? 0}>
                        {content.third_parties?.length > 0 ? (
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>Name</TableHead>
                                        <TableHead>Type</TableHead>
                                        <TableHead>Country</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {content.third_parties.map((tp) => (
                                        <TableRow key={tp.id}>
                                            <TableCell className="font-medium">{tp.name}</TableCell>
                                            <TableCell>{tp.type}</TableCell>
                                            <TableCell>{tp.country}</TableCell>
                                        </TableRow>
                                    ))}
                                </TableBody>
                            </Table>
                        ) : (
                            <p className="text-gray-400 text-sm py-2">No third parties registered.</p>
                        )}
                    </CollapsibleSection>

                    {/* Data Categories */}
                    <CollapsibleSection title="Data Categories" icon={Tag} count={content.data_categories?.length ?? 0}>
                        {content.data_categories?.length > 0 ? (
                            <div className="flex flex-wrap gap-2 py-2">
                                {content.data_categories.map((cat) => (
                                    <Badge key={cat} variant="outline" className="px-3 py-1 text-sm">
                                        {cat}
                                    </Badge>
                                ))}
                            </div>
                        ) : (
                            <p className="text-gray-400 text-sm py-2">No data categories found.</p>
                        )}
                    </CollapsibleSection>
                </>
            )}

            {/* Version History Toggle */}
            <div className="mt-8">
                <Button
                    onClick={() => setShowVersionHistory(!showVersionHistory)}
                    variant="outline"
                    className="mb-4"
                >
                    <Clock size={16} className="mr-2" />
                    {showVersionHistory ? 'Hide' : 'Show'} Version History
                </Button>

                {showVersionHistory && (
                    <Card>
                        <CardHeader className="p-4 border-b">
                            <h3 className="font-medium text-gray-900 flex items-center gap-2">
                                <Clock size={16} /> Version History
                            </h3>
                        </CardHeader>
                        <CardContent className="p-0">
                            {versionsData?.items?.length ? (
                                <div className="divide-y">
                                    {versionsData.items.map((v: RoPAVersion) => (
                                        <button
                                            key={v.id}
                                            onClick={() => handleViewVersion(v.version)}
                                            className={`w-full flex items-center justify-between p-4 hover:bg-gray-50 transition-colors text-left cursor-pointer ${viewingVersion?.id === v.id ? 'bg-blue-50 border-l-4 border-blue-500' : ''
                                                }`}
                                        >
                                            <div className="flex items-center gap-3">
                                                <Badge variant="outline" className="font-mono text-xs">
                                                    v{v.version}
                                                </Badge>
                                                <StatusBadge label={v.status} variant={STATUS_VARIANT[v.status]} />
                                                <span className="text-sm text-gray-500">
                                                    {v.generated_by === 'auto' ? '🤖 auto' : `👤 ${v.generated_by}`}
                                                </span>
                                            </div>
                                            <span className="text-xs text-gray-400">
                                                {new Date(v.created_at).toLocaleString()}
                                            </span>
                                        </button>
                                    ))}
                                </div>
                            ) : (
                                <p className="text-gray-400 text-sm p-4">No version history available.</p>
                            )}

                            {versionsData && versionsData.total_pages > 1 && (
                                <div className="flex justify-center gap-2 p-4 border-t">
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        disabled={versionPage <= 1}
                                        onClick={() => setVersionPage((p) => Math.max(1, p - 1))}
                                    >
                                        Previous
                                    </Button>
                                    <span className="text-sm text-gray-500 flex items-center">
                                        Page {versionPage} of {versionsData.total_pages}
                                    </span>
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        disabled={versionPage >= versionsData.total_pages}
                                        onClick={() => setVersionPage((p) => p + 1)}
                                    >
                                        Next
                                    </Button>
                                </div>
                            )}
                        </CardContent>
                    </Card>
                )}
            </div>

            {/* Save Changes Dialog */}
            <Dialog open={showSaveDialog} onOpenChange={setShowSaveDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Save Changes</DialogTitle>
                        <DialogDescription>
                            Your edits will be saved as a new minor version. Provide a brief summary of what changed.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="py-4">
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                            Change Summary
                        </label>
                        <Textarea
                            value={changeSummary}
                            onChange={(e) => setChangeSummary(e.target.value)}
                            placeholder="Describe what you changed..."
                            rows={3}
                        />
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setShowSaveDialog(false)}>
                            Cancel
                        </Button>
                        <Button
                            variant="primary"
                            className="bg-blue-600 hover:bg-blue-700 text-white"
                            onClick={handleSaveChanges}
                            disabled={!changeSummary.trim() || saveMutation.isPending}
                        >
                            {saveMutation.isPending ? 'Saving...' : 'Save as New Version'}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
};

export default RoPA;
