import { useQuery } from '@tanstack/react-query';
import { portalService } from '@/services/portalService';
import { ShieldAlert, AlertTriangle, AlertOctagon, Info, ChevronDown, ChevronUp } from 'lucide-react';
import { format } from 'date-fns';
import { useState } from 'react';
import { clsx } from 'clsx';
import type { BreachNotification } from '@/types/portal';

const SeverityBadge = ({ severity }: { severity: string }) => {
    const styles = {
        LOW: 'bg-blue-100 text-blue-800 border-blue-200',
        MEDIUM: 'bg-yellow-100 text-yellow-800 border-yellow-200',
        HIGH: 'bg-orange-100 text-orange-800 border-orange-200',
        CRITICAL: 'bg-red-100 text-red-800 border-red-200',
    };

    const icons = {
        LOW: Info,
        MEDIUM: AlertTriangle,
        HIGH: AlertOctagon,
        CRITICAL: ShieldAlert,
    };

    // @ts-ignore
    const Icon = icons[severity] || Info;
    // @ts-ignore
    const style = styles[severity] || styles.LOW;

    return (
        <span className={clsx('flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium border', style)}>
            <Icon size={14} />
            {severity}
        </span>
    );
};

const NotificationCard = ({ notification }: { notification: BreachNotification }) => {
    const [expanded, setExpanded] = useState(false);

    return (
        <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden transition-all hover:shadow-md">
            <div className="p-6 cursor-pointer" onClick={() => setExpanded(!expanded)}>
                <div className="flex justify-between items-start gap-4">
                    <div className="flex-1">
                        <div className="flex items-center gap-3 mb-2">
                            <SeverityBadge severity={notification.severity} />
                            <span className="text-sm text-gray-500">
                                {format(new Date(notification.created_at), 'MMM d, yyyy')}
                            </span>
                        </div>
                        <h3 className="text-lg font-semibold text-gray-900 mb-1">{notification.title}</h3>
                        <p className="text-gray-600 line-clamp-2">{notification.description}</p>
                    </div>
                    <button
                        className="text-gray-400 hover:text-gray-600 transition-colors p-1"
                        aria-label={expanded ? "Collapse" : "Expand"}
                    >
                        {expanded ? <ChevronUp size={20} /> : <ChevronDown size={20} />}
                    </button>
                </div>
            </div>

            {expanded && (
                <div className="px-6 pb-6 bg-gray-50 border-t border-gray-100 pt-4">
                    <div className="grid md:grid-cols-2 gap-6">
                        <div>
                            <h4 className="text-sm font-semibold text-gray-900 mb-2">Incident Details</h4>
                            <div className="space-y-3 text-sm">
                                <div>
                                    <span className="text-gray-500 block text-xs uppercase tracking-wider">Occurred At</span>
                                    <span className="text-gray-900 font-medium">
                                        {format(new Date(notification.occurred_at), 'MMM d, yyyy HH:mm')}
                                    </span>
                                </div>
                                <div>
                                    <span className="text-gray-500 block text-xs uppercase tracking-wider">Description</span>
                                    <p className="text-gray-700 mt-1 leading-relaxed">{notification.description}</p>
                                </div>
                            </div>
                        </div>

                        <div className="space-y-6">
                            <div>
                                <h4 className="text-sm font-semibold text-gray-900 mb-2">Affected Data</h4>
                                <div className="flex flex-wrap gap-2">
                                    {notification.affected_data?.map((item, i) => (
                                        <span key={i} className="px-2 py-1 bg-white border border-gray-200 rounded text-xs text-gray-700">
                                            {item}
                                        </span>
                                    ))}
                                </div>
                            </div>

                            <div>
                                <h4 className="text-sm font-semibold text-gray-900 mb-2">What We Are Doing</h4>
                                <p className="text-sm text-gray-700 leading-relaxed bg-white p-3 rounded border border-gray-200">
                                    {notification.what_we_are_doing}
                                </p>
                            </div>

                            {notification.contact_email && (
                                <div>
                                    <h4 className="text-sm font-semibold text-gray-900 mb-2">Contact</h4>
                                    <a href={`mailto:${notification.contact_email}`} className="text-sm text-blue-600 hover:underline">
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
        <div className="max-w-4xl mx-auto py-8">
            <div className="mb-8">
                <h1 className="text-2xl font-bold text-gray-900 mb-2">Data Breach Notifications</h1>
                <p className="text-gray-600">
                    Important alerts regarding processed data breaches that may impact you.
                    Under DPDP Rules R7(4), we are legally required to notify you of any personal data breaches.
                </p>
            </div>

            {isLoading ? (
                <div className="space-y-4">
                    {[1, 2, 3].map((i) => (
                        <div key={i} className="h-32 bg-gray-100 rounded-xl animate-pulse" />
                    ))}
                </div>
            ) : items.length === 0 ? (
                <div className="bg-white rounded-xl border border-gray-200 p-12 text-center">
                    <div className="w-16 h-16 bg-green-100 text-green-600 rounded-full flex items-center justify-center mx-auto mb-4">
                        <ShieldAlert size={32} />
                    </div>
                    <h3 className="text-lg font-semibold text-gray-900 mb-2">No Breach Notifications</h3>
                    <p className="text-gray-500">
                        Good news! There are no data breaches reported for your account.
                    </p>
                </div>
            ) : (
                <div className="space-y-4">
                    {items.map((notification) => (
                        <NotificationCard key={notification.id} notification={notification} />
                    ))}
                </div>
            )}
        </div>
    );
};

export default BreachNotifications;
