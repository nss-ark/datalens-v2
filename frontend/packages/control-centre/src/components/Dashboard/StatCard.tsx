import { type LucideIcon, ArrowUpRight, ArrowDownRight, Minus } from 'lucide-react';
import { Card, CardContent } from '@datalens/shared';
import { cn } from '@datalens/shared';

interface StatCardProps {
    title: string;
    value: string | number;
    icon: LucideIcon;
    color: 'primary' | 'success' | 'warning' | 'danger' | 'info';
    trend?: {
        value: number;
        label: string;
        direction: 'up' | 'down' | 'neutral';
    };
    loading?: boolean;
}

export const StatCard = ({ title, value, icon: Icon, color, trend, loading = false }: StatCardProps) => {
    if (loading) {
        return (
            <Card className="h-[140px] animate-pulse">
                <CardContent className="p-6">
                    <div className="flex items-center gap-4">
                        <div className="w-12 h-12 rounded-md bg-muted" />
                        <div className="space-y-2 flex-1">
                            <div className="h-4 bg-muted rounded w-1/2" />
                            <div className="h-8 bg-muted rounded w-3/4" />
                        </div>
                    </div>
                </CardContent>
            </Card>
        );
    }

    return (
        <Card className="hover:shadow-md transition-shadow">
            <CardContent className="p-6">
                <div className="flex items-start justify-between">
                    <div>
                        <p className="text-sm font-medium text-muted-foreground mb-1">{title}</p>
                        <h3 className="text-2xl font-bold text-foreground">{value}</h3>
                    </div>
                    <div className={cn(
                        "p-3 rounded-md",
                        color === 'primary' && "bg-blue-50 text-blue-600",
                        color === 'success' && "bg-emerald-50 text-emerald-600",
                        color === 'warning' && "bg-amber-50 text-amber-600",
                        color === 'danger' && "bg-red-50 text-red-600",
                        color === 'info' && "bg-indigo-50 text-indigo-600",
                    )}>
                        <Icon size={24} />
                    </div>
                </div>

                {trend && (
                    <div className="mt-4 flex items-center gap-2 text-sm">
                        <span className={cn(
                            "flex items-center font-medium",
                            trend.direction === 'up' && "text-emerald-600",
                            trend.direction === 'down' && "text-red-600",
                            trend.direction === 'neutral' && "text-muted-foreground",
                        )}>
                            {trend.direction === 'up' && <ArrowUpRight size={16} className="mr-1" />}
                            {trend.direction === 'down' && <ArrowDownRight size={16} className="mr-1" />}
                            {trend.direction === 'neutral' && <Minus size={16} className="mr-1" />}
                            {Math.abs(trend.value)}%
                        </span>
                        <span className="text-muted-foreground">{trend.label}</span>
                    </div>
                )}
            </CardContent>
        </Card>
    );
};
