import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Building, Users, Activity, Server, Plus, ArrowRight, Shield, Clock } from 'lucide-react';
import { AnalyticsCard, MotionList, cn } from '@datalens/shared';
import { adminService } from '@/services/adminService';
import type { AdminStats, Tenant } from '@/types/admin';

const SPARKLINE_TENANTS = [3, 5, 4, 7, 8, 10, 12];
const SPARKLINE_USERS = [800, 850, 900, 950, 1000, 1100, 1250];
const SPARKLINE_HEALTH = [99.8, 99.9, 99.7, 99.9, 100, 99.9, 99.9];
const SPARKLINE_STORAGE = [0.8, 0.85, 0.9, 0.95, 1.0, 1.1, 1.2];

const QUICK_ACTIONS = [
    { id: 'create-tenant', label: 'Create Tenant', description: 'Onboard a new organization', icon: Plus, path: '/tenants', color: 'bg-blue-500' },
    { id: 'view-users', label: 'Platform Users', description: 'Manage users across tenants', icon: Users, path: '/users', color: 'bg-emerald-500' },
    { id: 'dsr-overview', label: 'DSR Overview', description: 'Monitor data subject requests', icon: Shield, path: '/compliance/dsr', color: 'bg-purple-500' },
];

const AdminDashboard = () => {
    const navigate = useNavigate();
    const [stats, setStats] = useState<AdminStats | null>(null);
    const [recentTenants, setRecentTenants] = useState<Tenant[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const [statsData, tenantsData] = await Promise.all([
                    adminService.getStats(),
                    adminService.getTenants({ page: 1, limit: 5 }),
                ]);
                setStats(statsData);
                setRecentTenants(tenantsData.items.slice(0, 5));
            } catch (err) {
                console.error('Failed to fetch admin data:', err);
                setError('Failed to load dashboard data');
            } finally {
                setIsLoading(false);
            }
        };

        fetchData();
    }, []);

    if (isLoading) {
        return (
            <div className="space-y-6 animate-pulse">
                <div>
                    <div className="h-8 w-48 bg-zinc-200 dark:bg-zinc-700 rounded mb-2"></div>
                    <div className="h-4 w-64 bg-zinc-200 dark:bg-zinc-700 rounded"></div>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                    {[1, 2, 3, 4].map((i) => (
                        <div key={i} className="bg-white dark:bg-zinc-900 h-44 rounded-xl border border-zinc-200 dark:border-zinc-800"></div>
                    ))}
                </div>
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                    <div className="bg-white dark:bg-zinc-900 h-80 rounded-xl border border-zinc-200 dark:border-zinc-800"></div>
                    <div className="bg-white dark:bg-zinc-900 h-80 rounded-xl border border-zinc-200 dark:border-zinc-800"></div>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="p-6 bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-400 rounded-xl border border-red-200 dark:border-red-800">
                {error}
            </div>
        );
    }

    return (
        <div className="space-y-8">
            {/* Page Header */}
            <div>
                <h1 className="text-3xl font-bold text-zinc-900 dark:text-zinc-50 tracking-tight">
                    Platform Overview
                </h1>
                <p className="text-zinc-500 dark:text-zinc-400 mt-2 text-lg">
                    Monitor system health and tenant usage
                </p>
            </div>

            {/* Stats Grid with Sparklines */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <AnalyticsCard
                    title="Active Tenants"
                    value={stats?.active_tenants.toString() || "0"}
                    icon={<Building className="h-4 w-4" />}
                    description="Total active organizations"
                    trend={{ value: 12, label: "vs last month", direction: "up" }}
                    sparklineData={SPARKLINE_TENANTS}
                />
                <AnalyticsCard
                    title="Total Users"
                    value={stats?.total_users.toString() || "0"}
                    icon={<Users className="h-4 w-4" />}
                    description="Registered users across all tenants"
                    trend={{ value: 8, label: "vs last month", direction: "up" }}
                    sparklineData={SPARKLINE_USERS}
                />
                <AnalyticsCard
                    title="System Health"
                    value="99.9%"
                    icon={<Activity className="h-4 w-4" />}
                    description="Uptime last 30 days"
                    trend={{ value: 0, label: "Stable", direction: "neutral" }}
                    sparklineData={SPARKLINE_HEALTH}
                />
                <AnalyticsCard
                    title="Storage Used"
                    value="1.2 TB"
                    icon={<Server className="h-4 w-4" />}
                    description="Total storage consumed"
                    trend={{ value: 5, label: "vs last month", direction: "up" }}
                    sparklineData={SPARKLINE_STORAGE}
                />
            </div>

            {/* Bottom Panels: Recent Tenants + Quick Actions */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Recent Tenants */}
                <div className="lg:col-span-2 bg-white dark:bg-zinc-900 rounded-xl border border-zinc-200 dark:border-zinc-800 shadow-sm overflow-hidden">
                    <div className="flex items-center justify-between p-6 pb-4 border-b border-zinc-100 dark:border-zinc-800">
                        <div>
                            <h2 className="text-lg font-semibold text-zinc-900 dark:text-zinc-50">
                                Recent Tenants
                            </h2>
                            <p className="text-sm text-zinc-500 dark:text-zinc-400">
                                Latest organizations onboarded
                            </p>
                        </div>
                        <button
                            onClick={() => navigate('/tenants')}
                            className="text-sm text-blue-500 hover:text-blue-600 dark:text-blue-400 dark:hover:text-blue-300 flex items-center gap-1 font-medium transition-colors"
                        >
                            View All <ArrowRight className="h-4 w-4" />
                        </button>
                    </div>
                    <div className="p-4">
                        <MotionList
                            items={recentTenants}
                            className="space-y-2"
                            renderItem={(tenant) => (
                                <div className="flex items-center justify-between p-3 rounded-lg hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition-colors">
                                    <div className="flex items-center gap-3">
                                        <div className="w-10 h-10 rounded-lg bg-blue-50 dark:bg-blue-500/10 flex items-center justify-center text-blue-600 dark:text-blue-400 font-semibold text-sm">
                                            {tenant.name.charAt(0)}
                                        </div>
                                        <div>
                                            <p className="font-medium text-zinc-900 dark:text-zinc-100 text-sm">
                                                {tenant.name}
                                            </p>
                                            <p className="text-xs text-zinc-500 dark:text-zinc-400">
                                                {tenant.domain}.datalens.com
                                            </p>
                                        </div>
                                    </div>
                                    <div className="flex items-center gap-3">
                                        <span className={cn(
                                            "inline-flex items-center px-2 py-0.5 rounded text-xs font-medium",
                                            tenant.plan === 'ENTERPRISE' && "bg-purple-100 text-purple-700 dark:bg-purple-500/20 dark:text-purple-300",
                                            tenant.plan === 'PROFESSIONAL' && "bg-blue-100 text-blue-700 dark:bg-blue-500/20 dark:text-blue-300",
                                            tenant.plan === 'STARTER' && "bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300",
                                            tenant.plan === 'FREE' && "bg-zinc-100 text-zinc-600 dark:bg-zinc-700 dark:text-zinc-300",
                                        )}>
                                            {tenant.plan}
                                        </span>
                                        <span className={cn(
                                            "w-2 h-2 rounded-full",
                                            tenant.status === 'ACTIVE' ? "bg-emerald-500" : "bg-red-500"
                                        )} />
                                        <span className="text-xs text-zinc-400 dark:text-zinc-500 flex items-center gap-1">
                                            <Clock className="h-3 w-3" />
                                            {new Date(tenant.created_at).toLocaleDateString()}
                                        </span>
                                    </div>
                                </div>
                            )}
                        />
                    </div>
                </div>

                {/* Quick Actions */}
                <div className="bg-white dark:bg-zinc-900 rounded-xl border border-zinc-200 dark:border-zinc-800 shadow-sm">
                    <div className="p-6 pb-4 border-b border-zinc-100 dark:border-zinc-800">
                        <h2 className="text-lg font-semibold text-zinc-900 dark:text-zinc-50">
                            Quick Actions
                        </h2>
                        <p className="text-sm text-zinc-500 dark:text-zinc-400">
                            Common admin tasks
                        </p>
                    </div>
                    <div className="p-4 space-y-2">
                        {QUICK_ACTIONS.map((action) => (
                            <button
                                key={action.id}
                                onClick={() => navigate(action.path)}
                                className="w-full flex items-center gap-4 p-4 rounded-lg hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition-colors text-left group"
                            >
                                <div className={cn(
                                    "w-10 h-10 rounded-lg flex items-center justify-center text-white shadow-sm",
                                    action.color
                                )}>
                                    <action.icon className="h-5 w-5" />
                                </div>
                                <div className="flex-1 min-w-0">
                                    <p className="font-medium text-zinc-900 dark:text-zinc-100 text-sm">
                                        {action.label}
                                    </p>
                                    <p className="text-xs text-zinc-500 dark:text-zinc-400 truncate">
                                        {action.description}
                                    </p>
                                </div>
                                <ArrowRight className="h-4 w-4 text-zinc-300 dark:text-zinc-600 group-hover:text-zinc-500 dark:group-hover:text-zinc-400 transition-colors" />
                            </button>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default AdminDashboard;
