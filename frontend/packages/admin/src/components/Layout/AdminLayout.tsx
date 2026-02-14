import { Outlet } from 'react-router-dom';
import { AdminSidebar } from './AdminSidebar';
import { Header } from './Header';
import { useUIStore } from '@datalens/shared';

export const AdminLayout = () => {
    const { sidebarCollapsed } = useUIStore();
    const sidebarWidth = sidebarCollapsed ? 'var(--sidebar-collapsed-width)' : 'var(--sidebar-width)';

    return (
        <div style={{ display: 'flex', minHeight: '100vh', backgroundColor: '#f1f5f9' }}>
            <AdminSidebar />
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
                <main style={{ flex: 1, padding: '2rem', overflowY: 'auto' }}>
                    <Outlet />
                </main>
            </div>
        </div>
    );
};
