import { NavLink } from 'react-router-dom';
import {
    LayoutDashboard,
    Search,
    Database,
    FileSearch,
    CheckSquare,
    GitBranch,
    Users,
    FileText,
    ShieldCheck,
    BarChart3,
    AlertTriangle,
    Award,
    Briefcase,
    Building2,
    Globe,
    Clock,
    FileOutput,
    Settings,
    Gavel,
    ChevronLeft,
    ChevronRight,
    Bell,
    LogOut
} from 'lucide-react';
import { cn } from '@datalens/shared';
import styles from './Sidebar.module.css';
import { useAuthStore } from '@datalens/shared';
import { useUIStore } from '@datalens/shared';

const NAV_GROUPS = [
    {
        title: 'Overview',
        items: [
            { to: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
        ]
    },
    {
        title: 'Discovery',
        items: [
            { to: '/agents', label: 'Agents', icon: Search },
            { to: '/datasources', label: 'Data Sources', icon: Database },
            { to: '/pii/inventory', label: 'PII Inventory', icon: FileSearch },
            { to: '/pii/review', label: 'Review Queue', icon: CheckSquare },
            { to: '/governance/lineage', label: 'Data Lineage', icon: GitBranch },
        ]
    },
    {
        title: 'Subjects',
        items: [
            { to: '/subjects', label: 'Data Subjects', icon: Users },
        ]
    },
    {
        title: 'Compliance',
        items: [
            { to: '/dsr', label: 'DSR Requests', icon: FileText },
            { to: '/consent/notices', label: 'Privacy Notices', icon: FileText },
            { to: '/consent/widgets', label: 'Consent Widgets', icon: ShieldCheck },
            { to: '/consent', label: 'Consent Records', icon: CheckSquare },
            { to: '/compliance/analytics', label: 'Consent Analytics', icon: BarChart3 },
            { to: '/compliance/lab', label: 'Dark Pattern Lab', icon: ShieldCheck },
            { to: '/compliance/notifications', label: 'Notification History', icon: Bell },
            { to: '/compliance/grievances', label: 'Grievances', icon: AlertTriangle },
            { to: '/nominations', label: 'Nominations', icon: Award },
            { to: '/breach', label: 'Breach Management', icon: ShieldCheck },
            { to: '/compliance/settings/identity', label: 'Identity Verification', icon: ShieldCheck },
        ]
    },
    {
        title: 'Governance',
        items: [
            { to: '/departments', label: 'Departments', icon: Building2 },
            { to: '/third-parties', label: 'Third Parties', icon: Globe },
            { to: '/governance/purposes', label: 'Purpose Mapping', icon: Briefcase },
            { to: '/governance/policies', label: 'Policy Manager', icon: Building2 },
            { to: '/governance/violations', label: 'Compliance Issues', icon: AlertTriangle },
            { to: '/retention', label: 'Retention Policies', icon: Clock },
        ]
    },
    {
        title: 'Reporting',
        items: [
            { to: '/ropa', label: 'RoPA', icon: FileOutput },
            { to: '/reports', label: 'Reports', icon: FileText },
            { to: '/audit-logs', label: 'Audit Logs', icon: CheckSquare },
        ]
    },
    {
        title: 'Settings',
        items: [
            { to: '/users', label: 'User Management', icon: Users },
            { to: '/settings', label: 'General Settings', icon: Settings },
        ]
    }
];

export const Sidebar = () => {
    const { sidebarCollapsed, toggleSidebar } = useUIStore();
    const { user } = useAuthStore();

    return (
        <aside className={cn(styles.sidebar, sidebarCollapsed && styles.collapsed)}>
            <div className={styles.logo}>
                <Gavel className={styles.logoIcon} />
                <span>DataLens</span>
            </div>

            <nav className={styles.nav}>
                {NAV_GROUPS.map((group, groupIndex) => (
                    <div key={groupIndex} className={styles.group}>
                        <div className={styles.groupTitle}>{group.title}</div>
                        {group.items.map((item) => (
                            <NavLink
                                key={item.to}
                                to={item.to}
                                className={({ isActive }) => cn(styles.item, isActive && styles.active)}
                                title={sidebarCollapsed ? item.label : undefined}
                            >
                                <item.icon className={styles.itemIcon} />
                                {!sidebarCollapsed && <span>{item.label}</span>}
                            </NavLink>
                        ))}
                    </div>
                ))}
            </nav>

            {/* Admin Portal Switcher - FORCE VISIBLE */}
            <div style={{ padding: '0 12px 12px 12px' }}>
                <NavLink
                    to="/admin"
                    className={cn(styles.item)}
                    style={{
                        backgroundColor: '#eff6ff',
                        color: '#1d4ed8',
                        border: '1px dashed #bfdbfe',
                        justifyContent: sidebarCollapsed ? 'center' : 'flex-start'
                    }}
                    title="Switch to Admin Portal"
                >
                    <ShieldCheck className={styles.itemIcon} size={20} />
                    {!sidebarCollapsed && <span style={{ fontWeight: 600 }}>Admin Portal</span>}
                </NavLink>
            </div>


            <div className={styles.userProfile}>
                <div className={styles.userAvatar}>
                    {user?.name?.charAt(0) || 'U'}
                </div>
                {!sidebarCollapsed && (
                    <div className={styles.userInfo}>
                        <div className={styles.userName}>{user?.name || 'User'}</div>
                        <div className={styles.userRole}>{user?.role_ids?.[0] || 'Viewer'}</div>
                    </div>
                )}

                <div style={{ marginLeft: 'auto', display: 'flex', gap: '4px' }}>
                    <button
                        onClick={() => {
                            useAuthStore.getState().logout();
                            window.location.href = '/login';
                        }}
                        className={styles.collapseBtn}
                        style={{ padding: '4px' }}
                        title="Logout"
                    >
                        <LogOut size={16} />
                    </button>
                    <button
                        onClick={toggleSidebar}
                        className={styles.collapseBtn}
                        style={{ padding: '4px' }}
                    >
                        {sidebarCollapsed ? <ChevronRight size={16} /> : <ChevronLeft size={16} />}
                    </button>
                </div>
            </div>
        </aside>
    );
};
