import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Play, Pause, Trash2, Code, Copy, Check } from 'lucide-react';
import { useState } from 'react';
import { Button } from '@datalens/shared';
import { StatusBadge } from '@datalens/shared';
import { useWidget, useActivateWidget, usePauseWidget, useDeleteWidget } from '../hooks/useConsent';
import { Loader2 } from 'lucide-react';

export default function WidgetDetail() {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const { data: widget, isLoading, isError } = useWidget(id!);
    const activateMutation = useActivateWidget();
    const pauseMutation = usePauseWidget();
    const deleteMutation = useDeleteWidget();
    const [copied, setCopied] = useState(false);

    if (isLoading) return <div className="flex justify-center p-12"><Loader2 className="animate-spin" /></div>;
    if (isError || !widget) return <div className="p-8 text-center text-red-500">Widget not found</div>;

    const handleToggleStatus = async () => {
        if (widget.status === 'ACTIVE') {
            await pauseMutation.mutateAsync(widget.id);
        } else {
            await activateMutation.mutateAsync(widget.id);
        }
    };

    const handleDelete = async () => {
        if (confirm('Are you sure you want to delete this widget?')) {
            await deleteMutation.mutateAsync(widget.id);
            navigate('/consent/widgets');
        }
    };

    const embedCode = `<script src="https://cdn.datalens.io/consent.js" \n  data-widget-id="${widget.id}" \n  data-api-key="${widget.api_key || 'YOUR_API_KEY'}"></script>`;

    const handleCopy = () => {
        navigator.clipboard.writeText(embedCode);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    return (
        <div className="p-6 space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                    <Button variant="ghost" onClick={() => navigate('/consent/widgets')}>
                        <ArrowLeft size={20} />
                    </Button>
                    <div>
                        <h1 className="text-2xl font-bold flex items-center gap-2">
                            {widget.name}
                            <StatusBadge label={widget.status} />
                        </h1>
                        <p className="text-gray-500">{widget.domain} • {widget.type.replace('_', ' ')}</p>
                    </div>
                </div>
                <div className="flex gap-2">
                    <Button
                        variant="secondary"
                        onClick={handleToggleStatus}
                        icon={widget.status === 'ACTIVE' ? <Pause size={16} /> : <Play size={16} />}
                    >
                        {widget.status === 'ACTIVE' ? 'Pause' : 'Activate'}
                    </Button>
                    <Button
                        variant="ghost"
                        className="text-red-600 hover:bg-red-50"
                        onClick={handleDelete}
                        icon={<Trash2 size={16} />}
                    >
                        Delete
                    </Button>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Main Content - Embed Code & Preview */}
                <div className="lg:col-span-2 space-y-6">
                    <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-100">
                        <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
                            <Code size={20} className="text-blue-500" />
                            Embed Code
                        </h2>
                        <p className="text-sm text-gray-600 mb-4">
                            Copy and paste this snippet into the <code>&lt;head&gt;</code> of your website.
                        </p>
                        <div className="relative bg-gray-900 rounded-md p-4 overflow-x-auto group">
                            <pre className="text-gray-100 font-mono text-sm whitespace-pre-wrap break-all">
                                {embedCode}
                            </pre>
                            <button
                                onClick={handleCopy}
                                className="absolute top-2 right-2 p-2 bg-gray-700 rounded text-gray-300 hover:text-white opacity-0 group-hover:opacity-100 transition-opacity"
                            >
                                {copied ? <Check size={16} /> : <Copy size={16} />}
                            </button>
                        </div>
                    </div>

                    <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-100">
                        <h2 className="text-lg font-semibold mb-4">Configuration Summary</h2>
                        <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-4">
                            <div>
                                <dt className="text-sm font-medium text-gray-500">Layout</dt>
                                <dd className="text-sm text-gray-900">{widget.config.layout}</dd>
                            </div>
                            <div>
                                <dt className="text-sm font-medium text-gray-500">Default State</dt>
                                <dd className="text-sm text-gray-900">{widget.config.default_state}</dd>
                            </div>
                            <div>
                                <dt className="text-sm font-medium text-gray-500">Languages</dt>
                                <dd className="text-sm text-gray-900">{widget.config.languages.join(', ')}</dd>
                            </div>
                            <div>
                                <dt className="text-sm font-medium text-gray-500">Purposes</dt>
                                <dd className="text-sm text-gray-900">{widget.config.purpose_ids.length} purposes selected</dd>
                            </div>
                        </dl>
                    </div>
                </div>

                {/* Sidebar - Analytics Placeholder */}
                <div className="space-y-6">
                    <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-100">
                        <h2 className="text-lg font-semibold mb-4">Consent Analytics</h2>
                        <div className="space-y-4">
                            <div className="p-4 bg-gray-50 rounded-lg text-center">
                                <div className="text-2xl font-bold text-gray-900">0</div>
                                <div className="text-xs text-gray-500 uppercase tracking-wide">Total Consents</div>
                            </div>
                            <div className="p-4 bg-gray-50 rounded-lg text-center">
                                <div className="text-2xl font-bold text-gray-900">0%</div>
                                <div className="text-xs text-gray-500 uppercase tracking-wide">Opt-in Rate</div>
                            </div>
                            <p className="text-xs text-center text-gray-400">
                                Analytics data will appear here once the widget is active.
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
