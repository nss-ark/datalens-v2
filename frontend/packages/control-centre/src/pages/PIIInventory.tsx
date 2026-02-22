import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { FileSearch } from 'lucide-react';
import { DataTable, type Column } from '@datalens/shared';
import { Pagination } from '@datalens/shared';
import { StatusBadge } from '@datalens/shared';
import { discoveryService } from '../services/discovery';
import type { PIIClassification } from '../types/discovery';
import { cn } from '@datalens/shared';

const PAGE_SIZE = 20;

const PIIInventory = () => {
    const [page, setPage] = useState(1);

    // Fetch data
    const { data: result, isLoading } = useQuery({
        queryKey: ['pii-inventory', page],
        queryFn: () => discoveryService.listClassifications({ page, page_size: PAGE_SIZE }),
    });

    const classifications = result?.items ?? [];
    const total = result?.total ?? 0;

    // Confidence Badge Helper
    const ConfidenceBadge = ({ value }: { value: number }) => {
        const pct = Math.round(value * 100);
        const tier = pct >= 90 ? 'high' : pct >= 70 ? 'medium' : 'low';

        const colors = {
            high: "text-green-700 bg-green-50 ring-green-600/20",
            medium: "text-yellow-700 bg-yellow-50 ring-yellow-600/20",
            low: "text-red-700 bg-red-50 ring-red-600/20"
        };

        return (
            <span className={cn("inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset", colors[tier])}>
                {pct}%
            </span>
        );
    };

    // Columns
    const columns: Column<PIIClassification>[] = [
        {
            key: 'field_name',
            header: 'Field Name',
            sortable: true,
            render: (row) => (
                <div>
                    <div className="font-semibold text-gray-900">{row.field_name}</div>
                    <div className="text-xs text-gray-500 font-mono">{row.entity_name}</div>
                </div>
            ),
        },
        {
            key: 'category',
            header: 'Category',
            sortable: true,
            width: '150px',
            render: (row) => (
                <div className="flex flex-col">
                    <span className="text-sm font-medium text-gray-900">{row.category}</span>
                    <span className="text-xs text-gray-500">{row.type}</span>
                </div>
            ),
        },
        {
            key: 'sensitivity',
            header: 'Sensitivity',
            sortable: true,
            width: '120px',
            render: (row) => (
                <span className={cn(
                    "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium",
                    row.sensitivity === 'CRITICAL' ? "bg-red-100 text-red-700" :
                        row.sensitivity === 'HIGH' ? "bg-orange-100 text-orange-700" :
                            row.sensitivity === 'MEDIUM' ? "bg-yellow-100 text-yellow-800" :
                                "bg-blue-100 text-blue-700"
                )}>
                    {row.sensitivity}
                </span>
            ),
        },
        {
            key: 'confidence',
            header: 'Confidence',
            sortable: true,
            width: '100px',
            render: (row) => <ConfidenceBadge value={row.confidence} />,
        },
        {
            key: 'status',
            header: 'Status',
            sortable: true,
            width: '120px',
            render: (row) => <StatusBadge label={row.status} size="sm" />,
        },
        {
            key: 'detection_method',
            header: 'Method',
            width: '100px',
            render: (row) => <span className="text-xs text-gray-500 uppercase tracking-wider">{row.detection_method}</span>,
        }
    ];

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-start">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 mb-1">PII Inventory</h1>
                    <p className="text-sm text-gray-500">
                        Comprehensive list of all discovered PII across your data sources
                    </p>
                </div>
                {/* Placeholder for future export actions */}
            </div>

            <div className="bg-white rounded-lg border border-gray-200 shadow-sm overflow-hidden">
                <div className="p-4 border-b border-gray-200 bg-gray-50 flex items-center justify-between">
                    <div className="flex items-center gap-2 text-sm text-gray-600">
                        <FileSearch size={16} />
                        <span className="font-medium">All Classifications</span>
                        <span className="bg-gray-200 text-gray-700 py-0.5 px-2 rounded-full text-xs">{total}</span>
                    </div>
                    {/* Add filters if needed later */}
                </div>

                <DataTable
                    columns={columns}
                    data={classifications}
                    isLoading={isLoading}
                    keyExtractor={(row) => row.id}
                    emptyTitle="No PII found"
                    emptyDescription="Connect data sources and run scans to populate your inventory."
                />
            </div>

            {total > PAGE_SIZE && (
                <Pagination page={page} pageSize={PAGE_SIZE} total={total} onPageChange={setPage} />
            )}
        </div>
    );
};

export default PIIInventory;
