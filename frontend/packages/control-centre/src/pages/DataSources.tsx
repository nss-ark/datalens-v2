import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { Plus, Play, Database as DatabaseIcon, RefreshCw, History, Loader2, Globe, Trash2, FileUp, X, UploadCloud, Search, MoreHorizontal, CheckCircle2, Activity, Link as LinkIcon, Server, HardDrive, FileText, Cloud } from 'lucide-react';
import { Button } from '@datalens/shared';
import { Modal } from '@datalens/shared';
import { ScanHistoryModal } from '../components/DataSources/ScanHistoryModal';
import { useDataSources, useCreateDataSource, useScanDataSource, useScanStatus } from '../hooks/useDataSources';
import { toast } from '@datalens/shared';
import { dataSourceService } from '../services/datasource';
import type { DataSource, DataSourceType } from '../types/datasource';

const DS_TYPE_OPTIONS: { value: DataSourceType; label: string }[] = [
    { value: 'POSTGRESQL', label: 'PostgreSQL' },
    { value: 'MYSQL', label: 'MySQL' },
    { value: 'SQLSERVER', label: 'SQL Server' },
    { value: 'MONGODB', label: 'MongoDB' },
    { value: 'S3', label: 'Amazon S3' },
    { value: 'AZURE_BLOB', label: 'Azure Blob' },
    { value: 'MICROSOFT_365', label: 'Microsoft 365' },
    { value: 'GOOGLE_WORKSPACE', label: 'Google Workspace' },
    { value: 'FILE_UPLOAD', label: 'File Upload' },
];

const getDsIcon = (type: string) => {
    switch (type) {
        case 'POSTGRESQL':
        case 'MYSQL':
        case 'SQLSERVER': return <DatabaseIcon className="text-blue-600" size={24} />;
        case 'MONGODB': return <Server className="text-green-600" size={24} />;
        case 'S3':
        case 'AZURE_BLOB': return <Cloud className="text-blue-600" size={24} />;
        case 'MICROSOFT_365':
        case 'GOOGLE_WORKSPACE': return <Globe className="text-indigo-600" size={24} />;
        case 'FILE_UPLOAD': return <FileText className="text-orange-600" size={24} />;
        default: return <HardDrive className="text-gray-600" size={24} />;
    }
};

const getDsColor = (type: string) => {
    switch (type) {
        case 'POSTGRESQL': return 'bg-indigo-500';
        case 'MYSQL': return 'bg-blue-500';
        case 'MONGODB': return 'bg-green-500';
        default: return 'bg-gray-500';
    }
};

const DataSourceCard = ({ dataSource, onHistory, onDelete }: { dataSource: DataSource, onHistory: () => void, onDelete: () => void }) => {
    const { mutate: scanMutate, isPending: isStarting } = useScanDataSource();
    const { data: scanStatus } = useScanStatus(dataSource.id, true);

    const isScanning = scanStatus?.status === 'RUNNING' || scanStatus?.status === 'QUEUED';

    const handleScan = (e: React.MouseEvent) => {
        e.stopPropagation();
        scanMutate(dataSource.id, {
            onSuccess: () => toast.success('Scan started', 'The data source scan has been initiated.'),
            onError: () => toast.error('Scan failed', 'Could not initiate scan. Please try again.'),
        });
    };

    const typeLabel = DS_TYPE_OPTIONS.find(o => o.value === dataSource.type)?.label || dataSource.type;

    return (
        <div
            className="group relative overflow-hidden flex flex-col h-full transition-all duration-300"
            style={{
                backgroundColor: '#ffffff',
                borderRadius: '1rem',
                border: isScanning ? '1px solid rgba(59,130,246,0.3)' : '1px solid #e5e7eb',
                padding: '1.5rem',
                boxShadow: '0 4px 6px -1px rgba(0,0,0,0.02), 0 2px 4px -1px rgba(0,0,0,0.02)',
            }}
        >
            {/* More menu */}
            <div className="absolute top-0 right-0 p-4 opacity-0 group-hover:opacity-100 transition-opacity">
                <span className="text-gray-400 cursor-pointer"><MoreHorizontal size={20} /></span>
            </div>

            {/* Header: icon + name */}
            <div style={{ display: 'flex', alignItems: 'flex-start', gap: '1rem', marginBottom: isScanning ? '1rem' : '1.5rem' }}>
                <div style={{
                    width: '3rem', height: '3rem', borderRadius: '0.75rem',
                    backgroundColor: isScanning ? 'rgba(219,234,254,0.5)' : '#eff6ff',
                    display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
                }}>
                    {getDsIcon(dataSource.type)}
                </div>
                <div>
                    <h3 style={{ fontSize: '1.125rem', fontWeight: 600, color: '#111827', lineHeight: 1.3 }}>{dataSource.name}</h3>
                    <p style={{ fontSize: '0.75rem', color: '#6b7280', marginTop: '0.25rem' }}>{dataSource.description || 'No description'}</p>
                </div>
            </div>

            {/* Scan progress bar (when scanning) */}
            {isScanning && (
                <div style={{ marginBottom: '1.25rem' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end', marginBottom: '0.5rem' }}>
                        <span style={{ fontSize: '0.75rem', fontWeight: 600, color: '#2563eb' }} className="animate-pulse">Scanning Data...</span>
                        <span style={{ fontSize: '0.75rem', fontFamily: 'monospace', color: '#6b7280' }}>{scanStatus?.progress_percentage || 0}%</span>
                    </div>
                    <div style={{ width: '100%', height: '6px', backgroundColor: '#f3f4f6', borderRadius: '9999px', overflow: 'hidden' }}>
                        <div style={{ height: '100%', width: `${scanStatus?.progress_percentage || 5}%`, backgroundColor: '#3b82f6', borderRadius: '9999px', transition: 'width 0.5s' }}></div>
                    </div>
                </div>
            )}

            {/* Data rows */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', flex: 1 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '0.875rem' }}>
                    <span style={{ color: '#6b7280' }}>Type</span>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.375rem', fontWeight: 500, color: '#374151' }}>
                        <span className={`w-2 h-2 rounded-full ${getDsColor(dataSource.type)}`}></span>
                        {typeLabel}
                    </div>
                </div>
                {!isScanning && (
                    <>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '0.875rem' }}>
                            <span style={{ color: '#6b7280' }}>Status</span>
                            <span style={{
                                display: 'inline-flex', alignItems: 'center',
                                padding: '0.125rem 0.625rem', borderRadius: '9999px',
                                fontSize: '0.75rem', fontWeight: 500,
                                backgroundColor: dataSource.status === 'CONNECTED' ? '#dcfce7' : dataSource.status === 'ERROR' ? '#fee2e2' : '#fef9c3',
                                color: dataSource.status === 'CONNECTED' ? '#166534' : dataSource.status === 'ERROR' ? '#991b1b' : '#854d0e',
                            }}>
                                {dataSource.status === 'CONNECTED' ? 'Connected' : dataSource.status}
                            </span>
                        </div>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '0.875rem' }}>
                            <span style={{ color: '#6b7280' }}>Last Scanned</span>
                            <span style={{ color: '#374151', fontFamily: 'monospace', fontSize: '0.75rem' }}>
                                {dataSource.last_sync_at ? new Date(dataSource.last_sync_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) : 'Never'}
                            </span>
                        </div>
                    </>
                )}
            </div>

            {/* Footer buttons */}
            <div style={{ marginTop: '1.5rem', paddingTop: '1.5rem', borderTop: '1px solid #f3f4f6', display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                {isScanning ? (
                    <div
                        onClick={(e) => { e.stopPropagation(); }}
                        style={{ flex: 1, padding: '0.5rem 0.75rem', borderRadius: '0.5rem', color: 'rgba(239,68,68,0.8)', fontSize: '0.875rem', fontWeight: 500, cursor: 'pointer', textAlign: 'center' }}
                    >
                        Cancel Scan
                    </div>
                ) : (
                    <div
                        onClick={(e) => { e.stopPropagation(); handleScan(e); }}
                        style={{
                            flex: 1, padding: '0.5rem 0.75rem', borderRadius: '0.5rem',
                            backgroundColor: '#f3f4f6', color: '#374151',
                            fontSize: '0.875rem', fontWeight: 500, cursor: 'pointer',
                            display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem',
                        }}
                    >
                        {isStarting ? <Loader2 size={16} className="animate-spin" /> : <Play size={16} />}
                        Scan
                    </div>
                )}

                <div
                    onClick={(e) => { e.stopPropagation(); onHistory(); }}
                    style={{
                        flex: 1, padding: '0.5rem 0.75rem', borderRadius: '0.5rem',
                        border: '1px solid #e5e7eb', color: '#4b5563',
                        fontSize: '0.875rem', fontWeight: 500, cursor: 'pointer',
                        display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem',
                    }}
                >
                    <History size={16} />
                    History
                </div>
                <div
                    onClick={(e) => { e.stopPropagation(); onDelete(); }}
                    style={{
                        padding: '0.5rem', borderRadius: '0.5rem', color: '#ef4444', cursor: 'pointer',
                    }}
                >
                    <Trash2 size={18} />
                </div>
            </div>
        </div>
    );
};

const INITIAL_FORM = {
    name: '', type: 'POSTGRESQL' as DataSourceType, description: '',
    host: '', port: 5432, database: '', username: '', password: '', credentials: '',
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

    const [searchQuery, setSearchQuery] = useState('');
    const [typeFilter, setTypeFilter] = useState('All Types');

    const filteredDataSources = dataSources.filter(ds => {
        const matchesSearch = ds.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
            (ds.description && ds.description.toLowerCase().includes(searchQuery.toLowerCase()));

        let matchesType = true;
        if (typeFilter !== 'All Types') {
            const typeObj = DS_TYPE_OPTIONS.find(o => o.label === typeFilter);
            if (typeObj) matchesType = ds.type === typeObj.value;
        }

        return matchesSearch && matchesType;
    });

    const activeConnections = dataSources.filter(ds => ds.status === 'CONNECTED').length;

    // File Upload State
    const [uploadFile, setUploadFile] = useState<File | null>(null);
    const [uploadProgress, setUploadProgress] = useState(0);
    const [isUploading, setIsUploading] = useState(false);
    const [isDragOver, setIsDragOver] = useState(false);

    const handleCreate = async (e: FormEvent) => {
        e.preventDefault();

        // Handle File Upload
        if (form.type === 'FILE_UPLOAD') {
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
        const isGoogleServiceAccount = form.type === 'GOOGLE_WORKSPACE' && form.credentials && form.credentials.length > 2;

        if ((form.type === 'MICROSOFT_365' || form.type === 'GOOGLE_WORKSPACE') && !isGoogleServiceAccount) {
            const url = form.type === 'MICROSOFT_365'
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
        if (['POSTGRESQL', 'MYSQL', 'MONGODB', 'SQLSERVER', 'AZURE_SQL', 'RDS'].includes(form.type)) {
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

    return (
        <div style={{ width: '100%', paddingBottom: '2rem' }}>
            {/* Page Header */}
            <header style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '2.5rem', flexWrap: 'wrap', gap: '1.5rem' }}>
                <div>
                    <h1 style={{ fontSize: '1.875rem', fontWeight: 600, letterSpacing: '-0.025em', color: '#111827', marginBottom: '0.5rem' }}>Data Sources</h1>
                    <p style={{ color: '#6b7280', fontSize: '0.875rem', fontWeight: 500 }}>Manage your connected databases and storage systems securely.</p>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                    <div
                        onClick={() => refetch()}
                        style={{
                            display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem',
                            padding: '0.625rem 1rem', backgroundColor: '#ffffff', border: '1px solid #e5e7eb',
                            borderRadius: '0.75rem', boxShadow: '0 1px 2px rgba(0,0,0,0.05)',
                            color: '#374151', fontSize: '0.875rem', fontWeight: 500, cursor: 'pointer',
                        }}
                    >
                        <RefreshCw size={18} style={{ color: '#9ca3af' }} />
                        Refresh
                    </div>
                    <div
                        onClick={() => setShowModal(true)}
                        style={{
                            display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem',
                            padding: '0.625rem 1.25rem', backgroundColor: '#2563eb', color: '#ffffff',
                            borderRadius: '0.75rem', boxShadow: '0 4px 6px rgba(37,99,235,0.25)',
                            fontSize: '0.875rem', fontWeight: 500, cursor: 'pointer',
                        }}
                    >
                        <Plus size={18} />
                        Add Data Source
                    </div>
                </div>
            </header>

            {/* Status Cards */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '1.5rem', marginBottom: '2rem' }}>
                <div style={{ backgroundColor: '#fff', borderRadius: '1rem', padding: '1.5rem', border: '1px solid #f3f4f6', boxShadow: '0 4px 6px -1px rgba(0,0,0,0.02)' }}>
                    <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: '1rem' }}>
                        <div style={{ padding: '0.5rem', backgroundColor: '#f0fdf4', borderRadius: '0.5rem' }}>
                            <CheckCircle2 style={{ color: '#16a34a' }} size={24} />
                        </div>
                        <span style={{ fontSize: '0.75rem', fontWeight: 600, textTransform: 'uppercase', color: '#9ca3af', letterSpacing: '0.05em' }}>Status</span>
                    </div>
                    <div style={{ fontSize: '1.5rem', fontWeight: 700, color: '#111827', marginBottom: '0.25rem' }}>All Systems Normal</div>
                    <p style={{ fontSize: '0.875rem', color: '#6b7280' }}>{activeConnections} active connections monitored</p>
                </div>

                <div style={{ backgroundColor: '#fff', borderRadius: '1rem', padding: '1.5rem', border: '1px solid #f3f4f6', boxShadow: '0 4px 6px -1px rgba(0,0,0,0.02)' }}>
                    <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: '1rem' }}>
                        <div style={{ padding: '0.5rem', backgroundColor: '#eff6ff', borderRadius: '0.5rem' }}>
                            <Activity style={{ color: '#2563eb' }} size={24} />
                        </div>
                        <span style={{ fontSize: '0.75rem', fontWeight: 600, textTransform: 'uppercase', color: '#3b82f6', letterSpacing: '0.05em' }}>Active Scan</span>
                    </div>
                    <div style={{ fontSize: '1.5rem', fontWeight: 700, color: '#111827', marginBottom: '0.25rem' }}>Ready</div>
                    <p style={{ fontSize: '0.875rem', color: '#6b7280' }}>System idle and ready for scans</p>
                </div>

                <div
                    onClick={() => setShowModal(true)}
                    style={{
                        backgroundColor: '#fff', borderRadius: '1rem', padding: '1.5rem', border: '1px solid #f3f4f6',
                        boxShadow: '0 4px 6px -1px rgba(0,0,0,0.02)', display: 'flex', flexDirection: 'column',
                        justifyContent: 'center', alignItems: 'center', textAlign: 'center', cursor: 'pointer',
                    }}
                >
                    <div style={{ padding: '0.75rem', backgroundColor: '#eff6ff', borderRadius: '9999px', marginBottom: '0.75rem' }}>
                        <LinkIcon style={{ color: '#2563eb' }} size={24} />
                    </div>
                    <h3 style={{ fontWeight: 500, color: '#111827' }}>Connect New Source</h3>
                    <p style={{ fontSize: '0.75rem', color: '#6b7280', marginTop: '0.25rem' }}>Integrate DBs, S3 buckets, or APIs</p>
                </div>
            </div>

            {/* Search & Filter */}
            <div style={{ display: 'flex', gap: '1rem', marginBottom: '1.5rem', flexWrap: 'wrap' }}>
                <div style={{ position: 'relative', flex: 1, minWidth: '200px' }}>
                    <Search size={20} style={{ position: 'absolute', left: '0.75rem', top: '50%', transform: 'translateY(-50%)', color: '#9ca3af' }} />
                    <input
                        style={{
                            width: '100%', paddingLeft: '2.5rem', paddingRight: '1rem', paddingTop: '0.625rem', paddingBottom: '0.625rem',
                            borderRadius: '0.75rem', border: '1px solid #e5e7eb', backgroundColor: '#ffffff',
                            fontSize: '0.875rem', color: '#111827', outline: 'none', boxSizing: 'border-box',
                        }}
                        placeholder="Search databases..."
                        type="text"
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                    />
                </div>
                <div style={{ display: 'flex', gap: '0.75rem' }}>
                    <select
                        style={{
                            paddingLeft: '0.75rem', paddingRight: '2rem', paddingTop: '0.625rem', paddingBottom: '0.625rem',
                            borderRadius: '0.75rem', border: '1px solid #e5e7eb', backgroundColor: '#ffffff',
                            fontSize: '0.875rem', color: '#374151', outline: 'none',
                        }}
                        value={typeFilter}
                        onChange={(e) => setTypeFilter(e.target.value)}
                    >
                        <option>All Types</option>
                        {DS_TYPE_OPTIONS.map(opt => <option key={opt.value} value={opt.label}>{opt.label}</option>)}
                    </select>
                </div>
            </div>

            {/* Data Source Cards Grid */}
            {isLoading ? (
                <div style={{ display: 'flex', justifyContent: 'center', padding: '5rem 0' }}>
                    <Loader2 className="animate-spin" size={40} style={{ color: '#3b82f6' }} />
                </div>
            ) : filteredDataSources.length === 0 ? (
                <div style={{ textAlign: 'center', padding: '5rem 0', backgroundColor: '#ffffff', borderRadius: '1rem', border: '1px solid #e5e7eb' }}>
                    <DatabaseIcon size={48} style={{ margin: '0 auto 1rem', color: '#9ca3af' }} />
                    <h3 style={{ fontSize: '1.125rem', fontWeight: 500, color: '#111827', marginBottom: '0.5rem' }}>No data sources found</h3>
                    <p style={{ color: '#6b7280', marginBottom: '1.5rem' }}>Get started by connecting your first database or storage system.</p>
                    <Button onClick={() => setShowModal(true)} icon={<Plus size={16} />}>Connect Data Source</Button>
                </div>
            ) : (
                <>
                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))', gap: '1.5rem' }}>
                        {filteredDataSources.map(ds => (
                            <div key={ds.id} onClick={() => navigate(`/datasources/${ds.id}`)} style={{ cursor: 'pointer', height: '100%' }}>
                                <DataSourceCard
                                    dataSource={ds}
                                    onHistory={() => setHistoryId(ds.id)}
                                    onDelete={() => setDeleteId(ds.id)}
                                />
                            </div>
                        ))}
                    </div>

                    <div style={{ marginTop: '2.5rem', display: 'flex', alignItems: 'center', justifyContent: 'space-between', borderTop: '1px solid #e5e7eb', paddingTop: '1.5rem' }}>
                        <p style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                            Showing <span style={{ fontWeight: 500, color: '#111827' }}>1</span> to <span style={{ fontWeight: 500, color: '#111827' }}>{filteredDataSources.length}</span> of <span style={{ fontWeight: 500, color: '#111827' }}>{filteredDataSources.length}</span> results
                        </p>
                        <div style={{ display: 'flex', gap: '0.5rem' }}>
                            <span style={{ padding: '0.5rem 1rem', fontSize: '0.875rem', fontWeight: 500, color: '#6b7280', backgroundColor: '#ffffff', border: '1px solid #e5e7eb', borderRadius: '0.5rem', opacity: 0.5 }}>Previous</span>
                            <span style={{ padding: '0.5rem 1rem', fontSize: '0.875rem', fontWeight: 500, color: '#6b7280', backgroundColor: '#ffffff', border: '1px solid #e5e7eb', borderRadius: '0.5rem', opacity: 0.5 }}>Next</span>
                        </div>
                    </div>
                </>
            )}

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
                {import.meta.env.DEV && (
                    <div className="mb-5 bg-blue-50/50 border border-blue-100 rounded-lg p-3 text-sm">
                        <div className="font-semibold text-blue-900 mb-2 flex items-center gap-1.5">
                            <span className="w-2 h-2 rounded-full bg-blue-500 animate-pulse"></span>
                            Dev Environment Connectors
                        </div>
                        <div className="grid grid-cols-3 gap-2 text-xs">
                            <div className="bg-white p-2 rounded shadow-sm border border-blue-50">
                                <span className="font-semibold text-gray-700 block mb-1">PostgreSQL</span>
                                <span className="text-gray-500">Host:</span> localhost<br />
                                <span className="text-gray-500">Port:</span> 5434<br />
                                <span className="text-gray-500">DB:</span> customers_db<br />
                                <span className="text-gray-500">User:</span> postgres<br />
                                <span className="text-gray-500">Pass:</span> postgres
                            </div>
                            <div className="bg-white p-2 rounded shadow-sm border border-blue-50">
                                <span className="font-semibold text-gray-700 block mb-1">MySQL</span>
                                <span className="text-gray-500">Host:</span> localhost<br />
                                <span className="text-gray-500">Port:</span> 3307<br />
                                <span className="text-gray-500">DB:</span> inventory_db<br />
                                <span className="text-gray-500">User:</span> root<br />
                                <span className="text-gray-500">Pass:</span> root
                            </div>
                            <div className="bg-white p-2 rounded shadow-sm border border-blue-50">
                                <span className="font-semibold text-gray-700 block mb-1">MongoDB</span>
                                <span className="text-gray-500">Host:</span> localhost<br />
                                <span className="text-gray-500">Port:</span> 27018<br />
                                <span className="text-gray-500">DB:</span> admin<br />
                                <span className="text-gray-500">User:</span> admin<br />
                                <span className="text-gray-500">Pass:</span> password
                            </div>
                        </div>
                    </div>
                )}
                <form id="addDsForm" onSubmit={handleCreate} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                    <div>
                        <label style={labelStyle}>Data Source Type</label>
                        <select value={form.type} onChange={updateField('type')} style={inputStyle}>
                            {DS_TYPE_OPTIONS.map((opt) => (
                                <option key={opt.value} value={opt.value}>{opt.label}</option>
                            ))}
                        </select>
                    </div>

                    {(form.type === 'MICROSOFT_365' || form.type === 'GOOGLE_WORKSPACE') ? (
                        <div className="space-y-4">
                            {form.type === 'MICROSOFT_365' && (
                                <div className="bg-blue-50 p-4 rounded-lg border border-blue-100 text-sm text-blue-800">
                                    <h4 className="font-semibold mb-1">Configuration Required</h4>
                                    <p className="mb-2">Ensure your Azure AD Application is configured with this Redirect URI:</p>
                                    <code className="bg-white px-2 py-1 rounded border border-blue-200 block w-full mb-3 select-all">
                                        {window.location.origin}/api/v2/m365/callback
                                    </code>
                                    <p>Authentication uses the globally configured Client ID.</p>
                                </div>
                            )}

                            {form.type === 'GOOGLE_WORKSPACE' && (
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

                            {(form.type === 'GOOGLE_WORKSPACE' && form.credentials) ? (
                                <div>
                                    <label style={labelStyle}>Service Account Key (JSON)</label>
                                    <textarea
                                        value={form.credentials === '{}' ? '' : form.credentials}
                                        onChange={updateField('credentials')}
                                        style={{ ...inputStyle, height: '120px', padding: '0.75rem', resize: 'vertical', fontFamily: 'monospace', fontSize: '12px' }}
                                        placeholder='{"type":"service_account", ... }'
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
                                            You will be redirected to {form.type === 'MICROSOFT_365' ? 'Microsoft' : 'Google'} to authenticate and grant access.
                                        </p>
                                    </div>
                                    <Button
                                        type="submit"
                                        isLoading={isOAuthPending}
                                        className="w-full justify-center"
                                    >
                                        {isOAuthPending ? 'Connecting...' : `Connect ${form.type === 'MICROSOFT_365' ? 'Microsoft 365' : 'Google Workspace'}`}
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

                            {['POSTGRESQL', 'MYSQL', 'MONGODB', 'SQLSERVER', 'AZURE_SQL', 'RDS'].includes(form.type) ? (
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

                    {form.type === 'FILE_UPLOAD' && (
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
