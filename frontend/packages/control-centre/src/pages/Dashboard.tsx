import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Database, ShieldCheck, AlertTriangle, Activity, ArrowRight } from 'lucide-react';
import { Button, Feature09, Headline01, Card08 } from '@datalens/shared';
// Removed: Card, CardContent, CardHeader, CardTitle (replaced by Card08)
// Removed: StatCard (replaced by Feature09)
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

    // Transform stats for Bento Grid
    const bentoItems = [
        {
            title: "Total Data Sources",
            value: stats?.total_data_sources ?? 0,
            icon: Database,
            color: "primary" as const,
            description: "Connected sources",
        },
        {
            title: "Total Scans Run",
            value: stats?.total_scans ?? 0,
            icon: Activity,
            color: "info" as const,
            description: "Scans executed",
        },
        {
            title: "PII Fields Found",
            value: stats?.total_pii_fields ?? 0,
            icon: ShieldCheck,
            color: "danger" as const,
            description: "Sensitive fields detected",
        },
        {
            title: "Pending Reviews",
            value: stats?.pending_reviews ?? 0,
            icon: AlertTriangle,
            color: "warning" as const,
            description: "Items needing attention",
        },
    ];

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
        <div className="space-y-8 p-6">
            {/* Header Section */}
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div>
                    <Headline01 title="Dashboard" subtitle="Overview of your data privacy posture" />
                </div>
                {/* Actions moved to global toolbar, keeping legacy buttons for now if needed, or removing as per plan to use ActionToolbar globally */}
            </div>

            {/* Bento Grid Stats */}
            <Feature09 items={bentoItems} />

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Recent Scans */}
                <Card08 title="Recent Scans" className="lg:col-span-2" action={
                    <Button variant="ghost" size="sm" onClick={() => navigate('/datasources')}>
                        View All <ArrowRight size={16} className="ml-2" />
                    </Button>
                }>
                    <ErrorBoundary FallbackComponent={SectionErrorFallback}>
                        <DataTable
                            columns={recentScansColumns}
                            data={stats?.recent_scans ?? []}
                            isLoading={isLoading}
                            keyExtractor={(row) => row.id}
                            emptyTitle="No recent scans"
                            emptyDescription="Connect your first data source to begin scanning for PII. Results will appear here automatically."
                            loadingRows={3}
                        />
                    </ErrorBoundary>
                </Card08>

                {/* PII Distribution */}
                <Card08 title="PII by Category">
                    <ErrorBoundary FallbackComponent={SectionErrorFallback}>
                        <PIIChart data={stats?.pii_by_category ?? {}} loading={isLoading} />
                    </ErrorBoundary>
                </Card08>
            </div>
        </div>
    );
};

export default Dashboard;
