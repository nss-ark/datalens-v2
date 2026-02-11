import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { Plus, Play, Database as DatabaseIcon, RefreshCw, History, Loader2 } from 'lucide-react';
import { Button } from '../components/common/Button';
import { DataTable, type Column } from '../components/DataTable/DataTable';
import { StatusBadge } from '../components/common/StatusBadge';
import { Modal } from '../components/common/Modal';
import { ScanHistoryModal } from '../components/DataSources/ScanHistoryModal';
import { ErrorBoundary as SafeBoundary } from '../components/common/ErrorBoundary';
import { SectionErrorFallback } from '../components/common/ErrorFallbacks';
import { useDataSources, useCreateDataSource, useScanDataSource, useScanStatus } from '../hooks/useDataSources';
import { toast } from '../stores/toastStore';
import type { DataSource, DataSourceType } from '../types/datasource';

const DS_TYPE_OPTIONS: { value: DataSourceType; label: string }[] = [
    { value: 'postgresql', label: 'PostgreSQL' },
    { value: 'mysql', label: 'MySQL' },
    { value: 'mssql', label: 'SQL Server' },
    { value: 'mongodb', label: 'MongoDB' },
    { value: 'oracle', label: 'Oracle' },
    { value: 'sqlite', label: 'SQLite' },
    { value: 's3', label: 'Amazon S3' },
    { value: 'gcs', label: 'Google Cloud Storage' },
    { value: 'azure_blob', label: 'Azure Blob' },
];

const INITIAL_FORM = {
    name: '', type: 'postgresql' as DataSourceType, description: '',
    host: '', port: 5432, database: '', credentials: '',
};

// Component for the Scan Action button with polling
const ScanAction = ({ dataSource }: { dataSource: DataSource }) => {
    const { mutate: scanMutate, isPending: isStarting } = useScanDataSource();

    // Scan status polling is enabled for this data source
    const { data: scanStatus } = useScanStatus(dataSource.id, true);

    // Derived state
    const isScanning = scanStatus?.status === 'RUNNING' || scanStatus?.status === 'QUEUED';

    const handleScan = (e: React.MouseEvent) => {
        e.stopPropagation();
        scanMutate(dataSource.id, {
            onSuccess: () => toast.success('Scan started', 'The data source scan has been initiated.'),
            onError: () => toast.error('Scan failed', 'Could not initiate scan. Please try again.'),
        });
    };

    if (isStarting || isScanning) {
        return (
            <div className="flex items-center gap-2 text-sm text-indigo-600 font-medium bg-indigo-50 px-3 py-1.5 rounded-md">
                <Loader2 size={14} className="animate-spin" />
                {scanStatus?.progress_percentage ? `${scanStatus.progress_percentage}% ` : 'Scanning...'}
            </div>
        );
    }

    return (
        <Button
            variant="outline"
            size="sm"
            onClick={handleScan}
            icon={<Play size={14} />}
        >
            Scan
        </Button>
    );
};

const DataSources = () => {
    const navigate = useNavigate();
    const { data: dataSources = [], isLoading, refetch } = useDataSources();
    const { mutate: createMutate, isPending: isCreating } = useCreateDataSource();

    const [showModal, setShowModal] = useState(false);
    const [historyId, setHistoryId] = useState<string | null>(null);
    const [form, setForm] = useState(INITIAL_FORM);

    const handleCreate = (e: FormEvent) => {
        e.preventDefault();
        createMutate({
            name: form.name,
            type: form.type,
            description: form.description,
            host: form.host,
            port: form.port,
            database: form.database,
            credentials: form.credentials,
        }, {
            onSuccess: () => {
                toast.success('Data source added', `"${form.name}" has been created.`);
                setShowModal(false);
                setForm(INITIAL_FORM);
            },
            onError: () => toast.error('Failed to create', 'Could not add the data source.'),
        });
    };

    const updateField = (field: string) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        setForm((prev) => ({ ...prev, [field]: field === 'port' ? Number(e.target.value) : e.target.value }));
    };

    const columns: Column<DataSource>[] = [
        {
            key: 'name',
            header: 'Name',
            sortable: true,
            render: (row) => (
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                    <div style={{
                        width: '36px', height: '36px', borderRadius: 'var(--radius-md)',
                        backgroundColor: 'var(--primary-50)', color: 'var(--primary-600)',
                        display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
                    }}>
                        <DatabaseIcon size={18} />
                    </div>
                    <div>
                        <div style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{row.name}</div>
                        {row.description && (
                            <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', marginTop: '2px' }}>
                                {row.description}
                            </div>
                        )}
                    </div>
                </div>
            ),
        },
        {
            key: 'type',
            header: 'Type',
            sortable: true,
            width: '120px',
            render: (row) => (
                <span style={{
                    textTransform: 'uppercase', fontSize: '0.75rem', fontWeight: 600,
                    color: 'var(--text-secondary)', letterSpacing: '0.04em',
                }}>
                    {row.type}
                </span>
            ),
        },
        {
            key: 'status',
            header: 'Status',
            sortable: true,
            width: '140px',
            render: (row) => <StatusBadge label={row.status} />,
        },
        {
            key: 'last_sync_at',
            header: 'Last Scanned',
            sortable: true,
            width: '160px',
            render: (row) => (
                <span style={{ fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
                    {row.last_sync_at
                        ? new Date(row.last_sync_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
                        : 'Never'}
                </span>
            ),
        },
        {
            key: 'actions',
            header: '',
            width: '180px',
            render: (row) => (
                <div className="flex items-center gap-2 justify-end">
                    <Button
                        variant="ghost"
                        size="sm"
                        onClick={(e) => {
                            e.stopPropagation();
                            setHistoryId(row.id);
                        }}
                        icon={<History size={14} />}
                        title="View History"
                    />
                    <ScanAction dataSource={row} />
                </div>
            ),
        },
    ];

    return (
        <div>
            {/* Page Header */}
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
                <div>
                    <h1 style={{ fontSize: '1.5rem', fontWeight: 700, color: 'var(--text-primary)', marginBottom: '0.25rem' }}>
                        Data Sources
                    </h1>
                    <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
                        Manage your connected databases and storage systems
                    </p>
                </div>
                <div style={{ display: 'flex', gap: '0.75rem' }}>
                    <Button variant="outline" onClick={() => refetch()} icon={<RefreshCw size={16} />}>
                        Refresh
                    </Button>
                    <Button icon={<Plus size={16} />} onClick={() => setShowModal(true)}>
                        Add Data Source
                    </Button>
                </div>
            </div>

            {/* Table */}
            <SafeBoundary FallbackComponent={SectionErrorFallback}>
                <DataTable
                    columns={columns}
                    data={dataSources}
                    isLoading={isLoading}
                    keyExtractor={(row) => row.id}
                    onRowClick={(row) => navigate(`/datasources/${row.id}`)}
                    emptyTitle="No data sources yet"
                    emptyDescription="Connect your first database or storage system to start discovering PII."
                />
            </SafeBoundary>

            {/* Add Data Source Modal */}
            <Modal
                open={showModal}
                onClose={() => setShowModal(false)}
                title="Add Data Source"
                footer={
                    <>
                        <Button variant="outline" onClick={() => setShowModal(false)}>Cancel</Button>
                        <Button type="submit" form="addDsForm" isLoading={isCreating}>Add Source</Button>
                    </>
                }
            >
                <form id="addDsForm" onSubmit={handleCreate} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                    <div>
                        <label style={labelStyle}>Name</label>
                        <input type="text" value={form.name} onChange={updateField('name')} required style={inputStyle} placeholder="HR Database" />
                    </div>

                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                        <div>
                            <label style={labelStyle}>Type</label>
                            <select value={form.type} onChange={updateField('type')} style={inputStyle}>
                                {DS_TYPE_OPTIONS.map((opt) => (
                                    <option key={opt.value} value={opt.value}>{opt.label}</option>
                                ))}
                            </select>
                        </div>
                        <div>
                            <label style={labelStyle}>Port</label>
                            <input type="number" value={form.port} onChange={updateField('port')} style={inputStyle} />
                        </div>
                    </div>

                    <div>
                        <label style={labelStyle}>Host</label>
                        <input type="text" value={form.host} onChange={updateField('host')} required style={inputStyle} placeholder="db.example.com" />
                    </div>

                    <div>
                        <label style={labelStyle}>Database</label>
                        <input type="text" value={form.database} onChange={updateField('database')} required style={inputStyle} placeholder="production_db" />
                    </div>

                    <div>
                        <label style={labelStyle}>Description</label>
                        <textarea value={form.description} onChange={updateField('description')} style={{ ...inputStyle, height: '60px', padding: '0.5rem 0.875rem', resize: 'vertical' }} placeholder="Brief description..." />
                    </div>

                    <div>
                        <label style={labelStyle}>Credentials (connection string or password)</label>
                        <input type="password" value={form.credentials} onChange={updateField('credentials')} required style={inputStyle} placeholder="••••••••" />
                    </div>
                </form>
            </Modal>

            {/* Scan History Modal */}
            <ScanHistoryModal
                dataSourceId={historyId}
                onClose={() => setHistoryId(null)}
            />
        </div>
    );
};

const labelStyle: React.CSSProperties = {
    display: 'block', fontSize: '0.875rem', fontWeight: 500,
    color: 'var(--text-primary)', marginBottom: '0.375rem',
};

const inputStyle: React.CSSProperties = {
    width: '100%', height: '40px', padding: '0 0.875rem',
    borderRadius: 'var(--radius-md)', border: '1px solid var(--border-color)',
    fontSize: '0.875rem', outline: 'none', boxSizing: 'border-box',
};

export default DataSources;
