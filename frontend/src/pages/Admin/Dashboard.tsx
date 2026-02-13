import { useState, useEffect } from 'react';
import { Building, Users, Activity, Server } from 'lucide-react';
import { StatCard } from '../../components/Dashboard/StatCard';
import { adminService } from '../../services/adminService';
import type { AdminStats } from '../../types/admin';

const AdminDashboard = () => {
    const [stats, setStats] = useState<AdminStats | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchStats = async () => {
            try {
                const data = await adminService.getStats();
                setStats(data);
            } catch (err) {
                console.error('Failed to fetch admin stats:', err);
                setError('Failed to load dashboard statistics');
            } finally {
                setIsLoading(false);
            }
        };

        fetchStats();
    }, []);

    if (isLoading) {
        return (
            <div className="space-y-6 animate-pulse">
                <div>
                    <div className="h-8 w-48 bg-gray-200 rounded mb-2"></div>
                    <div className="h-4 w-64 bg-gray-200 rounded"></div>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                    {[1, 2, 3, 4].map((i) => (
                        <div key={i} className="bg-white h-32 rounded-lg border border-gray-200"></div>
                    ))}
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="p-6 bg-red-50 text-red-700 rounded-lg border border-red-200">
                {error}
            </div>
        );
    }

    return (
        <div className="space-y-6">
            <div>
                <h1 className="text-2xl font-bold text-gray-900">Platform Overview</h1>
                <p className="text-gray-500">Monitor system health and tenant usage</p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <StatCard
                    title="Active Tenants"
                    value={stats?.active_tenants.toString() || "0"}
                    icon={Building}
                    color="primary"
                // Trend removed as we don't have historical data yet
                />
                <StatCard
                    title="Total Users"
                    value={stats?.total_users.toString() || "0"}
                    icon={Users}
                    color="info"
                />
                <StatCard
                    title="System Health"
                    value="99.9%"
                    icon={Activity}
                    color="success"
                />
                <StatCard
                    title="Storage Used"
                    value="1.2 TB"
                    icon={Server}
                    color="warning"
                    trend={{ value: 5, label: "vs last month", direction: "up" }}
                />
            </div>

            {/* Placeholder for future widgets */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div className="bg-white p-6 rounded-lg border border-gray-200 shadow-sm h-64 flex items-center justify-center text-gray-400">
                    Tenant Growth Chart Placeholder
                </div>
                <div className="bg-white p-6 rounded-lg border border-gray-200 shadow-sm h-64 flex items-center justify-center text-gray-400">
                    Recent Activity Log Placeholder
                </div>
            </div>
        </div>
    );
};

export default AdminDashboard;
