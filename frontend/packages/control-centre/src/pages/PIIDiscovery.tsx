import { useState } from 'react';
import {
    Loader2, Download, RefreshCw, TrendingUp, Sparkles, AlertTriangle,
    Check, Pencil, X, Filter, Search, ChevronLeft, ChevronRight,
    CheckCircle2, XCircle, MoreVertical, Database, Cloud, HardDrive, Server,
} from 'lucide-react';
import { Modal } from '@datalens/shared';
import { Button } from '@datalens/shared';
import { useClassifications, useSubmitFeedback, useAccuracyStats } from '../hooks/useDiscovery';
import { toast } from '@datalens/shared';
import type {
    PIIClassification,
    PIICategory,
    PIIType,
    DetectionMethod,
    VerificationStatus,
} from '../types/discovery';
import { PII_CATEGORIES, PII_TYPES } from '../types/discovery';

const PAGE_SIZE = 20;

/* ─── Style helpers ──────────────────────────────────────────────────── */

const methodBadgeStyle = (method: string): React.CSSProperties => {
    switch (method) {
        case 'HEURISTIC':
            return { backgroundColor: '#f3e8ff', color: '#6b21a8', border: '1px solid #e9d5ff' };
        case 'AI':
        case 'AI_MODEL':
            return { backgroundColor: '#e0e7ff', color: '#3730a3', border: '1px solid #c7d2fe' };
        case 'REGEX':
            return { backgroundColor: '#dbeafe', color: '#1e40af', border: '1px solid #bfdbfe' };
        case 'INDUSTRY':
            return { backgroundColor: '#f0fdf4', color: '#166534', border: '1px solid #dcfce7' };
        case 'MANUAL':
            return { backgroundColor: '#fff7ed', color: '#9a3412', border: '1px solid #fed7aa' };
        default:
            return { backgroundColor: '#f1f5f9', color: '#475569', border: '1px solid #e2e8f0' };
    }
};

const statusBadgeStyle = (status: VerificationStatus): React.CSSProperties => {
    switch (status) {
        case 'VERIFIED':
            return { backgroundColor: '#dcfce7', color: '#166534', border: '1px solid #bbf7d0' };
        case 'REJECTED':
            return { backgroundColor: '#fee2e2', color: '#991b1b', border: '1px solid #fecaca' };
        case 'PENDING':
        default:
            return { backgroundColor: '#fef9c3', color: '#854d0e', border: '1px solid #fde68a' };
    }
};

const getConfidenceColor = (pct: number) => {
    if (pct >= 80) return '#10b981';
    if (pct >= 60) return '#f59e0b';
    return '#ef4444';
};

const getSourceIcon = (entityName: string) => {
    const lower = entityName.toLowerCase();
    if (lower.includes('mongo')) return <Server size={14} />;
    if (lower.includes('snowflake') || lower.includes('cloud') || lower.includes('s3')) return <Cloud size={14} />;
    if (lower.includes('file') || lower.includes('legacy')) return <HardDrive size={14} />;
    return <Database size={14} />;
};

/* ─── Component ──────────────────────────────────────────────────────── */

const PIIDiscovery = () => {
    // ── Filters ──
    const [statusFilter, setStatusFilter] = useState<VerificationStatus | ''>('');
    const [methodFilter, setMethodFilter] = useState<DetectionMethod | ''>('');
    const [page, setPage] = useState(1);
    const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
    const [searchQuery, setSearchQuery] = useState('');

    // ── Data ──
    const { data: result, isLoading } = useClassifications({
        status: statusFilter || undefined,
        detection_method: methodFilter || undefined,
        page,
        page_size: PAGE_SIZE,
    });

    // Also fetch pending count separately for stats
    const { data: pendingResult } = useClassifications({ status: 'PENDING', page: 1, page_size: 1 });

    const classifications = result?.items ?? [];
    const total = result?.total ?? 0;
    const totalPages = result?.total_pages ?? 1;

    // ── Accuracy stats ──
    const { data: accuracyStats } = useAccuracyStats('AI');

    // ── Feedback ──
    const { mutate: submitFeedback, isPending: isFeedbackPending } = useSubmitFeedback();

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

    const handleBulkVerify = () => {
        selectedIds.forEach((id) => {
            submitFeedback({ classification_id: id, feedback_type: 'VERIFIED', notes: '' });
        });
        toast.success('Bulk Verified', `${selectedIds.size} classifications verified.`);
        setSelectedIds(new Set());
    };

    const handleBulkReject = () => {
        selectedIds.forEach((id) => {
            submitFeedback({ classification_id: id, feedback_type: 'REJECTED', notes: 'Bulk rejected' });
        });
        toast.success('Bulk Rejected', `${selectedIds.size} classifications rejected.`);
        setSelectedIds(new Set());
    };

    // ── Selection helpers ──
    const toggleSelect = (id: string) => {
        setSelectedIds((prev) => {
            const next = new Set(prev);
            if (next.has(id)) next.delete(id);
            else next.add(id);
            return next;
        });
    };

    const toggleSelectAll = () => {
        if (selectedIds.size === filteredRows.length) {
            setSelectedIds(new Set());
        } else {
            setSelectedIds(new Set(filteredRows.map((c) => c.id)));
        }
    };

    // ── Search filter (client-side on current page) ──
    const filteredRows = searchQuery.trim()
        ? classifications.filter(
            (c) =>
                c.field_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                c.entity_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                c.category.toLowerCase().includes(searchQuery.toLowerCase()) ||
                c.type.toLowerCase().includes(searchQuery.toLowerCase())
        )
        : classifications;

    // ── Derived stats (all from REAL API data) ──
    const pendingCount = pendingResult?.total ?? 0;
    const aiConfidence = accuracyStats ? Math.round(accuracyStats.accuracy * 100) : 0;
    const highRiskCount = classifications.filter(
        (c) => c.sensitivity === 'CRITICAL' || c.sensitivity === 'HIGH'
    ).length;
    const pendingPct = total > 0 ? Math.round((pendingCount / total) * 100) : 0;

    const showFrom = total === 0 ? 0 : (page - 1) * PAGE_SIZE + 1;
    const showTo = Math.min(page * PAGE_SIZE, total);

    const splitPath = (entityName: string): [string, string] => {
        const idx = entityName.indexOf('.');
        if (idx === -1) return [entityName, ''];
        return [entityName.slice(0, idx), entityName.slice(idx + 1)];
    };

    return (
        <div style={{ maxWidth: '100rem', margin: '0 auto', padding: '2rem 1.5rem', paddingBottom: '10rem' }}>

            {/* ── Page Header ── */}
            <div style={{ marginBottom: '2rem', display: 'flex', flexWrap: 'wrap', alignItems: 'flex-end', justifyContent: 'space-between', gap: '1rem' }}>
                <div>
                    <h2 style={{ fontSize: '1.875rem', fontWeight: 700, color: '#1e293b', marginBottom: '0.5rem' }}>
                        Review Classifications
                    </h2>
                    <p style={{ color: '#64748b', maxWidth: '42rem' }}>
                        Review auto-detected PII classifications and provide feedback to improve the discovery engine.
                    </p>
                </div>
                <div style={{ display: 'flex', gap: '0.75rem' }}>
                    <div
                        role="button"
                        style={{
                            display: 'flex', alignItems: 'center', gap: '0.5rem',
                            padding: '0.5rem 1rem', fontSize: '0.875rem', fontWeight: 500,
                            color: '#64748b', border: '1px solid #e2e8f0', borderRadius: '0.5rem',
                            cursor: 'pointer', backgroundColor: '#ffffff',
                        }}
                        className="hover:bg-slate-50"
                    >
                        <Download size={16} />
                        Export
                    </div>
                    <div
                        role="button"
                        style={{
                            display: 'flex', alignItems: 'center', gap: '0.5rem',
                            padding: '0.5rem 1rem', fontSize: '0.875rem', fontWeight: 500,
                            backgroundColor: '#3b82f6', color: '#ffffff', borderRadius: '0.5rem',
                            cursor: 'pointer', boxShadow: '0 4px 14px rgba(59,130,246,0.25)',
                        }}
                    >
                        <RefreshCw size={16} />
                        Scan Now
                    </div>
                </div>
            </div>

            {/* ── Stats Cards ── */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '1.5rem', marginBottom: '2.5rem' }}>
                {/* Pending Review (2-col span) */}
                <div
                    style={{
                        gridColumn: 'span 2',
                        backgroundColor: '#ffffff', border: '1px solid #e2e8f0', borderRadius: '0.75rem',
                        padding: '1.5rem',
                        boxShadow: '0 4px 6px -1px rgba(0,0,0,0.05), 0 2px 4px -1px rgba(0,0,0,0.03)',
                        position: 'relative', overflow: 'hidden',
                    }}
                >
                    <div style={{ position: 'relative', zIndex: 10 }}>
                        <p style={{ fontSize: '0.75rem', fontWeight: 500, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '0.25rem' }}>
                            Pending Review
                        </p>
                        <div style={{ display: 'flex', alignItems: 'baseline', gap: '0.5rem' }}>
                            <span style={{ fontSize: '2.25rem', fontWeight: 700, color: '#1e293b' }}>{pendingCount}</span>
                            {pendingCount > 0 && (
                                <span style={{ fontSize: '0.875rem', color: '#10b981', fontWeight: 500, display: 'flex', alignItems: 'center', gap: '0.125rem' }}>
                                    <TrendingUp size={14} /> awaiting review
                                </span>
                            )}
                        </div>
                        <div style={{ marginTop: '1rem', width: '100%', height: '0.5rem', backgroundColor: '#f1f5f9', borderRadius: '9999px', overflow: 'hidden' }}>
                            <div style={{ height: '100%', width: `${pendingPct}%`, backgroundColor: '#3b82f6', borderRadius: '9999px', transition: 'width 0.5s' }} />
                        </div>
                        <p style={{ fontSize: '0.75rem', color: '#64748b', marginTop: '0.5rem' }}>
                            {pendingPct}% of total discovered fields
                        </p>
                    </div>
                </div>

                {/* AI Confidence */}
                <div
                    style={{
                        backgroundColor: '#ffffff', border: '1px solid #e2e8f0', borderRadius: '0.75rem',
                        padding: '1.5rem',
                        boxShadow: '0 4px 6px -1px rgba(0,0,0,0.05), 0 2px 4px -1px rgba(0,0,0,0.03)',
                        display: 'flex', flexDirection: 'column', justifyContent: 'space-between',
                    }}
                >
                    <div>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '0.5rem' }}>
                            <p style={{ fontSize: '0.75rem', fontWeight: 500, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                                AI Confidence
                            </p>
                            <div style={{ padding: '0.375rem', borderRadius: '0.375rem', backgroundColor: '#dcfce7' }}>
                                <Sparkles size={18} style={{ color: '#10b981' }} />
                            </div>
                        </div>
                        <span style={{ fontSize: '1.875rem', fontWeight: 700, color: '#1e293b' }}>
                            {aiConfidence > 0 ? `${aiConfidence}%` : '—'}
                        </span>
                    </div>
                    <div style={{ marginTop: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.875rem', color: '#64748b' }}>
                        <span style={{ width: '0.5rem', height: '0.5rem', borderRadius: '9999px', backgroundColor: aiConfidence >= 80 ? '#10b981' : '#f59e0b', display: 'inline-block' }} />
                        {aiConfidence >= 80 ? 'High Accuracy Mode' : 'Building Accuracy'}
                    </div>
                </div>

                {/* High Risk */}
                <div
                    style={{
                        backgroundColor: '#ffffff', borderRadius: '0.75rem',
                        padding: '1.5rem',
                        boxShadow: '0 4px 6px -1px rgba(0,0,0,0.05), 0 2px 4px -1px rgba(0,0,0,0.03)',
                        display: 'flex', flexDirection: 'column', justifyContent: 'space-between',
                        border: '1px solid #e2e8f0', borderLeft: '4px solid #ef4444',
                    }}
                >
                    <div>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '0.5rem' }}>
                            <p style={{ fontSize: '0.75rem', fontWeight: 500, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                                High Risk
                            </p>
                            <div style={{ padding: '0.375rem', borderRadius: '0.375rem', backgroundColor: '#fee2e2' }}>
                                <AlertTriangle size={18} style={{ color: '#ef4444' }} />
                            </div>
                        </div>
                        <span style={{ fontSize: '1.875rem', fontWeight: 700, color: '#1e293b' }}>
                            {highRiskCount}
                        </span>
                    </div>
                    {highRiskCount > 0 ? (
                        <div style={{ marginTop: '1rem', fontSize: '0.75rem', fontWeight: 500, padding: '0.25rem 0.5rem', backgroundColor: '#fef2f2', color: '#ef4444', borderRadius: '0.25rem', width: 'fit-content' }}>
                            Requires Immediate Action
                        </div>
                    ) : (
                        <div style={{ marginTop: '1rem', fontSize: '0.75rem', fontWeight: 500, padding: '0.25rem 0.5rem', backgroundColor: '#f0fdf4', color: '#16a34a', borderRadius: '0.25rem', width: 'fit-content' }}>
                            No critical items
                        </div>
                    )}
                </div>
            </div>

            {/* ── Filter / Sort Bar ── */}
            <div
                style={{
                    display: 'flex', flexWrap: 'wrap', alignItems: 'center', justifyContent: 'space-between',
                    gap: '1rem', marginBottom: '1.5rem',
                    backgroundColor: '#ffffff', padding: '0.5rem', borderRadius: '0.75rem',
                    border: '1px solid #e2e8f0', boxShadow: '0 1px 2px rgba(0,0,0,0.05)',
                }}
            >
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    {/* All Statuses */}
                    <div
                        role="button"
                        onClick={() => { setStatusFilter(''); setMethodFilter(''); setPage(1); }}
                        style={{
                            padding: '0.375rem 0.75rem', fontSize: '0.875rem', fontWeight: 500,
                            backgroundColor: !statusFilter && !methodFilter ? '#f1f5f9' : 'transparent',
                            borderRadius: '0.5rem', cursor: 'pointer',
                            color: !statusFilter && !methodFilter ? '#1e293b' : '#64748b',
                            display: 'flex', alignItems: 'center', gap: '0.5rem',
                        }}
                    >
                        <Filter size={14} />
                        All Statuses
                    </div>

                    <div style={{ height: '1.5rem', width: '1px', backgroundColor: '#e2e8f0', margin: '0 0.25rem' }} />

                    {/* Method filters */}
                    {[
                        { key: 'HEURISTIC' as DetectionMethod, label: 'Heuristic' },
                        { key: 'AI' as DetectionMethod, label: 'AI Model' },
                        { key: 'REGEX' as DetectionMethod, label: 'Regex' },
                    ].map(({ key, label }) => (
                        <div
                            key={key}
                            role="button"
                            onClick={() => { setMethodFilter(methodFilter === key ? '' : key); setPage(1); }}
                            style={{
                                padding: '0.375rem 0.75rem', fontSize: '0.875rem',
                                color: methodFilter === key ? '#3b82f6' : '#64748b',
                                cursor: 'pointer', fontWeight: methodFilter === key ? 500 : 400,
                                backgroundColor: methodFilter === key ? '#eff6ff' : 'transparent',
                                borderRadius: '0.375rem',
                            }}
                        >
                            {label}
                        </div>
                    ))}
                </div>

                <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', paddingRight: '0.5rem' }}>
                    {/* Search */}
                    <div style={{ position: 'relative' }}>
                        <Search size={14} style={{ position: 'absolute', left: '0.625rem', top: '50%', transform: 'translateY(-50%)', color: '#94a3b8' }} />
                        <input
                            type="text"
                            placeholder="Search fields..."
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            style={{
                                paddingLeft: '2rem', paddingRight: '0.75rem', paddingTop: '0.375rem', paddingBottom: '0.375rem',
                                fontSize: '0.875rem', backgroundColor: '#f8fafc', border: '1px solid #e2e8f0',
                                borderRadius: '0.375rem', outline: 'none', color: '#334155', width: '12rem',
                            }}
                        />
                    </div>

                    <span style={{ fontSize: '0.75rem', fontWeight: 500, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Sort by:</span>
                    <select
                        style={{
                            backgroundColor: 'transparent', border: 'none',
                            fontSize: '0.875rem', fontWeight: 500, color: '#1e293b', cursor: 'pointer', outline: 'none',
                        }}
                    >
                        <option>Confidence (High to Low)</option>
                        <option>Date Discovered</option>
                        <option>Risk Level</option>
                    </select>
                </div>
            </div>

            {/* ── Table Card ── */}
            <div
                style={{
                    backgroundColor: '#ffffff', border: '1px solid #e2e8f0', borderRadius: '0.75rem',
                    boxShadow: '0 4px 6px -1px rgba(0,0,0,0.05), 0 2px 4px -1px rgba(0,0,0,0.03)',
                    overflow: 'hidden',
                }}
            >
                {isLoading ? (
                    <div style={{ display: 'flex', justifyContent: 'center', padding: '5rem 0' }}>
                        <Loader2 className="animate-spin" size={40} style={{ color: '#3b82f6' }} />
                    </div>
                ) : filteredRows.length === 0 ? (
                    <div style={{ textAlign: 'center', padding: '5rem 1.5rem' }}>
                        <Search size={48} style={{ color: '#cbd5e1', margin: '0 auto 1rem' }} />
                        <h3 style={{ fontSize: '1.125rem', fontWeight: 500, color: '#1e293b', marginBottom: '0.5rem' }}>
                            No PII classifications found
                        </h3>
                        <p style={{ color: '#64748b', fontSize: '0.875rem' }}>
                            Run a scan on a data source to discover PII fields.
                        </p>
                    </div>
                ) : (
                    <>
                        <div style={{ overflowX: 'auto' }}>
                            <table style={{ width: '100%', textAlign: 'left', borderCollapse: 'collapse' }}>
                                <thead>
                                    <tr
                                        style={{
                                            backgroundColor: 'rgba(248,250,252,0.5)',
                                            fontSize: '0.75rem', textTransform: 'uppercase', letterSpacing: '0.05em',
                                            fontWeight: 600, color: '#64748b', borderBottom: '1px solid #e2e8f0',
                                        }}
                                    >
                                        <th style={{ padding: '1rem', width: '3rem', textAlign: 'center' }}>
                                            <input
                                                type="checkbox"
                                                checked={selectedIds.size === filteredRows.length && filteredRows.length > 0}
                                                onChange={toggleSelectAll}
                                                style={{ width: '1rem', height: '1rem', cursor: 'pointer', accentColor: '#3b82f6' }}
                                            />
                                        </th>
                                        <th style={{ padding: '1rem' }}>Field Name</th>
                                        <th style={{ padding: '1rem' }}>Data Source</th>
                                        <th style={{ padding: '1rem' }}>Category</th>
                                        <th style={{ padding: '1rem' }}>Type</th>
                                        <th style={{ padding: '1rem' }}>Confidence</th>
                                        <th style={{ padding: '1rem' }}>Method</th>
                                        <th style={{ padding: '1rem' }}>Status</th>
                                        <th style={{ padding: '1rem', textAlign: 'right' }}>Actions</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {filteredRows.map((row, idx) => {
                                        const isSelected = selectedIds.has(row.id);
                                        const confPct = Math.round(row.confidence * 100);
                                        const confColor = getConfidenceColor(confPct);
                                        const [source, table] = splitPath(row.entity_name);

                                        return (
                                            <tr
                                                key={row.id}
                                                className="group hover:bg-slate-50 transition-colors"
                                                style={{
                                                    borderBottom: idx < filteredRows.length - 1 ? '1px solid #e2e8f0' : 'none',
                                                    backgroundColor: isSelected ? 'rgba(239,246,255,0.4)' : undefined,
                                                }}
                                            >
                                                {/* Checkbox */}
                                                <td style={{ padding: '1rem', textAlign: 'center' }}>
                                                    <input
                                                        type="checkbox"
                                                        checked={isSelected}
                                                        onChange={() => toggleSelect(row.id)}
                                                        style={{ width: '1rem', height: '1rem', cursor: 'pointer', accentColor: '#3b82f6' }}
                                                    />
                                                </td>

                                                {/* Field Name */}
                                                <td style={{ padding: '1rem' }}>
                                                    <div style={{ fontWeight: 600, color: '#1e293b', fontSize: '0.875rem' }}>
                                                        {row.field_name}
                                                    </div>
                                                    <div style={{ fontSize: '0.75rem', color: '#64748b' }}>
                                                        {table || source}
                                                    </div>
                                                </td>

                                                {/* Data Source */}
                                                <td style={{ padding: '1rem' }}>
                                                    <div
                                                        style={{
                                                            display: 'inline-flex', alignItems: 'center', gap: '0.375rem',
                                                            fontSize: '0.75rem', color: '#64748b',
                                                            backgroundColor: '#f1f5f9', padding: '0.25rem 0.5rem',
                                                            borderRadius: '0.25rem', border: '1px solid #e2e8f0',
                                                            width: 'fit-content',
                                                        }}
                                                    >
                                                        {getSourceIcon(row.entity_name)}
                                                        {source}{table ? ` / ${table}` : ''}
                                                    </div>
                                                </td>

                                                {/* Category */}
                                                <td style={{ padding: '1rem', color: '#64748b', fontSize: '0.875rem' }}>
                                                    {row.category}
                                                </td>

                                                {/* Type */}
                                                <td style={{ padding: '1rem', fontWeight: 500, fontSize: '0.875rem', color: '#1e293b' }}>
                                                    {row.type}
                                                </td>

                                                {/* Confidence */}
                                                <td style={{ padding: '1rem' }}>
                                                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                                        <div style={{ width: '4rem', height: '0.375rem', backgroundColor: '#e2e8f0', borderRadius: '9999px', overflow: 'hidden' }}>
                                                            <div style={{ height: '100%', width: `${confPct}%`, backgroundColor: confColor, borderRadius: '9999px' }} />
                                                        </div>
                                                        <span style={{ fontWeight: 500, fontSize: '0.875rem', color: confColor }}>
                                                            {confPct}%
                                                        </span>
                                                    </div>
                                                </td>

                                                {/* Method */}
                                                <td style={{ padding: '1rem' }}>
                                                    <span
                                                        style={{
                                                            display: 'inline-flex', alignItems: 'center',
                                                            padding: '0.125rem 0.5rem', borderRadius: '0.25rem',
                                                            fontSize: '0.75rem', fontWeight: 500,
                                                            ...methodBadgeStyle(row.detection_method),
                                                        }}
                                                    >
                                                        {row.detection_method}
                                                    </span>
                                                </td>

                                                {/* Status */}
                                                <td style={{ padding: '1rem' }}>
                                                    <span
                                                        style={{
                                                            display: 'inline-flex', alignItems: 'center',
                                                            padding: '0.125rem 0.5rem', borderRadius: '0.25rem',
                                                            fontSize: '0.75rem', fontWeight: 500,
                                                            ...statusBadgeStyle(row.status),
                                                        }}
                                                    >
                                                        {row.status}
                                                    </span>
                                                </td>

                                                {/* Actions */}
                                                <td style={{ padding: '1rem', textAlign: 'right' }}>
                                                    {row.status === 'PENDING' ? (
                                                        <div
                                                            className="opacity-0 group-hover:opacity-100 transition-opacity"
                                                            style={{ display: 'flex', alignItems: 'center', justifyContent: 'flex-end', gap: '0.375rem' }}
                                                        >
                                                            <div
                                                                role="button"
                                                                title="Approve"
                                                                onClick={(e) => handleVerify(row, e)}
                                                                style={{ padding: '0.375rem', borderRadius: '0.375rem', color: '#10b981', cursor: 'pointer' }}
                                                                className="hover:bg-green-50"
                                                            >
                                                                <Check size={16} />
                                                            </div>
                                                            <div
                                                                role="button"
                                                                title="Edit"
                                                                onClick={(e) => openCorrectModal(row, e)}
                                                                style={{ padding: '0.375rem', borderRadius: '0.375rem', color: '#64748b', cursor: 'pointer' }}
                                                                className="hover:bg-slate-100"
                                                            >
                                                                <Pencil size={16} />
                                                            </div>
                                                            <div
                                                                role="button"
                                                                title="Reject"
                                                                onClick={(e) => openRejectModal(row, e)}
                                                                style={{ padding: '0.375rem', borderRadius: '0.375rem', color: '#ef4444', cursor: 'pointer' }}
                                                                className="hover:bg-red-50"
                                                            >
                                                                <X size={16} />
                                                            </div>
                                                        </div>
                                                    ) : (
                                                        <span style={{ fontSize: '0.75rem', color: '#94a3b8', fontStyle: 'italic' }}>
                                                            Reviewed
                                                        </span>
                                                    )}
                                                </td>
                                            </tr>
                                        );
                                    })}
                                </tbody>
                            </table>
                        </div>

                        {/* Pagination Footer */}
                        <div
                            style={{
                                display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                                padding: '0.75rem 1rem',
                                backgroundColor: 'rgba(248,250,252,0.5)',
                                borderTop: '1px solid #e2e8f0',
                            }}
                        >
                            <div style={{ fontSize: '0.875rem', color: '#64748b' }}>
                                Showing{' '}
                                <span style={{ fontWeight: 500, color: '#1e293b' }}>{showFrom}</span> to{' '}
                                <span style={{ fontWeight: 500, color: '#1e293b' }}>{showTo}</span> of{' '}
                                <span style={{ fontWeight: 500, color: '#1e293b' }}>{total}</span> results
                            </div>
                            <div style={{ display: 'flex', gap: '0.5rem' }}>
                                <div
                                    role="button"
                                    onClick={() => page > 1 && setPage(page - 1)}
                                    style={{
                                        display: 'flex', alignItems: 'center', gap: '0.25rem',
                                        padding: '0.25rem 0.75rem', fontSize: '0.875rem',
                                        borderRadius: '0.25rem', border: '1px solid #e2e8f0',
                                        color: page <= 1 ? '#cbd5e1' : '#64748b',
                                        cursor: page <= 1 ? 'not-allowed' : 'pointer',
                                        opacity: page <= 1 ? 0.5 : 1,
                                    }}
                                >
                                    <ChevronLeft size={14} /> Previous
                                </div>
                                <div
                                    role="button"
                                    onClick={() => page < totalPages && setPage(page + 1)}
                                    style={{
                                        display: 'flex', alignItems: 'center', gap: '0.25rem',
                                        padding: '0.25rem 0.75rem', fontSize: '0.875rem',
                                        borderRadius: '0.25rem', border: '1px solid #e2e8f0',
                                        color: page >= totalPages ? '#cbd5e1' : '#1e293b',
                                        cursor: page >= totalPages ? 'not-allowed' : 'pointer',
                                        opacity: page >= totalPages ? 0.5 : 1,
                                    }}
                                >
                                    Next <ChevronRight size={14} />
                                </div>
                            </div>
                        </div>
                    </>
                )}
            </div>

            {/* ── Floating Bulk Action Bar ── */}
            {selectedIds.size > 0 && (
                <div style={{ position: 'fixed', bottom: '1.5rem', left: '50%', transform: 'translateX(-50%)', zIndex: 50 }}>
                    <div
                        style={{
                            backgroundColor: 'rgba(15,23,42,0.92)',
                            backdropFilter: 'blur(12px)',
                            border: '1px solid rgba(255,255,255,0.1)',
                            borderRadius: '1rem',
                            boxShadow: '0 8px 32px rgba(0,0,0,0.3)',
                            padding: '0.5rem', display: 'flex', alignItems: 'center', gap: '1.5rem',
                        }}
                    >
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', padding: '0 1rem', borderRight: '1px solid rgba(255,255,255,0.1)' }}>
                            <div
                                style={{
                                    width: '1.25rem', height: '1.25rem', borderRadius: '0.25rem',
                                    backgroundColor: '#3b82f6', display: 'flex', alignItems: 'center',
                                    justifyContent: 'center', color: '#fff', fontSize: '0.75rem', fontWeight: 700,
                                }}
                            >
                                {selectedIds.size}
                            </div>
                            <span style={{ fontSize: '0.875rem', fontWeight: 500, color: '#ffffff', whiteSpace: 'nowrap' }}>fields selected</span>
                        </div>

                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', paddingRight: '0.5rem' }}>
                            <div
                                role="button"
                                onClick={handleBulkVerify}
                                style={{
                                    display: 'flex', alignItems: 'center', gap: '0.5rem',
                                    padding: '0.5rem 1rem', borderRadius: '0.75rem',
                                    fontSize: '0.875rem', fontWeight: 600,
                                    backgroundColor: 'rgba(16,185,129,0.1)', color: '#34d399',
                                    border: '1px solid rgba(16,185,129,0.2)', cursor: 'pointer',
                                    whiteSpace: 'nowrap',
                                }}
                            >
                                <CheckCircle2 size={16} />
                                Verify Selected
                            </div>

                            <div
                                role="button"
                                onClick={handleBulkReject}
                                style={{
                                    display: 'flex', alignItems: 'center', gap: '0.5rem',
                                    padding: '0.5rem 1rem', borderRadius: '0.75rem',
                                    fontSize: '0.875rem', fontWeight: 600,
                                    backgroundColor: 'rgba(239,68,68,0.1)', color: '#f87171',
                                    border: '1px solid rgba(239,68,68,0.2)', cursor: 'pointer',
                                    whiteSpace: 'nowrap',
                                }}
                            >
                                <XCircle size={16} />
                                Reject Selected
                            </div>

                            <div
                                role="button"
                                style={{ padding: '0.5rem', borderRadius: '0.75rem', color: '#94a3b8', cursor: 'pointer' }}
                            >
                                <MoreVertical size={18} />
                            </div>
                        </div>
                    </div>
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
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                        <p style={{ fontSize: '0.875rem', color: '#64748b' }}>
                            Correcting <strong>{correctModal.field_name}</strong> in <strong>{correctModal.entity_name}</strong>
                        </p>
                        <div>
                            <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: 500, color: '#334155', marginBottom: '0.375rem' }}>
                                Correct Category
                            </label>
                            <select
                                value={corrCategory}
                                onChange={(e) => setCorrCategory(e.target.value as PIICategory)}
                                style={{ width: '100%', height: '2.5rem', padding: '0 0.75rem', borderRadius: '0.375rem', border: '1px solid #e2e8f0', fontSize: '0.875rem', outline: 'none' }}
                            >
                                {PII_CATEGORIES.map((c) => (
                                    <option key={c.value} value={c.value}>{c.label}</option>
                                ))}
                            </select>
                        </div>
                        <div>
                            <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: 500, color: '#334155', marginBottom: '0.375rem' }}>
                                Correct Type
                            </label>
                            <select
                                value={corrType}
                                onChange={(e) => setCorrType(e.target.value as PIIType)}
                                style={{ width: '100%', height: '2.5rem', padding: '0 0.75rem', borderRadius: '0.375rem', border: '1px solid #e2e8f0', fontSize: '0.875rem', outline: 'none' }}
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
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                        <p style={{ fontSize: '0.875rem', color: '#64748b' }}>
                            Rejecting <strong>{rejectModal.field_name}</strong> ({rejectModal.category} → {rejectModal.type})
                        </p>
                        <div>
                            <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: 500, color: '#334155', marginBottom: '0.375rem' }}>
                                Reason / Notes
                            </label>
                            <textarea
                                value={rejectNotes}
                                onChange={(e) => setRejectNotes(e.target.value)}
                                placeholder="Why is this a false positive?"
                                style={{ width: '100%', height: '5rem', padding: '0.5rem 0.75rem', borderRadius: '0.375rem', border: '1px solid #e2e8f0', fontSize: '0.875rem', outline: 'none', resize: 'none' }}
                            />
                        </div>
                    </div>
                )}
            </Modal>
        </div>
    );
};

export default PIIDiscovery;
