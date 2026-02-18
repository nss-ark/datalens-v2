import { useUIStore, ActionToolbar } from '@datalens/shared';
import { useNavigate, Outlet } from 'react-router-dom';
import { Database, Plus, ShieldAlert } from 'lucide-react';
import { Sidebar } from './Sidebar';
import { Header } from './Header';

export const AppLayout = () => {
    const navigate = useNavigate();
    const { sidebarCollapsed } = useUIStore();
    const sidebarWidth = sidebarCollapsed ? 'var(--sidebar-collapsed-width)' : 'var(--sidebar-width)';

    const globalActions = [
        {
            id: 'new-datasource',
            label: 'Add Data Source',
            icon: Database,
            onClick: () => navigate('/datasources'),
            variant: 'ghost' as const,
        },
        {
            id: 'new-dsr',
            label: 'New Request',
            icon: Plus,
            onClick: () => navigate('/dsr'), // Assuming DSR creation is here or similar
            variant: 'primary' as const,
        },
        {
            id: 'breach-alert',
            label: 'Report Breach',
            icon: ShieldAlert,
            onClick: () => navigate('/breach/new'),
            variant: 'ghost' as const,
        }
    ];

    return (
        <div style={{ display: 'flex', minHeight: '100vh', backgroundColor: 'var(--bg-app)' }}>
            <Sidebar />
            <div
                style={{
                    marginLeft: sidebarWidth,
                    display: 'flex',
                    flexDirection: 'column',
                    transition: 'margin-left 0.3s ease',
                    width: `calc(100% - ${sidebarWidth})`,
                    minHeight: '100vh'
                }}
            >
                <Header />
                <main style={{ flex: 1, padding: '2rem', overflowY: 'auto', position: 'relative' }}>
                    <Outlet />
                    <ActionToolbar actions={globalActions} />
                </main>
            </div>
        </div>
    );
};
