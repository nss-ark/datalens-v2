import { useQuery } from '@tanstack/react-query';
import { portalService } from '@/services/portalService';
import { CheckCircle2, XCircle, Clock, AlertCircle, ScrollText } from 'lucide-react';
import { format } from 'date-fns';
import { MotionList01, MotionItem } from '@datalens/shared';

const PortalHistory = () => {
    const { data: history, isLoading } = useQuery({
        queryKey: ['portal-history'],
        queryFn: () => portalService.getHistory(),
    });

    const items = history?.items || [];

    const getIcon = (status: string) => {
        switch (status) {
            case 'GRANTED': return <CheckCircle2 className="w-5 h-5 text-emerald-600" />;
            case 'DENIED': return <XCircle className="w-5 h-5 text-red-500" />;
            case 'WITHDRAWN': return <AlertCircle className="w-5 h-5 text-orange-500" />;
            default: return <Clock className="w-5 h-5 text-slate-400" />;
        }
    };

    const getStatusColor = (status: string) => {
        switch (status) {
            case 'GRANTED': return 'bg-emerald-50 text-emerald-700 ring-emerald-200';
            case 'DENIED': return 'bg-red-50 text-red-700 ring-red-200';
            case 'WITHDRAWN': return 'bg-orange-50 text-orange-700 ring-orange-200';
            default: return 'bg-slate-50 text-slate-600 ring-slate-200';
        }
    };

    return (
        <div className="animate-fade-in">
            <div className="page-header">
                <h1>Consent History</h1>
                <p>Track all consent changes across your account over time.</p>
            </div>

            {isLoading ? (
                <div className="portal-card overflow-hidden">
                    {[1, 2, 3, 4].map(i => (
                        <div key={i} className="p-6 flex gap-4 border-b border-slate-100 last:border-0">
                            <div className="skeleton w-10 h-10 rounded-full flex-shrink-0" />
                            <div className="flex-1 space-y-2">
                                <div className="skeleton h-5 w-48" />
                                <div className="skeleton h-4 w-72" />
                                <div className="skeleton h-3 w-40" />
                            </div>
                        </div>
                    ))}
                </div>
            ) : items.length === 0 ? (
                <div className="portal-card p-12 text-center">
                    <div className="w-14 h-14 bg-slate-100 rounded-2xl flex items-center justify-center mx-auto mb-5">
                        <ScrollText className="w-7 h-7 text-slate-400" />
                    </div>
                    <h3 className="text-lg font-semibold text-slate-900 mb-1.5">No consent history</h3>
                    <p className="text-sm text-slate-500 max-w-sm mx-auto leading-relaxed">
                        Your consent changes will appear here as you grant or withdraw permissions.
                    </p>
                </div>
            ) : (
                <MotionList01 className="portal-card overflow-hidden divide-y divide-slate-100">
                    {items.map((entry, index) => (
                        <MotionItem key={entry.id} index={index}>
                            <div className="p-6 flex gap-4 hover:bg-slate-50/50 transition-colors">
                                <div className="mt-0.5 flex-shrink-0">
                                    <div className="w-10 h-10 rounded-full bg-slate-50 border border-slate-100 flex items-center justify-center">
                                        {getIcon(entry.new_status)}
                                    </div>
                                </div>
                                <div className="flex-1 min-w-0">
                                    <div className="flex flex-col sm:flex-row sm:justify-between sm:items-start gap-2">
                                        <div>
                                            <h3 className="font-semibold text-slate-900 text-sm">
                                                {entry.purpose_name}
                                            </h3>
                                            <p className="text-sm text-slate-500 mt-1 flex items-center gap-2 flex-wrap">
                                                <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ring-1 ${getStatusColor(entry.previous_status || 'NONE')}`}>
                                                    {entry.previous_status || 'NONE'}
                                                </span>
                                                <span className="text-slate-300">→</span>
                                                <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ring-1 ${getStatusColor(entry.new_status)}`}>
                                                    {entry.new_status}
                                                </span>
                                            </p>
                                        </div>
                                        <time className="text-xs text-slate-400 whitespace-nowrap">
                                            {format(new Date(entry.created_at || ''), 'MMM d, yyyy HH:mm')}
                                        </time>
                                    </div>
                                    <div className="mt-3 flex gap-3 text-xs text-slate-400">
                                        <span>Source: {entry.source}</span>
                                        <span className="text-slate-200">•</span>
                                        <span>Notice v{entry.notice_version}</span>
                                        <span className="text-slate-200">•</span>
                                        <span className="font-mono" title={entry.signature}>
                                            Sig: {entry.signature.substring(0, 16)}…
                                        </span>
                                    </div>
                                </div>
                            </div>
                        </MotionItem>
                    ))}
                </MotionList01>
            )}
        </div>
    );
};

export default PortalHistory;
