import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { portalService } from '@/services/portalService';
import { Plus, ShieldAlert, FileText } from 'lucide-react';
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
        <div className="animate-fade-in">
            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-8">
                <div className="page-header !mb-0">
                    <h1>My Requests</h1>
                    <p>Track and manage your data privacy requests.</p>
                </div>
                <button
                    onClick={() => navigate('/requests/new')}
                    className="flex items-center gap-2 px-5 py-2.5 bg-blue-600 text-white rounded-xl hover:bg-blue-700 transition-all duration-200 text-sm font-semibold shadow-sm hover:shadow-md active:scale-[0.98]"
                >
                    <Plus size={16} />
                    New Request
                </button>
            </div>

            <div className="portal-card overflow-hidden">
                {isLoading ? (
                    <div className="divide-y divide-slate-100">
                        {[1, 2, 3].map(i => (
                            <div key={i} className="p-6 flex gap-6 items-center">
                                <div className="skeleton h-5 w-24" />
                                <div className="skeleton h-4 w-48 flex-1" />
                                <div className="skeleton h-4 w-28" />
                                <div className="skeleton h-6 w-20 rounded-full" />
                            </div>
                        ))}
                    </div>
                ) : items.length === 0 ? (
                    <div className="p-12 text-center">
                        <div className="w-14 h-14 bg-slate-100 rounded-2xl flex items-center justify-center mx-auto mb-5">
                            <FileText className="w-7 h-7 text-slate-400" />
                        </div>
                        <h3 className="text-lg font-semibold text-slate-900 mb-1.5">No requests yet</h3>
                        <p className="text-sm text-slate-500 mb-5 max-w-sm mx-auto leading-relaxed">
                            You haven't submitted any data requests. Exercise your rights to access, correct, or erase your data.
                        </p>
                        <button
                            onClick={() => navigate('/requests/new')}
                            className="text-sm font-medium text-blue-600 hover:text-blue-700 transition-colors"
                        >
                            Submit your first request →
                        </button>
                    </div>
                ) : (
                    <div className="overflow-x-auto">
                        <table className="w-full text-left text-sm">
                            <thead>
                                <tr className="border-b border-slate-100 bg-slate-50/80">
                                    <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Type</th>
                                    <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Description</th>
                                    <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Date</th>
                                    <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Status</th>
                                    <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Actions</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-slate-100">
                                {items.map((req) => (
                                    <tr key={req.id} className="hover:bg-slate-50/50 transition-colors">
                                        <td className="px-6 py-4 font-medium text-slate-900 whitespace-nowrap">{req.type}</td>
                                        <td className="px-6 py-4 text-slate-500 max-w-xs truncate">{req.description}</td>
                                        <td className="px-6 py-4 text-slate-500 whitespace-nowrap">
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
                                                <span className="text-xs font-medium text-orange-600 bg-orange-50 px-2.5 py-1 rounded-full border border-orange-100 flex items-center w-fit gap-1">
                                                    <ShieldAlert size={12} />
                                                    Under Appeal
                                                </span>
                                            )}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
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
