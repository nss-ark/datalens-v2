import { useQuery } from '@tanstack/react-query';
import { portalService } from '@/services/portalService';
import { ShieldAlert, AlertTriangle, AlertOctagon, Info, ChevronDown, ChevronUp } from 'lucide-react';
import { format } from 'date-fns';
import { useState } from 'react';
import { clsx } from 'clsx';
import type { BreachNotification } from '@/types/portal';

const SeverityBadge = ({ severity }: { severity: string }) => {
    const styles: Record<string, string> = {
        LOW: 'bg-blue-50 text-blue-700 ring-blue-200',
        MEDIUM: 'bg-yellow-50 text-yellow-700 ring-yellow-200',
        HIGH: 'bg-orange-50 text-orange-700 ring-orange-200',
        CRITICAL: 'bg-red-50 text-red-700 ring-red-200',
    };

    const icons: Record<string, typeof Info> = {
        LOW: Info,
        MEDIUM: AlertTriangle,
        HIGH: AlertOctagon,
        CRITICAL: ShieldAlert,
    };

    const Icon = icons[severity] || Info;
    const style = styles[severity] || styles.LOW;

    return (
        <span className={clsx('flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold ring-1', style)}>
            <Icon size={13} />
            {severity}
        </span>
    );
};

const NotificationCard = ({ notification }: { notification: BreachNotification }) => {
    const [expanded, setExpanded] = useState(false);

    return (
        <div className="portal-card overflow-hidden transition-all">
            <div className="p-6 cursor-pointer hover:bg-slate-50/50 transition-colors" onClick={() => setExpanded(!expanded)}>
                <div className="flex justify-between items-start gap-4">
                    <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-3 mb-2.5">
                            <SeverityBadge severity={notification.severity} />
                            <span className="text-xs text-slate-400">
                                {format(new Date(notification.created_at), 'MMM d, yyyy')}
                            </span>
                        </div>
                        <h3 className="text-base font-semibold text-slate-900 mb-1">{notification.title}</h3>
                        <p className="text-sm text-slate-500 line-clamp-2 leading-relaxed">{notification.description}</p>
                    </div>
                    <button
                        className="text-slate-400 hover:text-slate-600 transition-colors p-1.5 rounded-lg hover:bg-slate-100 flex-shrink-0"
                        aria-label={expanded ? "Collapse" : "Expand"}
                    >
                        {expanded ? <ChevronUp size={18} /> : <ChevronDown size={18} />}
                    </button>
                </div>
            </div>

            {expanded && (
                <div className="px-6 pb-6 bg-slate-50/50 border-t border-slate-100 pt-5 animate-fade-in">
                    <div className="grid md:grid-cols-2 gap-6">
                        <div>
                            <h4 className="text-xs font-semibold text-slate-900 mb-3 uppercase tracking-wider">Incident Details</h4>
                            <div className="space-y-4 text-sm">
                                <div>
                                    <span className="text-slate-400 block text-xs uppercase tracking-wider font-semibold mb-1">Occurred At</span>
                                    <span className="text-slate-900 font-medium">
                                        {format(new Date(notification.occurred_at), 'MMM d, yyyy HH:mm')}
                                    </span>
                                </div>
                                <div>
                                    <span className="text-slate-400 block text-xs uppercase tracking-wider font-semibold mb-1">Description</span>
                                    <p className="text-slate-600 leading-relaxed">{notification.description}</p>
                                </div>
                            </div>
                        </div>

                        <div className="space-y-5">
                            <div>
                                <h4 className="text-xs font-semibold text-slate-900 mb-3 uppercase tracking-wider">Affected Data</h4>
                                <div className="flex flex-wrap gap-2">
                                    {notification.affected_data?.map((item, i) => (
                                        <span key={i} className="px-2.5 py-1 bg-white border border-slate-200 rounded-lg text-xs text-slate-700 font-medium">
                                            {item}
                                        </span>
                                    ))}
                                </div>
                            </div>

                            <div>
                                <h4 className="text-xs font-semibold text-slate-900 mb-3 uppercase tracking-wider">What We Are Doing</h4>
                                <p className="text-sm text-slate-600 leading-relaxed bg-white p-4 rounded-xl border border-slate-200">
                                    {notification.what_we_are_doing}
                                </p>
                            </div>

                            {notification.contact_email && (
                                <div>
                                    <h4 className="text-xs font-semibold text-slate-900 mb-2 uppercase tracking-wider">Contact</h4>
                                    <a href={`mailto:${notification.contact_email}`} className="text-sm text-blue-600 hover:text-blue-700 font-medium transition-colors">
                                        {notification.contact_email}
                                    </a>
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

const BreachNotifications = () => {
    const { data: notifications, isLoading } = useQuery({
        queryKey: ['breach-notifications'],
        queryFn: () => portalService.getBreachNotifications(),
    });

    const items = notifications?.items || [];

    return (
        <div className="animate-fade-in">
            <div className="page-header">
                <h1>Data Breach Notifications</h1>
                <p>
                    Important alerts regarding processed data breaches that may impact you.
                    Under DPDP Rules R7(4), we are legally required to notify you of any personal data breaches.
                </p>
            </div>

            {isLoading ? (
                <div className="space-y-4">
                    {[1, 2, 3].map((i) => (
                        <div key={i} className="portal-card p-6">
                            <div className="flex gap-3 mb-3">
                                <div className="skeleton h-6 w-20 rounded-full" />
                                <div className="skeleton h-4 w-24" />
                            </div>
                            <div className="skeleton h-5 w-72 mb-2" />
                            <div className="skeleton h-4 w-full" />
                        </div>
                    ))}
                </div>
            ) : items.length === 0 ? (
                <div className="portal-card p-12 text-center">
                    <div className="w-16 h-16 bg-emerald-50 text-emerald-600 rounded-2xl flex items-center justify-center mx-auto mb-5 ring-4 ring-emerald-50">
                        <ShieldAlert size={28} />
                    </div>
                    <h3 className="text-lg font-bold text-slate-900 mb-2">No Breach Notifications</h3>
                    <p className="text-sm text-slate-500 max-w-sm mx-auto leading-relaxed">
                        Good news! There are no data breaches reported for your account. We will notify you immediately if any incidents occur.
                    </p>
                </div>
            ) : (
                <div className="space-y-4 stagger-children">
                    {items.map((notification) => (
                        <NotificationCard key={notification.id} notification={notification} />
                    ))}
                </div>
            )}
        </div>
    );
};

export default BreachNotifications;
