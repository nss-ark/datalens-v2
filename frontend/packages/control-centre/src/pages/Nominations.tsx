import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Award, Filter, Info } from 'lucide-react';
import { DataTable, StatusBadge, Button } from '@datalens/shared';
import type { Column } from '@datalens/shared';
import { useQuery } from '@tanstack/react-query';
import { dsrService } from '../services/dsr';
import type { DSR } from '../types/dsr';

// ── Status filter options ────────────────────────────────────────────────

const STATUS_OPTIONS: { value: string; label: string }[] = [
    { value: '', label: 'All Statuses' },
    { value: 'PENDING', label: 'Pending' },
    { value: 'APPROVED', label: 'Approved' },
    { value: 'IN_PROGRESS', label: 'In Progress' },
    { value: 'COMPLETED', label: 'Completed' },
    { value: 'REJECTED', label: 'Rejected' },
];

// ── Page ─────────────────────────────────────────────────────────────────

const Nominations = () => {
    const navigate = useNavigate();
    const [page, setPage] = useState(1);
    const [statusFilter, setStatusFilter] = useState('');

    const { data, isLoading } = useQuery({
        queryKey: ['nominations', page, statusFilter],
        queryFn: () =>
            dsrService.list({
                page,
                page_size: 20,
                ...(statusFilter ? { status: statusFilter } : {}),
                type: 'NOMINATION',
            }),
    });

    const nominations = data?.items ?? [];

    const columns: Column<DSR>[] = [
        {
            key: 'id',
            header: 'ID',
            render: (row) => (
                <span style={{ fontSize: '0.75rem', fontFamily: 'monospace', color: '#6b7280' }}>
                    {row.id.slice(0, 8)}…
                </span>
            ),
        },
        {
            key: 'subject_name',
            header: 'Nominee Name',
            render: (row) => (
                <div>
                    <div style={{ fontWeight: 500, color: 'var(--text-primary)' }}>{row.subject_name}</div>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>{row.subject_email}</div>
                </div>
            ),
        },
        {
            key: 'status',
            header: 'Status',
            render: (row) => <StatusBadge label={row.status} />,
        },
        {
            key: 'priority',
            header: 'Priority',
            render: (row) => (
                <span style={{
                    fontSize: '0.75rem', fontWeight: 600,
                    color: row.priority === 'HIGH' ? 'var(--status-danger)' : row.priority === 'MEDIUM' ? 'var(--status-warning)' : 'var(--text-tertiary)',
                }}>
                    {row.priority}
                </span>
            ),
        },
        {
            key: 'created_at',
            header: 'Created',
            render: (row) => (
                <span style={{ fontSize: '0.8125rem', color: 'var(--text-secondary)' }}>
                    {new Date(row.created_at).toLocaleDateString()}
                </span>
            ),
        },
    ];

    const selectStyle: React.CSSProperties = {
        padding: '6px 12px',
        border: '1px solid var(--border-primary)',
        borderRadius: '6px',
        fontSize: '0.8125rem',
        backgroundColor: 'var(--bg-secondary)',
        color: 'var(--text-primary)',
    };

    return (
        <div style={{ padding: '24px 32px' }}>
            {/* Header */}
            <div style={{ marginBottom: '24px' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '4px' }}>
                    <Award size={24} style={{ color: '#8b5cf6' }} />
                    <h1 style={{ fontSize: '1.5rem', fontWeight: 700, color: 'var(--text-primary)' }}>
                        Nominations
                    </h1>
                </div>
                <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', marginBottom: '16px' }}>
                    DPDPA Section 14 — Right to Nominate
                </p>

                {/* Explainer */}
                <div style={{
                    display: 'flex', gap: '12px', alignItems: 'flex-start',
                    padding: '14px 18px', borderRadius: '12px',
                    backgroundColor: '#f5f3ff', border: '1px solid #e9e5ff',
                }}>
                    <Info size={18} style={{ color: '#7c3aed', flexShrink: 0, marginTop: 2 }} />
                    <div style={{ fontSize: '0.8125rem', color: '#4c1d95', lineHeight: 1.6 }}>
                        Under <strong>DPDPA Section 14</strong>, a Data Principal may nominate another individual
                        to exercise their data rights on their behalf — particularly in cases of the
                        Principal&apos;s incapacity or death. Nominations are tracked as a special category of
                        Data Subject Requests.
                    </div>
                </div>
            </div>

            {/* Filters */}
            <div style={{ display: 'flex', gap: '12px', marginBottom: '16px', alignItems: 'center' }}>
                <Filter size={16} style={{ color: 'var(--text-tertiary)' }} />
                <select
                    value={statusFilter}
                    onChange={e => { setStatusFilter(e.target.value); setPage(1); }}
                    style={selectStyle}
                >
                    {STATUS_OPTIONS.map(s => <option key={s.value} value={s.value}>{s.label}</option>)}
                </select>
                {data && (
                    <span style={{ marginLeft: 'auto', fontSize: '0.8125rem', color: 'var(--text-tertiary)' }}>
                        {data.total} nomination{data.total !== 1 ? 's' : ''}
                    </span>
                )}
            </div>

            {/* Table */}
            <DataTable
                columns={columns}
                data={nominations}
                isLoading={isLoading}
                keyExtractor={(row) => row.id}
                onRowClick={(row) => navigate(`/dsr/${row.id}`)}
                emptyTitle="No Nominations"
                emptyDescription="No nomination requests have been submitted yet. When a Data Principal nominates someone under DPDPA Section 14, it will appear here."
            />

            {/* Pagination */}
            {data && data.total_pages > 1 && (
                <div style={{ display: 'flex', justifyContent: 'center', gap: '8px', marginTop: '16px' }}>
                    <Button variant="ghost" size="sm" onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page <= 1}>
                        Previous
                    </Button>
                    <span style={{ display: 'flex', alignItems: 'center', fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
                        Page {page} of {data.total_pages}
                    </span>
                    <Button variant="ghost" size="sm" onClick={() => setPage(p => p + 1)} disabled={page >= data.total_pages}>
                        Next
                    </Button>
                </div>
            )}
        </div>
    );
};

export default Nominations;
