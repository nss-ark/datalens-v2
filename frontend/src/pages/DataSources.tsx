import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { Plus, Play, Database as DatabaseIcon, RefreshCw, History, Loader2, Globe, Trash2, FileUp, X, UploadCloud } from 'lucide-react';
import { Button } from '../components/common/Button';
import { DataTable, type Column } from '../components/DataTable/DataTable';
import { StatusBadge } from '../components/common/StatusBadge';
import { Modal } from '../components/common/Modal';
import { ScanHistoryModal } from '../components/DataSources/ScanHistoryModal';
import { ErrorBoundary as SafeBoundary } from '../components/common/ErrorBoundary';
import { SectionErrorFallback } from '../components/common/ErrorFallbacks';
import { useDataSources, useCreateDataSource, useScanDataSource, useScanStatus } from '../hooks/useDataSources';
import { toast } from '../stores/toastStore';
import { dataSourceService } from '../services/datasource';
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
    { value: 'm365', label: 'Microsoft 365' },
    { value: 'google_workspace', label: 'Google Workspace' },
    { value: 'local_file', label: 'File Upload' },
];

const INITIAL_FORM = {
    name: '', type: 'postgresql' as DataSourceType, description: '',
    host: '', port: 5432, database: '', username: '', password: '', credentials: '',
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
    const { data: apiData, isLoading, refetch } = useDataSources();
    const dataSources = apiData || [];
    const { mutate: createMutate, isPending: isCreating } = useCreateDataSource();

    const [showModal, setShowModal] = useState(false);
    const [historyId, setHistoryId] = useState<string | null>(null);
    const [deleteId, setDeleteId] = useState<string | null>(null); // For delete confirmation
    const [isDeleting, setIsDeleting] = useState(false);
    const [form, setForm] = useState(INITIAL_FORM);
    const [isOAuthPending, setIsOAuthPending] = useState(false);

    // File Upload State
    const [uploadFile, setUploadFile] = useState<File | null>(null);
    const [uploadProgress, setUploadProgress] = useState(0);
    const [isUploading, setIsUploading] = useState(false);
    const [isDragOver, setIsDragOver] = useState(false);

    const handleCreate = async (e: FormEvent) => {
        e.preventDefault();

        // Handle File Upload
        if (form.type === 'local_file') {
            if (!uploadFile) {
                toast.error('No file selected', 'Please select a file to upload.');
                return;
            }

            setIsUploading(true);
            try {
                await dataSourceService.upload(uploadFile, (percent) => setUploadProgress(percent));
                toast.success('File Uploaded', 'Data source created from file.');
                setShowModal(false);
                setForm(INITIAL_FORM);
                setUploadFile(null);
                setUploadProgress(0);
                refetch();
            } catch (error) {
                toast.error('Upload Failed', 'Could not upload file.');
                console.error(error);
            } finally {
                setIsUploading(false);
            }
            return;
        }

        // Handle OAuth Flows
        // Exception: Google Service Account (if credentials provided)
        const isGoogleServiceAccount = form.type === 'google_workspace' && form.credentials && form.credentials.length > 2;

        if ((form.type === 'm365' || form.type === 'google_workspace') && !isGoogleServiceAccount) {
            const url = form.type === 'm365'
                ? dataSourceService.getM365AuthUrl()
                : dataSourceService.getGoogleAuthUrl();

            // Open Popup
            const width = 600;
            const height = 700;
            const left = window.screen.width / 2 - width / 2;
            const top = window.screen.height / 2 - height / 2;

            const popup = window.open(
                url,
                'Connect Data Source',
                `width=${width},height=${height},top=${top},left=${left}`
            );

            if (popup) {
                setIsOAuthPending(true);
                const timer = setInterval(() => {
                    if (popup.closed) {
                        clearInterval(timer);
                        setIsOAuthPending(false);
                        setShowModal(false);
                        refetch(); // Refresh list to see new DS
                        toast.success('Connection Attempted', 'Check the list for the new data source.');
                    }
                }, 1000);
            } else {
                toast.error('Popup Blocked', 'Please allow popups to connect this data source.');
            }
            return;
        }

        // Construct credentials string based on type
        let finalCredentials = form.credentials;
        if (['postgresql', 'mysql', 'mongodb', 'mssql', 'oracle', 'sqlite', 'azure_sql', 'rds'].includes(form.type)) {
            finalCredentials = `${form.username}:${form.password}`;
        }

        // Handle Standard Database/Storage Flows
        createMutate({
            name: form.name,
            type: form.type,
            description: form.description,
            host: form.host,
            port: form.port,
            database: form.database,
            credentials: finalCredentials,
        }, {
            onSuccess: () => {
                toast.success('Data source added', `"${form.name}" has been created.`);
                setShowModal(false);
                setForm(INITIAL_FORM);
            },
            onError: () => toast.error('Failed to create', 'Could not add the data source.'),
        });
    };

    const handleDelete = async () => {
        if (!deleteId) return;
        setIsDeleting(true);
        try {
            await dataSourceService.delete(deleteId);
            toast.success('Data Source Deleted', 'The data source has been removed.');
            setDeleteId(null);
            refetch();
        } catch (error) {
            toast.error('Delete Failed', 'Could not delete data source.');
            console.error(error);
        } finally {
            setIsDeleting(false);
        }
    };

    const handleDrop = (e: React.DragEvent) => {
        e.preventDefault();
        setIsDragOver(false);
        if (e.dataTransfer.files && e.dataTransfer.files[0]) {
            setUploadFile(e.dataTransfer.files[0]);
        }
    };

    const handleDragOver = (e: React.DragEvent) => {
        e.preventDefault();
        setIsDragOver(true);
    };

    const handleDragLeave = (e: React.DragEvent) => {
        e.preventDefault();
        setIsDragOver(false);
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
                    <Button
                        variant="ghost"
                        size="sm"
                        onClick={(e) => {
                            e.stopPropagation();
                            setDeleteId(row.id);
                        }}
                        icon={<Trash2 size={14} className="text-red-500" />}
                        title="Delete"
                    />
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
                        <label style={labelStyle}>Data Source Type</label>
                        <select value={form.type} onChange={updateField('type')} style={inputStyle}>
                            {DS_TYPE_OPTIONS.map((opt) => (
                                <option key={opt.value} value={opt.value}>{opt.label}</option>
                            ))}
                        </select>
                    </div>

                    {(form.type === 'm365' || form.type === 'google_workspace') ? (
                        <div className="space-y-4">
                            {form.type === 'm365' && (
                                <div className="bg-blue-50 p-4 rounded-lg border border-blue-100 text-sm text-blue-800">
                                    <h4 className="font-semibold mb-1">Configuration Required</h4>
                                    <p className="mb-2">Ensure your Azure AD Application is configured with this Redirect URI:</p>
                                    <code className="bg-white px-2 py-1 rounded border border-blue-200 block w-full mb-3 select-all">
                                        {window.location.origin}/api/v2/m365/callback
                                    </code>
                                    <p>Authentication uses the globally configured Client ID.</p>
                                </div>
                            )}

                            {form.type === 'google_workspace' && (
                                <div className="flex gap-4 p-1 bg-gray-100 rounded-lg select-none mb-4">
                                    <button
                                        type="button"
                                        onClick={() => setForm(f => ({ ...f, credentials: '' }))}
                                        className={`flex-1 py-1.5 text-sm font-medium rounded-md transition-colors ${!form.credentials ? 'bg-white shadow-sm text-gray-900' : 'text-gray-500 hover:text-gray-700'}`}
                                    >
                                        OAuth (User)
                                    </button>
                                    <button
                                        type="button"
                                        onClick={() => setForm(f => ({ ...f, credentials: '{}' }))}
                                        className={`flex-1 py-1.5 text-sm font-medium rounded-md transition-colors ${form.credentials ? 'bg-white shadow-sm text-gray-900' : 'text-gray-500 hover:text-gray-700'}`}
                                    >
                                        Service Account
                                    </button>
                                </div>
                            )}

                            {(form.type === 'google_workspace' && form.credentials) ? (
                                <div>
                                    <label style={labelStyle}>Service Account Key (JSON)</label>
                                    <textarea
                                        value={form.credentials === '{}' ? '' : form.credentials}
                                        onChange={updateField('credentials')}
                                        style={{ ...inputStyle, height: '120px', padding: '0.75rem', resize: 'vertical', fontFamily: 'monospace', fontSize: '12px' }}
                                        placeholder='{ "type": "service_account", ... }'
                                        required
                                    />
                                    <div className="mt-4 flex justify-end">
                                        <Button type="submit" isLoading={isCreating}>Connect Service Account</Button>
                                    </div>
                                </div>
                            ) : (
                                <div className="bg-gray-50 p-6 rounded-lg border border-gray-200 text-center">
                                    <div className="mb-4">
                                        <div className="mx-auto w-12 h-12 bg-white rounded-full flex items-center justify-center shadow-sm mb-3">
                                            <Globe className="text-blue-600" size={24} />
                                        </div>
                                        <h3 className="text-gray-900 font-medium">Connect via OAuth</h3>
                                        <p className="text-sm text-gray-500 mt-1 max-w-xs mx-auto">
                                            You will be redirected to {form.type === 'm365' ? 'Microsoft' : 'Google'} to authenticate and grant access.
                                        </p>
                                    </div>
                                    <Button
                                        type="submit"
                                        isLoading={isOAuthPending}
                                        className="w-full justify-center"
                                    >
                                        {isOAuthPending ? 'Connecting...' : `Connect ${form.type === 'm365' ? 'Microsoft 365' : 'Google Workspace'}`}
                                    </Button>
                                </div>
                            )}
                        </div>
                    ) : (
                        <>
                            <div>
                                <label style={labelStyle}>Name</label>
                                <input type="text" value={form.name} onChange={updateField('name')} required style={inputStyle} placeholder="HR Database" />
                            </div>

                            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                                <div>
                                    <label style={labelStyle}>Host</label>
                                    <input type="text" value={form.host} onChange={updateField('host')} required style={inputStyle} placeholder="db.example.com" />
                                </div>
                                <div>
                                    <label style={labelStyle}>Port</label>
                                    <input type="number" value={form.port} onChange={updateField('port')} style={inputStyle} />
                                </div>
                            </div>

                            <div>
                                <label style={labelStyle}>Database / Bucket</label>
                                <input type="text" value={form.database} onChange={updateField('database')} required style={inputStyle} placeholder="production_db" />
                            </div>

                            <div>
                                <label style={labelStyle}>Description</label>
                                <textarea value={form.description} onChange={updateField('description')} style={{ ...inputStyle, height: '60px', padding: '0.5rem 0.875rem', resize: 'vertical' }} placeholder="Brief description..." />
                            </div>

                            {['postgresql', 'mysql', 'mongodb', 'mssql', 'oracle', 'sqlite', 'azure_sql', 'rds'].includes(form.type) ? (
                                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                                    <div>
                                        <label style={labelStyle}>Username</label>
                                        <input type="text" value={form.username} onChange={updateField('username')} required style={inputStyle} placeholder="admin" />
                                    </div>
                                    <div>
                                        <label style={labelStyle}>Password</label>
                                        <input type="password" value={form.password} onChange={updateField('password')} required style={inputStyle} placeholder="••••••••" />
                                    </div>
                                </div>
                            ) : (
                                <div>
                                    <label style={labelStyle}>Credentials</label>
                                    <input type="password" value={form.credentials} onChange={updateField('credentials')} required style={inputStyle} placeholder="Connection string or password" />
                                </div>
                            )}
                        </>
                    )}

                    {form.type === 'local_file' && (
                        <div>
                            <label style={labelStyle}>Upload File</label>
                            <div
                                onDrop={handleDrop}
                                onDragOver={handleDragOver}
                                onDragLeave={handleDragLeave}
                                className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${isDragOver ? 'border-primary-500 bg-primary-50' : 'border-gray-300 hover:border-primary-400'
                                    }`}
                            >
                                {uploadFile ? (
                                    <div className="flex items-center justify-center gap-3">
                                        <FileUp className="text-primary-600" size={32} />
                                        <div className="text-left">
                                            <div className="font-medium text-gray-900">{uploadFile.name}</div>
                                            <div className="text-sm text-gray-500">{(uploadFile.size / 1024).toFixed(1)} KB</div>
                                        </div>
                                        <button
                                            type="button"
                                            onClick={() => setUploadFile(null)}
                                            className="ml-2 p-1 text-gray-400 hover:text-red-500"
                                        >
                                            <X size={16} />
                                        </button>
                                    </div>
                                ) : (
                                    <div className="space-y-2">
                                        <div className="mx-auto w-10 h-10 bg-gray-100 rounded-full flex items-center justify-center">
                                            <UploadCloud className="text-gray-400" size={24} />
                                        </div>
                                        <div className="text-sm text-gray-600">
                                            <span className="text-primary-600 font-medium cursor-pointer">Click to upload</span> or drag and drop
                                        </div>
                                        <p className="text-xs text-gray-500">PDF, DOCX, XLSX, CSV up to 10MB</p>
                                        <input
                                            type="file"
                                            className="hidden"
                                            id="file-upload"
                                            onChange={(e) => {
                                                if (e.target.files?.[0]) setUploadFile(e.target.files[0]);
                                            }}
                                            accept=".pdf,.docx,.xlsx,.csv"
                                        />
                                        <label htmlFor="file-upload" className="absolute inset-0 cursor-pointer opacity-0" />
                                    </div>
                                )}
                            </div>
                            {isUploading && (
                                <div className="mt-4">
                                    <div className="flex justify-between text-xs mb-1">
                                        <span>Uploading...</span>
                                        <span>{uploadProgress}%</span>
                                    </div>
                                    <div className="w-full bg-gray-200 rounded-full h-1.5">
                                        <div
                                            className="bg-primary-600 h-1.5 rounded-full transition-all duration-300"
                                            style={{ width: `${uploadProgress}%` }}
                                        />
                                    </div>
                                </div>
                            )}
                        </div>
                    )}
                </form>
            </Modal>

            {/* Delete Confirmation Modal */}
            <Modal
                open={!!deleteId}
                onClose={() => setDeleteId(null)}
                title="Delete Data Source"
                footer={
                    <>
                        <Button variant="outline" onClick={() => setDeleteId(null)}>Cancel</Button>
                        <Button
                            className="bg-red-600 hover:bg-red-700 text-white"
                            onClick={handleDelete}
                            isLoading={isDeleting}
                        >
                            Delete
                        </Button>
                    </>
                }
            >
                <div className="p-4">
                    <p className="text-gray-700">
                        Are you sure you want to delete this data source? This action cannot be undone.
                        All scanned data and history associated with this source will be permanently removed.
                    </p>
                </div>
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
