import { cn } from '../../utils/cn';
import styles from './StatusBadge.module.css';

type BadgeVariant = 'success' | 'warning' | 'danger' | 'info' | 'neutral';

interface StatusBadgeProps {
    label: string;
    variant?: BadgeVariant;
    showDot?: boolean;
}

const STATUS_MAP: Record<string, BadgeVariant> = {
    CONNECTED: 'success',
    ACTIVE: 'success',
    COMPLETED: 'success',
    VERIFIED: 'success',
    RUNNING: 'info',
    SCANNING: 'info',
    TESTING: 'info',
    PENDING: 'warning',
    INVITED: 'warning',
    DISCONNECTED: 'neutral',
    ERROR: 'danger',
    FAILED: 'danger',
    SUSPENDED: 'danger',
    CANCELLED: 'neutral',
};

export function StatusBadge({ label, variant, showDot = true }: StatusBadgeProps) {
    const resolvedVariant = variant || STATUS_MAP[label] || 'neutral';

    return (
        <span className={cn(styles.badge, styles[resolvedVariant])}>
            {showDot && <span className={styles.dot} />}
            {label}
        </span>
    );
}
