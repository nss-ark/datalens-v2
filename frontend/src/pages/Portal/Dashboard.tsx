import { usePortalAuthStore } from '../../stores/portalAuthStore';
import { IdentityCard } from '../../components/Portal/IdentityCard';

const PortalDashboard = () => {
    const profile = usePortalAuthStore(state => state.profile);

    return (
        <div>
            <div className="mb-8">
                <h1 className="text-3xl font-bold text-gray-900">Hello, {profile?.email}</h1>
                <p className="text-gray-500 mt-2">Manage your privacy preferences and data rights.</p>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
                {/* Identity Verification Card - Takes full width on mobile, 1 col on large screens */}
                <div className="lg:col-span-1">
                    <IdentityCard />
                </div>

                {/* Stats Cards - Spans 2 cols */}
                <div className="lg:col-span-2 grid grid-cols-1 sm:grid-cols-3 gap-6">
                    <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm">
                        <div className="text-sm font-medium text-gray-500 mb-1">Privacy Score</div>
                        <div className="text-3xl font-bold text-green-600">Good</div>
                    </div>
                    <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm">
                        <div className="text-sm font-medium text-gray-500 mb-1">Active Consents</div>
                        <div className="text-3xl font-bold text-blue-600">3</div>
                    </div>
                    <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm">
                        <div className="text-sm font-medium text-gray-500 mb-1">Open Requests</div>
                        <div className="text-3xl font-bold text-orange-600">1</div>
                    </div>
                </div>
            </div>

            <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-6">
                <h3 className="text-lg font-bold text-gray-900 mb-4">Quick Actions</h3>
                <div className="flex gap-4">
                    <button className="px-4 py-2 bg-blue-50 text-blue-700 rounded-lg font-medium hover:bg-blue-100 transition-colors">
                        Manage Consents
                    </button>
                    <button className="px-4 py-2 bg-blue-50 text-blue-700 rounded-lg font-medium hover:bg-blue-100 transition-colors">
                        Download My Data
                    </button>
                    <button className="px-4 py-2 bg-blue-50 text-blue-700 rounded-lg font-medium hover:bg-blue-100 transition-colors">
                        Submit Request
                    </button>
                </div>
            </div>
        </div>
    );
};

export default PortalDashboard;
