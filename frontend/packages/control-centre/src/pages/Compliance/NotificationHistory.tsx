import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { DataTable } from '@datalens/shared';
import { Button } from '@datalens/shared';
import { Modal } from '@datalens/shared';
import { StatusBadge } from '@datalens/shared';
import { notificationService } from '../../services/notificationService';
import { format } from 'date-fns';
import { Eye, RotateCcw } from 'lucide-react';
import type { ConsentNotification, NotificationStatus, NotificationChannel } from '../../types/notification';

export default function NotificationHistory() {
    const [page] = useState(1); // TODO: Add pagination controls
    const [selectedNotification, setSelectedNotification] = useState<ConsentNotification | null>(null);
    const [filters, setFilters] = useState<{
        status?: NotificationStatus;
        channel?: NotificationChannel;
        event_type?: string;
    }>({});

    const { data, isLoading, refetch } = useQuery({
        queryKey: ['notifications', page, filters],
        queryFn: () => notificationService.listNotifications({
            page,
            page_size: 10,
            ...filters
        })
    });

    const columns = [
        {
            key: 'event_type',
            header: 'Event',
            render: (row: ConsentNotification) => <span className="font-medium text-gray-900">{row.event_type}</span>
        },
        {
            key: 'recipient',
            header: 'Recipient',
            render: (row: ConsentNotification) => (
                <div>
                    <div className="text-gray-900">{row.recipient_id}</div>
                    <div className="text-xs text-gray-500">{row.recipient_type}</div>
                </div>
            )
        },
        {
            key: 'channel',
            header: 'Channel',
            render: (row: ConsentNotification) => <div className="text-xs uppercase font-bold text-gray-500">{row.channel}</div>
        },
        {
            key: 'status',
            header: 'Status',
            render: (row: ConsentNotification) => <StatusBadge label={row.status} />
        },
        {
            key: 'sent_at',
            header: 'Time',
            render: (row: ConsentNotification) => (
                <div className="text-sm text-gray-600">
                    {row.sent_at ? format(new Date(row.sent_at), 'MMM d, HH:mm') : '-'}
                </div>
            )
        },
        {
            key: 'actions',
            header: 'Actions',
            render: (row: ConsentNotification) => (
                <div className="flex space-x-2">
                    <Button
                        size="sm"
                        variant="secondary"
                        icon={<Eye size={14} />}
                        onClick={() => setSelectedNotification(row)}
                        title="View Payload"
                    />
                    {row.status === 'FAILED' && (
                        <Button
                            size="sm"
                            variant="secondary"
                            icon={<RotateCcw size={14} />}
                            onClick={async () => {
                                await notificationService.resendNotification(row.id);
                                refetch();
                            }}
                            title="Resend"
                        />
                    )}
                </div>
            )
        }
    ];

    return (
        <div className="p-6">
            <div className="flex justify-between items-center mb-6">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900">Notification History</h1>
                    <p className="text-gray-500">Audit log of all consent-related notifications sent to users.</p>
                </div>
            </div>

            <div className="flex space-x-4 mb-4">
                <select
                    className="border rounded px-3 py-2 text-sm"
                    value={filters.channel || ''}
                    onChange={e => setFilters(prev => ({ ...prev, channel: (e.target.value as NotificationChannel) || undefined }))}
                >
                    <option value="">All Channels</option>
                    <option value="EMAIL">Email</option>
                    <option value="SMS">SMS</option>
                    <option value="WEBHOOK">Webhook</option>
                </select>

                <select
                    className="border rounded px-3 py-2 text-sm"
                    value={filters.status || ''}
                    onChange={e => setFilters(prev => ({ ...prev, status: (e.target.value as NotificationStatus) || undefined }))}
                >
                    <option value="">All Statuses</option>
                    <option value="SENT">Sent</option>
                    <option value="DELIVERED">Delivered</option>
                    <option value="FAILED">Failed</option>
                </select>
            </div>

            <DataTable
                columns={columns}
                data={data?.items || []}
                isLoading={isLoading}
                keyExtractor={(row) => row.id}
                emptyTitle="No notifications found"
                emptyDescription="No notifications match your filters."
            />

            {/* Payload Viewer Modal */}
            <Modal
                open={!!selectedNotification}
                onClose={() => setSelectedNotification(null)}
                title="Notification Details"
                size="lg"
            >
                {selectedNotification && (
                    <div className="space-y-4">
                        <div className="grid grid-cols-2 gap-4 text-sm">
                            <div>
                                <label className="block text-gray-500 text-xs uppercase">Event</label>
                                <div className="font-medium">{selectedNotification.event_type}</div>
                            </div>
                            <div>
                                <label className="block text-gray-500 text-xs uppercase">Channel</label>
                                <div className="font-medium">{selectedNotification.channel}</div>
                            </div>
                            <div>
                                <label className="block text-gray-500 text-xs uppercase">Recipient</label>
                                <div className="font-medium">{selectedNotification.recipient_id}</div>
                            </div>
                            <div>
                                <label className="block text-gray-500 text-xs uppercase">Sent At</label>
                                <div className="font-medium">
                                    {selectedNotification.sent_at ? format(new Date(selectedNotification.sent_at), 'PPpp') : '-'}
                                </div>
                            </div>
                        </div>

                        {selectedNotification.failure_reason && (
                            <div className="bg-red-50 p-3 rounded border border-red-200 text-red-700 text-sm">
                                <strong>Failure Reason:</strong> {selectedNotification.failure_reason}
                            </div>
                        )}

                        <div>
                            <label className="block text-gray-500 text-xs uppercase mb-1">Payload</label>
                            <pre className="bg-gray-900 text-gray-100 p-4 rounded overflow-x-auto text-xs font-mono">
                                {JSON.stringify(selectedNotification.payload, null, 2)}
                            </pre>
                        </div>
                    </div>
                )}
            </Modal>
        </div>
    );
}
