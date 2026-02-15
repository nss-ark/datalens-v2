import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { portalService } from '@/services/portalService';
import { Plus, ShieldAlert } from 'lucide-react';
import { format } from 'date-fns';
import { StatusBadge, Button } from '@datalens/shared';
import { useNavigate } from 'react-router-dom';
import { AppealModal } from '@/components/AppealModal';

const PortalRequests = () => {
    const navigate = useNavigate();
    const [appealId, setAppealId] = useState<string | null>(null);
    const [isAppealOpen, setIsAppealOpen] = useState(false);

    const { data: requests, isLoading, refetch } = useQuery({
        queryKey: ['portal-requests'],
        queryFn: () => portalService.listRequests(),
    });

    const items = requests?.items || [];

    const handleOpenAppeal = (id: string) => {
        setAppealId(id);
        setIsAppealOpen(true);
    };

    return (
        <div>
            <div className="flex justify-between items-center mb-6">
                <h1 className="text-2xl font-bold text-gray-900">My Requests</h1>
                <button
                    onClick={() => navigate('/portal/requests/new')}
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
                            onClick={() => navigate('/portal/requests/new')}
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
                                <th className="px-6 py-4 font-medium text-gray-500">Actions</th>
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
                                    <td className="px-6 py-4">
                                        {req.status === 'REJECTED' && (
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                onClick={() => handleOpenAppeal(req.id)}
                                                className="text-orange-600 border-orange-200 hover:bg-orange-50 hover:text-orange-700"
                                            >
                                                <ShieldAlert size={14} className="mr-1" />
                                                Appeal
                                            </Button>
                                        )}
                                        {req.status === 'APPEALED' && (
                                            <span className="text-xs font-medium text-orange-600 bg-orange-50 px-2 py-1 rounded-full border border-orange-100 flex items-center w-fit gap-1">
                                                <ShieldAlert size={12} />
                                                Under Appeal
                                            </span>
                                        )}
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )}
            </div>

            {appealId && (
                <AppealModal
                    isOpen={isAppealOpen}
                    dprId={appealId}
                    onClose={() => setIsAppealOpen(false)}
                    onSuccess={refetch}
                />
            )}
        </div>
    );
};

export default PortalRequests;
