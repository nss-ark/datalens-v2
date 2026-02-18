import { cn } from '../../lib/utils';
import { ArrowDown, ArrowUp, Minus } from 'lucide-react';

interface AnalyticsCardProps {
    title: string;
    value: string | number;
    trend?: {
        value: number;
        label?: string;
        direction: 'up' | 'down' | 'neutral';
    };
    icon?: React.ReactNode;
    className?: string;
    description?: string;
    sparklineData?: number[];
}

function Sparkline({ data, className }: { data: number[]; className?: string }) {
    if (!data || data.length < 2) return null;

    const width = 100;
    const height = 32;
    const padding = 2;

    const min = Math.min(...data);
    const max = Math.max(...data);
    const range = max - min || 1;

    const points = data
        .map((val, i) => {
            const x = padding + (i / (data.length - 1)) * (width - padding * 2);
            const y = height - padding - ((val - min) / range) * (height - padding * 2);
            return `${x},${y}`;
        })
        .join(' ');

    // Build area fill path (polygon from line to bottom)
    const areaPoints = [
        `${padding},${height - padding}`,
        ...data.map((val, i) => {
            const x = padding + (i / (data.length - 1)) * (width - padding * 2);
            const y = height - padding - ((val - min) / range) * (height - padding * 2);
            return `${x},${y}`;
        }),
        `${width - padding},${height - padding}`,
    ].join(' ');

    return (
        <svg
            viewBox={`0 0 ${width} ${height}`}
            className={cn("w-full h-8", className)}
            preserveAspectRatio="none"
        >
            <defs>
                <linearGradient id="sparkline-fill" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="currentColor" stopOpacity="0.2" />
                    <stop offset="100%" stopColor="currentColor" stopOpacity="0.02" />
                </linearGradient>
            </defs>
            <polygon
                points={areaPoints}
                fill="url(#sparkline-fill)"
            />
            <polyline
                points={points}
                fill="none"
                stroke="currentColor"
                strokeWidth="1.5"
                strokeLinecap="round"
                strokeLinejoin="round"
                vectorEffect="non-scaling-stroke"
            />
        </svg>
    );
}

export function AnalyticsCard({ title, value, trend, icon, className, description, sparklineData }: AnalyticsCardProps) {
    return (
        <div className={cn(
            "overflow-hidden relative rounded-xl border border-zinc-200 bg-white shadow-sm",
            "dark:border-zinc-800 dark:bg-zinc-900",
            "transition-shadow hover:shadow-md",
            className
        )}>
            <div className="flex flex-row items-center justify-between space-y-0 p-6 pb-2">
                <h3 className="text-sm font-medium text-zinc-500 dark:text-zinc-400">
                    {title}
                </h3>
                {icon && <div className="text-zinc-500 dark:text-zinc-400">{icon}</div>}
            </div>
            <div className="p-6 pt-0">
                <div className="text-2xl font-bold text-zinc-900 dark:text-zinc-50">{value}</div>
                {(trend || description) && (
                    <div className="flex items-center text-xs text-zinc-500 dark:text-zinc-400 mt-1">
                        {trend ? (
                            <span
                                className={cn(
                                    "flex items-center font-medium mr-2",
                                    trend.direction === 'up' && "text-emerald-500",
                                    trend.direction === 'down' && "text-rose-500",
                                    trend.direction === 'neutral' && "text-zinc-500"
                                )}
                            >
                                {trend.direction === 'up' && <ArrowUp className="h-3 w-3 mr-1" />}
                                {trend.direction === 'down' && <ArrowDown className="h-3 w-3 mr-1" />}
                                {trend.direction === 'neutral' && <Minus className="h-3 w-3 mr-1" />}
                                {Math.abs(trend.value)}%
                            </span>
                        ) : null}
                        {trend?.label && <span>{trend.label}</span>}
                        {!trend && description && <span>{description}</span>}
                    </div>
                )}
            </div>

            {/* Sparkline overlay at bottom */}
            {sparklineData && sparklineData.length >= 2 && (
                <div className="px-4 pb-3 text-blue-500/60 dark:text-blue-400/50">
                    <Sparkline data={sparklineData} />
                </div>
            )}

            {/* Decorative background element */}
            <div className="absolute -bottom-4 -right-4 h-24 w-24 rounded-full bg-blue-500/5 blur-2xl" />
        </div>
    );
}
