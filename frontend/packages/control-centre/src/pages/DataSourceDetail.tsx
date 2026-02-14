import { useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import {
    ArrowLeft, Database, Play, History, Settings,
    Clock
} from 'lucide-react';
import { Button } from '@datalens/shared';
import { StatusBadge } from '@datalens/shared';
import { ScanHistoryModal } from '../components/DataSources/ScanHistoryModal';
import { useDataSource, useScanDataSource, useScanStatus } from '../hooks/useDataSources';
import { toast } from '@datalens/shared';

const DataSourceDetail = () => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const { data: dataSource, isLoading } = useDataSource(id || '');
    const { mutate: scanMutate, isPending: isStartingScan } = useScanDataSource();

    // Poll for scan status
    const { data: scanStatus } = useScanStatus(id || '', true);

    const [showHistory, setShowHistory] = useState(false);

    const handleScan = () => {
        if (!id) return;
        scanMutate(id, {
            onSuccess: () => toast.success('Scan started', 'The data source scan has been initiated.'),
            onError: () => toast.error('Scan failed', 'Could not initiate scan. Please try again.'),
        });
    };

    if (isLoading) {
        return <div className="p-8"><div className="animate-pulse space-y-4">{[1, 2, 3].map(i => <div key={i} className="h-12 bg-gray-100 rounded" />)}</div></div>;
    }

    if (!dataSource) {
        return (
            <div className="p-8 text-center bg-white rounded-lg border border-gray-200 m-8">
                <h2 className="text-xl font-semibold text-gray-900 mb-2">Data Source Not Found</h2>
                <Button variant="ghost" onClick={() => navigate('/datasources')}>Back to List</Button>
            </div>
        );
    }

    const isM365 = ['onedrive', 'sharepoint', 'outlook', 'm365'].includes(dataSource.type);
    const isScanning = scanStatus?.status === 'RUNNING' || scanStatus?.status === 'QUEUED';

    // Helper to format date
    const formatDate = (dateStr?: string | null) => {
        if (!dateStr) return 'Never';
        return new Date(dateStr).toLocaleString();
    };

    return (
        <div className="p-6 max-w-7xl mx-auto">
            {/* Header */}
            <div className="mb-6">
                <button
                    onClick={() => navigate('/datasources')}
                    className="flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-900 mb-4 transition-colors"
                >
                    <ArrowLeft size={16} /> Back to Data Sources
                </button>

                <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                    <div className="flex items-center gap-4">
                        <div className="w-12 h-12 rounded-lg bg-blue-50 text-blue-600 flex items-center justify-center shrink-0">
                            <Database size={24} />
                        </div>
                        <div>
                            <h1 className="text-2xl font-bold text-gray-900">{dataSource.name}</h1>
                            <div className="flex items-center gap-2 mt-1">
                                <StatusBadge label={dataSource.status} />
                                <span className="text-sm text-gray-500 uppercase tracking-wider font-medium px-2 border-l border-gray-200">
                                    {dataSource.type}
                                </span>
                            </div>
                        </div>
                    </div>

                    <div className="flex items-center gap-3">
                        <Button
                            variant="ghost"
                            icon={<History size={16} />}
                            onClick={() => setShowHistory(true)}
                        >
                            History
                        </Button>

                        {isM365 && (
                            <Link to={`/datasources/${dataSource.id}/config`}>
                                <Button variant="outline" icon={<Settings size={16} />}>
                                    Configure Scope
                                </Button>
                            </Link>
                        )}

                        <Button
                            icon={isScanning ? <Clock size={16} className="animate-spin" /> : <Play size={16} />}
                            onClick={handleScan}
                            disabled={isScanning || isStartingScan}
                        >
                            {isScanning ? 'Scanning...' : 'Scan Now'}
                        </Button>
                    </div>
                </div>
            </div>

            {/* Content Grid */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">

                {/* Main Details */}
                <div className="lg:col-span-2 space-y-6">
                    {/* Connection Info */}
                    <section className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm">
                        <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
                            <Settings size={18} className="text-gray-400" />
                            Connection Details
                        </h3>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-y-4 gap-x-8">
                            <DetailRow label="Host" value={dataSource.host} />
                            <DetailRow label="Port" value={dataSource.port?.toString()} />
                            <DetailRow label="Database / Bucket" value={dataSource.database} />
                            <DetailRow label="Description" value={dataSource.description} fullWidth />
                        </div>
                    </section>
                </div>

                {/* Sidebar Stats */}
                <div className="space-y-6">
                    {/* Scan Status Card */}
                    <div className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm">
                        <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wider mb-4">
                            Last Scan
                        </h3>

                        {scanStatus ? (
                            <div className="space-y-4">
                                <div className="flex items-center justify-between">
                                    <span className="text-sm text-gray-600">Status</span>
                                    <StatusBadge label={scanStatus.status} />
                                </div>

                                {scanStatus.status === 'RUNNING' && (
                                    <div className="space-y-1">
                                        <div className="flex justify-between text-xs text-gray-500">
                                            <span>Progress</span>
                                            <span>{scanStatus.progress_percentage}%</span>
                                        </div>
                                        <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
                                            <div
                                                className="h-full bg-blue-600 transition-all duration-500"
                                                style={{ width: `${scanStatus.progress_percentage}%` }}
                                            />
                                        </div>
                                        {scanStatus.current_table && (
                                            <div className="text-xs text-gray-400 truncate">
                                                Scanning: {scanStatus.current_table}
                                            </div>
                                        )}
                                    </div>
                                )}

                                <div className="pt-4 border-t border-gray-100 grid grid-cols-2 gap-4">
                                    <div>
                                        <div className="text-2xl font-bold text-gray-900">{scanStatus.tables_processed || 0}</div>
                                        <div className="text-xs text-gray-500">Files/Tables</div>
                                    </div>
                                    <div>
                                        <div className="text-2xl font-bold text-gray-900 text-purple-600">{scanStatus.pii_found || 0}</div>
                                        <div className="text-xs text-gray-500">PII Detected</div>
                                    </div>
                                </div>
                            </div>
                        ) : (
                            <div className="text-sm text-gray-500 italic">No recent scan activity</div>
                        )}

                        <div className="mt-4 pt-4 border-t border-gray-100 text-xs text-gray-400 flex items-center gap-1.5">
                            <Clock size={12} />
                            Last synced: {formatDate(dataSource.last_sync_at)}
                        </div>
                    </div>
                </div>
            </div>

            <ScanHistoryModal
                dataSourceId={showHistory ? id ?? null : null}
                onClose={() => setShowHistory(false)}
            />
        </div>
    );
};

const DetailRow = ({ label, value, fullWidth = false }: { label: string; value?: string | null; fullWidth?: boolean }) => (
    <div className={fullWidth ? 'col-span-1 md:col-span-2' : ''}>
        <dt className="text-xs font-medium text-gray-500 mb-1">{label}</dt>
        <dd className="text-sm text-gray-900 font-medium break-words">{value || <span className="text-gray-400 italic">Not set</span>}</dd>
    </div>
);

export default DataSourceDetail;
