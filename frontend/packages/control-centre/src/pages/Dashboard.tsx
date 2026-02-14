import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Database, ShieldCheck, AlertTriangle, Activity, ArrowRight, Plus } from 'lucide-react';
import { Button } from '@datalens/shared';
import { Card, CardContent, CardHeader, CardTitle } from '@datalens/shared';
import { StatCard } from '../components/Dashboard/StatCard';
import { PIIChart } from '../components/Dashboard/PIIChart';
import { DataTable } from '@datalens/shared';
import { StatusBadge } from '@datalens/shared';
import { ErrorBoundary } from '@datalens/shared';
import { SectionErrorFallback } from '@datalens/shared';
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
        <div className="space-y-6 p-6">
            {/* Actions Header (Title removed to avoid redundancy) */}
            <div className="flex justify-end gap-3">
                <Button variant="outline" onClick={() => navigate('/discovery')}>
                    Review PII
                </Button>
                <Button onClick={() => navigate('/datasources')}>
                    <Plus size={16} className="mr-2" />
                    Add Data Source
                </Button>
            </div>

            {/* Stat Cards - Improved Grid */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
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
                <Card className="lg:col-span-2">
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-lg font-semibold">Recent Scans</CardTitle>
                        <Button variant="ghost" size="sm" onClick={() => navigate('/datasources')}>
                            View All <ArrowRight size={16} className="ml-2" />
                        </Button>
                    </CardHeader>
                    <CardContent>
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
                    </CardContent>
                </Card>

                {/* PII Distribution */}
                <Card>
                    <CardHeader>
                        <CardTitle className="text-lg font-semibold">PII by Category</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <ErrorBoundary FallbackComponent={SectionErrorFallback}>
                            <PIIChart data={stats?.pii_by_category ?? {}} loading={isLoading} />
                        </ErrorBoundary>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
};

export default Dashboard;
