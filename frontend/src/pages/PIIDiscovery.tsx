import { useState } from 'react';
import { Check, Pencil, X, Filter } from 'lucide-react';
import { DataTable, type Column } from '../components/DataTable/DataTable';
import { Pagination } from '../components/DataTable/Pagination';
import { StatusBadge } from '../components/common/StatusBadge';
import { Modal } from '../components/common/Modal';
import { Button } from '../components/common/Button';
import { useClassifications, useSubmitFeedback, useAccuracyStats } from '../hooks/useDiscovery';
import { toast } from '../stores/toastStore';
import { cn } from '../utils/cn';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import type {
    PIIClassification,
    PIICategory,
    PIIType,
    DetectionMethod,
    VerificationStatus,
} from '../types/discovery';
import { PII_CATEGORIES, PII_TYPES, DETECTION_METHODS } from '../types/discovery';

const PAGE_SIZE = 20;

const PIIDiscovery = () => {
    // ── Filters ──
    const [statusFilter, setStatusFilter] = useState<VerificationStatus | ''>('');
    const [methodFilter, setMethodFilter] = useState<DetectionMethod | ''>('');
    const [page, setPage] = useState(1);

    // ── Data ──
    const { data: result, isLoading } = useClassifications({
        status: statusFilter || undefined,
        detection_method: methodFilter || undefined,
        page,
        page_size: PAGE_SIZE,
    });
    const classifications = result?.data ?? [];
    const total = result?.total ?? 0;

    // ── Feedback ──
    const { mutate: submitFeedback, isPending: isFeedbackPending } = useSubmitFeedback();

    // ── Accuracy Stats ──
    const [statsMethod, setStatsMethod] = useState<DetectionMethod>('AI');
    const { data: accuracyStats } = useAccuracyStats(statsMethod);

    // ── Modal State ──
    const [correctModal, setCorrectModal] = useState<PIIClassification | null>(null);
    const [rejectModal, setRejectModal] = useState<PIIClassification | null>(null);
    const [corrCategory, setCorrCategory] = useState<PIICategory>('IDENTITY');
    const [corrType, setCorrType] = useState<PIIType>('NAME');
    const [rejectNotes, setRejectNotes] = useState('');

    // ── Handlers ──
    const handleVerify = (row: PIIClassification, e: React.MouseEvent) => {
        e.stopPropagation();
        submitFeedback(
            { classification_id: row.id, feedback_type: 'VERIFIED', notes: '' },
            {
                onSuccess: () => toast.success('Verified', `"${row.field_name}" marked as verified PII.`),
                onError: () => toast.error('Error', 'Failed to verify classification.'),
            }
        );
    };

    const openCorrectModal = (row: PIIClassification, e: React.MouseEvent) => {
        e.stopPropagation();
        setCorrCategory(row.category);
        setCorrType(row.type);
        setCorrectModal(row);
    };

    const handleCorrect = () => {
        if (!correctModal) return;
        submitFeedback(
            {
                classification_id: correctModal.id,
                feedback_type: 'CORRECTED',
                corrected_category: corrCategory,
                corrected_type: corrType,
                notes: '',
            },
            {
                onSuccess: () => {
                    toast.success('Corrected', `"${correctModal.field_name}" classification updated.`);
                    setCorrectModal(null);
                },
                onError: () => toast.error('Error', 'Failed to correct classification.'),
            }
        );
    };

    const openRejectModal = (row: PIIClassification, e: React.MouseEvent) => {
        e.stopPropagation();
        setRejectNotes('');
        setRejectModal(row);
    };

    const handleReject = () => {
        if (!rejectModal) return;
        submitFeedback(
            {
                classification_id: rejectModal.id,
                feedback_type: 'REJECTED',
                notes: rejectNotes,
            },
            {
                onSuccess: () => {
                    toast.success('Rejected', `"${rejectModal.field_name}" marked as false positive.`);
                    setRejectModal(null);
                },
                onError: () => toast.error('Error', 'Failed to reject classification.'),
            }
        );
    };

    const clearFilters = () => {
        setStatusFilter('');
        setMethodFilter('');
        setPage(1);
    };

    const hasFilters = statusFilter || methodFilter;

    // ── Confidence badge helper ──
    const ConfidenceBadge = ({ value }: { value: number }) => {
        const pct = Math.round(value * 100);
        const tier = pct >= 90 ? 'high' : pct >= 70 ? 'medium' : 'low';

        const colors = {
            high: "text-green-700",
            medium: "text-yellow-700",
            low: "text-red-700"
        };

        const fills = {
            high: "bg-green-500",
            medium: "bg-yellow-500",
            low: "bg-red-500"
        };

        return (
            <span className={cn("inline-flex items-center gap-1.5 text-xs font-semibold", colors[tier])}>
                <span className="w-10 h-1.5 bg-slate-100 rounded-full overflow-hidden">
                    <span className={cn("block h-full rounded-full transition-all duration-300", fills[tier])} style={{ width: `${pct}%` }} />
                </span>
                {pct}%
            </span>
        );
    };

    // ── Columns ──
    const columns: Column<PIIClassification>[] = [
        {
            key: 'field_name',
            header: 'Field',
            sortable: true,
            render: (row) => (
                <div>
                    <div className="font-semibold text-gray-900">{row.field_name}</div>
                    <div className="text-xs text-gray-500">{row.entity_name}</div>
                </div>
            ),
        },
        {
            key: 'category',
            header: 'Category',
            sortable: true,
            width: '120px',
            render: (row) => (
                <span className="text-xs font-medium">{row.category}</span>
            ),
        },
        {
            key: 'type',
            header: 'Type',
            sortable: true,
            width: '110px',
            render: (row) => (
                <span className="text-xs">{row.type}</span>
            ),
        },
        {
            key: 'confidence',
            header: 'Confidence',
            sortable: true,
            width: '120px',
            render: (row) => <ConfidenceBadge value={row.confidence} />,
        },
        {
            key: 'detection_method',
            header: 'Method',
            sortable: true,
            width: '100px',
            render: (row) => (
                <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold bg-slate-100 text-slate-600 tracking-wide uppercase">
                    {row.detection_method}
                </span>
            ),
        },
        {
            key: 'status',
            header: 'Status',
            sortable: true,
            width: '110px',
            render: (row) => <StatusBadge label={row.status} size="sm" />,
        },
        {
            key: 'actions',
            header: '',
            width: '110px',
            render: (row) =>
                row.status === 'PENDING' ? (
                    <div className="flex gap-1">
                        <button
                            className="inline-flex items-center justify-center w-7 h-7 rounded-md border border-gray-200 bg-white hover:bg-green-50 hover:text-green-600 hover:border-green-200 transition-colors"
                            onClick={(e) => handleVerify(row, e)}
                            title="Verify"
                        >
                            <Check size={14} />
                        </button>
                        <button
                            className="inline-flex items-center justify-center w-7 h-7 rounded-md border border-gray-200 bg-white hover:bg-yellow-50 hover:text-yellow-600 hover:border-yellow-200 transition-colors"
                            onClick={(e) => openCorrectModal(row, e)}
                            title="Correct"
                        >
                            <Pencil size={14} />
                        </button>
                        <button
                            className="inline-flex items-center justify-center w-7 h-7 rounded-md border border-gray-200 bg-white hover:bg-red-50 hover:text-red-600 hover:border-red-200 transition-colors"
                            onClick={(e) => openRejectModal(row, e)}
                            title="Reject"
                        >
                            <X size={14} />
                        </button>
                    </div>
                ) : (
                    <span className="text-xs text-gray-400">Reviewed</span>
                ),
        },
    ];

    return (
        <div className="space-y-6">
            {/* Page Header */}
            <div className="flex justify-between items-start">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 mb-1">PII Discovery</h1>
                    <p className="text-sm text-gray-500">
                        Review auto-detected PII classifications and provide feedback
                    </p>
                </div>
            </div>

            {/* Accuracy Stats Panel */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                {DETECTION_METHODS.map((m) => (
                    <StatsCard
                        key={m.value}
                        method={m.value}
                        label={m.label}
                        active={statsMethod === m.value}
                        onClick={() => setStatsMethod(m.value)}
                        stats={statsMethod === m.value ? accuracyStats : undefined}
                    />
                ))}
            </div>

            {/* Filter Bar */}
            <div className="flex flex-wrap gap-3 items-center bg-white p-3 rounded-lg border border-gray-200 shadow-sm">
                <Filter size={16} className="text-gray-400 ml-1" />
                <select
                    className="h-9 rounded-md border border-gray-300 bg-white px-3 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    value={statusFilter}
                    onChange={(e) => { setStatusFilter(e.target.value as VerificationStatus | ''); setPage(1); }}
                >
                    <option value="">All Statuses</option>
                    <option value="PENDING">Pending</option>
                    <option value="VERIFIED">Verified</option>
                    <option value="REJECTED">Rejected</option>
                </select>

                <select
                    className="h-9 rounded-md border border-gray-300 bg-white px-3 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    value={methodFilter}
                    onChange={(e) => { setMethodFilter(e.target.value as DetectionMethod | ''); setPage(1); }}
                >
                    <option value="">All Methods</option>
                    {DETECTION_METHODS.map((m) => (
                        <option key={m.value} value={m.value}>{m.label}</option>
                    ))}
                </select>

                {hasFilters && (
                    <button
                        className="text-xs font-medium text-primary-600 hover:text-primary-700 ml-1"
                        onClick={clearFilters}
                    >
                        Clear filters
                    </button>
                )}

                <span className="text-xs text-gray-400 ml-auto mr-1">{total} classifications</span>
            </div>

            {/* Data Table */}
            <div className="bg-white rounded-lg border border-gray-200 shadow-sm overflow-hidden">
                <DataTable
                    columns={columns}
                    data={classifications}
                    isLoading={isLoading}
                    keyExtractor={(row) => row.id}
                    emptyTitle="No PII classifications found"
                    emptyDescription="Run a scan on a data source to discover PII fields."
                />
            </div>

            {/* Pagination */}
            {total > PAGE_SIZE && (
                <div className="mt-4">
                    <Pagination page={page} pageSize={PAGE_SIZE} total={total} onPageChange={setPage} />
                </div>
            )}

            {/* ── Correct Modal ── */}
            <Modal
                open={!!correctModal}
                onClose={() => setCorrectModal(null)}
                title="Correct Classification"
                footer={
                    <>
                        <Button variant="outline" onClick={() => setCorrectModal(null)}>Cancel</Button>
                        <Button onClick={handleCorrect} isLoading={isFeedbackPending}>Apply Correction</Button>
                    </>
                }
            >
                {correctModal && (
                    <div className="space-y-4">
                        <p className="text-sm text-gray-500">
                            Correcting <strong>{correctModal.field_name}</strong> in <strong>{correctModal.entity_name}</strong>
                        </p>

                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1.5">Correct Category</label>
                            <select
                                className="w-full h-10 px-3 rounded-md border border-gray-300 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
                                value={corrCategory}
                                onChange={(e) => setCorrCategory(e.target.value as PIICategory)}
                            >
                                {PII_CATEGORIES.map((c) => (
                                    <option key={c.value} value={c.value}>{c.label}</option>
                                ))}
                            </select>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1.5">Correct Type</label>
                            <select
                                className="w-full h-10 px-3 rounded-md border border-gray-300 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
                                value={corrType}
                                onChange={(e) => setCorrType(e.target.value as PIIType)}
                            >
                                {PII_TYPES.map((t) => (
                                    <option key={t.value} value={t.value}>{t.label}</option>
                                ))}
                            </select>
                        </div>
                    </div>
                )}
            </Modal>

            {/* ── Reject Modal ── */}
            <Modal
                open={!!rejectModal}
                onClose={() => setRejectModal(null)}
                title="Reject as False Positive"
                footer={
                    <>
                        <Button variant="outline" onClick={() => setRejectModal(null)}>Cancel</Button>
                        <Button variant="danger" onClick={handleReject} isLoading={isFeedbackPending}>Reject</Button>
                    </>
                }
            >
                {rejectModal && (
                    <div className="space-y-4">
                        <p className="text-sm text-gray-500">
                            Rejecting <strong>{rejectModal.field_name}</strong> ({rejectModal.category} → {rejectModal.type})
                        </p>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1.5">Reason / Notes</label>
                            <textarea
                                className="w-full h-20 px-3 py-2 rounded-md border border-gray-300 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 resize-none"
                                value={rejectNotes}
                                onChange={(e) => setRejectNotes(e.target.value)}
                                placeholder="Why is this a false positive?"
                            />
                        </div>
                    </div>
                )}
            </Modal>
        </div>
    );
};

// ── Stats Card Sub-component ──
function StatsCard({
    label, active, onClick, stats,
}: {
    method: string; label: string; active: boolean;
    onClick: () => void; stats?: { total: number; verified: number; corrected: number; rejected: number; accuracy: number };
}) {
    const rawAcc = stats ? stats.accuracy * 100 : 0;
    const acc = Number.isFinite(rawAcc) ? Math.round(rawAcc) : 0;

    return (
        <Card
            className={cn(
                "cursor-pointer transition-all hover:shadow-md",
                active ? "border-primary-500 ring-1 ring-primary-200" : ""
            )}
            onClick={onClick}
        >
            <CardHeader className="p-4 pb-2">
                <CardTitle className="text-xs font-semibold uppercase tracking-wider text-gray-400">
                    {label}
                </CardTitle>
            </CardHeader>
            <CardContent className="p-4 pt-0">
                {stats ? (
                    <>
                        <div className="text-2xl font-bold text-gray-900">{acc}%</div>
                        <div className="text-xs text-gray-500 mt-1">
                            {stats.verified}✓ {stats.corrected}✎ {stats.rejected}✗ of {stats.total}
                        </div>
                        <div className="h-1.5 w-full bg-slate-100 rounded-full mt-2 overflow-hidden">
                            <div className="h-full bg-primary-600 rounded-full transition-all duration-300" style={{ width: `${acc}%` }} />
                        </div>
                    </>
                ) : (
                    <div className="text-xs text-gray-400 mt-1">Click to load</div>
                )}
            </CardContent>
        </Card>
    );
}

export default PIIDiscovery;
