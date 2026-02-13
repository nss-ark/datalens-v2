import { Building, Users, Activity, Server } from 'lucide-react';
import { StatCard } from '../../components/Dashboard/StatCard';

const AdminDashboard = () => {
    return (
        <div className="space-y-6">
            <div>
                <h1 className="text-2xl font-bold text-gray-900">Platform Overview</h1>
                <p className="text-gray-500">Monitor system health and tenant usage</p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <StatCard
                    title="Active Tenants"
                    value="12"
                    icon={Building}
                    color="primary"
                    trend={{ value: 8, label: "vs last month", direction: "up" }}
                />
                <StatCard
                    title="Total Users"
                    value="148"
                    icon={Users}
                    color="info"
                    trend={{ value: 12, label: "vs last month", direction: "up" }}
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
