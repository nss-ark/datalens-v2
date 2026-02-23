import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
    Loader2, Search, Filter, Download, Database, Cloud, HardDrive, Server,
    ChevronLeft, ChevronRight, SearchX,
} from 'lucide-react';
import { discoveryService } from '../services/discovery';
import type { PIIClassification, SensitivityLevel, VerificationStatus } from '../types/discovery';
import { useDataSources } from '../hooks/useDataSources';

const PAGE_SIZE = 20;

/* ─── Helpers ────────────────────────────────────────────────────────── */

const splitPath = (entityName: string): [string, string] => {
    const idx = entityName.indexOf('.');
    if (idx === -1) return [entityName, ''];
    return [entityName.slice(0, idx), entityName.slice(idx + 1)];
};

const sensitivityStyle = (level: SensitivityLevel): React.CSSProperties => {
    switch (level) {
        case 'CRITICAL':
            return { backgroundColor: '#fef2f2', color: '#991b1b', border: '1px solid #fee2e2' };
        case 'HIGH':
            return { backgroundColor: '#fff7ed', color: '#9a3412', border: '1px solid #fed7aa' };
        case 'MEDIUM':
            return { backgroundColor: '#fffbeb', color: '#92400e', border: '1px solid #fef3c7' };
        case 'LOW':
        default:
            return { backgroundColor: '#eff6ff', color: '#1e40af', border: '1px solid #dbeafe' };
    }
};

const statusStyle = (status: VerificationStatus): { bg: string; text: string; dotBg: string; borderColor: string } => {
    switch (status) {
        case 'VERIFIED':
            return { bg: '#f0fdf4', text: '#166534', dotBg: '#22c55e', borderColor: '#dcfce7' };
        case 'REJECTED':
            return { bg: '#fef2f2', text: '#991b1b', dotBg: '#ef4444', borderColor: '#fee2e2' };
        case 'PENDING':
        default:
            return { bg: '#fff7ed', text: '#9a3412', dotBg: '#fb923c', borderColor: '#fed7aa' };
    }
};

const getSourceIcon = (entityName: string) => {
    const lower = entityName.toLowerCase();
    if (lower.includes('mongo')) return <Server size={14} style={{ color: '#ea580c' }} />;
    if (lower.includes('snowflake') || lower.includes('cloud') || lower.includes('s3'))
        return <Cloud size={14} style={{ color: '#0284c7' }} />;
    if (lower.includes('mysql') || lower.includes('legacy'))
        return <HardDrive size={14} style={{ color: '#9333ea' }} />;
    return <Database size={14} style={{ color: '#4f46e5' }} />;
};

const getSourceLabel = (entityName: string): string => {
    const lower = entityName.toLowerCase();
    if (lower.includes('mongo')) return 'Mongo_Prod';
    if (lower.includes('mysql') || lower.includes('legacy')) return 'MySQL_Legacy';
    if (lower.includes('snowflake') || lower.includes('marketing')) return 'Snowflake';
    return 'PostgreSQL';
};

const getSourceIconBg = (entityName: string): string => {
    const lower = entityName.toLowerCase();
    if (lower.includes('mongo')) return '#fff7ed';
    if (lower.includes('snowflake') || lower.includes('cloud') || lower.includes('s3')) return '#f0f9ff';
    if (lower.includes('mysql') || lower.includes('legacy')) return '#faf5ff';
    return '#eef2ff';
};

const getDsIconComponent = (type: string) => {
    switch (type) {
        case 'MONGODB': return <Server size={14} />;
        case 'S3':
        case 'SNOWFLAKE':
        case 'AZURE_BLOB': return <Cloud size={14} />;
        case 'MYSQL':
        case 'SQLSERVER': return <HardDrive size={14} />;
        default: return <Database size={14} />;
    }
};

/* ─── Component ──────────────────────────────────────────────────────── */

const PIIInventory = () => {
    const [page, setPage] = useState(1);
    const [search, setSearch] = useState('');
    const [sourceFilter, setSourceFilter] = useState<string | null>(null);

    const { data: result, isLoading } = useQuery({
        queryKey: ['pii-inventory', page, sourceFilter],
        queryFn: () =>
            discoveryService.listClassifications({
                page,
                page_size: PAGE_SIZE,
                ...(sourceFilter ? { data_source_id: sourceFilter } : {}),
            }),
    });

    const { data: dataSources } = useDataSources();

    const classifications = result?.items ?? [];
    const total = result?.total ?? 0;
    const totalPages = result?.total_pages ?? 1;

    const filtered = search.trim()
        ? classifications.filter(
            (c) =>
                c.field_name.toLowerCase().includes(search.toLowerCase()) ||
                c.entity_name.toLowerCase().includes(search.toLowerCase()) ||
                c.category.toLowerCase().includes(search.toLowerCase())
        )
        : classifications;

    const showFrom = total === 0 ? 0 : (page - 1) * PAGE_SIZE + 1;
    const showTo = Math.min(page * PAGE_SIZE, total);

    return (
        <div style={{ maxWidth: '80rem', margin: '0 auto', padding: '2rem 1rem' }}>
            {/* Page Header */}
            <div style={{ marginBottom: '2rem' }}>
                <h1 style={{ fontSize: '1.875rem', fontWeight: 600, letterSpacing: '-0.025em', color: '#0f172a', marginBottom: '0.5rem' }}>
                    PII Inventory
                </h1>
                <p style={{ color: '#64748b', fontWeight: 300 }}>
                    Comprehensive list of all discovered PII across your data sources.
                </p>
            </div>

            {/* Filter Bar */}
            <div
                style={{
                    backgroundColor: '#ffffff', border: '1px solid #e2e8f0', borderRadius: '0.75rem',
                    boxShadow: '0 1px 2px rgba(0,0,0,0.05)', padding: '1rem', marginBottom: '1.5rem',
                    display: 'flex', flexWrap: 'wrap', alignItems: 'center', justifyContent: 'space-between', gap: '1rem',
                }}
            >
                {/* Left: source pills */}
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', overflowX: 'auto' }}>
                    <span style={{ fontSize: '0.875rem', fontWeight: 500, color: '#64748b', whiteSpace: 'nowrap', marginRight: '0.5rem' }}>
                        Data Sources:
                    </span>

                    <div
                        role="button"
                        onClick={() => { setSourceFilter(null); setPage(1); }}
                        style={{
                            display: 'flex', alignItems: 'center', gap: '0.5rem',
                            padding: '0.375rem 0.75rem', borderRadius: '9999px',
                            fontSize: '0.75rem', fontWeight: 500, cursor: 'pointer',
                            backgroundColor: sourceFilter === null ? '#0f172a' : '#f1f5f9',
                            color: sourceFilter === null ? '#ffffff' : '#475569',
                            border: sourceFilter === null ? 'none' : '1px solid transparent',
                        }}
                    >
                        <Database size={12} />
                        All Sources
                    </div>

                    {(dataSources || []).map((ds) => (
                        <div
                            key={ds.id}
                            role="button"
                            onClick={() => { setSourceFilter(ds.id); setPage(1); }}
                            style={{
                                display: 'flex', alignItems: 'center', gap: '0.5rem',
                                padding: '0.375rem 0.75rem', borderRadius: '9999px',
                                fontSize: '0.75rem', fontWeight: 500, cursor: 'pointer',
                                backgroundColor: sourceFilter === ds.id ? '#0f172a' : '#f1f5f9',
                                color: sourceFilter === ds.id ? '#ffffff' : '#475569',
                                border: sourceFilter === ds.id ? 'none' : '1px solid transparent',
                            }}
                        >
                            {getDsIconComponent(ds.type)}
                            {ds.name}
                        </div>
                    ))}
                </div>

                {/* Right: search + icons */}
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                    <div style={{ position: 'relative' }}>
                        <Search size={14} style={{ position: 'absolute', left: '0.75rem', top: '50%', transform: 'translateY(-50%)', color: '#94a3b8' }} />
                        <input
                            type="text"
                            placeholder="Search fields..."
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                            style={{
                                paddingLeft: '2.25rem', paddingRight: '1rem', paddingTop: '0.375rem', paddingBottom: '0.375rem',
                                fontSize: '0.875rem', backgroundColor: '#f8fafc', border: '1px solid #e2e8f0',
                                borderRadius: '0.5rem', outline: 'none', color: '#334155', width: '16rem',
                            }}
                        />
                    </div>

                    <div role="button" style={{ padding: '0.5rem', color: '#94a3b8', cursor: 'pointer' }} className="hover:text-slate-800">
                        <Filter size={18} />
                    </div>

                    <div role="button" style={{ padding: '0.5rem', color: '#94a3b8', cursor: 'pointer' }} className="hover:text-slate-800">
                        <Download size={18} />
                    </div>
                </div>
            </div>

            {/* Table Card */}
            <div
                style={{
                    backgroundColor: '#ffffff', border: '1px solid #e2e8f0', borderRadius: '0.75rem',
                    boxShadow: '0 1px 2px rgba(0,0,0,0.05)', overflow: 'hidden',
                }}
            >
                {isLoading ? (
                    <div style={{ display: 'flex', justifyContent: 'center', padding: '5rem 0' }}>
                        <Loader2 className="animate-spin" size={40} style={{ color: '#3b82f6' }} />
                    </div>
                ) : filtered.length === 0 ? (
                    <div style={{ textAlign: 'center', padding: '5rem 1.5rem' }}>
                        <SearchX size={48} style={{ color: '#cbd5e1', margin: '0 auto 1rem' }} />
                        <h3 style={{ fontSize: '1.125rem', fontWeight: 500, color: '#0f172a', marginBottom: '0.5rem' }}>
                            No PII found
                        </h3>
                        <p style={{ color: '#64748b', fontSize: '0.875rem' }}>
                            Connect data sources and run scans to populate your inventory.
                        </p>
                    </div>
                ) : (
                    <>
                        <div style={{ overflowX: 'auto' }}>
                            <table style={{ width: '100%', textAlign: 'left', borderCollapse: 'collapse' }}>
                                <thead>
                                    <tr style={{ borderBottom: '1px solid #e2e8f0', backgroundColor: 'rgba(248,250,252,0.5)' }}>
                                        {['Field Name / Path', 'Source', 'Category', 'Sensitivity', 'Confidence', 'Status', 'Method'].map(
                                            (h, i) => (
                                                <th
                                                    key={h}
                                                    style={{
                                                        padding: '1rem 1.5rem', fontSize: '0.75rem', fontWeight: 600,
                                                        color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em',
                                                        ...(i === 0 ? { width: '25%' } : {}),
                                                        ...(i === 6 ? { textAlign: 'right' as const } : {}),
                                                    }}
                                                >
                                                    {h}
                                                </th>
                                            )
                                        )}
                                    </tr>
                                </thead>
                                <tbody>
                                    {filtered.map((row, idx) => {
                                        const [sourceName, tableName] = splitPath(row.entity_name);
                                        const sens = sensitivityStyle(row.sensitivity);
                                        const st = statusStyle(row.status);
                                        const confPct = Math.round(row.confidence * 100);
                                        const isHighConf = confPct >= 90;

                                        return (
                                            <tr
                                                key={row.id}
                                                className="group hover:bg-slate-50 transition-colors"
                                                style={{ borderBottom: idx < filtered.length - 1 ? '1px solid #e2e8f0' : 'none' }}
                                            >
                                                {/* Field Name / Path */}
                                                <td style={{ padding: '1rem 1.5rem' }}>
                                                    <div style={{ display: 'flex', flexDirection: 'column' }}>
                                                        <span style={{ fontWeight: 500, color: '#0f172a', fontSize: '0.875rem' }}>
                                                            {row.field_name}
                                                        </span>
                                                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.25rem', marginTop: '0.25rem', fontSize: '0.75rem', color: '#64748b', fontFamily: 'ui-monospace, monospace' }}>
                                                            <span>{sourceName}</span>
                                                            {tableName && (
                                                                <>
                                                                    <span style={{ color: '#cbd5e1' }}>/</span>
                                                                    <span>{tableName}</span>
                                                                </>
                                                            )}
                                                        </div>
                                                    </div>
                                                </td>

                                                {/* Source */}
                                                <td style={{ padding: '1rem 1.5rem' }}>
                                                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#475569' }}>
                                                        <div
                                                            style={{
                                                                padding: '0.375rem', borderRadius: '0.375rem',
                                                                backgroundColor: getSourceIconBg(row.entity_name),
                                                                display: 'flex', alignItems: 'center', justifyContent: 'center',
                                                            }}
                                                        >
                                                            {getSourceIcon(row.entity_name)}
                                                        </div>
                                                        <span style={{ fontSize: '0.875rem', fontWeight: 500 }}>
                                                            {getSourceLabel(row.entity_name)}
                                                        </span>
                                                    </div>
                                                </td>

                                                {/* Category */}
                                                <td style={{ padding: '1rem 1.5rem' }}>
                                                    <div style={{ display: 'flex', flexDirection: 'column' }}>
                                                        <span style={{ fontSize: '0.875rem', fontWeight: 500, color: '#1e293b' }}>{row.category}</span>
                                                        <span style={{ fontSize: '0.75rem', color: '#64748b' }}>{row.type}</span>
                                                    </div>
                                                </td>

                                                {/* Sensitivity */}
                                                <td style={{ padding: '1rem 1.5rem' }}>
                                                    <span style={{ display: 'inline-flex', alignItems: 'center', padding: '0.25rem 0.625rem', borderRadius: '0.375rem', fontSize: '0.75rem', fontWeight: 500, ...sens }}>
                                                        {row.sensitivity}
                                                    </span>
                                                </td>

                                                {/* Confidence */}
                                                <td style={{ padding: '1rem 1.5rem' }}>
                                                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                                        <div style={{ width: '4rem', height: '0.375rem', backgroundColor: '#f1f5f9', borderRadius: '9999px', overflow: 'hidden' }}>
                                                            <div style={{ height: '100%', width: `${confPct}%`, backgroundColor: isHighConf ? '#10b981' : '#22c55e', borderRadius: '9999px' }} />
                                                        </div>
                                                        <span style={{ fontSize: '0.75rem', fontWeight: 500, color: isHighConf ? '#059669' : '#475569' }}>
                                                            {confPct}%
                                                        </span>
                                                    </div>
                                                </td>

                                                {/* Status */}
                                                <td style={{ padding: '1rem 1.5rem' }}>
                                                    <span
                                                        style={{
                                                            display: 'inline-flex', alignItems: 'center', gap: '0.25rem',
                                                            padding: '0.125rem 0.5rem', borderRadius: '9999px',
                                                            fontSize: '0.75rem', fontWeight: 500,
                                                            backgroundColor: st.bg, color: st.text, border: `1px solid ${st.borderColor}`,
                                                        }}
                                                    >
                                                        <span
                                                            className={row.status === 'PENDING' ? 'animate-pulse' : ''}
                                                            style={{ width: '0.375rem', height: '0.375rem', borderRadius: '9999px', backgroundColor: st.dotBg, display: 'inline-block' }}
                                                        />
                                                        {row.status}
                                                    </span>
                                                </td>

                                                {/* Method */}
                                                <td style={{ padding: '1rem 1.5rem', textAlign: 'right' }}>
                                                    <span style={{ fontSize: '0.75rem', fontFamily: 'ui-monospace, monospace', color: '#64748b' }}>
                                                        {row.detection_method}
                                                    </span>
                                                </td>
                                            </tr>
                                        );
                                    })}
                                </tbody>
                            </table>
                        </div>

                        {/* Pagination Footer */}
                        <div style={{ padding: '1rem 1.5rem', borderTop: '1px solid #e2e8f0', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                            <p style={{ fontSize: '0.875rem', color: '#64748b' }}>
                                Showing <span style={{ fontWeight: 500, color: '#0f172a' }}>{showFrom}</span> to{' '}
                                <span style={{ fontWeight: 500, color: '#0f172a' }}>{showTo}</span> of{' '}
                                <span style={{ fontWeight: 500, color: '#0f172a' }}>{total}</span> results
                            </p>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                <div
                                    role="button"
                                    onClick={() => page > 1 && setPage(page - 1)}
                                    style={{
                                        padding: '0.5rem', borderRadius: '0.5rem', border: '1px solid #e2e8f0',
                                        color: page <= 1 ? '#cbd5e1' : '#64748b', cursor: page <= 1 ? 'not-allowed' : 'pointer',
                                        opacity: page <= 1 ? 0.5 : 1, display: 'flex', alignItems: 'center', justifyContent: 'center',
                                    }}
                                >
                                    <ChevronLeft size={16} />
                                </div>
                                <div
                                    role="button"
                                    onClick={() => page < totalPages && setPage(page + 1)}
                                    style={{
                                        padding: '0.5rem', borderRadius: '0.5rem', border: '1px solid #e2e8f0',
                                        color: page >= totalPages ? '#cbd5e1' : '#64748b', cursor: page >= totalPages ? 'not-allowed' : 'pointer',
                                        opacity: page >= totalPages ? 0.5 : 1, display: 'flex', alignItems: 'center', justifyContent: 'center',
                                    }}
                                >
                                    <ChevronRight size={16} />
                                </div>
                            </div>
                        </div>
                    </>
                )}
            </div>
        </div>
    );
};

export default PIIInventory;
