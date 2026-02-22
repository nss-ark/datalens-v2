import { useUIStore } from '@datalens/shared';
import { Outlet } from 'react-router-dom';
import { Sidebar } from './Sidebar';
import { Header } from './Header';

export const AppLayout = () => {
    const { sidebarCollapsed } = useUIStore();
    const sidebarWidth = sidebarCollapsed ? 'var(--sidebar-collapsed-width)' : 'var(--sidebar-width)';

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
                </main>
            </div>
        </div>
    );
};
