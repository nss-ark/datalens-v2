import { useQuery } from '@tanstack/react-query';
import { portalService } from '../../services/portalService';
import { CheckCircle2, XCircle, Clock, AlertCircle } from 'lucide-react';
import { format } from 'date-fns';

const PortalHistory = () => {
    const { data: history, isLoading } = useQuery({
        queryKey: ['portal-history'],
        queryFn: () => portalService.getHistory(),
    });

    if (isLoading) return <div className="p-8 text-center text-gray-500">Loading history...</div>;

    const items = history?.items || [];

    const getIcon = (status: string) => {
        switch (status) {
            case 'GRANTED': return <CheckCircle2 className="w-5 h-5 text-green-600" />;
            case 'DENIED': return <XCircle className="w-5 h-5 text-red-600" />;
            case 'WITHDRAWN': return <AlertCircle className="w-5 h-5 text-orange-600" />;
            default: return <Clock className="w-5 h-5 text-gray-400" />;
        }
    };

    return (
        <div>
            <h1 className="text-2xl font-bold text-gray-900 mb-6">Consent History</h1>

            <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
                {items.length === 0 ? (
                    <div className="p-8 text-center text-gray-500">No history found.</div>
                ) : (
                    <div className="divide-y divide-gray-200">
                        {items.map((entry) => (
                            <div key={entry.id} className="p-6 flex gap-4">
                                <div className="mt-1">
                                    {getIcon(entry.new_status)}
                                </div>
                                <div className="flex-1">
                                    <div className="flex justify-between items-start">
                                        <div>
                                            <h3 className="font-medium text-gray-900">
                                                {entry.purpose_name}
                                            </h3>
                                            <p className="text-sm text-gray-500 mt-1">
                                                Status change: <span className="font-medium">{entry.previous_status || 'NONE'}</span> → <span className="font-medium">{entry.new_status}</span>
                                            </p>
                                        </div>
                                        <time className="text-sm text-gray-500">
                                            {format(new Date(entry.created_at || ''), 'MMM d, yyyy HH:mm')}
                                        </time>
                                    </div>
                                    <div className="mt-3 flex gap-4 text-xs text-gray-400">
                                        <span>Source: {entry.source}</span>
                                        <span>•</span>
                                        <span>Notice v{entry.notice_version}</span>
                                        <span>•</span>
                                        <span className="font-mono" title={entry.signature}>
                                            Sig: {entry.signature.substring(0, 16)}...
                                        </span>
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
};

export default PortalHistory;
