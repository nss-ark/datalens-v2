import { cn } from '../lib/utils';
import { Badge } from '../ui/badge';

type BadgeVariant = 'success' | 'warning' | 'danger' | 'info' | 'neutral';

interface StatusBadgeProps {
    label: string;
    variant?: BadgeVariant;
    showDot?: boolean;
    size?: 'sm' | 'md';
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

const VARIANT_STYLES: Record<BadgeVariant, string> = {
    success: "bg-green-100 text-green-800 hover:bg-green-100 border-green-200",
    warning: "bg-yellow-100 text-yellow-800 hover:bg-yellow-100 border-yellow-200",
    danger: "bg-red-100 text-red-800 hover:bg-red-100 border-red-200",
    info: "bg-blue-100 text-blue-800 hover:bg-blue-100 border-blue-200",
    neutral: "bg-gray-100 text-gray-800 hover:bg-gray-100 border-gray-200",
};

const DOT_COLORS: Record<BadgeVariant, string> = {
    success: "bg-green-600",
    warning: "bg-yellow-600",
    danger: "bg-red-600",
    info: "bg-blue-600",
    neutral: "bg-gray-500",
};

export function StatusBadge({ label, variant, showDot = true, size = 'md' }: StatusBadgeProps) {
    const resolvedVariant = variant || STATUS_MAP[label] || 'neutral';
    const colors = VARIANT_STYLES[resolvedVariant];
    const dotColor = DOT_COLORS[resolvedVariant];

    return (
        <Badge
            variant="outline"
            className={cn(
                "font-medium border shadow-sm",
                colors,
                size === 'sm' ? "text-[10px] px-1.5 py-0 h-5" : "text-xs px-2.5 py-0.5"
            )}
        >
            {showDot && (
                <span className={cn("mr-1.5 h-1.5 w-1.5 rounded-full shrink-0", dotColor)} />
            )}
            {label}
        </Badge>
    );
}
