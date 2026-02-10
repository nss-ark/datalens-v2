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
import type {
    PIIClassification,
    PIICategory,
    PIIType,
    DetectionMethod,
    VerificationStatus,
} from '../types/discovery';
import { PII_CATEGORIES, PII_TYPES, DETECTION_METHODS } from '../types/discovery';
import styles from './PIIDiscovery.module.css';

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
        const tier = pct >= 90 ? 'confHigh' : pct >= 70 ? 'confMedium' : 'confLow';
        return (
            <span className={cn(styles.confidence, styles[tier])}>
                <span className={styles.confBar}>
                    <span className={styles.confFill} style={{ width: `${pct}%` }} />
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
                    <div style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{row.field_name}</div>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>{row.entity_name}</div>
                </div>
            ),
        },
        {
            key: 'category',
            header: 'Category',
            sortable: true,
            width: '120px',
            render: (row) => (
                <span style={{ fontSize: '0.8125rem', fontWeight: 500 }}>{row.category}</span>
            ),
        },
        {
            key: 'type',
            header: 'Type',
            sortable: true,
            width: '110px',
            render: (row) => (
                <span style={{ fontSize: '0.8125rem' }}>{row.type}</span>
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
            render: (row) => <span className={styles.methodBadge}>{row.detection_method}</span>,
        },
        {
            key: 'status',
            header: 'Status',
            sortable: true,
            width: '110px',
            render: (row) => <StatusBadge label={row.status} />,
        },
        {
            key: 'actions',
            header: '',
            width: '110px',
            render: (row) =>
                row.status === 'PENDING' ? (
                    <div className={styles.actions}>
                        <button
                            className={cn(styles.actionBtn, styles.verifyBtn)}
                            onClick={(e) => handleVerify(row, e)}
                            title="Verify"
                        >
                            <Check size={14} />
                        </button>
                        <button
                            className={cn(styles.actionBtn, styles.correctBtn)}
                            onClick={(e) => openCorrectModal(row, e)}
                            title="Correct"
                        >
                            <Pencil size={14} />
                        </button>
                        <button
                            className={cn(styles.actionBtn, styles.rejectBtn)}
                            onClick={(e) => openRejectModal(row, e)}
                            title="Reject"
                        >
                            <X size={14} />
                        </button>
                    </div>
                ) : (
                    <span style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>Reviewed</span>
                ),
        },
    ];

    return (
        <div>
            {/* Page Header */}
            <div className={styles.pageHeader}>
                <div>
                    <h1 className={styles.title}>PII Discovery</h1>
                    <p className={styles.subtitle}>
                        Review auto-detected PII classifications and provide feedback
                    </p>
                </div>
            </div>

            {/* Accuracy Stats Panel */}
            <div className={styles.statsPanel}>
                {DETECTION_METHODS.map((m) => (
                    <StatsCard key={m.value} method={m.value} label={m.label} active={statsMethod === m.value} onClick={() => setStatsMethod(m.value)} stats={statsMethod === m.value ? accuracyStats : undefined} />
                ))}
            </div>

            {/* Filter Bar */}
            <div className={styles.filterBar}>
                <Filter size={16} style={{ color: 'var(--text-tertiary)' }} />
                <select
                    className={styles.filterSelect}
                    value={statusFilter}
                    onChange={(e) => { setStatusFilter(e.target.value as VerificationStatus | ''); setPage(1); }}
                >
                    <option value="">All Statuses</option>
                    <option value="PENDING">Pending</option>
                    <option value="VERIFIED">Verified</option>
                    <option value="REJECTED">Rejected</option>
                </select>

                <select
                    className={styles.filterSelect}
                    value={methodFilter}
                    onChange={(e) => { setMethodFilter(e.target.value as DetectionMethod | ''); setPage(1); }}
                >
                    <option value="">All Methods</option>
                    {DETECTION_METHODS.map((m) => (
                        <option key={m.value} value={m.value}>{m.label}</option>
                    ))}
                </select>

                {hasFilters && (
                    <button className={styles.clearBtn} onClick={clearFilters}>Clear filters</button>
                )}

                <span className={styles.activeFilters}>{total} classifications</span>
            </div>

            {/* Data Table */}
            <DataTable
                columns={columns}
                data={classifications}
                isLoading={isLoading}
                keyExtractor={(row) => row.id}
                emptyTitle="No PII classifications found"
                emptyDescription="Run a scan on a data source to discover PII fields."
            />

            {/* Pagination */}
            {total > PAGE_SIZE && (
                <Pagination page={page} pageSize={PAGE_SIZE} total={total} onPageChange={setPage} />
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
                    <div className={styles.modalForm}>
                        <p style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>
                            Correcting <strong>{correctModal.field_name}</strong> in <strong>{correctModal.entity_name}</strong>
                        </p>

                        <div>
                            <label className={styles.modalLabel}>Correct Category</label>
                            <select className={styles.modalSelect} value={corrCategory} onChange={(e) => setCorrCategory(e.target.value as PIICategory)}>
                                {PII_CATEGORIES.map((c) => (
                                    <option key={c.value} value={c.value}>{c.label}</option>
                                ))}
                            </select>
                        </div>

                        <div>
                            <label className={styles.modalLabel}>Correct Type</label>
                            <select className={styles.modalSelect} value={corrType} onChange={(e) => setCorrType(e.target.value as PIIType)}>
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
                    <div className={styles.modalForm}>
                        <p style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>
                            Rejecting <strong>{rejectModal.field_name}</strong> ({rejectModal.category} → {rejectModal.type})
                        </p>
                        <div>
                            <label className={styles.modalLabel}>Reason / Notes</label>
                            <textarea
                                className={styles.modalTextarea}
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
    const acc = stats ? Math.round(stats.accuracy * 100) : 0;
    return (
        <div
            className={styles.statCard}
            style={{
                cursor: 'pointer',
                borderColor: active ? 'var(--primary-400)' : undefined,
                boxShadow: active ? '0 0 0 1px var(--primary-200)' : undefined,
            }}
            onClick={onClick}
        >
            <div className={styles.statLabel}>{label}</div>
            {stats ? (
                <>
                    <div className={styles.statValue}>{acc}%</div>
                    <div className={styles.statSub}>
                        {stats.verified}✓ {stats.corrected}✎ {stats.rejected}✗ of {stats.total}
                    </div>
                    <div className={styles.progressBar}>
                        <div className={styles.progressFill} style={{ width: `${acc}%` }} />
                    </div>
                </>
            ) : (
                <div className={styles.statSub} style={{ marginTop: '0.25rem' }}>Click to load</div>
            )}
        </div>
    );
}

export default PIIDiscovery;
