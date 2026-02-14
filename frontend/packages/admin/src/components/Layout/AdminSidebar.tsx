import { NavLink } from 'react-router-dom';
import {
    Users,
    Settings,
    Building,
    Shield,
    ChevronLeft,
    ChevronRight,
} from 'lucide-react';
import { cn } from '@datalens/shared';
import styles from './Sidebar.module.css'; // Reusing Sidebar styles for now, but we can override colors inline or with a new module
import { useAuthStore } from '@datalens/shared';
import { useUIStore } from '@datalens/shared';

const ADMIN_NAV_ITEMS = [
    { to: '/admin/tenants', label: 'Tenants', icon: Building },
    { to: '/admin/users', label: 'Platform Users', icon: Users },
    { to: '/admin/compliance/dsr', label: 'DSR Requests', icon: Shield },
    { to: '/admin/settings', label: 'Settings', icon: Settings },
];

export const AdminSidebar = () => {
    const { sidebarCollapsed, toggleSidebar } = useUIStore();
    const { user } = useAuthStore();

    return (
        <aside className={cn(styles.sidebar, sidebarCollapsed && styles.collapsed)} style={{ backgroundColor: '#1e293b', borderRight: '1px solid #0f172a' }}>
            <div className={styles.logo} style={{ color: 'white' }}>
                <Shield className={styles.logoIcon} style={{ color: '#3b82f6' }} />
                <span>Admin Portal</span>
            </div>

            <nav className={styles.nav}>
                <div className={styles.group}>
                    <div className={styles.groupTitle} style={{ color: '#94a3b8' }}>Platform</div>
                    {ADMIN_NAV_ITEMS.map((item) => (
                        <NavLink
                            key={item.to}
                            to={item.to}
                            className={({ isActive }) => cn(styles.item, isActive && styles.active)}
                            style={({ isActive }) => ({
                                color: isActive ? 'white' : '#cbd5e1',
                                backgroundColor: isActive ? 'rgba(59, 130, 246, 0.2)' : 'transparent'
                            })}
                            title={sidebarCollapsed ? item.label : undefined}
                        >
                            <item.icon className={styles.itemIcon} />
                            {!sidebarCollapsed && <span>{item.label}</span>}
                        </NavLink>
                    ))}
                </div>
            </nav>

            <div className={styles.userProfile} style={{ borderTop: '1px solid #334155' }}>
                <div className={styles.userAvatar} style={{ backgroundColor: '#3b82f6', color: 'white' }}>
                    {user?.name?.charAt(0) || 'A'}
                </div>
                {!sidebarCollapsed && (
                    <div className={styles.userInfo}>
                        <div className={styles.userName} style={{ color: 'white' }}>{user?.name || 'Admin'}</div>
                        <div className={styles.userRole} style={{ color: '#94a3b8' }}>Platform Admin</div>
                    </div>
                )}
                <button
                    onClick={toggleSidebar}
                    className={styles.collapseBtn}
                    style={{ marginLeft: 'auto', padding: '4px', color: 'white' }}
                >
                    {sidebarCollapsed ? <ChevronRight size={16} /> : <ChevronLeft size={16} />}
                </button>
            </div>
        </aside>
    );
};
