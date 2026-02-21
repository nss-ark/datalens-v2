import { useState, useCallback } from 'react';
import { useQuery } from '@tanstack/react-query';
import { DataTable, StatusBadge, Pagination, Button } from '@datalens/shared';
import { auditService } from '../services/auditService';
import type { AuditLog, AuditLogFilters } from '../services/auditService';
import { format } from 'date-fns';
import { FileText, X, ChevronDown, ChevronUp } from 'lucide-react';

// -- Constants --

const ENTITY_TYPES = ['DSR', 'CONSENT', 'BREACH', 'DATA_SOURCE', 'POLICY', 'USER', 'NOTICE', 'WIDGET'] as const;
const ACTIONS = ['CREATE', 'UPDATE', 'DELETE', 'APPROVE', 'REJECT'] as const;

const ACTION_VARIANT: Record<string, 'success' | 'danger' | 'info' | 'warning' | 'neutral'> = {
    CREATE: 'success',
    APPROVE: 'success',
    UPDATE: 'info',
    DELETE: 'danger',
    REJECT: 'danger',
};

const PAGE_SIZE = 20;

// -- Detail Expander --

function DetailCell({ log }: { log: AuditLog }) {
    const [expanded, setExpanded] = useState(false);
    const hasDetails = log.old_values || log.new_values;

    if (!hasDetails) {
        return <span className="text-xs text-gray-400">—</span>;
    }

    return (
        <div>
            <button
                onClick={(e) => { e.stopPropagation(); setExpanded(!expanded); }}
                className="inline-flex items-center gap-1 text-xs text-blue-600 hover:text-blue-800 font-medium"
            >
                {expanded ? <ChevronUp size={12} /> : <ChevronDown size={12} />}
                {expanded ? 'Hide' : 'Details'}
            </button>
            {expanded && (
                <div className="mt-2 space-y-1 text-[11px] font-mono bg-gray-50 rounded p-2 max-w-xs overflow-x-auto">
                    {log.old_values && (
                        <div>
                            <span className="font-semibold text-red-600">Old:</span>{' '}
                            <span className="text-gray-600">{JSON.stringify(log.old_values, null, 1)}</span>
                        </div>
                    )}
                    {log.new_values && (
                        <div>
                            <span className="font-semibold text-green-600">New:</span>{' '}
                            <span className="text-gray-600">{JSON.stringify(log.new_values, null, 1)}</span>
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}

// -- Page --

export default function AuditLogs() {
    const [page, setPage] = useState(1);
    const [entityType, setEntityType] = useState('');
    const [action, setAction] = useState('');
    const [startDate, setStartDate] = useState('');
    const [endDate, setEndDate] = useState('');

    const filters: AuditLogFilters = {
        page,
        page_size: PAGE_SIZE,
        ...(entityType && { entity_type: entityType }),
        ...(action && { action }),
        ...(startDate && { start_date: new Date(startDate).toISOString() }),
        ...(endDate && { end_date: new Date(endDate).toISOString() }),
    };

    const { data, isLoading } = useQuery({
        queryKey: ['audit-logs', filters],
        queryFn: () => auditService.list(filters),
    });

    const clearFilters = useCallback(() => {
        setEntityType('');
        setAction('');
        setStartDate('');
        setEndDate('');
        setPage(1);
    }, []);

    const hasFilters = entityType || action || startDate || endDate;

    const columns = [
        {
            key: 'created_at',
            header: 'Timestamp',
            sortable: true,
            width: '160px',
            render: (row: AuditLog) => (
                <span className="text-sm text-gray-700 whitespace-nowrap">
                    {format(new Date(row.created_at), 'MMM d, yyyy HH:mm')}
                </span>
            ),
        },
        {
            key: 'action',
            header: 'Action',
            width: '120px',
            render: (row: AuditLog) => (
                <StatusBadge
                    label={row.action}
                    variant={ACTION_VARIANT[row.action] || 'neutral'}
                />
            ),
        },
        {
            key: 'resource_type',
            header: 'Entity Type',
            width: '120px',
            render: (row: AuditLog) => (
                <span className="text-xs px-2 py-1 bg-gray-100 rounded-full text-gray-600 font-medium">
                    {row.resource_type}
                </span>
            ),
        },
        {
            key: 'resource_id',
            header: 'Resource ID',
            width: '120px',
            render: (row: AuditLog) => (
                <span className="text-xs font-mono text-gray-500" title={row.resource_id}>
                    {row.resource_id ? `${row.resource_id.slice(0, 8)}…` : '—'}
                </span>
            ),
        },
        {
            key: 'user_id',
            header: 'User',
            width: '120px',
            render: (row: AuditLog) => (
                <span className="text-xs font-mono text-gray-500" title={row.user_id}>
                    {row.user_id ? `${row.user_id.slice(0, 8)}…` : '—'}
                </span>
            ),
        },
        {
            key: 'ip_address',
            header: 'IP Address',
            width: '120px',
            render: (row: AuditLog) => (
                <span className="text-sm text-gray-600">{row.ip_address || '—'}</span>
            ),
        },
        {
            key: 'details',
            header: 'Details',
            width: '140px',
            render: (row: AuditLog) => <DetailCell log={row} />,
        },
    ];

    return (
        <div className="p-6">
            {/* Header */}
            <div className="flex justify-between items-center mb-6">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
                        <FileText className="text-blue-600" size={24} />
                        Audit Logs
                    </h1>
                    <p className="text-gray-500 mt-1">Track all system activity and changes</p>
                </div>
            </div>

            {/* Filter Bar */}
            <div className="flex flex-wrap items-end gap-4 mb-6 p-4 bg-white rounded-lg border border-gray-200 shadow-sm">
                {/* Entity Type */}
                <div className="flex flex-col gap-1">
                    <label htmlFor="filter-entity" className="text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Entity Type
                    </label>
                    <select
                        id="filter-entity"
                        value={entityType}
                        onChange={(e) => { setEntityType(e.target.value); setPage(1); }}
                        className="h-9 rounded-md border border-gray-300 bg-white px-3 text-sm text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    >
                        <option value="">All Types</option>
                        {ENTITY_TYPES.map((t) => (
                            <option key={t} value={t}>{t}</option>
                        ))}
                    </select>
                </div>

                {/* Action */}
                <div className="flex flex-col gap-1">
                    <label htmlFor="filter-action" className="text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Action
                    </label>
                    <select
                        id="filter-action"
                        value={action}
                        onChange={(e) => { setAction(e.target.value); setPage(1); }}
                        className="h-9 rounded-md border border-gray-300 bg-white px-3 text-sm text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    >
                        <option value="">All Actions</option>
                        {ACTIONS.map((a) => (
                            <option key={a} value={a}>{a}</option>
                        ))}
                    </select>
                </div>

                {/* Start Date */}
                <div className="flex flex-col gap-1">
                    <label htmlFor="filter-start" className="text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Start Date
                    </label>
                    <input
                        id="filter-start"
                        type="date"
                        value={startDate}
                        onChange={(e) => { setStartDate(e.target.value); setPage(1); }}
                        className="h-9 rounded-md border border-gray-300 bg-white px-3 text-sm text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                </div>

                {/* End Date */}
                <div className="flex flex-col gap-1">
                    <label htmlFor="filter-end" className="text-xs font-medium text-gray-500 uppercase tracking-wider">
                        End Date
                    </label>
                    <input
                        id="filter-end"
                        type="date"
                        value={endDate}
                        onChange={(e) => { setEndDate(e.target.value); setPage(1); }}
                        className="h-9 rounded-md border border-gray-300 bg-white px-3 text-sm text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                </div>

                {/* Clear */}
                {hasFilters && (
                    <Button
                        variant="secondary"
                        size="sm"
                        icon={<X size={14} />}
                        onClick={clearFilters}
                    >
                        Clear Filters
                    </Button>
                )}
            </div>

            {/* Table */}
            <DataTable<AuditLog>
                columns={columns}
                data={data?.items || []}
                isLoading={isLoading}
                keyExtractor={(row) => row.id}
                emptyTitle="No audit log entries found"
                emptyDescription="There are no audit log entries matching your filters."
            />

            {/* Pagination */}
            {data && data.total > 0 && (
                <div className="mt-4">
                    <Pagination
                        page={data.page}
                        pageSize={data.page_size}
                        total={data.total}
                        onPageChange={setPage}
                    />
                </div>
            )}
        </div>
    );
}
