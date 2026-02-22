import { useState, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Briefcase, CheckCircle, Search, Plus, Trash2, ArrowDown } from 'lucide-react';
import { governanceService } from '../../services/governance';
import { purposeAssignmentService } from '../../services/purposeAssignmentService';
import type { PurposeAssignment } from '../../services/purposeAssignmentService';
import { SuggestionCard } from '../../components/Governance/SuggestionCard';
import { toast } from '@datalens/shared';
import { Button, Badge, Input } from '@datalens/shared';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@datalens/shared';
import {
    Tabs, TabsList, TabsTrigger, TabsContent,
} from '@datalens/shared';
import {
    Card, CardContent,
    Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription,
    Table, TableHeader, TableHead, TableBody, TableRow, TableCell,
} from '@datalens/shared';
import type { PurposeSuggestion } from '../../types/governance';

// ─── Scope Type Options ─────────────────────────────────────────────────────

const SCOPE_TYPES = ['SERVER', 'DATABASE', 'SCHEMA', 'TABLE', 'COLUMN'] as const;

const SCOPE_PLACEHOLDERS: Record<string, string> = {
    SERVER: 'server-name',
    DATABASE: 'db_name',
    SCHEMA: 'db_name.schema_name',
    TABLE: 'db_name.schema_name.table_name',
    COLUMN: 'db_name.schema_name.table_name.column_name',
};

// ─── Scope Assignments Sub-Component ────────────────────────────────────────

function ScopeAssignments() {
    const queryClient = useQueryClient();

    // Filters
    const [scopeType, setScopeType] = useState<string>('');
    const [scopeId, setScopeId] = useState('');
    const [showInherited, setShowInherited] = useState(false);
    const [hasSearched, setHasSearched] = useState(false);

    // Add dialog
    const [showAddDialog, setShowAddDialog] = useState(false);
    const [newPurposeId, setNewPurposeId] = useState('');
    const [newScopeType, setNewScopeType] = useState<string>('SERVER');
    const [newScopeId, setNewScopeId] = useState('');
    const [newScopeName, setNewScopeName] = useState('');

    // Default: fetch all assignments
    const { data: allAssignments = [], isLoading: isLoadingAll } = useQuery({
        queryKey: ['purposeAssignments', 'all'],
        queryFn: purposeAssignmentService.getAll,
    });

    // Scope search results
    const { data: scopeResults, isLoading: isLoadingScope, refetch: refetchScope } = useQuery({
        queryKey: ['purposeAssignments', 'scope', scopeType, scopeId, showInherited],
        queryFn: () =>
            showInherited
                ? purposeAssignmentService.getEffective(scopeType, scopeId)
                : purposeAssignmentService.getByScope(scopeType, scopeId),
        enabled: hasSearched && !!scopeType && !!scopeId,
    });

    // Fetch purposes for the add dialog
    const { data: purposes = [] } = useQuery({
        queryKey: ['purposeSuggestions-list'],
        queryFn: async () => {
            // Use governance suggestions as a proxy for purpose list, or call any available purpose endpoint
            const suggestions = await governanceService.getPurposeSuggestions();
            // Extract unique purpose names as simple objects
            return suggestions;
        },
        enabled: showAddDialog,
    });

    // Mutations
    const assignMutation = useMutation({
        mutationFn: purposeAssignmentService.assign,
        onSuccess: () => {
            toast.success('Assigned', 'Purpose has been assigned to the scope.');
            queryClient.invalidateQueries({ queryKey: ['purposeAssignments'] });
            setShowAddDialog(false);
            resetAddForm();
        },
        onError: () => toast.error('Error', 'Failed to assign purpose.'),
    });

    const removeMutation = useMutation({
        mutationFn: purposeAssignmentService.remove,
        onSuccess: () => {
            toast.success('Removed', 'Assignment has been removed.');
            queryClient.invalidateQueries({ queryKey: ['purposeAssignments'] });
        },
        onError: () => toast.error('Error', 'Failed to remove assignment.'),
    });

    const resetAddForm = () => {
        setNewPurposeId('');
        setNewScopeType('SERVER');
        setNewScopeId('');
        setNewScopeName('');
    };

    const handleSearch = useCallback(() => {
        if (!scopeType || !scopeId) {
            toast.info('Scope Required', 'Please select a scope type and enter a scope ID.');
            return;
        }
        setHasSearched(true);
        refetchScope();
    }, [scopeType, scopeId, refetchScope]);

    const handleClearSearch = useCallback(() => {
        setScopeType('');
        setScopeId('');
        setHasSearched(false);
    }, []);

    const handleAdd = useCallback(() => {
        if (!newPurposeId || !newScopeType || !newScopeId) {
            toast.info('Missing Fields', 'Please fill in purpose, scope type, and scope ID.');
            return;
        }
        assignMutation.mutate({
            purpose_id: newPurposeId,
            scope_type: newScopeType,
            scope_id: newScopeId,
            scope_name: newScopeName || undefined,
        });
    }, [newPurposeId, newScopeType, newScopeId, newScopeName, assignMutation]);

    // Which data to display
    const displayData: PurposeAssignment[] = hasSearched ? (scopeResults ?? []) : allAssignments;
    const isLoading = hasSearched ? isLoadingScope : isLoadingAll;

    return (
        <div className="mt-6 space-y-6">
            {/* Search Bar */}
            <Card>
                <CardContent className="p-4">
                    <div className="flex flex-wrap items-end gap-3">
                        <div className="flex-shrink-0 w-[160px]">
                            <label className="block text-xs font-medium text-gray-500 mb-1">Scope Type</label>
                            <Select value={scopeType} onValueChange={setScopeType}>
                                <SelectTrigger>
                                    <SelectValue placeholder="Select scope..." />
                                </SelectTrigger>
                                <SelectContent>
                                    {SCOPE_TYPES.map((st) => (
                                        <SelectItem key={st} value={st}>{st}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="flex-1 min-w-[200px]">
                            <label className="block text-xs font-medium text-gray-500 mb-1">Scope ID</label>
                            <Input
                                value={scopeId}
                                onChange={(e) => setScopeId(e.target.value)}
                                placeholder={scopeType ? SCOPE_PLACEHOLDERS[scopeType] : 'Select scope type first...'}
                            />
                        </div>
                        <Button onClick={handleSearch} variant="primary" className="bg-blue-600 hover:bg-blue-700 text-white">
                            <Search size={16} className="mr-1" /> Search
                        </Button>
                        {hasSearched && (
                            <Button onClick={handleClearSearch} variant="outline">
                                Clear
                            </Button>
                        )}
                    </div>

                    <div className="mt-3 flex items-center gap-4">
                        <label className="flex items-center gap-2 text-sm text-gray-600 cursor-pointer">
                            <input
                                type="checkbox"
                                checked={showInherited}
                                onChange={(e) => {
                                    setShowInherited(e.target.checked);
                                    if (hasSearched) refetchScope();
                                }}
                                className="rounded border-gray-300"
                            />
                            Show inherited (effective view)
                        </label>
                    </div>
                </CardContent>
            </Card>

            {/* Add Assignment Button */}
            <div className="flex justify-end">
                <Button onClick={() => setShowAddDialog(true)} variant="primary" className="bg-blue-600 hover:bg-blue-700 text-white">
                    <Plus size={16} className="mr-1" /> Add Assignment
                </Button>
            </div>

            {/* Assignments Table */}
            {isLoading ? (
                <div className="flex justify-center py-12">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
                </div>
            ) : displayData.length === 0 ? (
                <Card className="p-8 text-center">
                    <p className="text-gray-400">
                        {hasSearched
                            ? 'No assignments found for this scope.'
                            : 'No purpose assignments yet. Click "Add Assignment" to get started.'}
                    </p>
                </Card>
            ) : (
                <Card>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Purpose ID</TableHead>
                                <TableHead>Scope</TableHead>
                                <TableHead>Scope ID</TableHead>
                                <TableHead>Source</TableHead>
                                <TableHead>Date</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {displayData.map((a) => (
                                <TableRow
                                    key={a.id}
                                    className={a.inherited ? 'opacity-60' : ''}
                                >
                                    <TableCell className="font-medium font-mono text-xs">
                                        {a.scope_name || a.purpose_id.slice(0, 8)}
                                    </TableCell>
                                    <TableCell>
                                        <Badge variant="outline" className="text-xs">{a.scope_type}</Badge>
                                    </TableCell>
                                    <TableCell className="font-mono text-xs max-w-xs truncate">
                                        {a.scope_id}
                                    </TableCell>
                                    <TableCell>
                                        {a.inherited ? (
                                            <Badge variant="secondary" className="bg-gray-100 text-gray-500">
                                                <ArrowDown size={12} className="mr-1" /> Inherited
                                            </Badge>
                                        ) : (
                                            <Badge variant="default">Direct</Badge>
                                        )}
                                    </TableCell>
                                    <TableCell className="text-xs text-gray-500">
                                        {new Date(a.assigned_at).toLocaleDateString()}
                                    </TableCell>
                                    <TableCell className="text-right">
                                        {!a.inherited && (
                                            <Button
                                                onClick={() => {
                                                    if (window.confirm('Remove this assignment?')) {
                                                        removeMutation.mutate(a.id);
                                                    }
                                                }}
                                                variant="ghost"
                                                size="sm"
                                                className="text-red-500 hover:text-red-700"
                                            >
                                                <Trash2 size={14} />
                                            </Button>
                                        )}
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </Card>
            )}

            {/* Add Assignment Dialog */}
            <Dialog open={showAddDialog} onOpenChange={setShowAddDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Add Purpose Assignment</DialogTitle>
                        <DialogDescription>
                            Assign a purpose to a specific scope level. Assignments inherit down the hierarchy.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Purpose ID</label>
                            <Input
                                value={newPurposeId}
                                onChange={(e) => setNewPurposeId(e.target.value)}
                                placeholder="Enter purpose UUID"
                            />
                            {purposes.length > 0 && (
                                <p className="text-xs text-gray-400 mt-1">
                                    Tip: Copy a purpose ID from the Governance Purposes page.
                                </p>
                            )}
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Scope Type</label>
                            <Select value={newScopeType} onValueChange={setNewScopeType}>
                                <SelectTrigger>
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    {SCOPE_TYPES.map((st) => (
                                        <SelectItem key={st} value={st}>{st}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Scope ID</label>
                            <Input
                                value={newScopeId}
                                onChange={(e) => setNewScopeId(e.target.value)}
                                placeholder={SCOPE_PLACEHOLDERS[newScopeType] ?? 'Enter scope ID'}
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Scope Name (optional)</label>
                            <Input
                                value={newScopeName}
                                onChange={(e) => setNewScopeName(e.target.value)}
                                placeholder="Human-readable name for the scope"
                            />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => { setShowAddDialog(false); resetAddForm(); }}>
                            Cancel
                        </Button>
                        <Button
                            variant="primary"
                            className="bg-blue-600 hover:bg-blue-700 text-white"
                            onClick={handleAdd}
                            disabled={assignMutation.isPending}
                        >
                            {assignMutation.isPending ? 'Assigning...' : 'Assign'}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}

// ─── Main Page Component ────────────────────────────────────────────────────

const PurposeMapping = () => {
    const queryClient = useQueryClient();
    const [filter, setFilter] = useState<'all' | 'high_confidence'>('all');

    // Fetch suggestions
    const { data: suggestions = [], isLoading } = useQuery({
        queryKey: ['purposeSuggestions'],
        queryFn: governanceService.getPurposeSuggestions,
    });

    // Accept mutation
    const acceptMutation = useMutation({
        mutationFn: governanceService.acceptSuggestion,
        onSuccess: () => {
            toast.success('Purpose accepted', 'The purpose mapping has been updated.');
            queryClient.invalidateQueries({ queryKey: ['purposeSuggestions'] });
        },
        onError: () => {
            toast.error('Error', 'Failed to accept suggestion.');
        }
    });

    // Reject mutation
    const rejectMutation = useMutation({
        mutationFn: governanceService.rejectSuggestion,
        onSuccess: () => {
            toast.info('Suggestion rejected', 'The suggestion has been dismissed.');
            queryClient.invalidateQueries({ queryKey: ['purposeSuggestions'] });
        },
        onError: () => {
            toast.error('Error', 'Failed to reject suggestion.');
        }
    });

    const filteredSuggestions = suggestions.filter((s: PurposeSuggestion) => {
        if (filter === 'high_confidence') return s.confidenceScore >= 0.8;
        return true;
    });

    const handleAcceptAllHighConfidence = async () => {
        const highConfidence = suggestions.filter((s: PurposeSuggestion) => s.confidenceScore >= 0.8);
        if (highConfidence.length === 0) {
            toast.info('No high confidence suggestions', 'There are no suggestions with >80% confidence.');
            return;
        }

        try {
            await Promise.all(highConfidence.map((s: PurposeSuggestion) => governanceService.acceptSuggestion(s.id)));
            toast.success('Batch Accept Complete', `Accepted ${highConfidence.length} suggestions.`);
            queryClient.invalidateQueries({ queryKey: ['purposeSuggestions'] });
        } catch (error) {
            toast.error('Batch Error', 'Failed to accept some suggestions.');
            console.error(error);
        }
    };

    return (
        <div className="p-6 max-w-7xl mx-auto">
            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-8">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
                        <Briefcase className="text-blue-600" />
                        Purpose Mapping
                    </h1>
                    <p className="text-gray-500 mt-1">
                        Manage AI suggestions and scope-level purpose assignments.
                    </p>
                </div>
            </div>

            <Tabs defaultValue="suggestions" className="mt-2">
                <TabsList>
                    <TabsTrigger value="suggestions">AI Suggestions</TabsTrigger>
                    <TabsTrigger value="assignments">Scope Assignments</TabsTrigger>
                </TabsList>

                {/* Tab 1: AI Suggestions (existing content) */}
                <TabsContent value="suggestions">
                    <div className="flex flex-wrap items-center gap-3 mt-4 mb-6">
                        <Select
                            value={filter}
                            onValueChange={(val) => setFilter(val as 'all' | 'high_confidence')}
                        >
                            <SelectTrigger className="w-[180px]">
                                <SelectValue placeholder="Filter suggestions" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="all">All Suggestions</SelectItem>
                                <SelectItem value="high_confidence">High Confidence Only</SelectItem>
                            </SelectContent>
                        </Select>

                        <Button
                            onClick={handleAcceptAllHighConfidence}
                            variant="primary"
                            className="bg-green-600 hover:bg-green-700 text-white"
                        >
                            <CheckCircle size={16} className="mr-2" />
                            Accept High Confidence
                        </Button>
                    </div>

                    {isLoading ? (
                        <div className="flex justify-center py-12">
                            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                        </div>
                    ) : filteredSuggestions.length === 0 ? (
                        <div className="text-center py-12 bg-white rounded-lg shadow-sm border border-gray-200">
                            <Briefcase className="mx-auto h-12 w-12 text-gray-300" />
                            <h3 className="mt-2 text-sm font-medium text-gray-900">No suggestions found</h3>
                            <p className="mt-1 text-sm text-gray-500">
                                Good job! All data elements are currently mapped.
                            </p>
                        </div>
                    ) : (
                        <div className="grid grid-cols-1 gap-6">
                            {filteredSuggestions.map((suggestion: PurposeSuggestion) => (
                                <SuggestionCard
                                    key={suggestion.id}
                                    suggestion={suggestion}
                                    onAccept={(id) => acceptMutation.mutate(id)}
                                    onReject={(id) => rejectMutation.mutate(id)}
                                />
                            ))}
                        </div>
                    )}
                </TabsContent>

                {/* Tab 2: Scope Assignments (new) */}
                <TabsContent value="assignments">
                    <ScopeAssignments />
                </TabsContent>
            </Tabs>
        </div>
    );
};

export default PurposeMapping;
