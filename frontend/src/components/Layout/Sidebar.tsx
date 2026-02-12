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

    Clock,
    FileOutput,
    Settings,
    Gavel,
    ChevronLeft,
    ChevronRight
} from 'lucide-react';
import { cn } from '../../utils/cn';
import styles from './Sidebar.module.css';
import { useAuthStore } from '../../stores/authStore';
import { useUIStore } from '../../stores/uiStore';

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
            { to: '/lineage', label: 'Data Lineage', icon: GitBranch },
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
            { to: '/consent/widgets', label: 'Consent Widgets', icon: ShieldCheck },
            { to: '/consent', label: 'Consent Records', icon: CheckSquare },
            { to: '/consent/analytics', label: 'Consent Analytics', icon: BarChart3 },
            { to: '/grievances', label: 'Grievances', icon: AlertTriangle },
            { to: '/nominations', label: 'Nominations', icon: Award },
            { to: '/breach', label: 'Breach Incidents', icon: ShieldCheck },
        ]
    },
    {
        title: 'Governance',
        items: [
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
                <button
                    onClick={toggleSidebar}
                    className={styles.collapseBtn}
                    style={{ marginLeft: 'auto', padding: '4px' }}
                >
                    {sidebarCollapsed ? <ChevronRight size={16} /> : <ChevronLeft size={16} />}
                </button>
            </div>
        </aside>
    );
};
