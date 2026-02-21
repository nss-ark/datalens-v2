import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { format } from 'date-fns';
import { Search, ShieldAlert, ListIcon, LayoutList } from 'lucide-react';
import { DataTable, StatusBadge, Button, Input } from '@datalens/shared';
import { dataSubjectService } from '../services/dataSubjectService';
import type { DataPrincipalProfile } from '../services/dataSubjectService';

// Add a simple hook for debouncing values
function useDebounce<T>(value: T, delay: number): T {
    const [debouncedValue, setDebouncedValue] = useState<T>(value);

    useEffect(() => {
        const handler = setTimeout(() => {
            setDebouncedValue(value);
        }, delay);

        return () => {
            clearTimeout(handler);
        };
    }, [value, delay]);

    return debouncedValue;
}

export default function DataSubjects() {
    const navigate = useNavigate();
    const [page, setPage] = useState(1);
    const [searchTerm, setSearchTerm] = useState('');
    const debouncedSearchTerm = useDebounce(searchTerm, 300);

    const { data, isLoading, isError } = useQuery({
        queryKey: ['subjects', page, debouncedSearchTerm],
        queryFn: () => dataSubjectService.listSubjects({
            page,
            page_size: 10,
            search: debouncedSearchTerm || undefined
        })
    });

    const columns = [
        {
            key: 'email',
            header: 'Email / Phone',
            render: (row: DataPrincipalProfile) => (
                <div className="flex flex-col">
                    <span className="font-medium text-gray-900">{row.email || 'N/A'}</span>
                    {row.phone && <span className="text-xs text-gray-500">{row.phone}</span>}
                </div>
            )
        },
        {
            key: 'status',
            header: 'Verification Status',
            render: (row: DataPrincipalProfile) => <StatusBadge label={row.verification_status} />
        },
        {
            key: 'method',
            header: 'Method',
            render: (row: DataPrincipalProfile) => (
                <span className="text-sm text-gray-600">
                    {row.verification_method || '-'}
                </span>
            )
        },
        {
            key: 'minor',
            header: 'Status',
            render: (row: DataPrincipalProfile) => (
                row.is_minor ? (
                    <span className="inline-flex items-center gap-1.5 px-2 py-1 rounded-md text-xs font-medium bg-purple-100 text-purple-700">
                        <ShieldAlert size={12} />
                        Minor
                    </span>
                ) : (
                    <span className="text-sm text-gray-500">Adult</span>
                )
            )
        },
        {
            key: 'last_access',
            header: 'Last Access',
            render: (row: DataPrincipalProfile) => (
                <span className="text-sm text-gray-500">
                    {row.last_access_at ? format(new Date(row.last_access_at), 'MMM d, yyyy HH:mm') : 'Never'}
                </span>
            )
        },
        {
            key: 'actions',
            header: 'Actions',
            render: (row: DataPrincipalProfile) => (
                <div className="flex gap-2">
                    <Button
                        size="sm"
                        variant="secondary"
                        icon={<ListIcon size={14} />}
                        onClick={() => navigate(`/dsr?subject_id=${row.subject_id}`)}
                    >
                        DSRs
                    </Button>
                    <Button
                        size="sm"
                        variant="secondary"
                        icon={<LayoutList size={14} />}
                        onClick={() => navigate(`/consent?subject_id=${row.subject_id}`)}
                    >
                        Consent
                    </Button>
                </div>
            )
        }
    ];

    return (
        <div className="p-6 max-w-7xl mx-auto space-y-6">
            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900">Data Subjects</h1>
                    <p className="text-gray-500 mt-1">
                        View and manage verified data principals, their requests, and consent records.
                    </p>
                </div>
            </div>

            <div className="bg-white p-4 rounded-lg border shadow-sm flex flex-col sm:flex-row gap-4 justify-between items-center">
                <div className="relative w-full max-w-md">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={18} />
                    <Input
                        placeholder="Search by email or phone..."
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                        className="pl-10 w-full"
                    />
                </div>
            </div>

            {isError ? (
                <div className="p-8 text-center bg-red-50 border border-red-100 rounded-lg text-red-600">
                    Failed to load data subjects. Please try again.
                </div>
            ) : (
                <div className="bg-white border rounded-lg shadow-sm overflow-hidden">
                    <DataTable
                        columns={columns}
                        data={data?.items || []}
                        isLoading={isLoading}
                        keyExtractor={(row) => row.id}
                        emptyTitle="No Data Subjects Found"
                        emptyDescription={
                            searchTerm
                                ? "No matching subjects found for your search."
                                : "No data subjects have been verified for this tenant yet."
                        }
                    />

                    {data && data.total_pages > 1 && (
                        <div className="p-4 border-t flex items-center justify-between">
                            <span className="text-sm text-gray-500">
                                Showing {(page - 1) * 10 + 1} to {Math.min(page * 10, data.total)} of {data.total} subjects
                            </span>
                            <div className="flex space-x-2">
                                <Button
                                    variant="outline"
                                    size="sm"
                                    disabled={page === 1}
                                    onClick={() => setPage(p => Math.max(1, p - 1))}
                                >
                                    Previous
                                </Button>
                                <Button
                                    variant="outline"
                                    size="sm"
                                    disabled={page === data.total_pages}
                                    onClick={() => setPage(p => Math.min(data.total_pages, p + 1))}
                                >
                                    Next
                                </Button>
                            </div>
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}
