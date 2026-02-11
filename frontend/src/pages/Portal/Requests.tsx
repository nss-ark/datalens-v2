import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { portalService } from '../../services/portalService';
import { Plus } from 'lucide-react';
import { format } from 'date-fns';
import { DPRRequestModal } from '../../components/Portal/DPRRequestModal';
import { StatusBadge } from '../../components/common/StatusBadge';

const PortalRequests = () => {
    const [isCreateOpen, setCreateOpen] = useState(false);
    const { data: requests, isLoading, refetch } = useQuery({
        queryKey: ['portal-requests'],
        queryFn: () => portalService.listRequests(),
    });

    const items = requests?.items || [];

    return (
        <div>
            <div className="flex justify-between items-center mb-6">
                <h1 className="text-2xl font-bold text-gray-900">My Requests</h1>
                <button
                    onClick={() => setCreateOpen(true)}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                >
                    <Plus size={18} />
                    New Request
                </button>
            </div>

            <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
                {isLoading ? (
                    <div className="p-8 text-center text-gray-500">Loading requests...</div>
                ) : items.length === 0 ? (
                    <div className="p-12 text-center">
                        <p className="text-gray-500 mb-4">You haven't submitted any requests yet.</p>
                        <button
                            onClick={() => setCreateOpen(true)}
                            className="text-blue-600 hover:underline font-medium"
                        >
                            Submit your first request
                        </button>
                    </div>
                ) : (
                    <table className="w-full text-left text-sm">
                        <thead className="bg-gray-50 border-b border-gray-200">
                            <tr>
                                <th className="px-6 py-4 font-medium text-gray-500">Type</th>
                                <th className="px-6 py-4 font-medium text-gray-500">Description</th>
                                <th className="px-6 py-4 font-medium text-gray-500">Date</th>
                                <th className="px-6 py-4 font-medium text-gray-500">Status</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-200">
                            {items.map((req) => (
                                <tr key={req.id} className="hover:bg-gray-50">
                                    <td className="px-6 py-4 font-medium text-gray-900">{req.type}</td>
                                    <td className="px-6 py-4 text-gray-500 max-w-xs truncate">{req.description}</td>
                                    <td className="px-6 py-4 text-gray-500">
                                        {format(new Date(req.submitted_at), 'MMM d, yyyy')}
                                    </td>
                                    <td className="px-6 py-4">
                                        <StatusBadge label={req.status} />
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )}
            </div>

            <DPRRequestModal
                isOpen={isCreateOpen}
                onClose={() => setCreateOpen(false)}
                onSuccess={() => {
                    setCreateOpen(false);
                    refetch();
                }}
            />
        </div>
    );
};

export default PortalRequests;
