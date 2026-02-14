import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
    ArrowLeft, Save, Loader2, User, Globe
} from 'lucide-react';
import { Button } from '@datalens/shared';
import { useDataSource, useM365Users, useSharePointSites, useUpdateScope } from '../hooks/useDataSources';
import { toast } from '@datalens/shared';
import { GoogleScopeConfig } from '../components/DataSources/GoogleScopeConfig';
import type { M365ScopeConfig, M365User, SharePointSite } from '../types/datasource';

const DataSourceConfig = () => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();

    const { data: dataSource, isLoading: isDsLoading } = useDataSource(id || '');

    // Only fetch M365 data if type is M365
    // Note: React Query hooks will run but enable logic should prevent fetch if id is missing.
    // Ideally we'd pass `enabled: isM365` to the hooks, but for now we'll rely on the component conditional rendering.
    const { data: users = [], isLoading: isUsersLoading } = useM365Users(id || '');
    const { data: sites = [], isLoading: isSitesLoading } = useSharePointSites(id || '');

    const { mutate: updateScope, isPending: isSaving } = useUpdateScope();

    // Local State for M365 toggles
    const [localUsers, setLocalUsers] = useState<M365User[]>([]);
    const [localSites, setLocalSites] = useState<SharePointSite[]>([]);
    const [activeTab, setActiveTab] = useState<'users' | 'sites'>('users');

    // Sync local state with fetched data
    useEffect(() => {
        if (users.length > 0 && localUsers.length === 0) {
            setLocalUsers(users);
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [users]);

    useEffect(() => {
        if (sites.length > 0 && localSites.length === 0) {
            setLocalSites(sites);
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [sites]);

    // Handle Toggles for M365
    const toggleUserScan = (userId: string, field: 'scanOneDrive' | 'scanOutlook') => {
        setLocalUsers(prev => prev.map(u =>
            u.id === userId ? { ...u, [field]: !u[field] } : u
        ));
    };

    const toggleSiteScan = (siteId: string) => {
        setLocalSites(prev => prev.map(s =>
            s.id === siteId ? { ...s, scanDocuments: !s.scanDocuments } : s
        ));
    };

    // Save Changes for M365
    const handleSave = () => {
        if (!id) return;

        const config: M365ScopeConfig = {
            users: localUsers,
            sites: localSites,
            scanAllUsers: false, // Default for now
            scanAllSites: false
        };

        updateScope({ id, config }, {
            onSuccess: () => toast.success('Configuration Saved', 'Scan scope has been updated.'),
            onError: () => toast.error('Save Failed', 'Could not update configuration.')
        });
    };

    if (isDsLoading) {
        return <div className="p-8"><div className="animate-pulse space-y-4">{[1, 2, 3].map(i => <div key={i} className="h-12 bg-gray-100 rounded" />)}</div></div>;
    }

    if (!dataSource) {
        return (
            <div className="p-8 text-center">
                <h2 className="text-xl font-semibold">Data Source Not Found</h2>
                <Button variant="ghost" onClick={() => navigate('/datasources')}>Back</Button>
            </div>
        );
    }

    // Google Workspace Handler
    if (dataSource.type === 'google_workspace') {
        return (
            <div className="p-6 max-w-5xl mx-auto">
                <div className="mb-6">
                    <button
                        onClick={() => navigate(`/datasources/${id}`)}
                        className="flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-900 mb-2 transition-colors"
                    >
                        <ArrowLeft size={16} /> Back to Details
                    </button>
                    <h1 className="text-2xl font-bold text-gray-900">Configure Google Workspace</h1>
                    <p className="text-sm text-gray-500 mt-1">
                        Select which Google services to scan for <strong>{dataSource.name}</strong>
                    </p>
                </div>
                <GoogleScopeConfig dataSource={dataSource} />
            </div>
        );
    }

    // M365 Handler
    if (['onedrive', 'sharepoint', 'outlook', 'm365'].includes(dataSource.type)) {
        return (
            <div className="p-6 max-w-5xl mx-auto">
                {/* Header */}
                <div className="mb-6 flex items-center justify-between">
                    <div>
                        <button
                            onClick={() => navigate(`/datasources/${id}`)}
                            className="flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-900 mb-2 transition-colors"
                        >
                            <ArrowLeft size={16} /> Back to Details
                        </button>
                        <h1 className="text-2xl font-bold text-gray-900">Configure Scan Scope</h1>
                        <p className="text-sm text-gray-500 mt-1">
                            Select which users and sites to include in the scan for <strong>{dataSource.name}</strong>
                        </p>
                    </div>
                    <Button
                        onClick={handleSave}
                        isLoading={isSaving}
                        icon={<Save size={16} />}
                    >
                        Save Changes
                    </Button>
                </div>

                {/* Tabs */}
                <div className="flex gap-4 border-b border-gray-200 mb-6">
                    <TabButton
                        active={activeTab === 'users'}
                        onClick={() => setActiveTab('users')}
                        icon={<User size={16} />}
                        label="M365 Users"
                        count={localUsers.length}
                    />
                    <TabButton
                        active={activeTab === 'sites'}
                        onClick={() => setActiveTab('sites')}
                        icon={<Globe size={16} />}
                        label="SharePoint Sites"
                        count={localSites.length}
                    />
                </div>

                {/* Content */}
                <div className="bg-white border border-gray-200 rounded-xl shadow-sm overflow-hidden">
                    {activeTab === 'users' && (
                        <div className="min-h-[400px]">
                            {isUsersLoading ? (
                                <div className="p-8 flex justify-center"><Loader2 className="animate-spin text-gray-400" /></div>
                            ) : (
                                <table className="w-full text-left border-collapse">
                                    <thead>
                                        <tr className="bg-gray-50 border-b border-gray-100 text-xs text-gray-500 uppercase tracking-wider">
                                            <th className="px-6 py-3 font-medium">User</th>
                                            <th className="px-6 py-3 font-medium text-center w-32">OneDrive</th>
                                            <th className="px-6 py-3 font-medium text-center w-32">Outlook</th>
                                        </tr>
                                    </thead>
                                    <tbody className="divide-y divide-gray-100">
                                        {localUsers.map(user => (
                                            <tr key={user.id} className="hover:bg-gray-50">
                                                <td className="px-6 py-4">
                                                    <div className="font-medium text-gray-900">{user.displayName}</div>
                                                    <div className="text-xs text-gray-500">{user.email}</div>
                                                </td>
                                                <td className="px-6 py-4 text-center">
                                                    <Toggle
                                                        checked={user.scanOneDrive}
                                                        onChange={() => toggleUserScan(user.id, 'scanOneDrive')}
                                                    />
                                                </td>
                                                <td className="px-6 py-4 text-center">
                                                    <Toggle
                                                        checked={user.scanOutlook}
                                                        onChange={() => toggleUserScan(user.id, 'scanOutlook')}
                                                    />
                                                </td>
                                            </tr>
                                        ))}
                                        {localUsers.length === 0 && (
                                            <tr>
                                                <td colSpan={3} className="px-6 py-8 text-center text-gray-500 italic">No users found.</td>
                                            </tr>
                                        )}
                                    </tbody>
                                </table>
                            )}
                        </div>
                    )}

                    {activeTab === 'sites' && (
                        <div className="min-h-[400px]">
                            {isSitesLoading ? (
                                <div className="p-8 flex justify-center"><Loader2 className="animate-spin text-gray-400" /></div>
                            ) : (
                                <table className="w-full text-left border-collapse">
                                    <thead>
                                        <tr className="bg-gray-50 border-b border-gray-100 text-xs text-gray-500 uppercase tracking-wider">
                                            <th className="px-6 py-3 font-medium">Site Name</th>
                                            <th className="px-6 py-3 font-medium">URL</th>
                                            <th className="px-6 py-3 font-medium text-center w-32">Scan Docs</th>
                                        </tr>
                                    </thead>
                                    <tbody className="divide-y divide-gray-100">
                                        {localSites.map(site => (
                                            <tr key={site.id} className="hover:bg-gray-50">
                                                <td className="px-6 py-4 font-medium text-gray-900">{site.name}</td>
                                                <td className="px-6 py-4 text-xs text-blue-600 truncate max-w-xs block">
                                                    <a href={site.url} target="_blank" rel="noopener noreferrer" className="hover:underline">
                                                        {site.url}
                                                    </a>
                                                </td>
                                                <td className="px-6 py-4 text-center">
                                                    <Toggle
                                                        checked={site.scanDocuments}
                                                        onChange={() => toggleSiteScan(site.id)}
                                                    />
                                                </td>
                                            </tr>
                                        ))}
                                        {localSites.length === 0 && (
                                            <tr>
                                                <td colSpan={3} className="px-6 py-8 text-center text-gray-500 italic">No sites found.</td>
                                            </tr>
                                        )}
                                    </tbody>
                                </table>
                            )}
                        </div>
                    )}
                </div>
            </div>
        );
    }

    // Default fallback
    return (
        <div className="p-8 text-center text-gray-500">
            This data source type ({dataSource.type}) does not support scope configuration.
            <div className="mt-4">
                <Button variant="outline" onClick={() => navigate(`/datasources/${id}`)}>Back to Details</Button>
            </div>
        </div>
    );
};

interface TabButtonProps {
    active: boolean;
    onClick: () => void;
    icon: React.ReactNode;
    label: string;
    count: number;
}

const TabButton = ({ active, onClick, icon, label, count }: TabButtonProps) => (
    <button
        onClick={onClick}
        className={`flex items-center gap-2 px-4 py-3 text-sm font-medium border-b-2 transition-colors ${active
            ? 'border-blue-600 text-blue-600'
            : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
    >
        {icon}
        {label}
        <span className="bg-gray-100 text-gray-600 px-2 py-0.5 rounded-full text-xs">{count}</span>
    </button>
);

const Toggle = ({ checked, onChange }: { checked: boolean; onChange: () => void }) => (
    <button
        onClick={onChange}
        className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 ${checked ? 'bg-blue-600' : 'bg-gray-200'
            }`}
    >
        <span
            className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${checked ? 'translate-x-6' : 'translate-x-1'
                }`}
        />
    </button>
);

export default DataSourceConfig;
