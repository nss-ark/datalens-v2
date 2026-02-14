import { Clock } from 'lucide-react';
import { cn } from '@datalens/shared';

interface SLATimerProps {
    deadline: string;
    label: string;
    isOverdue?: boolean;
    className?: string;
}

export const SLATimer = ({ label, deadline, className }: SLATimerProps) => {
    // We can rely on the server's pre-calculated "time_remaining" or calculate client-side.
    // Since the server sends deadlines, let's calculate rough hours left for display.

    const deadlineDate = new Date(deadline);
    const now = new Date();
    const diffMs = deadlineDate.getTime() - now.getTime();

    // Format duration
    const isExpired = diffMs < 0;
    const absDiff = Math.abs(diffMs);
    const hours = Math.floor(absDiff / (1000 * 60 * 60));
    const minutes = Math.floor((absDiff % (1000 * 60 * 60)) / (1000 * 60));

    const timeString = `${hours}h ${minutes} m`;
    const displayText = isExpired ? `Overdue by ${timeString} ` : `${timeString} remaining`;

    const isUrgent = !isExpired && hours < 4; // Warning if < 4 hours

    return (
        <div className={cn(
            "flex items-center gap-2 p-3 rounded-lg border",
            isExpired ? "bg-red-50 border-red-200 text-red-700" :
                isUrgent ? "bg-orange-50 border-orange-200 text-orange-700" :
                    "bg-blue-50 border-blue-200 text-blue-700",
            className
        )}>
            <Clock size={16} className={isExpired || isUrgent ? "animate-pulse" : ""} />
            <div className="flex flex-col">
                <span className="text-xs font-semibold uppercase tracking-wider opacity-80">{label}</span>
                <span className="font-mono font-medium">{displayText}</span>
            </div>
        </div>
    );
};
