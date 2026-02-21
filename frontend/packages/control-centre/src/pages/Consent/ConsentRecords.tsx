import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { DataTable, StatusBadge, Button } from '@datalens/shared';
import { consentRecordService } from '../../services/consentRecordService';
import type { ConsentSession } from '../../services/consentRecordService';
import { useNavigate } from 'react-router-dom';
import { format } from 'date-fns';

export default function ConsentRecords() {
    const navigate = useNavigate();
    const [page, setPage] = useState(1);
    const [statusFilter, setStatusFilter] = useState('');

    const { data, isLoading } = useQuery({
        queryKey: ['consent-records', page, statusFilter],
        queryFn: () => consentRecordService.list({
            page,
            page_size: 10,
            status: statusFilter || undefined
        })
    });

    const columns = [
        {
            key: 'id',
            header: 'Session ID',
            render: (row: ConsentSession) => (
                <span className="font-mono text-xs text-gray-500" title={row.id}>
                    {row.id.substring(0, 8)}...
                </span>
            )
        },
        {
            key: 'subject',
            header: 'Subject',
            render: (row: ConsentSession) => (
                <div className="max-w-xs truncate font-medium text-gray-900" title={row.subject_id}>
                    {row.subject_id ? `${row.subject_id.substring(0, 8)}...` : 'Unknown'}
                </div>
            )
        },
        {
            key: 'status',
            header: 'Status',
            render: (row: ConsentSession) => {
                const statusLabel = row.status || (row.decisions?.every(d => !d.granted) ? 'WITHDRAWN' : 'GRANTED');
                return <StatusBadge label={statusLabel} />;
            }
        },
        {
            key: 'purposes',
            header: 'Purposes',
            render: (row: ConsentSession) => (
                <span className="text-sm text-gray-600">
                    {row.decisions?.length || 0} purposes
                </span>
            )
        },
        {
            key: 'widget',
            header: 'Widget',
            render: (row: ConsentSession) => (
                <button
                    onClick={() => navigate(`/consent/widgets/${row.widget_id}`)}
                    className="text-blue-600 hover:text-blue-800 hover:underline text-sm font-medium transition-colors cursor-pointer text-left"
                >
                    {row.widget_id ? `${row.widget_id.substring(0, 8)}...` : 'N/A'}
                </button>
            )
        },
        {
            key: 'created_at',
            header: 'Timestamp',
            render: (row: ConsentSession) => (
                <span className="text-sm text-gray-500">
                    {row.created_at ? format(new Date(row.created_at), 'MMM d, yyyy, HH:mm') : ''}
                </span>
            )
        }
    ];

    return (
        <div className="p-6 max-w-7xl mx-auto space-y-6 py-12">
            <div className="flex justify-between items-center mb-6">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900">Consent Records</h1>
                    <p className="text-gray-500 mt-1">View all consent sessions across your organization.</p>
                </div>
            </div>

            <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6 space-y-6 flex flex-col min-h-[500px]">
                <div className="flex space-x-2">
                    {['', 'GRANTED', 'WITHDRAWN', 'EXPIRED'].map(status => (
                        <button
                            key={status}
                            onClick={() => {
                                setStatusFilter(status);
                                setPage(1);
                            }}
                            className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${statusFilter === status
                                ? 'bg-blue-100 text-blue-700'
                                : 'bg-white text-gray-600 hover:bg-gray-50 border border-gray-200'
                                }`}
                        >
                            {status === '' ? 'All' : status}
                        </button>
                    ))}
                </div>

                <div className="flex-1">
                    <DataTable
                        columns={columns}
                        data={data?.items || []}
                        isLoading={isLoading}
                        keyExtractor={(row) => row.id}
                        emptyTitle="No consent records found"
                        emptyDescription="There are no consent sessions matching your criteria."
                    />
                </div>

                {data && data.total_pages > 1 && (
                    <div className="mt-4 flex justify-between items-center pt-4 border-t border-gray-100">
                        <Button
                            variant="outline"
                            disabled={page === 1}
                            onClick={() => setPage(p => p - 1)}
                        >
                            Previous
                        </Button>
                        <span className="text-sm text-gray-500">Page {page} of {data.total_pages}</span>
                        <Button
                            variant="outline"
                            disabled={page >= data.total_pages}
                            onClick={() => setPage(p => p + 1)}
                        >
                            Next
                        </Button>
                    </div>
                )}
            </div>
        </div>
    );
}
