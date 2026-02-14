import { Modal } from '@datalens/shared';
import { DataTable } from '@datalens/shared';
import { StatusBadge } from '@datalens/shared';
import { useScanHistory } from '../../hooks/useDataSources';
import type { ScanHistoryItem } from '../../types/datasource';
import type { Column } from '@datalens/shared';

interface ScanHistoryModalProps {
    dataSourceId: string | null;
    onClose: () => void;
}

export const ScanHistoryModal = ({ dataSourceId, onClose }: ScanHistoryModalProps) => {
    const { data: history = [], isLoading } = useScanHistory(dataSourceId || '');

    const columns: Column<ScanHistoryItem>[] = [
        {
            key: 'started_at',
            header: 'Started At',
            render: (row) => new Date(row.started_at).toLocaleString(),
            width: '180px',
        },
        {
            key: 'status',
            header: 'Status',
            render: (row) => <StatusBadge label={row.status} />,
            width: '120px',
        },
        {
            key: 'tables_scanned',
            header: 'Tables',
            render: (row) => <span>{row.tables_scanned}</span>,
            width: '100px',
        },
        {
            key: 'pii_found',
            header: 'PII Found',
            render: (row) => (
                <span className={row.pii_found > 0 ? 'text-amber-600 font-medium' : 'text-gray-500'}>
                    {row.pii_found}
                </span>
            ),
            width: '100px',
        },
        {
            key: 'duration',
            header: 'Duration',
            render: (row) => {
                if (!row.completed_at) return '-';
                const start = new Date(row.started_at).getTime();
                const end = new Date(row.completed_at).getTime();
                const diff = Math.round((end - start) / 1000);
                return `${diff}s`;
            },
            width: '100px',
        },
    ];

    return (
        <Modal
            open={!!dataSourceId}
            onClose={onClose}
            title="Scan History"
        >
            <div className="min-h-[300px]">
                <DataTable
                    columns={columns}
                    data={history}
                    isLoading={isLoading}
                    keyExtractor={(row) => row.id}
                    emptyTitle="No scan history"
                    emptyDescription="This data source has not been scanned yet."
                    loadingRows={5}
                />
            </div>
        </Modal>
    );
};
