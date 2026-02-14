import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Database, ShieldCheck, AlertTriangle, Activity, ArrowRight, Plus } from 'lucide-react';
import { Button } from '../components/common/Button';
import { StatCard } from '../components/Dashboard/StatCard';
import { PIIChart } from '../components/Dashboard/PIIChart';
import { DataTable } from '../components/DataTable/DataTable';
import { StatusBadge } from '../components/common/StatusBadge';
import { ErrorBoundary } from '../components/common/ErrorBoundary';
import { SectionErrorFallback } from '../components/common/ErrorFallbacks';
import { dashboardService } from '../services/dashboard';
import type { ScanSummary } from '../types/dashboard';

const Dashboard = () => {
    const navigate = useNavigate();
    const { data: stats, isLoading } = useQuery({
        queryKey: ['dashboardStats'],
        queryFn: dashboardService.getStats,
        refetchInterval: 30000, // Refresh every 30s
    });

    const recentScansColumns = [
        {
            key: 'data_source_name',
            header: 'Data Source',
            render: (row: ScanSummary) => (
                <div className="font-medium text-gray-900">{row.data_source_name || 'Unknown Source'}</div>
            ),
        },
        {
            key: 'status',
            header: 'Status',
            render: (row: ScanSummary) => <StatusBadge label={row.status} />,
        },
        {
            key: 'tables_scanned',
            header: 'Tables',
            render: (row: ScanSummary) => <span>{row.tables_scanned}</span>,
        },
        {
            key: 'pii_found',
            header: 'PII Found',
            render: (row: ScanSummary) => (
                <span className={row.pii_found > 0 ? 'text-amber-600 font-medium' : 'text-gray-500'}>
                    {row.pii_found}
                </span>
            ),
        },
        {
            key: 'started_at',
            header: 'Time',
            render: (row: ScanSummary) => (
                <span className="text-gray-500 text-sm">
                    {new Date(row.started_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                </span>
            ),
        },
    ];

    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 tracking-tight">Dashboard</h1>
                    <p className="text-gray-500 mt-1">Overview of your data compliance posture</p>
                </div>
                <div className="flex gap-3">
                    <Button variant="outline" onClick={() => navigate('/discovery')}>
                        Review PII
                    </Button>
                    <Button icon={<Plus size={16} />} onClick={() => navigate('/datasources')}>
                        Add Data Source
                    </Button>
                </div>
            </div>

            {/* Stat Cards */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <StatCard
                    title="Total Data Sources"
                    value={stats?.total_data_sources ?? 0}
                    icon={Database}
                    color="primary"
                    loading={isLoading}
                />
                <StatCard
                    title="Total Scans Run"
                    value={stats?.total_scans ?? 0}
                    icon={Activity}
                    color="info"
                    loading={isLoading}
                />
                <StatCard
                    title="PII Fields Found"
                    value={stats?.total_pii_fields ?? 0}
                    icon={ShieldCheck}
                    color="danger"
                    loading={isLoading}
                />
                <StatCard
                    title="Pending Reviews"
                    value={stats?.pending_reviews ?? 0}
                    icon={AlertTriangle}
                    color="warning"
                    loading={isLoading}
                />
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Recent Scans */}
                <div className="lg:col-span-2 bg-white rounded-lg border border-gray-200 shadow-sm p-6">
                    <div className="flex justify-between items-center mb-6">
                        <h2 className="text-lg font-semibold text-gray-900">Recent Scans</h2>
                        <Button variant="ghost" size="sm" rightIcon={<ArrowRight size={16} />} onClick={() => navigate('/datasources')}>
                            View All
                        </Button>
                    </div>
                    <ErrorBoundary FallbackComponent={SectionErrorFallback}>
                        <DataTable
                            columns={recentScansColumns}
                            data={stats?.recent_scans ?? []}
                            isLoading={isLoading}
                            keyExtractor={(row) => row.id}
                            emptyTitle="No recent scans"
                            emptyDescription="Start a scan from the Data Sources page"
                            loadingRows={3}
                        />
                    </ErrorBoundary>
                </div>

                {/* PII Distribution */}
                <div className="bg-white rounded-lg border border-gray-200 shadow-sm p-6">
                    <div className="flex justify-between items-center mb-6">
                        <h2 className="text-lg font-semibold text-gray-900">PII by Category</h2>
                    </div>
                    <ErrorBoundary FallbackComponent={SectionErrorFallback}>
                        <PIIChart data={stats?.pii_by_category ?? {}} loading={isLoading} />
                    </ErrorBoundary>
                </div>
            </div>
        </div>
    );
};

export default Dashboard;
