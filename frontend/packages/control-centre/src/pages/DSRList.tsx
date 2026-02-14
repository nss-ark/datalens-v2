import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { FileText, Plus, Filter } from 'lucide-react';
import { DataTable } from '@datalens/shared';
import { Button } from '@datalens/shared';
import { StatusBadge } from '@datalens/shared';
import { CreateDSRModal } from '../components/DSR/CreateDSRModal';
import { useDSRs } from '../hooks/useDSR';
import type { DSR, DSRRequestType } from '../types/dsr';
import type { Column } from '@datalens/shared';

// SLA helpers
function getDaysRemaining(deadline: string): number {
    const now = new Date();
    const sla = new Date(deadline);
    const diff = sla.getTime() - now.getTime();
    return Math.ceil(diff / (1000 * 60 * 60 * 24));
}

function formatSLA(deadline: string): string {
    const days = getDaysRemaining(deadline);
    if (days < 0) return `${Math.abs(days)}d overdue`;
    if (days === 0) return 'Due today';
    return `${days}d remaining`;
}

const STATUS_OPTIONS: { value: string; label: string }[] = [
    { value: '', label: 'All Statuses' },
    { value: 'PENDING', label: 'Pending' },
    { value: 'APPROVED', label: 'Approved' },
    { value: 'IN_PROGRESS', label: 'In Progress' },
    { value: 'COMPLETED', label: 'Completed' },
    { value: 'REJECTED', label: 'Rejected' },
    { value: 'FAILED', label: 'Failed' },
];

const TYPE_OPTIONS: { value: string; label: string }[] = [
    { value: '', label: 'All Types' },
    { value: 'ACCESS', label: 'Access' },
    { value: 'ERASURE', label: 'Erasure' },
    { value: 'CORRECTION', label: 'Correction' },
    { value: 'PORTABILITY', label: 'Portability' },
];

const DSRList = () => {
    const navigate = useNavigate();
    const [showCreate, setShowCreate] = useState(false);
    const [statusFilter, setStatusFilter] = useState('');
    const [typeFilter, setTypeFilter] = useState('');
    const [page, setPage] = useState(1);

    const { data, isLoading } = useDSRs({
        page,
        page_size: 20,
        ...(statusFilter ? { status: statusFilter } : {}),
    });

    const dsrs = data?.items ?? [];

    // Client-side type filter (backend only supports status filter)
    const filtered = typeFilter
        ? dsrs.filter((d: DSR) => d.request_type === typeFilter)
        : dsrs;

    const columns: Column<DSR>[] = [
        {
            key: 'subject_name',
            header: 'Subject',
            render: (row) => (
                <div>
                    <div style={{ fontWeight: 500, color: 'var(--text-primary)' }}>{row.subject_name}</div>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>{row.subject_email}</div>
                </div>
            ),
        },
        {
            key: 'request_type',
            header: 'Type',
            render: (row) => (
                <span style={{
                    padding: '2px 8px',
                    borderRadius: '4px',
                    fontSize: '0.75rem',
                    fontWeight: 600,
                    backgroundColor: getTypeColor(row.request_type).bg,
                    color: getTypeColor(row.request_type).text,
                }}>
                    {row.request_type}
                </span>
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
                    fontSize: '0.75rem',
                    fontWeight: 600,
                    color: row.priority === 'HIGH' ? 'var(--status-danger)' : row.priority === 'MEDIUM' ? 'var(--status-warning)' : 'var(--text-tertiary)',
                }}>
                    {row.priority}
                </span>
            ),
        },
        {
            key: 'sla_deadline',
            header: 'SLA Deadline',
            render: (row) => {
                const days = getDaysRemaining(row.sla_deadline);
                const isOverdue = days < 0;
                const isUrgent = days >= 0 && days <= 3;
                return (
                    <div>
                        <div style={{
                            fontWeight: 500,
                            color: isOverdue ? 'var(--status-danger)' : isUrgent ? 'var(--status-warning)' : 'var(--text-primary)',
                        }}>
                            {formatSLA(row.sla_deadline)}
                        </div>
                        <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>
                            {new Date(row.sla_deadline).toLocaleDateString()}
                        </div>
                    </div>
                );
            },
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
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '24px' }}>
                <div>
                    <h1 style={{ fontSize: '1.5rem', fontWeight: 700, color: 'var(--text-primary)', display: 'flex', alignItems: 'center', gap: '10px' }}>
                        <FileText size={24} />
                        DSR Management
                    </h1>
                    <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', marginTop: '4px' }}>
                        Manage Data Subject Requests — access, erasure, correction, and portability
                    </p>
                </div>
                <Button icon={<Plus size={16} />} onClick={() => setShowCreate(true)}>
                    New DSR
                </Button>
            </div>

            {/* Filters */}
            <div style={{ display: 'flex', gap: '12px', marginBottom: '16px', alignItems: 'center' }}>
                <Filter size={16} style={{ color: 'var(--text-tertiary)' }} />
                <select value={statusFilter} onChange={e => { setStatusFilter(e.target.value); setPage(1); }} style={selectStyle}>
                    {STATUS_OPTIONS.map(s => <option key={s.value} value={s.value}>{s.label}</option>)}
                </select>
                <select value={typeFilter} onChange={e => { setTypeFilter(e.target.value); setPage(1); }} style={selectStyle}>
                    {TYPE_OPTIONS.map(t => <option key={t.value} value={t.value}>{t.label}</option>)}
                </select>
                {data && (
                    <span style={{ marginLeft: 'auto', fontSize: '0.8125rem', color: 'var(--text-tertiary)' }}>
                        {data.total} total requests
                    </span>
                )}
            </div>

            {/* Table */}
            <DataTable
                columns={columns}
                data={filtered}
                isLoading={isLoading}
                keyExtractor={(row) => row.id}
                onRowClick={(row) => navigate(`/dsr/${row.id}`)}
                emptyTitle="No DSR Requests"
                emptyDescription="Create a new Data Subject Request to get started."
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

            {/* Create Modal */}
            <CreateDSRModal open={showCreate} onClose={() => setShowCreate(false)} />
        </div>
    );
};

function getTypeColor(type: DSRRequestType): { bg: string; text: string } {
    switch (type) {
        case 'ACCESS': return { bg: '#dbeafe', text: '#1d4ed8' };
        case 'ERASURE': return { bg: '#fee2e2', text: '#b91c1c' };
        case 'CORRECTION': return { bg: '#fef3c7', text: '#b45309' };
        case 'PORTABILITY': return { bg: '#e0e7ff', text: '#4338ca' };
    }
}

export default DSRList;
